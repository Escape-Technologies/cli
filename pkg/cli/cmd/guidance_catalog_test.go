package cmd

import (
	"regexp"
	"testing"

	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
)

// guidanceToolToken matches the snake_case words guidance tells a model to
// call. Dotted fields (overall.statuses) and camelCase (assetId) do not match.
var guidanceToolToken = regexp.MustCompile(`[a-z]+_[a-z_]+`)

// guidanceAPILiterals are snake_case values the instructions quote because
// the API requires that exact string. They are not tool names. A new
// snake_case word that is neither a tool nor listed here fails the test.
var guidanceAPILiterals = map[string]struct{}{
	"read_only": {},
}

func TestGuidanceNamesOnlyCatalogTools(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	catalog := map[string]struct{}{}
	for _, spec := range specs {
		catalog[spec.Name] = struct{}{}
	}

	for _, name := range climcp.BuiltinToolNames() {
		catalog[name] = struct{}{}
	}

	for _, name := range climcp.KnowledgeToolNames() {
		catalog[name] = struct{}{}
	}

	texts, err := climcp.GuidanceTexts()
	if err != nil {
		t.Fatalf("guidance texts: %v", err)
	}

	seen := map[string]struct{}{}
	for _, text := range texts {
		for _, token := range guidanceToolToken.FindAllString(text, -1) {
			seen[token] = struct{}{}
			if _, ok := catalog[token]; ok {
				continue
			}

			if _, ok := guidanceAPILiterals[token]; ok {
				continue
			}

			t.Errorf("guidance names %q, which is not a catalog tool", token)
		}
	}

	if len(seen) == 0 {
		t.Fatal("guidance contained no tool-shaped tokens")
	}

	for _, name := range []string{"scans_start", "scans_get", "escape_get_tool_spec", "profiles_update_configuration"} {
		if _, ok := seen[name]; !ok {
			t.Errorf("guidance does not name %s", name)
		}
	}

	if _, ok := seen["scans_watch"]; ok {
		t.Fatal("guidance still names scans_watch")
	}
}
