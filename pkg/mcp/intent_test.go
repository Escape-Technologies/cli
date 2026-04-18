package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// minimalLibraryToolsListResponse mimics the JSON-RPC envelope that the
// mark3labs/mcp-go server would emit for tools/list, carrying a mixed set
// of command-backed tools (shown with full RawInputSchema) and built-ins
// (which the interceptor must preserve verbatim).
func minimalLibraryToolsListResponse(t *testing.T, tools []mcpgo.Tool) []byte {
	t.Helper()
	raw := make([]json.RawMessage, 0, len(tools))
	for _, tool := range tools {
		encoded, err := json.Marshal(tool)
		if err != nil {
			t.Fatalf("marshal tool: %v", err)
		}
		raw = append(raw, encoded)
	}
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"result":  map[string]any{"tools": raw},
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return body
}

func specWithBody(t *testing.T, name, description string) ToolSpec {
	t.Helper()
	rawInputSchema, err := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"body": map[string]any{
				"type":       "object",
				"properties": map[string]any{"giant": map[string]any{"type": "string"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal raw schema: %v", err)
	}
	return ToolSpec{
		Name:         name,
		Description:  description,
		BodyProperty: "body",
		Tool:         mcpgo.NewToolWithRawSchema(name, description, rawInputSchema),
	}
}

func fakeHandler(body []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
}

type mockClassifier struct {
	result []string
	err    error
	called bool
}

func (m *mockClassifier) Rank(
	_ context.Context,
	_ ChatContext,
	_ []ToolDigest,
) ([]string, error) {
	m.called = true
	return m.result, m.err
}

func (m *mockClassifier) TopK() int { return 15 }

func callToolsList(t *testing.T, handler http.Handler, header string) *httptest.ResponseRecorder {
	t.Helper()
	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if header != "" {
		req.Header.Set(chatContextHeader, header)
	}
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}

func parseToolsFromResponse(t *testing.T, body []byte) []map[string]any {
	t.Helper()
	var envelope struct {
		Result struct {
			Tools []map[string]any `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("unmarshal response: %v\n%s", err, body)
	}
	return envelope.Result.Tools
}

func TestIntentMiddleware_OffPassesThrough(t *testing.T) {
	t.Parallel()

	specA := specWithBody(t, "alpha_create", "create an alpha")
	libraryBody := minimalLibraryToolsListResponse(t, []mcpgo.Tool{specA.Tool})

	handler := NewIntentMiddleware(
		fakeHandler(libraryBody),
		IntentOptions{Mode: IntentModeOff, Specs: []ToolSpec{specA}},
	)

	resp := callToolsList(t, handler, "")
	tools := parseToolsFromResponse(t, resp.Body.Bytes())
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	// Verify the schema was NOT replaced — Off mode must be a no-op.
	inputSchema, _ := tools[0]["inputSchema"].(map[string]any)
	props, _ := inputSchema["properties"].(map[string]any)
	body, _ := props["body"].(map[string]any)
	if body["additionalProperties"] != nil {
		t.Fatalf("Off mode altered inputSchema: %v", body)
	}
}

func TestIntentMiddleware_CompactOnlyNoContext(t *testing.T) {
	t.Parallel()

	specA := specWithBody(t, "alpha_create", "create an alpha")
	specB := specWithBody(t, "beta_list", "list betas")
	libraryBody := minimalLibraryToolsListResponse(t, []mcpgo.Tool{specA.Tool, specB.Tool})

	handler := NewIntentMiddleware(
		fakeHandler(libraryBody),
		IntentOptions{
			Mode:  IntentModeCompactOnly,
			Specs: []ToolSpec{specA, specB},
		},
	)

	resp := callToolsList(t, handler, "")
	tools := parseToolsFromResponse(t, resp.Body.Bytes())
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	for _, tool := range tools {
		inputSchema, _ := tool["inputSchema"].(map[string]any)
		props, _ := inputSchema["properties"].(map[string]any)
		body, _ := props["body"].(map[string]any)
		if body["additionalProperties"] != true {
			t.Fatalf("expected stubbed body for %q, got %v", tool["name"], body)
		}
	}
}

func TestIntentMiddleware_OnSelectsTopK(t *testing.T) {
	t.Parallel()

	specA := specWithBody(t, "issues_list", "list issues")
	specB := specWithBody(t, "profiles_create", "create profile")
	libraryBody := minimalLibraryToolsListResponse(t, []mcpgo.Tool{specA.Tool, specB.Tool})

	classifier := &mockClassifier{result: []string{"issues_list"}}
	handler := NewIntentMiddleware(
		fakeHandler(libraryBody),
		IntentOptions{
			Mode:       IntentModeOn,
			Classifier: classifier,
			Specs:      []ToolSpec{specA, specB},
		},
	)

	header, _ := json.Marshal(ChatContext{Current: "show me the findings"})
	resp := callToolsList(t, handler, string(header))

	if !classifier.called {
		t.Fatalf("classifier should have been called")
	}
	tools := parseToolsFromResponse(t, resp.Body.Bytes())
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	byName := map[string]map[string]any{}
	for _, tool := range tools {
		name, _ := tool["name"].(string)
		byName[name] = tool
	}

	selected, _ := byName["issues_list"]["inputSchema"].(map[string]any)
	selectedBody, _ := selected["properties"].(map[string]any)["body"].(map[string]any)
	if selectedBody["additionalProperties"] != nil {
		t.Fatalf("selected tool body should keep its full schema, got %v", selectedBody)
	}

	stubbed, _ := byName["profiles_create"]["inputSchema"].(map[string]any)
	stubbedBody, _ := stubbed["properties"].(map[string]any)["body"].(map[string]any)
	if stubbedBody["additionalProperties"] != true {
		t.Fatalf("unselected tool should be stubbed, got %v", stubbedBody)
	}
}

func TestIntentMiddleware_ClassifierErrorFallsBackToCompact(t *testing.T) {
	t.Parallel()

	specA := specWithBody(t, "alpha_create", "a")
	libraryBody := minimalLibraryToolsListResponse(t, []mcpgo.Tool{specA.Tool})

	classifier := &mockClassifier{err: errors.New("boom")}
	handler := NewIntentMiddleware(
		fakeHandler(libraryBody),
		IntentOptions{
			Mode:       IntentModeOn,
			Classifier: classifier,
			Specs:      []ToolSpec{specA},
		},
	)

	header, _ := json.Marshal(ChatContext{Current: "hi"})
	resp := callToolsList(t, handler, string(header))
	tools := parseToolsFromResponse(t, resp.Body.Bytes())
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	tool := tools[0]
	inputSchema, _ := tool["inputSchema"].(map[string]any)
	body, _ := inputSchema["properties"].(map[string]any)["body"].(map[string]any)
	if body["additionalProperties"] != true {
		t.Fatalf("expected classifier error to fall back to stub, got %v", body)
	}
}

func TestIntentMiddleware_PassesThroughNonToolsList(t *testing.T) {
	t.Parallel()

	downstreamCalled := false
	downstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstreamCalled = true
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "initialize") {
			t.Fatalf("downstream received unexpected body: %s", body)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`))
	})

	handler := NewIntentMiddleware(downstream, IntentOptions{
		Mode:  IntentModeCompactOnly,
		Specs: []ToolSpec{},
	})

	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if !downstreamCalled {
		t.Fatalf("initialize should pass through to library handler")
	}
}

func TestParseChatContext_RejectsInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in string
		ok bool
	}{
		{"", false},
		{"not-json", false},
		{`{}`, false},
		{`{"current":"","history":[]}`, false},
		{`{"current":"hello"}`, true},
		{`{"history":[{"role":"user","content":"hi"}]}`, true},
	}
	for _, tc := range cases {
		_, ok := parseChatContext(tc.in)
		if ok != tc.ok {
			t.Errorf("parseChatContext(%q) = %v, want %v", tc.in, ok, tc.ok)
		}
	}
}

func TestModeFromEnv(t *testing.T) {
	cases := []struct {
		set  string
		want IntentMode
	}{
		{"", IntentModeCompactOnly},
		{"on", IntentModeOn},
		{"On", IntentModeOn},
		{"compact_only", IntentModeCompactOnly},
		{"off", IntentModeOff},
		{"garbage", IntentModeCompactOnly},
	}
	for _, tc := range cases {
		t.Run(tc.set, func(t *testing.T) {
			t.Setenv(intentModeEnv, tc.set)
			got := ModeFromEnv(IntentModeCompactOnly)
			if got != tc.want {
				t.Errorf("ModeFromEnv(%q) = %q, want %q", tc.set, got, tc.want)
			}
		})
	}
}

func TestParseClassifierResult(t *testing.T) {
	t.Parallel()

	names, err := parseClassifierResult(`{"tools":["a","b"]}`)
	if err != nil {
		t.Fatalf("object form: %v", err)
	}
	if !reflectEqualStrings(names, []string{"a", "b"}) {
		t.Fatalf("object form names: %v", names)
	}

	names, err = parseClassifierResult(`["c","d"]`)
	if err != nil {
		t.Fatalf("array form: %v", err)
	}
	if !reflectEqualStrings(names, []string{"c", "d"}) {
		t.Fatalf("array form names: %v", names)
	}

	if _, err := parseClassifierResult(``); err == nil {
		t.Fatal("empty input should error")
	}
	if _, err := parseClassifierResult(`nonsense`); err == nil {
		t.Fatal("nonsense input should error")
	}
}

func reflectEqualStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
