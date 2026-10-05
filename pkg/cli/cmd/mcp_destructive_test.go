package cmd

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
)

func TestCommandSchemaRegistryIncludesDestructiveCommands(t *testing.T) {
	registry := CommandSchemaRegistry()

	wantOutput := map[string]any{
		"escape-cli profiles delete":     v3.DeleteProfile200Response{},
		"escape-cli assets delete":       v3.DeleteProfile200Response{},
		"escape-cli assets bulk-delete":  v3.BulkUpdateAssets200Response{},
		"escape-cli assets bulk-update":  v3.BulkUpdateAssets200Response{},
		"escape-cli tags delete":         v3.DeleteProfile200Response{},
		"escape-cli custom-rules delete": v3.DeleteCustomRule200Response{},
		"escape-cli locations delete":    v3.DeleteLocation200Response{},
		"escape-cli workflows delete":    v3.CreateWorkflow200Response{},
		"escape-cli integrations delete": v3.CreateakamaiIntegration200Response{},
		"escape-cli issues bulk-update":  v3.BulkUpdateIssues200Response{},
	}

	for path, output := range wantOutput {
		schemas, ok := registry[path]
		if !ok {
			t.Fatalf("expected %q in CommandSchemaRegistry", path)
		}

		if output == nil {
			if schemas.Output != nil {
				t.Fatalf("%s: expected no output schema, got %T", path, schemas.Output)
			}

			continue
		}

		if reflect.TypeOf(schemas.Output) != reflect.TypeOf(output) {
			t.Fatalf("%s: output type %T, want %T", path, schemas.Output, output)
		}
	}
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
