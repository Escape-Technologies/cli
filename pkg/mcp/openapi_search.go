package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultPublicAPIOpenAPIURL = "https://public.escape.tech/v3/openapi.json"
	defaultOpenAPIIndexTTL     = 5 * time.Minute
	publicAPIOpenAPIEnv        = "PUBLIC_API_OPENAPI_URL"

	openapiHTTPTimeout        = 15 * time.Second
	openapiErrorBodyPeekBytes = 256
	openapiMaxResultsPerQuery = 3
	openapiMaxNormalizedField = 4096
	openapiTermCoverageWeight = 3
	openapiExpectedMethods    = 4
)

// OpenAPISearchIndexOptions configures the OpenAPI-spec index.
type OpenAPISearchIndexOptions struct {
	SpecURL    string
	TTL        time.Duration
	HTTPClient *http.Client
}

// OpenAPISearchIndex caches and queries the public-API OpenAPI spec.
type OpenAPISearchIndex struct {
	specURL    string
	ttl        time.Duration
	httpClient *http.Client

	mu          sync.Mutex
	cache       []indexedOperation
	specServers []string
	cacheUntil  time.Time
	inFlight    chan struct{}
	inFlightErr error
	inFlightRes []indexedOperation
	inFlightSrv []string
}

// openapiParameter is the projected shape of an OpenAPI parameter object.
type openapiParameter struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Required    bool           `json:"required"`
	Description string         `json:"description"`
	Schema      *openapiSchema `json:"schema,omitempty"`
}

// openapiRequestBody is the projected shape of a JSON-content request body.
type openapiRequestBody struct {
	Required bool           `json:"required"`
	Schema   *openapiSchema `json:"schema,omitempty"`
}

// openapiSchema is the subset of JSON-schema fields used by the renderer. Each
// field stays optional so unknown shapes degrade gracefully.
type openapiSchema struct {
	Ref         string                    `json:"$ref,omitempty"`
	Type        string                    `json:"type,omitempty"`
	Format      string                    `json:"format,omitempty"`
	Default     any                       `json:"default,omitempty"`
	Enum        []any                     `json:"enum,omitempty"`
	Description string                    `json:"description,omitempty"`
	Properties  map[string]*openapiSchema `json:"properties,omitempty"`
	Required    []string                  `json:"required,omitempty"`
	Items       *openapiSchema            `json:"items,omitempty"`
	AnyOf       []*openapiSchema          `json:"anyOf,omitempty"`
	OneOf       []*openapiSchema          `json:"oneOf,omitempty"`
	AllOf       []*openapiSchema          `json:"allOf,omitempty"`
}

func (s *openapiSchema) UnmarshalJSON(data []byte) error {
	type alias openapiSchema
	var raw struct {
		*alias
		Type any `json:"type,omitempty"`
	}
	raw.alias = (*alias)(s)
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode openapi schema: %w", err)
	}
	switch typed := raw.Type.(type) {
	case string:
		s.Type = typed
	case []any:
		for _, item := range typed {
			value, ok := item.(string)
			if ok && value != "null" {
				s.Type = value
				break
			}
		}
	}
	return nil
}

// indexedOperation is a single OpenAPI operation flattened for search.
type indexedOperation struct {
	OperationID string
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Parameters  []openapiParameter
	RequestBody *openapiRequestBody

	normalizedSummary     string
	normalizedOperationID string
	normalizedPath        string
	normalizedDescription string
	normalizedTags        string
}

// rawOpenAPISpec is the projection of the OpenAPI document used by the
// indexer. References are resolved post-decode.
type rawOpenAPISpec struct {
	Servers    []rawOpenAPIServer            `json:"servers"`
	Paths      map[string]rawOpenAPIPathItem `json:"paths"`
	Components struct {
		Schemas map[string]*openapiSchema `json:"schemas"`
	} `json:"components"`
}

type rawOpenAPIServer struct {
	URL string `json:"url"`
}

type rawOpenAPIPathItem map[string]json.RawMessage

type rawOpenAPIOperation struct {
	OperationID string                 `json:"operationId"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Parameters  []openapiParameter     `json:"parameters"`
	RequestBody *rawOpenAPIRequestBody `json:"requestBody"`
}

type rawOpenAPIRequestBody struct {
	Required bool                           `json:"required"`
	Content  map[string]rawOpenAPIMediaType `json:"content"`
}

type rawOpenAPIMediaType struct {
	Schema *openapiSchema `json:"schema"`
}

// httpMethodSet enumerates the OpenAPI path-item method keys recognised by
// the indexer. Other keys (parameters, summary, $ref) are ignored.
var httpMethodSet = map[string]struct{}{
	"get": {}, "post": {}, "put": {}, "patch": {}, "delete": {},
	"head": {}, "options": {}, "trace": {},
}

// NewOpenAPISearchIndex builds an OpenAPI search index with sensible defaults.
func NewOpenAPISearchIndex(options OpenAPISearchIndexOptions) *OpenAPISearchIndex {
	index := &OpenAPISearchIndex{
		specURL:    firstNonEmpty(options.SpecURL, defaultPublicAPIOpenAPIURL),
		ttl:        firstNonZero(options.TTL, defaultOpenAPIIndexTTL),
		httpClient: options.HTTPClient,
	}
	if index.httpClient == nil {
		index.httpClient = &http.Client{Timeout: openapiHTTPTimeout}
	}
	return index
}

// search returns the top-N operations matching the query. Limit is clamped to
// [1, openapiMaxResultsPerQuery].
func (o *OpenAPISearchIndex) search(ctx context.Context, query string, limit int) ([]indexedOperation, []string, error) {
	if limit < 1 {
		limit = 1
	}
	if limit > openapiMaxResultsPerQuery {
		limit = openapiMaxResultsPerQuery
	}

	spec := toDocsQuery(query)
	if spec.normalizedQuery == "" {
		return nil, nil, nil
	}

	ops, servers, err := o.loadOperations(ctx)
	if err != nil {
		return nil, nil, err
	}

	type scored struct {
		op    indexedOperation
		score float64
	}
	ranked := make([]scored, 0, len(ops))
	for _, op := range ops {
		s := scoreOperation(op, spec)
		if s > 0 {
			ranked = append(ranked, scored{op: op, score: s})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].score != ranked[j].score {
			return ranked[i].score > ranked[j].score
		}
		// Prefer shorter paths (closer match to a resource root) on ties.
		if len(ranked[i].op.Path) != len(ranked[j].op.Path) {
			return len(ranked[i].op.Path) < len(ranked[j].op.Path)
		}
		return ranked[i].op.OperationID < ranked[j].op.OperationID
	})

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	out := make([]indexedOperation, 0, len(ranked))
	for _, entry := range ranked {
		out = append(out, entry.op)
	}
	return out, servers, nil
}

// scoreOperation weights matches against the operation's structured fields.
// Summary + operationId behave like docs titles (heaviest); path is medium
// since it anchors on resource names; description and tags are lighter.
func scoreOperation(op indexedOperation, query docsQuerySpec) float64 {
	if query.normalizedQuery == "" || len(query.terms) == 0 {
		return 0
	}

	var score float64

	if op.normalizedSummary == query.normalizedQuery ||
		op.normalizedSummary == query.normalizedQuery+"s" {
		score += 4
	} else if strings.HasPrefix(op.normalizedSummary, query.normalizedQuery) {
		score += 2
	}

	if strings.Contains(op.normalizedSummary, query.normalizedQuery) {
		score += 8
	}
	if strings.Contains(op.normalizedOperationID, query.normalizedQuery) {
		score += 7
	}
	if strings.Contains(op.normalizedPath, query.normalizedQuery) {
		score += 6
	}
	if strings.Contains(op.normalizedDescription, query.normalizedQuery) {
		score += 3
	}
	if strings.Contains(op.normalizedTags, query.normalizedQuery) {
		score += 2
	}

	matched := 0
	for _, term := range query.terms {
		hit := false
		if strings.Contains(op.normalizedSummary, term) {
			score += 2.5
			hit = true
		}
		if strings.Contains(op.normalizedOperationID, term) {
			score += 2
			hit = true
		}
		if strings.Contains(op.normalizedPath, term) {
			score += 1.75
			hit = true
		}
		if strings.Contains(op.normalizedDescription, term) {
			score++
			hit = true
		}
		if strings.Contains(op.normalizedTags, term) {
			score += 0.75
			hit = true
		}
		if hit {
			matched++
		}
	}
	score += (float64(matched) / float64(len(query.terms))) * openapiTermCoverageWeight

	return score
}

func (o *OpenAPISearchIndex) loadOperations(ctx context.Context) ([]indexedOperation, []string, error) {
	o.mu.Lock()
	if len(o.cache) > 0 && time.Now().Before(o.cacheUntil) {
		defer o.mu.Unlock()
		return o.cache, o.specServers, nil
	}
	if o.inFlight != nil {
		ch := o.inFlight
		o.mu.Unlock()
		select {
		case <-ch:
			o.mu.Lock()
			defer o.mu.Unlock()
			if o.inFlightErr != nil {
				return nil, nil, o.inFlightErr
			}
			return o.inFlightRes, o.inFlightSrv, nil
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("openapi index wait: %w", ctx.Err())
		}
	}

	o.inFlight = make(chan struct{})
	o.mu.Unlock()

	ops, servers, err := o.fetchOperations(ctx)

	o.mu.Lock()
	o.inFlightErr = err
	o.inFlightRes = ops
	o.inFlightSrv = servers
	close(o.inFlight)
	o.inFlight = nil
	if err == nil {
		o.cache = ops
		o.specServers = servers
		o.cacheUntil = time.Now().Add(o.ttl)
	}
	o.mu.Unlock()
	return ops, servers, err
}

func (o *OpenAPISearchIndex) fetchOperations(ctx context.Context) ([]indexedOperation, []string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.specURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("new openapi request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("openapi fetch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, openapiErrorBodyPeekBytes))
		return nil, nil, fmt.Errorf("openapi http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read openapi body: %w", err)
	}
	return parseOpenAPISpec(body)
}

// parseOpenAPISpec is exported via tests (lowercase but same package).
func parseOpenAPISpec(body []byte) ([]indexedOperation, []string, error) {
	var raw rawOpenAPISpec
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, nil, fmt.Errorf("decode openapi: %w", err)
	}

	servers := make([]string, 0, len(raw.Servers))
	for _, s := range raw.Servers {
		if strings.TrimSpace(s.URL) != "" {
			servers = append(servers, strings.TrimRight(s.URL, "/"))
		}
	}

	ops := make([]indexedOperation, 0, len(raw.Paths)*openapiExpectedMethods)
	for path, item := range raw.Paths {
		pathParameters := decodePathParameters(item["parameters"])
		for method, rawOp := range item {
			lower := strings.ToLower(method)
			if _, ok := httpMethodSet[lower]; !ok {
				continue
			}
			var op rawOpenAPIOperation
			if err := json.Unmarshal(rawOp, &op); err != nil {
				continue
			}
			if len(pathParameters) > 0 {
				op.Parameters = append(append([]openapiParameter{}, pathParameters...), op.Parameters...)
			}
			indexed := toIndexedOperation(strings.ToUpper(lower), path, op, raw.Components.Schemas)
			ops = append(ops, indexed)
		}
	}
	return ops, servers, nil
}

func decodePathParameters(raw json.RawMessage) []openapiParameter {
	if len(raw) == 0 {
		return nil
	}
	var params []openapiParameter
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil
	}
	return params
}

func toIndexedOperation(
	method, path string,
	op rawOpenAPIOperation,
	schemas map[string]*openapiSchema,
) indexedOperation {
	resolvedParams := make([]openapiParameter, 0, len(op.Parameters))
	for _, p := range op.Parameters {
		if p.Schema != nil {
			p.Schema = resolveSchemaRef(p.Schema, schemas)
		}
		resolvedParams = append(resolvedParams, p)
	}

	var body *openapiRequestBody
	if op.RequestBody != nil {
		if media, ok := op.RequestBody.Content["application/json"]; ok && media.Schema != nil {
			body = &openapiRequestBody{
				Required: op.RequestBody.Required,
				Schema:   resolveSchemaRef(media.Schema, schemas),
			}
		}
	}

	return indexedOperation{
		OperationID:           op.OperationID,
		Method:                method,
		Path:                  path,
		Summary:               op.Summary,
		Description:           op.Description,
		Tags:                  op.Tags,
		Parameters:            resolvedParams,
		RequestBody:           body,
		normalizedSummary:     truncateField(normalizeForSearch(op.Summary)),
		normalizedOperationID: truncateField(normalizeForSearch(op.OperationID)),
		normalizedPath:        truncateField(normalizeForSearch(path)),
		normalizedDescription: truncateField(normalizeForSearch(op.Description)),
		normalizedTags:        truncateField(normalizeForSearch(strings.Join(op.Tags, " "))),
	}
}

// resolveSchemaRef walks a schema and replaces any local component reference
// with a deep copy of the referenced schema. Cycles are broken by capping
// recursion at the renderer level (renderJSONSkeleton's depth cap).
func resolveSchemaRef(schema *openapiSchema, schemas map[string]*openapiSchema) *openapiSchema {
	return resolveSchemaRefSeen(schema, schemas, map[string]struct{}{})
}

func resolveSchemaRefSeen(
	schema *openapiSchema,
	schemas map[string]*openapiSchema,
	seen map[string]struct{},
) *openapiSchema {
	if schema == nil {
		return nil
	}
	if schema.Ref != "" {
		name := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
		if _, cycle := seen[name]; cycle {
			return schema
		}
		if resolved, ok := schemas[name]; ok && resolved != nil {
			nextSeen := make(map[string]struct{}, len(seen)+1)
			for k, v := range seen {
				nextSeen[k] = v
			}
			nextSeen[name] = struct{}{}
			cp := *resolved
			return resolveSchemaRefSeen(&cp, schemas, nextSeen)
		}
		return schema
	}
	if len(schema.Properties) > 0 {
		resolvedProps := make(map[string]*openapiSchema, len(schema.Properties))
		for k, v := range schema.Properties {
			resolvedProps[k] = resolveSchemaRefSeen(v, schemas, seen)
		}
		schema.Properties = resolvedProps
	}
	if schema.Items != nil {
		schema.Items = resolveSchemaRefSeen(schema.Items, schemas, seen)
	}
	for i, branch := range schema.AnyOf {
		schema.AnyOf[i] = resolveSchemaRefSeen(branch, schemas, seen)
	}
	for i, branch := range schema.OneOf {
		schema.OneOf[i] = resolveSchemaRefSeen(branch, schemas, seen)
	}
	for i, branch := range schema.AllOf {
		schema.AllOf[i] = resolveSchemaRefSeen(branch, schemas, seen)
	}
	return schema
}

func truncateField(value string) string {
	if len(value) <= openapiMaxNormalizedField {
		return value
	}
	return value[:openapiMaxNormalizedField]
}

// ErrOpenAPIUnavailable is returned when the OpenAPI spec cannot be fetched.
var ErrOpenAPIUnavailable = errors.New("openapi spec unavailable")
