package cmd

import (
	"slices"
	"testing"
)

func TestCommandSchemaRegistryIncludesIssueMutationTools(t *testing.T) {
	registry := CommandSchemaRegistry()

	for _, path := range []string{
		"escape-cli issues update",
		"escape-cli issues bulk-update",
	} {
		if _, ok := registry[path]; !ok {
			t.Fatalf("expected %q in CommandSchemaRegistry", path)
		}
	}
}

func TestBuildMCPToolSpecsIncludesIssueMutationTools(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}

	for _, want := range []string{"issues_update", "issues_bulk_update"} {
		if !slices.Contains(names, want) {
			t.Fatalf("expected MCP tool %q, got %v", want, names)
		}
	}
}
