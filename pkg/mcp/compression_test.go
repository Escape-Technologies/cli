package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func TestBuildStubToolKeepsFlagDescriptionAndEnum(t *testing.T) {
	t.Parallel()

	long := "status filter\n" + strings.Repeat("token ", 40)
	rawInput, err := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status": map[string]any{
				"type":        "array",
				"description": long,
				"items": map[string]any{
					"type": "string",
					"enum": []string{"OPEN", "RESOLVED", "FALSE_POSITIVE"},
				},
			},
			"sort_direction": map[string]any{
				"type":        "string",
				"description": "sort direction: asc, desc",
				"enum":        []string{"asc", "desc"},
			},
			"watch": map[string]any{
				"type":        "boolean",
				"description": "watch the\njob",
			},
		},
		"required": []string{"status"},
	})
	if err != nil {
		t.Fatal(err)
	}

	spec := ToolSpec{
		Name:        "issues_list",
		Description: "List issues",
		PositionalArgs: []string{
			"issue_id",
		},
		FlagBindings: []FlagBinding{
			{Property: "status", FlagName: "status", Kind: "stringSlice"},
			{Property: "sort_direction", FlagName: "sort-direction", Kind: "string"},
			{Property: "watch", FlagName: "watch", Kind: "bool"},
		},
		Tool: mcpgo.NewToolWithRawSchema("issues_list", "List issues", rawInput),
	}

	stub, err := BuildStubTool(spec)
	if err != nil {
		t.Fatal(err)
	}

	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
		Required   []string                  `json:"required"`
	}
	if err := json.Unmarshal(stub.RawInputSchema, &schema); err != nil {
		t.Fatal(err)
	}

	status := schema.Properties["status"]
	if status == nil {
		t.Fatal("missing status")
	}

	desc, _ := status["description"].(string)
	if desc == "" || strings.Contains(desc, "\n") {
		t.Fatalf("status description = %q", desc)
	}

	if len([]rune(desc)) > stubFlagDescriptionRunes {
		t.Fatalf("status description length %d: %q", len([]rune(desc)), desc)
	}

	if !strings.HasPrefix(desc, "status filter token") || !strings.HasSuffix(desc, "...") {
		t.Fatalf("status description = %q", desc)
	}

	items, _ := status["items"].(map[string]any)
	if items == nil {
		t.Fatal("missing status items")
	}

	gotEnum := enumStrings(t, items["enum"])
	wantEnum := []string{"OPEN", "RESOLVED", "FALSE_POSITIVE"}
	if strings.Join(gotEnum, ",") != strings.Join(wantEnum, ",") {
		t.Fatalf("status enum = %#v", gotEnum)
	}

	direction := schema.Properties["sort_direction"]
	if direction["description"] != "sort direction: asc, desc" {
		t.Fatalf("direction description = %#v", direction["description"])
	}

	if strings.Join(enumStrings(t, direction["enum"]), ",") != "asc,desc" {
		t.Fatalf("direction enum = %#v", direction["enum"])
	}

	watch := schema.Properties["watch"]
	if watch["description"] != "watch the job" || watch["type"] != "boolean" {
		t.Fatalf("watch = %#v", watch)
	}

	if !stringSliceContains(schema.Required, "status") || !stringSliceContains(schema.Required, "issue_id") {
		t.Fatalf("required = %#v", schema.Required)
	}
}

func TestFlagStubSchemaCopiesDefault(t *testing.T) {
	t.Parallel()

	usage := "Deep-merge stdin into the stored configuration. CLI default off (full replace). MCP default: true"
	if len([]rune(usage)) > stubFlagDescriptionRunes {
		t.Fatalf("usage length %d exceeds stub cap %d", len([]rune(usage)), stubFlagDescriptionRunes)
	}

	got := flagStubSchema("bool", map[string]any{
		"type":        "boolean",
		"description": usage,
		"default":     true,
	})
	if got["default"] != true {
		t.Fatalf("default = %#v", got["default"])
	}

	desc, _ := got["description"].(string)
	if !strings.Contains(desc, "MCP default: true") {
		t.Fatalf("description = %q", desc)
	}
}

func TestCompactDescriptionCollapsesWhitespace(t *testing.T) {
	t.Parallel()

	if got := compactDescription("  sort\tdirection: asc, desc  "); got != "sort direction: asc, desc" {
		t.Fatalf("got %q", got)
	}

	long := strings.Repeat("a", stubFlagDescriptionRunes+1)
	got := compactDescription(long)
	if len([]rune(got)) != stubFlagDescriptionRunes {
		t.Fatalf("len %d: %q", len([]rune(got)), got)
	}

	if !strings.HasSuffix(got, "...") {
		t.Fatalf("got %q", got)
	}
}

func enumStrings(t *testing.T, value any) []string {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("enum type %T", value)
	}

	out := make([]string, len(items))
	for i, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("enum value %T", item)
		}

		out[i] = text
	}

	return out
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}
