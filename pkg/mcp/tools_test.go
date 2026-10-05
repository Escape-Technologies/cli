package mcp

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
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

func TestBuildCommandArgsRejectsEmptyRequiredFlag(t *testing.T) {
	t.Parallel()

	spec := ToolSpec{
		FlagBindings: []FlagBinding{{
			Property: "asset_id",
			FlagName: "asset-id",
			Kind:     "stringSlice",
			Required: true,
		}},
	}

	for _, args := range []map[string]any{
		{},
		{"asset_id": []any{}},
		{"asset_id": []string{}},
		{"asset_id": ""},
		{"asset_id": nil},
	} {
		_, _, err := buildCommandArgs(spec, args)
		if err == nil {
			t.Fatalf("expected required asset_id to be refused, args %#v", args)
		}
	}

	got, _, err := buildCommandArgs(spec, map[string]any{"asset_id": []any{"asset-1"}})
	if err != nil {
		t.Fatalf("buildCommandArgs: %v", err)
	}

	want := []string{"--asset-id", "asset-1"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("args = %#v, want %#v", got, want)
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

func TestBuildCommandArgsAppliesListPageDefault(t *testing.T) {
	t.Parallel()

	spec := ToolSpec{
		FlagBindings: []FlagBinding{
			{Property: "size", FlagName: "size", Kind: "int"},
			{Property: "cursor", FlagName: "cursor", Kind: "string"},
		},
		DefaultArgs: map[string]any{"size": DefaultListPageSize},
	}

	t.Run("omitted size and cursor", func(t *testing.T) {
		t.Parallel()

		raw := map[string]any{}
		args, _, err := buildCommandArgs(spec, raw)
		if err != nil {
			t.Fatalf("buildCommandArgs: %v", err)
		}

		if !reflect.DeepEqual(args, []string{"--size", "50"}) {
			t.Fatalf("args = %#v", args)
		}

		if _, ok := raw["size"]; ok {
			t.Fatal("default injection must not mutate the caller map")
		}
	})

	t.Run("cursor selects the page", func(t *testing.T) {
		t.Parallel()

		args, _, err := buildCommandArgs(spec, map[string]any{"cursor": "abc"})
		if err != nil {
			t.Fatalf("buildCommandArgs: %v", err)
		}

		if !reflect.DeepEqual(args, []string{"--cursor", "abc"}) {
			t.Fatalf("args = %#v", args)
		}
	})

	t.Run("explicit size", func(t *testing.T) {
		t.Parallel()

		args, _, err := buildCommandArgs(spec, map[string]any{"size": float64(10)})
		if err != nil {
			t.Fatalf("buildCommandArgs: %v", err)
		}

		if !reflect.DeepEqual(args, []string{"--size", "10"}) {
			t.Fatalf("args = %#v", args)
		}
	})

	t.Run("no defaults", func(t *testing.T) {
		t.Parallel()

		plain := spec
		plain.DefaultArgs = nil
		args, _, err := buildCommandArgs(plain, map[string]any{})
		if err != nil {
			t.Fatalf("buildCommandArgs: %v", err)
		}

		if len(args) != 0 {
			t.Fatalf("args = %#v", args)
		}
	})
}

func TestTruncatedToolResultMarksError(t *testing.T) {
	t.Parallel()

	message := truncatedStdoutMessage(maxStdoutBytes)
	result := truncatedToolResult(&ExecutionResult{
		Stdout: message,
		Payload: map[string]any{
			"truncated": true,
			"error":     message,
		},
	}, nil)
	if !result.IsError {
		t.Fatal("expected tool error")
	}

	payload, ok := result.StructuredContent.(map[string]any)
	if !ok || payload["truncated"] != true || payload["error"] != message {
		t.Fatalf("structured = %#v", result.StructuredContent)
	}

	text, ok := result.Content[0].(mcpgo.TextContent)
	if !ok || text.Text != message {
		t.Fatalf("content = %#v", result.Content)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("structured payload is not valid JSON: %v", err)
	}

	if !json.Valid(encoded) || !strings.Contains(string(encoded), `"truncated":true`) {
		t.Fatalf("encoded = %s", encoded)
	}
}

func TestTruncatedToolResultKeepsFailureCause(t *testing.T) {
	t.Parallel()

	message := truncatedStdoutMessage(maxStdoutBytes)
	result := truncatedToolResult(&ExecutionResult{
		Stdout:          message,
		Stderr:          "Error:\n  Invalid API Key.",
		StdoutTruncated: true,
		ExitCode:        1,
		Payload: map[string]any{
			"truncated": true,
			"error":     message,
		},
	}, errors.New("exit status 1"))
	if !result.IsError {
		t.Fatal("expected tool error")
	}

	payload, ok := result.StructuredContent.(map[string]any)
	if !ok || payload["truncated"] != true || payload["error"] != message {
		t.Fatalf("structured = %#v", result.StructuredContent)
	}

	if payload["exitError"] != "exit status 1" || payload["stderr"] != "Error:\n  Invalid API Key." {
		t.Fatalf("structured payload lost the failure cause: %#v", payload)
	}

	text, ok := result.Content[0].(mcpgo.TextContent)
	if !ok || !strings.Contains(text.Text, "Invalid API Key.") || !strings.Contains(text.Text, message) {
		t.Fatalf("content = %#v", result.Content)
	}
}

func TestBuildCommandArgsAppendsForcedArgsLast(t *testing.T) {
	t.Parallel()

	args, _, err := buildCommandArgs(ToolSpec{
		FlagBindings: []FlagBinding{{Property: "timeout", FlagName: "timeout", Kind: "string"}},
		ForcedArgs:   []string{"--timeout", "25s", "--pending-on-timeout=true"},
	}, map[string]any{"timeout": "10m"})
	if err != nil {
		t.Fatalf("expected command args, got error: %v", err)
	}

	want := []string{"--timeout", "10m", "--timeout", "25s", "--pending-on-timeout=true"}
	if len(args) != len(want) {
		t.Fatalf("expected %#v, got %#v", want, args)
	}

	for index, value := range want {
		if args[index] != value {
			t.Fatalf("expected %#v, got %#v", want, args)
		}
	}
}

func TestCommandFailureTextHidesKillOnDeadline(t *testing.T) {
	t.Parallel()

	err := executionTimeoutFailure(ExecutionOptions{
		DisplayCommand: []string{"scans", "get"},
		Timeout:        defaultToolExecutionTimeout,
	})
	got := commandFailureText(err, &ExecutionResult{
		Stderr:   "signal: killed",
		Stdout:   "partial",
		ExitCode: -1,
	})
	if got != err.Error() {
		t.Fatalf("expected only the deadline text, got %q", got)
	}

	if strings.Contains(got, "signal: killed") || strings.Contains(got, "exit code") || strings.Contains(got, "partial") {
		t.Fatalf("deadline text includes subprocess noise: %q", got)
	}
}

func TestCommandFailureTextKeepsOutputForOrdinaryFailures(t *testing.T) {
	t.Parallel()

	got := commandFailureText(errors.New(`command "scans get" failed with exit code 1`), &ExecutionResult{
		Stderr: "bad id",
		Stdout: "partial",
	})
	for _, part := range []string{`command "scans get" failed with exit code 1`, "bad id", "partial"} {
		if !strings.Contains(got, part) {
			t.Fatalf("expected %q in %q", part, got)
		}
	}
}

func TestEmailsWaitTimeoutFitsToolBudget(t *testing.T) {
	t.Parallel()

	if EmailsWaitTimeout <= 0 || EmailsWaitTimeout >= defaultToolExecutionTimeout {
		t.Fatalf("emails wait budget %s must be inside %s", EmailsWaitTimeout, defaultToolExecutionTimeout)
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

func TestBuildCommandArgsAppliesFlagDefaultWhenOmitted(t *testing.T) {
	t.Parallel()

	args, _, err := buildCommandArgs(ToolSpec{
		PositionalArgs: []string{"profile_id"},
		FlagBindings: []FlagBinding{
			{Property: "merge", FlagName: "merge", Kind: "bool", Default: true},
		},
	}, map[string]any{"profile_id": "prof-1"})
	if err != nil {
		t.Fatalf("buildCommandArgs: %v", err)
	}

	want := []string{"prof-1", "--merge=true"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestBuildCommandArgsExplicitFalseOverridesFlagDefault(t *testing.T) {
	t.Parallel()

	args, _, err := buildCommandArgs(ToolSpec{
		FlagBindings: []FlagBinding{
			{Property: "merge", FlagName: "merge", Kind: "bool", Default: true},
		},
	}, map[string]any{"merge": false})
	if err != nil {
		t.Fatalf("buildCommandArgs: %v", err)
	}

	want := []string{"--merge=false"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}
