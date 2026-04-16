package mcp

import "testing"

func TestBuildCommandArgs(t *testing.T) {
	t.Parallel()

	spec := ToolSpec{
		PositionalArgs: []string{"scan_id"},
		FlagBindings: []FlagBinding{
			{Property: "watch", FlagName: "watch", Kind: "bool"},
			{Property: "status", FlagName: "status", Kind: "stringSlice"},
			{Property: "limit", FlagName: "limit", Kind: "int"},
		},
		BodyProperty: "body",
	}

	args, body, err := buildCommandArgs(spec, map[string]any{
		"scan_id": "scan-1",
		"watch":   true,
		"status":  []any{"RUNNING", "FAILED"},
		"limit":   float64(10),
		"args":    []any{"tail"},
		"body": map[string]any{
			"name": "demo",
		},
	})
	if err != nil {
		t.Fatalf("expected command args, got error: %v", err)
	}

	expected := []string{
		"scan-1",
		"tail",
		"--watch=true",
		"--status", "RUNNING",
		"--status", "FAILED",
		"--limit", "10",
	}
	if len(args) != len(expected) {
		t.Fatalf("expected %d args, got %d: %#v", len(expected), len(args), args)
	}
	for index, value := range expected {
		if args[index] != value {
			t.Fatalf("expected arg %d to be %q, got %q", index, value, args[index])
		}
	}

	bodyMap, ok := body.(map[string]any)
	if !ok {
		t.Fatalf("expected body map, got %T", body)
	}
	if bodyMap["name"] != "demo" {
		t.Fatalf("expected body name demo, got %#v", bodyMap["name"])
	}
}

func TestBuildCommandArgsRequiresNamedPositionals(t *testing.T) {
	t.Parallel()

	_, _, err := buildCommandArgs(ToolSpec{PositionalArgs: []string{"scan_id"}}, map[string]any{})
	if err == nil {
		t.Fatal("expected missing positional arg error")
	}
}
