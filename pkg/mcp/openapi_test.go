package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func newAuthedContext() context.Context {
	return context.WithValue(context.Background(), authContextKey{}, Auth{
		APIKey: "test-key",
		Method: AuthMethodAPIKeyHeader,
	})
}

func newPublicAPIHandlerWithFixture(t *testing.T) (handler func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error), specURL string) {
	t.Helper()
	body := loadFixtureSpec(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	index := NewOpenAPISearchIndex(OpenAPISearchIndexOptions{SpecURL: srv.URL})
	return buildPublicAPIHandler(index, "https://public.escape.tech"), srv.URL
}

func newCallToolRequest(args map[string]any) mcpgo.CallToolRequest {
	return mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{Arguments: args},
	}
}

func TestPublicAPIHandler_AuthGate(t *testing.T) {
	t.Parallel()

	handler, _ := newPublicAPIHandlerWithFixture(t)
	res, err := handler(context.Background(), newCallToolRequest(map[string]any{
		"question": "list scans",
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected auth error, got: %+v", res)
	}
}

func TestPublicAPIHandler_RejectsEmptyQuestion(t *testing.T) {
	t.Parallel()

	handler, _ := newPublicAPIHandlerWithFixture(t)
	res, err := handler(newAuthedContext(), newCallToolRequest(map[string]any{
		"question": "   ",
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected validation error for empty question, got: %+v", res)
	}
}

func TestPublicAPIHandler_ReturnsCurlBlock(t *testing.T) {
	t.Parallel()

	handler, _ := newPublicAPIHandlerWithFixture(t)
	res, err := handler(newAuthedContext(), newCallToolRequest(map[string]any{
		"question": "how do I list scans from the last 3 days",
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %+v", res)
	}

	text := textFromResult(t, res)
	if !strings.Contains(text, "```bash") {
		t.Fatalf("expected fenced bash block, got:\n%s", text)
	}
	if !strings.Contains(text, "X-ESCAPE-API-KEY: $ESCAPE_API_KEY") {
		t.Fatalf("expected API key placeholder header, got:\n%s", text)
	}
	if !strings.Contains(text, "## GET /scans") {
		t.Fatalf("expected GET /scans heading, got:\n%s", text)
	}

	payload := structuredFromResult(t, res)
	matches, ok := payload["matches"].([]any)
	if !ok || len(matches) == 0 {
		t.Fatalf("expected non-empty matches in payload, got %T %v", payload["matches"], payload["matches"])
	}
	first := matches[0].(map[string]any)
	if first["method"] != "GET" || first["path"] != "/scans" {
		t.Fatalf("unexpected first match: %+v", first)
	}
	if curl, _ := first["curl"].(string); !strings.Contains(curl, "--data-urlencode 'after=") {
		t.Fatalf("expected after= placeholder in curl, got:\n%s", curl)
	}
}

func TestPublicAPIHandler_LimitClampedToThree(t *testing.T) {
	t.Parallel()

	handler, _ := newPublicAPIHandlerWithFixture(t)
	res, err := handler(newAuthedContext(), newCallToolRequest(map[string]any{
		"question": "scans",
		"limit":    float64(99),
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}

	payload := structuredFromResult(t, res)
	matches, _ := payload["matches"].([]any)
	if len(matches) > publicAPIMaxLimit {
		t.Fatalf("expected at most %d matches, got %d", publicAPIMaxLimit, len(matches))
	}
}

func TestPublicAPIHandler_FallbackOnSpecUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	index := NewOpenAPISearchIndex(OpenAPISearchIndexOptions{SpecURL: srv.URL})
	handler := buildPublicAPIHandler(index, "")
	res, err := handler(newAuthedContext(), newCallToolRequest(map[string]any{
		"question": "list scans",
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("fallback should be a normal result, not an error")
	}
	payload := structuredFromResult(t, res)
	if fallback, _ := payload["fallback"].(bool); !fallback {
		t.Fatalf("expected fallback=true, got %v", payload)
	}
	if specOff, _ := payload["specUnavailable"].(bool); !specOff {
		t.Fatalf("expected specUnavailable=true, got %v", payload)
	}
}

func TestResolveBaseURL_Precedence(t *testing.T) {
	t.Parallel()

	// Configured wins, with /v3 suffixed when missing.
	if got := resolveBaseURL("https://staging.escape.tech", []string{"https://ignored/v3"}); got != "https://staging.escape.tech/v3" {
		t.Fatalf("configured w/o /v3: got %q", got)
	}
	// Configured already ending in /v3 stays as-is.
	if got := resolveBaseURL("https://staging.escape.tech/v3", nil); got != "https://staging.escape.tech/v3" {
		t.Fatalf("configured /v3: got %q", got)
	}
	// Falls back to spec server.
	if got := resolveBaseURL("", []string{"https://public.escape.tech/v3"}); got != "https://public.escape.tech/v3" {
		t.Fatalf("spec server fallback: got %q", got)
	}
	// Final default.
	if got := resolveBaseURL("", nil); got != publicAPIDefaultBaseURL {
		t.Fatalf("default fallback: got %q", got)
	}
}

func textFromResult(t *testing.T, res *mcpgo.CallToolResult) string {
	t.Helper()
	for _, c := range res.Content {
		if tc, ok := c.(mcpgo.TextContent); ok {
			return tc.Text
		}
	}
	t.Fatalf("no TextContent in result: %+v", res)
	return ""
}

func structuredFromResult(t *testing.T, res *mcpgo.CallToolResult) map[string]any {
	t.Helper()
	if res.StructuredContent == nil {
		t.Fatalf("missing structured content")
	}
	encoded, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(encoded, &out); err != nil {
		t.Fatalf("unmarshal structured: %v", err)
	}
	return out
}
