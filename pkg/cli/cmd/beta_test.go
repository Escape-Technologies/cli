package cmd

import (
	"strings"
	"testing"
)

// Every Beta tool path must be a real MCP tool and its catalog description must say
// so: the map is hand-maintained, so a renamed command would silently drop the
// marker from the model's context.
func TestBetaMCPToolsCarryTheMarker(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	registered := make(map[string]string, len(specs))
	for _, spec := range specs {
		registered[spec.Path] = spec.Description
	}

	if len(betaMCPTools) == 0 {
		t.Fatal("no Beta tool is declared")
	}

	for path := range betaMCPTools {
		description, ok := registered[path]
		if !ok {
			t.Errorf("%s is declared Beta but is not an MCP tool", path)
			continue
		}

		if !strings.Contains(strings.ToLower(description), "beta") {
			t.Errorf("%s description does not mention Beta: %q", path, description)
		}
	}
}
