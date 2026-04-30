package mcp

import (
	"strings"
	"testing"
)

func TestRenderCurl_GETUsesDataUrlencodePerLine(t *testing.T) {
	t.Parallel()

	op := indexedOperation{
		Method: "GET",
		Path:   "/scans",
		Parameters: []openapiParameter{
			{Name: "after", In: "query", Schema: &openapiSchema{Type: "string", Format: "date-time"}},
			{Name: "before", In: "query", Schema: &openapiSchema{Type: "string", Format: "date-time"}},
			{Name: "size", In: "query", Schema: &openapiSchema{Type: "integer", Default: 50}},
		},
	}
	got := RenderCurl(op, "https://public.escape.tech/v3")

	requireContains(t, got, "curl -X GET 'https://public.escape.tech/v3/scans' \\")
	requireContains(t, got, "  --get \\")
	requireContains(t, got, "  --data-urlencode 'after=<ISO8601>' \\")
	requireContains(t, got, "  --data-urlencode 'before=<ISO8601>' \\")
	requireContains(t, got, "  --data-urlencode 'size=50' \\")
	requireContains(t, got, "  -H 'X-ESCAPE-API-KEY: $ESCAPE_API_KEY' \\")
	requireContains(t, got, "  -H 'Accept: application/json'")
	requireMissing(t, got, "Authorization:")
}

func TestRenderCurl_UsesPreferredComposedSchemaForParams(t *testing.T) {
	t.Parallel()

	op := indexedOperation{
		Method: "GET",
		Path:   "/scans",
		Parameters: []openapiParameter{
			{
				Name: "after",
				In:   "query",
				Schema: &openapiSchema{AnyOf: []*openapiSchema{
					{Type: "string", Format: "date-time"},
					{Type: "null"},
				}},
			},
		},
	}
	got := RenderCurl(op, "https://public.escape.tech/v3")

	requireContains(t, got, "--data-urlencode 'after=<ISO8601>'")
}

func TestRenderCurl_PathParamsAreInlinedAsPlaceholders(t *testing.T) {
	t.Parallel()

	op := indexedOperation{
		Method: "GET",
		Path:   "/scans/{scanId}/issues",
		Parameters: []openapiParameter{
			{Name: "scanId", In: "path", Required: true, Schema: &openapiSchema{Type: "string"}},
		},
	}
	got := RenderCurl(op, "https://public.escape.tech/v3")

	requireContains(t, got, "https://public.escape.tech/v3/scans/<scanId>/issues")
}

func TestRenderCurl_NonGetQueryParamsStayInURL(t *testing.T) {
	t.Parallel()

	op := indexedOperation{
		Method: "POST",
		Path:   "/exports",
		Parameters: []openapiParameter{
			{Name: "projectId", In: "query", Required: true, Schema: &openapiSchema{Type: "string", Format: "uuid"}},
		},
	}
	got := RenderCurl(op, "https://public.escape.tech/v3")

	requireContains(t, got, "curl -X POST 'https://public.escape.tech/v3/exports?projectId=%3Cuuid%3E' \\")
	requireMissing(t, got, "--get")
	requireMissing(t, got, "--data-urlencode")
}

func TestRenderCurl_PostBodyIsPrettyPrintedJSON(t *testing.T) {
	t.Parallel()

	op := indexedOperation{
		Method: "POST",
		Path:   "/assets",
		RequestBody: &openapiRequestBody{
			Required: true,
			Schema: &openapiSchema{
				Type:     "object",
				Required: []string{"name", "kind"},
				Properties: map[string]*openapiSchema{
					"name": {Type: "string"},
					"kind": {Type: "string", Enum: []any{"REST", "GRAPHQL", "WEBAPP"}},
					"url":  {Type: "string", Format: "uri"},
					"metadata": {
						Type: "object",
						Properties: map[string]*openapiSchema{
							"owner": {Type: "string"},
						},
					},
				},
			},
		},
	}
	got := RenderCurl(op, "https://public.escape.tech/v3")

	requireContains(t, got, "curl -X POST 'https://public.escape.tech/v3/assets' \\")
	requireContains(t, got, "  -H 'Content-Type: application/json' \\")
	// Pretty-printed: 2-space indent, multi-line.
	requireContains(t, got, `"name": "<string>"`)
	requireContains(t, got, `"kind": "REST"`)
	// Required keys come first; alphabetical within each group.
	nameIdx := strings.Index(got, `"name"`)
	urlIdx := strings.Index(got, `"url"`)
	if nameIdx == -1 || urlIdx == -1 || nameIdx > urlIdx {
		t.Fatalf("expected required 'name' before optional 'url'\n%s", got)
	}
	// Body line itself is on its own line, not minified inline with --data.
	if !strings.Contains(got, "  --data '{\n") {
		t.Fatalf("expected pretty-printed multi-line body after --data, got:\n%s", got)
	}
}

func TestRenderCurl_DepthCapAvoidsRunaway(t *testing.T) {
	t.Parallel()

	// Build a 6-level deep nested object — depth cap should kick in at 4.
	deepest := &openapiSchema{Type: "object", Properties: map[string]*openapiSchema{
		"leaf": {Type: "string"},
	}}
	current := deepest
	for i := 0; i < 5; i++ {
		current = &openapiSchema{Type: "object", Properties: map[string]*openapiSchema{
			"child": current,
		}}
	}
	op := indexedOperation{
		Method:      "POST",
		Path:        "/x",
		RequestBody: &openapiRequestBody{Schema: current},
	}
	got := RenderCurl(op, "https://api.example/v3")
	requireContains(t, got, `"<...>"`)
}

func TestRenderCurl_NeverEmitsRealAPIKey(t *testing.T) {
	t.Parallel()

	op := indexedOperation{Method: "GET", Path: "/scans"}
	got := RenderCurl(op, "https://public.escape.tech/v3")
	if strings.Contains(strings.ToLower(got), "real-key") {
		t.Fatalf("rendered cURL leaked a value: %s", got)
	}
	requireContains(t, got, "$ESCAPE_API_KEY")
}

func TestRenderCurl_LineContinuationsExceptLast(t *testing.T) {
	t.Parallel()

	op := indexedOperation{
		Method: "GET",
		Path:   "/scans",
		Parameters: []openapiParameter{
			{Name: "size", In: "query", Schema: &openapiSchema{Type: "integer"}},
		},
	}
	got := RenderCurl(op, "https://public.escape.tech/v3")
	lines := strings.Split(got, "\n")
	if len(lines) < 4 {
		t.Fatalf("expected multi-line cURL, got:\n%s", got)
	}
	for i, line := range lines {
		if i == len(lines)-1 {
			if strings.HasSuffix(line, ` \`) {
				t.Fatalf("last line should not end with continuation: %q", line)
			}
		} else if !strings.HasSuffix(line, ` \`) {
			t.Fatalf("intermediate line %d missing continuation: %q", i, line)
		}
	}
}

func requireContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected substring %q in:\n%s", needle, haystack)
	}
}

func requireMissing(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Fatalf("did not expect substring %q in:\n%s", needle, haystack)
	}
}
