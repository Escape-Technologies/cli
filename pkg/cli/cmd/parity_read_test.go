package cmd

import (
	"slices"
	"strings"
	"testing"
)

// parityReadToolPaths are the command paths added for PRD-2299. Every one of
// them must be an MCP tool: the hosted catalog is derived from the registry, so
// a missing registration is a silently unavailable capability.
var parityReadToolPaths = []string{
	"escape-cli security-tests list",
	"escape-cli security-tests get",
	"escape-cli compliance list",
	"escape-cli compliance enabled",
	"escape-cli compliance per-application",
	"escape-cli scans compliance",
	"escape-cli statistics endpoints",
	"escape-cli statistics issues",
	"escape-cli statistics trends",
	"escape-cli permissions list",
	"escape-cli validations start",
	"escape-cli validations get",
	"escape-cli ai-chats list",
	"escape-cli ai-chats get",
	"escape-cli ai-chats messages",
	"escape-cli ai-chats create",
	"escape-cli ai-chats message",
	"escape-cli jobs list",
}

// parityBetaReadToolPaths are the ones whose public-API operation carries the
// Beta tag in services/public-api/src/v3/routes/beta-policy.ts. The catalog
// description has to say so, because that text is what a model keeps in
// context for every tool.
var parityBetaReadToolPaths = []string{
	"escape-cli validations start",
	"escape-cli validations get",
	"escape-cli ai-chats list",
	"escape-cli ai-chats get",
	"escape-cli ai-chats messages",
	"escape-cli ai-chats create",
	"escape-cli ai-chats message",
	"escape-cli jobs list",
}

func TestParityReadGroupsAreRegisteredTools(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	byPath := make(map[string]string, len(specs))
	for _, spec := range specs {
		byPath[spec.Path] = spec.Description
	}

	for _, path := range parityReadToolPaths {
		description, ok := byPath[path]
		if !ok {
			t.Errorf("%s is registered but is not an MCP tool", path)

			continue
		}

		// A truncated description dropped either the page contract or the
		// Beta marker, so a model never saw the rule it carries.
		if strings.HasSuffix(description, "...") {
			t.Errorf("%s description was cut by the rune cap, so its marker is lost: %q", path, description)
		}

		if !slices.Contains(parityBetaReadToolPaths, path) && strings.Contains(strings.ToLower(description), "beta") {
			t.Errorf("%s is not a Beta operation but its description says Beta: %q", path, description)
		}
	}

	for _, path := range parityBetaReadToolPaths {
		description, ok := byPath[path]
		if !ok {
			t.Errorf("%s is declared Beta but is not an MCP tool", path)

			continue
		}

		if !strings.Contains(strings.ToLower(description), "beta") {
			t.Errorf("%s description does not mention Beta: %q", path, description)
		}
	}
}
