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

// commonOptionalQueryParams lists query-parameter names that are usually worth
// surfacing even when they are not flagged required. Keeps the rendered cURL
// useful for pagination + date filtering without dumping every option.
var commonOptionalQueryParams = map[string]struct{}{
	"limit": {}, "size": {}, "page": {}, "cursor": {},
	"from": {}, "to": {}, "after": {}, "before": {},
	"since": {}, "until": {},
}

// RenderCurl produces a multi-line, copy-paste-ready cURL example for the
// given operation against baseURL. The output never contains the caller's
// real credentials — only the placeholder `$ESCAPE_API_KEY`.
func RenderCurl(op indexedOperation, baseURL string) string {
	method := strings.ToUpper(op.Method)
	if method == "" {
		method = "GET"
	}

	pathParams, queryParams := splitParameters(op.Parameters)
	urlPath := fillPathParams(op.Path, pathParams)
	fullURL := joinURL(baseURL, urlPath)

	useGet := method == "GET" && len(queryParams) > 0
	if !useGet && len(queryParams) > 0 {
		fullURL = appendQueryParams(fullURL, queryParams)
	}

	lines := []string{fmt.Sprintf("curl -X %s '%s'", method, fullURL)}
	if useGet {
		lines = append(lines, "  --get")
		for _, p := range queryParams {
			lines = append(lines, fmt.Sprintf("  --data-urlencode '%s=%s'", p.Name, paramPlaceholder(p)))
		}
	}

	lines = append(lines,
		fmt.Sprintf("  -H '%s'", curlAPIKeyHeader),
		fmt.Sprintf("  -H '%s'", curlAcceptJSON),
	)

	if op.RequestBody != nil && op.RequestBody.Schema != nil {
		bodyJSON := renderJSONSkeleton(op.RequestBody.Schema, 0)
		lines = append(lines, fmt.Sprintf("  -H '%s'", curlContentTypeJSON))
		// Newlines survive inside a shell single-quoted string. Single quotes
		// inside the body are pre-escaped.
		escaped := strings.ReplaceAll(bodyJSON, `'`, `'\''`)
		lines = append(lines, fmt.Sprintf("  --data '%s'", escaped))
	}

	return strings.Join(joinWithContinuations(lines), "\n")
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
			if !p.Required {
				if _, common := commonOptionalQueryParams[strings.ToLower(p.Name)]; !common {
					continue
				}
			}
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
