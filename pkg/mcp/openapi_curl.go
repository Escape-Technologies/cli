package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const (
	// curlAPIKeyHeader is the placeholder echoed back to the caller. We never
	// emit the real API key — the cURL block is shown through the assistant
	// transcript and may be persisted.
	curlAPIKeyHeader       = "X-ESCAPE-API-KEY: $ESCAPE_API_KEY"
	curlAcceptJSON         = "Accept: application/json"
	curlContentTypeJSON    = "Content-Type: application/json"
	curlBodyMaxDepth       = 4
	curlPlaceholderTimeISO = "<ISO8601>"
	curlIndent             = "  "
)

// optionalQueryHeading introduces query parameters that are not part of the
// cURL command. They stay out of the URL so a copied command does not send
// a placeholder that filters to nothing or flips a flag such as dryRun.
const optionalQueryHeading = "Optional query parameters (not sent; add one only when you need it):"

// RenderCurl produces a copy-paste-ready cURL command and, when the operation
// has query parameters the command does not send, a short list of those
// parameters (name, type, enum). The command includes required parameters and
// pagination parameters that declare a default. Other optional parameters are
// listed in the second result, not given a placeholder value. The output
// never contains the caller's real credentials — only `$ESCAPE_API_KEY`.
func RenderCurl(op indexedOperation, baseURL string) (string, string) {
	method := strings.ToUpper(op.Method)
	if method == "" {
		method = "GET"
	}

	pathParams, queryParams := splitParameters(op.Parameters)
	rendered, optional := partitionQueryParams(queryParams)
	urlPath := fillPathParams(op.Path, pathParams)
	fullURL := joinURL(baseURL, urlPath)

	useGet := method == "GET" && len(rendered) > 0
	if !useGet && len(rendered) > 0 {
		fullURL = appendQueryParams(fullURL, rendered)
	}

	lines := []string{fmt.Sprintf("curl -X %s %s", method, shellQuote(fullURL))}
	if useGet {
		lines = append(lines, "  --get")
		for _, p := range rendered {
			lines = append(lines, "  --data-urlencode "+shellQuote(fmt.Sprintf("%s=%s", p.Name, paramPlaceholder(p))))
		}
	}

	lines = append(lines,
		"  -H "+shellQuote(curlAPIKeyHeader),
		"  -H "+shellQuote(curlAcceptJSON),
	)

	if op.RequestBody != nil && op.RequestBody.Schema != nil {
		bodyJSON := renderJSONSkeleton(op.RequestBody.Schema, 0)
		lines = append(lines, "  -H "+shellQuote(curlContentTypeJSON))
		lines = append(lines, "  --data "+shellQuote(bodyJSON))
	}

	return strings.Join(joinWithContinuations(lines), "\n"), renderOptionalParams(optional)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// joinWithContinuations appends ` \` to every line except the last so the
// rendered cURL is a single shell command spread across lines.
func joinWithContinuations(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		if i == len(lines)-1 {
			out[i] = line
			continue
		}

		out[i] = line + ` \`
	}

	return out
}

func splitParameters(params []openapiParameter) (pathParams, queryParams []openapiParameter) {
	for _, p := range params {
		switch strings.ToLower(p.In) {
		case "path":
			pathParams = append(pathParams, p)
		case "query":
			queryParams = append(queryParams, p)
		}
	}

	// Stable order: required first, then alphabetical. Keeps output diff-stable.
	sort.SliceStable(queryParams, func(i, j int) bool {
		if queryParams[i].Required != queryParams[j].Required {
			return queryParams[i].Required
		}

		return queryParams[i].Name < queryParams[j].Name
	})

	return pathParams, queryParams
}

// partitionQueryParams splits the sorted query list into parameters the
// command sends and parameters that are only documented. Pagination with a
// schema default is sent so the example shows the page size the server would
// apply. An optional boolean or enum is not, because the placeholder would
// change the result (dryRun=true, sortType=<first enum>, search=<search>).
func partitionQueryParams(params []openapiParameter) (rendered, optional []openapiParameter) {
	for _, p := range params {
		if p.Required || defaultedPagination(p) {
			rendered = append(rendered, p)
			continue
		}

		optional = append(optional, p)
	}

	return rendered, optional
}

func defaultedPagination(p openapiParameter) bool {
	if !isPageParam(p.Name) || p.Schema == nil {
		return false
	}

	schema := preferredSchema(p.Schema)

	return schema != nil && schema.Default != nil
}

func isPageParam(name string) bool {
	switch strings.ToLower(name) {
	case "size", "cursor", "page", "limit", "offset":
		return true
	default:
		return false
	}
}

func renderOptionalParams(params []openapiParameter) string {
	if len(params) == 0 {
		return ""
	}

	lines := make([]string, 0, len(params)+1)
	lines = append(lines, optionalQueryHeading)
	for _, p := range params {
		lines = append(lines, describeOptionalParam(p))
	}

	return strings.Join(lines, "\n")
}

func describeOptionalParam(p openapiParameter) string {
	var schema *openapiSchema
	if p.Schema != nil {
		schema = preferredSchema(p.Schema)
	}

	typeName := "value"
	if schema != nil && schema.Type != "" {
		typeName = schema.Type
	}

	detail := typeName
	if schema != nil && schema.Format != "" {
		detail += ", " + schema.Format
	}

	line := fmt.Sprintf("- %s (%s)", p.Name, detail)
	if schema != nil && len(schema.Enum) > 0 {
		parts := make([]string, len(schema.Enum))
		for i, value := range schema.Enum {
			parts[i] = fmt.Sprintf("%v", value)
		}

		line += " enum: " + strings.Join(parts, ", ")
	}

	if schema != nil && schema.Default != nil {
		line += fmt.Sprintf(" default: %v", schema.Default)
	}

	return line
}

func fillPathParams(path string, params []openapiParameter) string {
	for _, p := range params {
		path = strings.ReplaceAll(path, "{"+p.Name+"}", "<"+p.Name+">")
	}

	return path
}

func joinURL(baseURL, path string) string {
	base := strings.TrimRight(baseURL, "/")
	if path == "" {
		return base
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return base + path
}

func appendQueryParams(rawURL string, params []openapiParameter) string {
	if len(params) == 0 {
		return rawURL
	}

	separator := "?"
	if strings.Contains(rawURL, "?") {
		separator = "&"
	}

	parts := make([]string, 0, len(params))
	for _, p := range params {
		parts = append(parts, url.QueryEscape(p.Name)+"="+url.QueryEscape(paramPlaceholder(p)))
	}

	return rawURL + separator + strings.Join(parts, "&")
}

// paramPlaceholder returns a representative placeholder for a query/path
// parameter, honouring enums, formats, and defaults.
func paramPlaceholder(p openapiParameter) string {
	if p.Schema == nil {
		return "<" + p.Name + ">"
	}

	schema := preferredSchema(p.Schema)
	if schema.Default != nil {
		return fmt.Sprintf("%v", schema.Default)
	}

	if len(schema.Enum) > 0 {
		return fmt.Sprintf("%v", schema.Enum[0])
	}

	switch schema.Format {
	case "date-time":
		return curlPlaceholderTimeISO
	case "uri":
		return "https://example.com"
	case "uuid":
		return "<uuid>"
	}

	switch schema.Type {
	case "integer", "number":
		return "<" + p.Name + ":number>"
	case "boolean":
		return "true"
	default:
		return "<" + p.Name + ">"
	}
}

// renderJSONSkeleton produces a pretty-printed JSON example from a schema,
// capped at curlBodyMaxDepth to keep output readable. Uses json.Encoder with
// SetEscapeHTML(false) so literal `<placeholder>` markers stay readable
// instead of getting escaped to `<placeholder>`.
func renderJSONSkeleton(schema *openapiSchema, depth int) string {
	value := buildSkeletonValue(schema, depth)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", curlIndent)
	if err := enc.Encode(value); err != nil {
		return "{}"
	}

	return strings.TrimRight(buf.String(), "\n")
}

func buildSkeletonValue(schema *openapiSchema, depth int) any {
	if schema == nil {
		return nil
	}

	schema = preferredSchema(schema)
	if depth >= curlBodyMaxDepth {
		return "<...>"
	}

	if schema.Default != nil {
		return schema.Default
	}

	if len(schema.Enum) > 0 {
		return schema.Enum[0]
	}

	switch schema.Type {
	case "object":
		return buildObjectSkeleton(schema, depth)
	case "array":
		return []any{buildSkeletonValue(schema.Items, depth+1)}
	case "string":
		switch schema.Format {
		case "date-time":
			return curlPlaceholderTimeISO
		case "uri":
			return "https://example.com"
		case "uuid":
			return "<uuid>"
		}

		return "<string>"
	case "integer", "number":
		return 0
	case "boolean":
		return false
	}

	if len(schema.Properties) > 0 {
		return buildObjectSkeleton(schema, depth)
	}

	return nil
}

func preferredSchema(schema *openapiSchema) *openapiSchema {
	if schema == nil {
		return nil
	}

	for _, branch := range [][]*openapiSchema{schema.AllOf, schema.OneOf, schema.AnyOf} {
		for _, candidate := range branch {
			if candidate == nil || candidate.Type == "null" {
				continue
			}

			return candidate
		}
	}

	return schema
}

// orderedMap preserves insertion order when marshalled to JSON. We sort keys
// (required first, then alphabetical) so the body skeleton reads top-down.
type orderedMap struct {
	keys   []string
	values map[string]any
}

func (m orderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		if err := encodeNoHTML(&buf, k); err != nil {
			return nil, fmt.Errorf("marshal key %q: %w", k, err)
		}

		buf.WriteByte(':')
		if err := encodeNoHTML(&buf, m.values[k]); err != nil {
			return nil, fmt.Errorf("marshal value for %q: %w", k, err)
		}
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func encodeNoHTML(buf *bytes.Buffer, value any) error {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	// Encoder appends a trailing newline; strip it so we can keep the value
	// inline within the surrounding JSON.
	bytes := buf.Bytes()
	if n := len(bytes); n > 0 && bytes[n-1] == '\n' {
		buf.Truncate(n - 1)
	}

	return nil
}

func buildObjectSkeleton(schema *openapiSchema, depth int) any {
	if len(schema.Properties) == 0 {
		return map[string]any{}
	}

	requiredSet := make(map[string]struct{}, len(schema.Required))
	for _, name := range schema.Required {
		requiredSet[name] = struct{}{}
	}

	keys := make([]string, 0, len(schema.Properties))
	for k := range schema.Properties {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		_, leftReq := requiredSet[keys[i]]
		_, rightReq := requiredSet[keys[j]]
		if leftReq != rightReq {
			return leftReq
		}

		return keys[i] < keys[j]
	})

	values := make(map[string]any, len(keys))
	for _, k := range keys {
		values[k] = buildSkeletonValue(schema.Properties[k], depth+1)
	}

	return orderedMap{keys: keys, values: values}
}
