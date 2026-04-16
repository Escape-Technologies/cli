package mcp

import (
	"strings"
	"testing"
)

func TestBuildCommandEnvStripsInheritedAuthVars(t *testing.T) {
	t.Setenv("ESCAPE_API_URL", "https://parent.example.com")
	t.Setenv("ESCAPE_API_KEY", "parent-key")
	t.Setenv("ESCAPE_AUTHORIZATION", "Key parent")
	t.Setenv("UNRELATED_VAR", "keep-me")

	env := buildCommandEnv(ExecutionOptions{
		PublicAPIURL: "https://request.example.com",
		Auth:         Auth{APIKey: "request-key"},
	})

	counts := map[string]int{
		"ESCAPE_API_URL=":        0,
		"ESCAPE_API_KEY=":        0,
		"ESCAPE_AUTHORIZATION=":  0,
		"ESCAPE_COLOR_DISABLED=": 0,
		"UNRELATED_VAR=":         0,
	}
	for _, entry := range env {
		for prefix := range counts {
			if strings.HasPrefix(entry, prefix) {
				counts[prefix]++
			}
		}
	}

	if counts["ESCAPE_API_URL="] != 1 {
		t.Fatalf("expected exactly one ESCAPE_API_URL entry, got %d", counts["ESCAPE_API_URL="])
	}
	if counts["ESCAPE_API_KEY="] != 1 {
		t.Fatalf("expected exactly one ESCAPE_API_KEY entry, got %d", counts["ESCAPE_API_KEY="])
	}
	if counts["ESCAPE_AUTHORIZATION="] != 0 {
		t.Fatalf("expected ESCAPE_AUTHORIZATION to be stripped when not provided by request, got %d", counts["ESCAPE_AUTHORIZATION="])
	}
	if counts["UNRELATED_VAR="] != 1 {
		t.Fatalf("expected unrelated parent vars to be preserved, got %d", counts["UNRELATED_VAR="])
	}
	if counts["ESCAPE_COLOR_DISABLED="] != 1 {
		t.Fatalf("expected ESCAPE_COLOR_DISABLED to be set, got %d", counts["ESCAPE_COLOR_DISABLED="])
	}

	want := map[string]string{
		"ESCAPE_API_URL=": "https://request.example.com",
		"ESCAPE_API_KEY=": "request-key",
	}
	for _, entry := range env {
		for prefix, expected := range want {
			if strings.HasPrefix(entry, prefix) {
				value := strings.TrimPrefix(entry, prefix)
				if value != expected {
					t.Fatalf("expected %s%q, got %s%q", prefix, expected, prefix, value)
				}
			}
		}
	}
}

func TestBuildCommandEnvWithoutRequestAuthDoesNotLeakParent(t *testing.T) {
	t.Setenv("ESCAPE_API_URL", "https://parent.example.com")
	t.Setenv("ESCAPE_API_KEY", "parent-key")
	t.Setenv("ESCAPE_AUTHORIZATION", "Key parent")

	env := buildCommandEnv(ExecutionOptions{})

	for _, entry := range env {
		if strings.HasPrefix(entry, "ESCAPE_API_URL=") ||
			strings.HasPrefix(entry, "ESCAPE_API_KEY=") ||
			strings.HasPrefix(entry, "ESCAPE_AUTHORIZATION=") {
			t.Fatalf("expected no inherited ESCAPE auth vars when request is empty, got %q", entry)
		}
	}
}
