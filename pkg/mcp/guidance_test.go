package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func TestInitializeIncludesInstructions(t *testing.T) {
	t.Parallel()

	if len(ServerInstructions) == 0 || len(ServerInstructions) >= serverInstructionsLimit {
		t.Fatalf("instructions length %d, want 1..%d", len(ServerInstructions), serverInstructionsLimit-1)
	}

	for _, phrase := range []string{
		"profiles_update_configuration",
		"defaults merge to true",
		"profiles_get",
		"read_only",
		"scans_coverage",
		"overall.statuses",
		"targetsTruncated",
		"byUser",
		"scans_targets",
		"not a page size",
		"scans_reasoning",
		"poll scans_get",
		"role changes",
		"authentications_start",
		"escape_get_tool_spec",
	} {
		if !strings.Contains(ServerInstructions, phrase) {
			t.Errorf("instructions missing %q", phrase)
		}
	}

	if strings.Contains(ServerInstructions, "scans_watch") {
		t.Fatal("instructions name scans_watch, which is not a tool")
	}

	server := testProtocolServer(t)
	body := callMCP(t, server, `{
		"jsonrpc":"2.0",
		"id":1,
		"method":"initialize",
		"params":{
			"protocolVersion":"2025-03-26",
			"capabilities":{},
			"clientInfo":{"name":"test","version":"0"}
		}
	}`)

	var parsed struct {
		Result struct {
			Instructions string                     `json:"instructions"`
			Capabilities map[string]json.RawMessage `json:"capabilities"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal initialize: %v\n%s", err, body)
	}

	if parsed.Error != nil {
		t.Fatalf("initialize error: %s", parsed.Error.Message)
	}

	if parsed.Result.Instructions != ServerInstructions {
		t.Fatalf("initialize instructions = %q", parsed.Result.Instructions)
	}

	if _, ok := parsed.Result.Capabilities["prompts"]; !ok {
		t.Fatalf("initialize capabilities missing prompts: %s", body)
	}
}

func TestPromptsListAndRender(t *testing.T) {
	t.Parallel()

	server := testProtocolServer(t)
	body := callMCP(t, server, `{"jsonrpc":"2.0","id":2,"method":"prompts/list"}`)

	var listed struct {
		Result struct {
			Prompts []struct {
				Name      string `json:"name"`
				Arguments []struct {
					Name     string `json:"name"`
					Required bool   `json:"required"`
				} `json:"arguments"`
			} `json:"prompts"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &listed); err != nil {
		t.Fatalf("unmarshal prompts/list: %v\n%s", err, body)
	}

	if listed.Error != nil {
		t.Fatalf("prompts/list error: %s", listed.Error.Message)
	}

	byName := map[string][]string{}
	for _, prompt := range listed.Result.Prompts {
		required := make([]string, 0)
		for _, arg := range prompt.Arguments {
			if arg.Required {
				required = append(required, arg.Name)
			}
		}

		byName[prompt.Name] = required
	}

	wantArgs := map[string][]string{
		"triage_findings":             nil,
		"fix_finding":                 {"issue_id"},
		"setup_dast_profile":          {"url"},
		"start_and_summarize_pentest": {"profile_id"},
		"explain_coverage":            {"scan_id"},
	}
	if len(byName) != len(wantArgs) {
		t.Fatalf("prompts = %#v", byName)
	}

	for name, args := range wantArgs {
		got, ok := byName[name]
		if !ok {
			t.Fatalf("missing prompt %s", name)
		}

		if strings.Join(got, ",") != strings.Join(args, ",") {
			t.Fatalf("prompt %s args = %v, want %v", name, got, args)
		}
	}

	cases := []struct {
		name   string
		args   map[string]string
		want   []string
		absent []string
	}{
		{
			name: "triage_findings",
			want: []string{"issues_list", "issues_get_with_events", "issues_update"},
		},
		{
			name: "fix_finding",
			args: map[string]string{"issue_id": "iss-1"},
			want: []string{"iss-1", "issues_get_with_events", "issues_comment", "issues_update"},
		},
		{
			name: "setup_dast_profile",
			args: map[string]string{"url": "https://app.example.com"},
			want: []string{
				"https://app.example.com",
				"assets_list",
				"assets_create",
				"profiles_create_webapp",
				"profiles_create_rest",
				"profiles_update_configuration",
				"merge",
				"assetId and name",
				"Set mode",
				"schema error",
				"retry once",
			},
			absent: []string{"on that tool first", "scans_watch"},
		},
		{
			name:   "start_and_summarize_pentest",
			args:   map[string]string{"profile_id": "prof-1"},
			want:   []string{"prof-1", "profiles_get", "scans_start", "scans_get", "scans_coverage", "scans_issues"},
			absent: []string{"scans_watch", "watch"},
		},
		{
			name: "explain_coverage",
			args: map[string]string{"scan_id": "scan-1"},
			want: []string{"scan-1", "scans_coverage", "overall.statuses", "byUser", "scans_targets", "not a page size"},
		},
	}
	for _, tc := range cases {
		text := renderPrompt(t, server, tc.name, tc.args)
		for _, phrase := range tc.want {
			if !strings.Contains(text, phrase) {
				t.Errorf("prompt %s missing %q\n%s", tc.name, phrase, text)
			}
		}

		for _, phrase := range tc.absent {
			if strings.Contains(text, phrase) {
				t.Errorf("prompt %s still contains %q\n%s", tc.name, phrase, text)
			}
		}
	}
}

func TestPromptRequiresArgument(t *testing.T) {
	t.Parallel()

	server := testProtocolServer(t)
	body := callMCP(t, server, `{"jsonrpc":"2.0","id":4,"method":"prompts/get","params":{"name":"fix_finding"}}`)
	if !strings.Contains(string(body), "missing required argument") {
		t.Fatalf("expected missing argument error, got %s", body)
	}
}

func testProtocolServer(t *testing.T) *mcpserver.MCPServer {
	t.Helper()
	server, err := NewServer(ServerOptions{Version: "test"}).protocolServer()
	if err != nil {
		t.Fatalf("protocol server: %v", err)
	}

	return server
}

func callMCP(t *testing.T, server *mcpserver.MCPServer, message string) []byte {
	t.Helper()
	resp := server.HandleMessage(context.Background(), json.RawMessage(message))
	if resp == nil {
		t.Fatal("nil MCP response")
	}

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal MCP response: %v", err)
	}

	return body
}

func renderPrompt(t *testing.T, server *mcpserver.MCPServer, name string, args map[string]string) string {
	t.Helper()
	params := map[string]any{"name": name}
	if args != nil {
		params["arguments"] = args
	}

	raw, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "prompts/get",
		"params":  params,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	body := callMCP(t, server, string(raw))

	var parsed struct {
		Result struct {
			Messages []struct {
				Content struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"messages"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal prompts/get %s: %v\n%s", name, err, body)
	}

	if parsed.Error != nil {
		t.Fatalf("prompts/get %s: %s", name, parsed.Error.Message)
	}

	if len(parsed.Result.Messages) != 1 || parsed.Result.Messages[0].Content.Text == "" {
		t.Fatalf("prompts/get %s returned no text: %s", name, body)
	}

	return parsed.Result.Messages[0].Content.Text
}
