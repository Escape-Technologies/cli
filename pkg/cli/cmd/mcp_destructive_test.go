package cmd

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
)

func TestCommandSchemaRegistryIncludesDestructiveCommands(t *testing.T) {
	registry := CommandSchemaRegistry()

	// confirm.go prints a different document on --dry-run than the mutation
	// result, so those commands declare both. workflows delete prints the
	// workflow either way. issues bulk-update forwards the API's own dryRun
	// field on the same response. The expectation is the schema of those
	// types, not the providedSchema wrapper they happen to share.
	wantOutput := map[string]any{
		"escape-cli profiles delete":         outputWithDryRun(v3.DeleteProfile200Response{}, v3.GetProfile200Response{}),
		"escape-cli assets delete":           outputWithDryRun(v3.DeleteProfile200Response{}, v3.AssetDetailed1{}),
		"escape-cli assets bulk-delete":      outputWithDryRun(escape.BulkOperationResult{}, out.Message{}),
		"escape-cli assets bulk-update":      outputWithDryRun(escape.BulkOperationResult{}, out.Message{}),
		"escape-cli tags delete":             outputWithDryRun(v3.DeleteProfile200Response{}, v3.CreateTag200Response{}),
		"escape-cli custom-rules delete":     outputWithDryRun(v3.DeleteCustomRule200Response{}, v3.CreateCustomRule200Response{}),
		"escape-cli locations delete":        outputWithDryRun(v3.DeleteLocation200Response{}, v3.GetLocation200Response{}),
		"escape-cli workflows delete":        v3.CreateWorkflow200Response{},
		"escape-cli integrations delete":     outputWithDryRun(v3.CreateakamaiIntegration200Response{}, map[string]interface{}{}),
		"escape-cli issues bulk-update":      v3.BulkUpdateIssues200Response{},
		"escape-cli regression-tests delete": outputWithDryRun(v3.DeleteProfile200Response{}, v3.CreateRegressionTest200Response{}),
		"escape-cli users deactivate":        outputWithDryRun(v3.ListUsers200ResponseInner{}, v3.GetUser200Response{}),
	}

	for path, output := range wantOutput {
		schemas, ok := registry[path]
		if !ok {
			t.Fatalf("expected %q in CommandSchemaRegistry", path)
		}

		if got, want := outputSchemaDocument(t, schemas.Output), outputSchemaDocument(t, output); got != want {
			t.Fatalf("%s declared schema does not match the result and dry-run preview", path)
		}
	}
}

func outputSchemaDocument(t *testing.T, value any) string {
	t.Helper()

	encoded, err := json.Marshal(clischema.Generate(value))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	return string(encoded)
}

func TestBuildMCPToolSpecsRequiresConfirmOnDestructiveTools(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	byName := make(map[string]climcp.ToolSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}

	for _, name := range []string{
		"profiles_delete",
		"assets_delete",
		"assets_bulk_delete",
		"assets_bulk_update",
		"tags_delete",
		"custom_rules_delete",
		"locations_delete",
		"workflows_delete",
		"roles_unbind",
		"integrations_delete",
		"issues_bulk_update",
	} {
		spec, ok := byName[name]
		if !ok {
			t.Fatalf("expected MCP tool %q", name)
		}

		if !spec.Destructive {
			t.Fatalf("%s: expected Destructive", name)
		}

		schema := mcpInputSchema(t, spec)
		props, _ := schema["properties"].(map[string]any)
		confirm, _ := props[climcp.ConfirmProperty].(map[string]any)
		if confirm["type"] != "boolean" {
			t.Fatalf("%s: confirm type = %#v", name, confirm["type"])
		}

		if confirm["description"] != climcp.ConfirmPropertyDescription {
			t.Fatalf("%s: confirm description = %#v", name, confirm["description"])
		}

		required, _ := schema["required"].([]any)
		if !slices.Contains(required, any(climcp.ConfirmProperty)) {
			t.Fatalf("%s: confirm is not required, got %#v", name, required)
		}

		for _, binding := range spec.FlagBindings {
			if binding.Property == climcp.ConfirmProperty || binding.FlagName == climcp.ConfirmProperty {
				t.Fatalf("%s: confirm is bound as CLI flag %#v", name, binding)
			}
		}
	}

	assetsGet, ok := byName["assets_get"]
	if !ok {
		t.Fatal("expected assets_get tool")
	}

	if assetsGet.Destructive {
		t.Fatal("assets_get should not be destructive")
	}

	props, _ := mcpInputSchema(t, assetsGet)["properties"].(map[string]any)
	if _, exists := props[climcp.ConfirmProperty]; exists {
		t.Fatal("assets_get schema includes confirm")
	}
}

func TestProjectsAndRolesDeleteAreNotMCPTools(t *testing.T) {
	registry := CommandSchemaRegistry()
	for _, path := range []string{"escape-cli projects delete", "escape-cli roles delete"} {
		if _, ok := registry[path]; ok {
			t.Errorf("%s is still in CommandSchemaRegistry", path)
		}

		if isDestructiveMCPCommand(path) {
			t.Errorf("%s is still marked destructive for MCP", path)
		}
	}

	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	for _, spec := range specs {
		if spec.Name == "projects_delete" || spec.Name == "roles_delete" {
			t.Errorf("MCP tool %s is still registered", spec.Name)
		}
	}
}

func TestIntegrationsDeleteExposesKindFlag(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	var spec climcp.ToolSpec
	for _, candidate := range specs {
		if candidate.Name == "integrations_delete" {
			spec = candidate
			break
		}
	}

	if spec.Name == "" {
		t.Fatal("missing integrations_delete")
	}

	for _, binding := range spec.FlagBindings {
		if binding.FlagName == "kind" && binding.Property == "kind" && binding.Kind == "string" {
			return
		}
	}

	t.Fatalf("integrations_delete missing --kind binding, got %#v", spec.FlagBindings)
}

func mcpInputSchema(t *testing.T, spec climcp.ToolSpec) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
		t.Fatalf("unmarshal schema for %s: %v", spec.Name, err)
	}

	return schema
}
