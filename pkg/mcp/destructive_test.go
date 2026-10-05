package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

func TestDestructiveHandlerRefusesWithoutConfirm(t *testing.T) {
	t.Parallel()

	spec := ToolSpec{
		Name:           "assets_delete",
		Command:        []string{"assets", "delete"},
		PositionalArgs: []string{"asset_id"},
		Destructive:    true,
	}
	handler := buildToolHandler(spec, CommandExecutionOptions{})

	cases := []map[string]any{
		{"asset_id": "asset-1"},
		{"asset_id": "asset-1", ConfirmProperty: false},
		{"asset_id": "asset-1", ConfirmProperty: "true"},
	}
	for _, args := range cases {
		res, err := handler(context.Background(), newCallToolRequest(args))
		if err != nil {
			t.Fatalf("handler error: %v", err)
		}

		if res == nil || !res.IsError {
			t.Fatalf("expected tool error for %#v, got %#v", args, res)
		}

		text := textFromResult(t, res)
		if !strings.Contains(text, "asset-1") || !strings.Contains(text, "permanently delete the asset") {
			t.Fatalf("refusal does not explain the delete: %s", text)
		}

		if !strings.Contains(text, ConfirmPropertyDescription) {
			t.Fatalf("refusal does not ask for confirmation: %s", text)
		}
	}
}

func TestDestructiveHandlerConfirmTrueReachesAuth(t *testing.T) {
	t.Parallel()

	spec := ToolSpec{
		Name:           "assets_delete",
		Command:        []string{"assets", "delete"},
		PositionalArgs: []string{"asset_id"},
		Destructive:    true,
	}
	handler := buildToolHandler(spec, CommandExecutionOptions{})
	res, err := handler(context.Background(), newCallToolRequest(map[string]any{
		"asset_id":      "asset-1",
		ConfirmProperty: true,
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	text := textFromResult(t, res)
	if !res.IsError || !strings.Contains(text, "missing authentication") {
		t.Fatalf("expected auth failure after confirm, got %s", text)
	}

	if strings.Contains(text, "refusing destructive") {
		t.Fatalf("confirm=true was treated as unconfirmed: %s", text)
	}
}

func TestDestructiveHandlerExplainsBulkUpdate(t *testing.T) {
	t.Parallel()

	spec := ToolSpec{
		Name:    "assets_bulk_update",
		Command: []string{"assets", "bulk-update"},
		FlagBindings: []FlagBinding{
			{Property: "asset_id", FlagName: "asset-id", Kind: "stringSlice"},
			{Property: "status", FlagName: "status", Kind: "string"},
		},
		Destructive: true,
	}
	handler := buildToolHandler(spec, CommandExecutionOptions{})
	res, err := handler(context.Background(), newCallToolRequest(map[string]any{
		"asset_id": []any{"asset-1", "asset-2"},
		"status":   "ARCHIVED",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}

	text := textFromResult(t, res)
	if !res.IsError {
		t.Fatalf("expected refusal, got %s", text)
	}

	for _, part := range []string{
		"update every asset matching",
		"asset_id=asset-1,asset-2",
		"status=ARCHIVED",
		ConfirmPropertyDescription,
	} {
		if !strings.Contains(text, part) {
			t.Fatalf("refusal %q missing %q", text, part)
		}
	}
}

func TestBuildCommandArgsDoesNotForwardConfirm(t *testing.T) {
	t.Parallel()

	args, _, err := buildCommandArgs(ToolSpec{
		PositionalArgs: []string{"integration_id"},
		FlagBindings: []FlagBinding{
			{Property: ConfirmProperty, FlagName: ConfirmProperty, Kind: "bool"},
			{Property: "kind", FlagName: "kind", Kind: "string"},
		},
		Destructive: true,
	}, map[string]any{
		"integration_id": "int-1",
		ConfirmProperty:  true,
		"kind":           "jira",
	})
	if err != nil {
		t.Fatalf("buildCommandArgs: %v", err)
	}

	expected := []string{"int-1", "--kind", "jira"}
	if len(args) != len(expected) {
		t.Fatalf("args = %#v, want %#v", args, expected)
	}

	for i, want := range expected {
		if args[i] != want {
			t.Fatalf("args = %#v, want %#v", args, expected)
		}
	}

	for _, arg := range args {
		if strings.Contains(arg, "confirm") {
			t.Fatalf("confirm was forwarded: %#v", args)
		}
	}
}

func TestBuildStubToolKeepsDestructiveConfirm(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"asset_id": map[string]any{"type": "string"},
			ConfirmProperty: map[string]any{
				"type":        "boolean",
				"description": ConfirmPropertyDescription,
			},
		},
		"required": []any{"asset_id", ConfirmProperty},
	})
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	spec := ToolSpec{
		Name:           "assets_delete",
		Description:    "Delete an asset from your inventory",
		PositionalArgs: []string{"asset_id"},
		Destructive:    true,
		Tool:           mcpgo.NewToolWithRawSchema("assets_delete", "Delete an asset from your inventory", raw),
	}
	stub, err := BuildStubTool(spec)
	if err != nil {
		t.Fatalf("BuildStubTool: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(stub.RawInputSchema, &schema); err != nil {
		t.Fatalf("unmarshal stub: %v", err)
	}

	props, _ := schema["properties"].(map[string]any)
	confirm, _ := props[ConfirmProperty].(map[string]any)
	if confirm["type"] != "boolean" || confirm["description"] != ConfirmPropertyDescription {
		t.Fatalf("stub confirm = %#v", confirm)
	}

	required, _ := schema["required"].([]any)
	found := false
	for _, name := range required {
		if name == ConfirmProperty {
			found = true
		}
	}

	if !found {
		t.Fatalf("stub required = %#v", required)
	}
}
