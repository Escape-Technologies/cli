package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func TestCapabilityScore_ExpandsSynonyms(t *testing.T) {
	t.Parallel()

	issues := ToolSpec{Name: "issues_list", Path: "escape-cli issues list", Description: "List issues"}
	assets := ToolSpec{Name: "assets_list", Path: "escape-cli assets list", Description: "List assets"}
	profiles := ToolSpec{Name: "profiles_list", Path: "escape-cli profiles list", Description: "List API profiles"}
	scans := ToolSpec{Name: "scans_start", Path: "escape-cli scans start", Description: "Start a scan"}
	pentest := ToolSpec{Name: "profiles_create_ai_pentest", Path: "escape-cli profiles create-ai-pentest", Description: "Create an AI pentest profile"}
	locations := ToolSpec{Name: "locations_list", Path: "escape-cli locations list", Description: "List private locations"}

	cases := []struct {
		objective string
		spec      ToolSpec
		wantMatch bool
	}{
		{"show my findings", issues, true},
		{"open vulnerabilities", issues, true},
		{"any bugs?", issues, true},
		{"list domains", assets, true},
		{"my apps", profiles, true},
		{"run a pentest", scans, true},
		{"pentesting", pentest, true},
		{"triage findings", issues, true},
		// Negative guardrail: synonyms must not drag unrelated tools in.
		{"findings", assets, false},
		{"findings", locations, false},
		{"domains", locations, false},
	}
	for _, tc := range cases {
		score := capabilityScore(tc.spec, strings.ToLower(tc.objective))
		if tc.wantMatch && score == 0 {
			t.Errorf("capabilityScore(%q, %q) = 0, want > 0", tc.spec.Name, tc.objective)
		}

		if !tc.wantMatch && score != 0 {
			t.Errorf("capabilityScore(%q, %q) = %d, want 0", tc.spec.Name, tc.objective, score)
		}
	}
}

func TestListCapabilitiesHandler_RanksSynonymObjectiveFirst(t *testing.T) {
	t.Parallel()

	specs := []ToolSpec{
		{Name: "locations_list", Path: "escape-cli locations list", Description: "List private locations"},
		{Name: "issues_list", Path: "escape-cli issues list", Description: "List issues"},
		{Name: "assets_list", Path: "escape-cli assets list", Description: "List assets"},
	}
	handler := buildListCapabilitiesHandler(specs)

	res, err := handler(newAuthedContext(), newCallToolRequest(map[string]any{
		"objective": "top findings",
	}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}

	if res.IsError {
		t.Fatalf("unexpected error result: %+v", res)
	}

	payload := structuredFromResult(t, res)
	// items is a []map[string]any, so wrapStructuredPayload nests it under
	// "value" (only top-level []any lands under "items").
	ranked, ok := payload["value"].([]any)
	if !ok || len(ranked) != len(specs) {
		t.Fatalf("expected %d ranked items, got %T %v", len(specs), payload["value"], payload["value"])
	}

	items := ranked
	first, _ := items[0].(map[string]any)
	if first["name"] != "issues_list" {
		t.Fatalf("expected issues_list first for 'findings' objective, got %+v", items)
	}

	// Without an objective the ordering falls back to plain name sort.
	res, _ = handler(newAuthedContext(), newCallToolRequest(map[string]any{}))
	payload = structuredFromResult(t, res)
	items, _ = payload["value"].([]any)
	first, _ = items[0].(map[string]any)
	if first["name"] != "assets_list" {
		t.Fatalf("expected alphabetical order without objective, got %+v", items)
	}
}

func TestToolSpecResultIsCompactAndNotDuplicated(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"type":"object","properties":{"status":{"type":"string","enum":["OPEN","RESOLVED"]}}}`)
	tool := mcpgo.NewToolWithRawSchema("issues_list", "list", raw)
	res, err := toolSpecResult(ToolSpec{
		Name:        "issues_list",
		Description: "List issues",
		Tool:        tool,
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.StructuredContent != nil {
		t.Fatalf("structured content duplicates the schema: %#v", res.StructuredContent)
	}

	text := textFromResult(t, res)
	if strings.Contains(text, "\n") {
		t.Fatalf("spec is indented: %s", text)
	}

	if strings.Count(text, "RESOLVED") != 1 {
		t.Fatalf("enum value count = %d in %s", strings.Count(text, "RESOLVED"), text)
	}

	var payload struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		InputSchema map[string]any `json:"inputSchema"`
	}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatal(err)
	}

	if payload.Name != "issues_list" || payload.Description != "List issues" {
		t.Fatalf("payload = %#v", payload)
	}

	if payload.InputSchema["type"] != "object" {
		t.Fatalf("inputSchema = %#v", payload.InputSchema)
	}
}

func TestListCapabilitiesHandler_AuthGate(t *testing.T) {
	t.Parallel()

	handler := buildListCapabilitiesHandler([]ToolSpec{{Name: "issues_list"}})
	res, err := handler(context.Background(), newCallToolRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}

	if !res.IsError {
		t.Fatalf("expected auth error, got: %+v", res)
	}
}
