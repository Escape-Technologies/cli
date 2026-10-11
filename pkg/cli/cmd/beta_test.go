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

// TestBetaCommandsMatchThePublicAPITags lists the CLI commands whose public
// API operations are tagged Beta and were missing from betaMCPTools. The
// OpenAPI document is the source of truth; this table is the CLI side of it.
// Operations with no CLI command (scenarios, security-test writes, exports)
// are not listed.
func TestBetaCommandsMatchThePublicAPITags(t *testing.T) {
	paths := []string{
		"escape-cli retests list",
		"escape-cli retests start",
		"escape-cli retests get",
		"escape-cli scans agents",
		"escape-cli scans reasoning",
		"escape-cli org-settings update",
	}
	for _, path := range paths {
		if betaMCPToolNoteFor(path) == "" {
			t.Errorf("%s is missing from betaMCPTools", path)
		}

		command, _, err := rootCmd.Find(strings.Fields(strings.TrimPrefix(path, "escape-cli ")))
		if err != nil {
			t.Errorf("find %s: %v", path, err)
			continue
		}

		if command.CommandPath() != path {
			t.Errorf("find %s resolved to %s", path, command.CommandPath())
		}

		if !strings.Contains(command.Short, "(Beta)") {
			t.Errorf("%s Short = %q, want (Beta)", path, command.Short)
		}
	}
}
