package mcp

import (
	"reflect"
	"testing"
)

func TestBuildCommandArgs(t *testing.T) {
	t.Parallel()

	spec := ToolSpec{
		PositionalArgs: []string{"scan_id"},
		FlagBindings: []FlagBinding{
			{Property: "watch", FlagName: "watch", Kind: "bool"},
			{Property: "status", FlagName: "status", Kind: "stringSlice"},
			{Property: "limit", FlagName: "limit", Kind: "int"},
		},
		BodyProperty:   "body",
		AllowExtraArgs: true,
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

func TestBuildCommandArgsIgnoresExtraArgsByDefault(t *testing.T) {
	t.Parallel()

	args, _, err := buildCommandArgs(
		ToolSpec{PositionalArgs: []string{"scan_id"}},
		map[string]any{
			"scan_id": "scan-1",
			"args":    []any{"--injected"},
		},
	)
	if err != nil {
		t.Fatalf("expected command args, got error: %v", err)
	}
	for _, value := range args {
		if value == "--injected" {
			t.Fatalf("expected free-form args to be ignored without AllowExtraArgs, got %#v", args)
		}
	}
}

func TestBuildCommandArgsRejectsDashPrefixedPositionals(t *testing.T) {
	t.Parallel()

	_, _, err := buildCommandArgs(
		ToolSpec{PositionalArgs: []string{"scan_id"}},
		map[string]any{"scan_id": "--watch=true"},
	)
	if err == nil {
		t.Fatal("expected positional guard error")
	}
	if err.Error() != `positional "scan_id" must not start with '-'` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWrapStructuredPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "object passes through",
			input:    map[string]any{"id": "abc"},
			expected: map[string]any{"id": "abc"},
		},
		{
			name:     "array wraps under items",
			input:    []any{map[string]any{"id": "abc"}, map[string]any{"id": "def"}},
			expected: map[string]any{"items": []any{map[string]any{"id": "abc"}, map[string]any{"id": "def"}}},
		},
		{
			name:     "primitive wraps under value",
			input:    "hello",
			expected: map[string]any{"value": "hello"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			actual := wrapStructuredPayload(testCase.input)
			if !reflect.DeepEqual(actual, testCase.expected) {
				t.Fatalf("expected %#v, got %#v", testCase.expected, actual)
			}
		})
	}
}

func TestBuildCommandArgsRejectsOutputOverrideInExtraArgs(t *testing.T) {
	t.Parallel()

	_, _, err := buildCommandArgs(
		ToolSpec{AllowExtraArgs: true},
		map[string]any{"args": []any{"--output=json"}},
	)
	if err == nil {
		t.Fatal("expected output override error")
	}
	if err.Error() != `args must not override the injected "--output" flag` {
		t.Fatalf("unexpected error: %v", err)
	}
}
