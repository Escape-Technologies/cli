package cmd

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
)

// mcpToolJSONSchema is the subset of the tool's JSON Schema documents the
// tests assert against.
type mcpToolJSONSchema struct {
	Type       string                    `json:"type"`
	Required   []string                  `json:"required"`
	Properties map[string]map[string]any `json:"properties"`
}

func decodeMCPToolSchema(t *testing.T, name string, raw []byte) mcpToolJSONSchema {
	t.Helper()

	var schema mcpToolJSONSchema
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal schema for %q: %v", name, err)
	}

	return schema
}

func specByName(t *testing.T, specs []climcp.ToolSpec, name string) climcp.ToolSpec {
	t.Helper()

	for _, spec := range specs {
		if spec.Name == name {
			return spec
		}
	}

	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}

	t.Fatalf("expected MCP tool %q, got %v", name, names)

	return climcp.ToolSpec{}
}

func TestCommandSchemaRegistryIncludesScanCancelIgnoreAndProblemTools(t *testing.T) {
	registry := CommandSchemaRegistry()

	tests := []struct {
		path       string
		wantOutput any
	}{
		{"escape-cli scans cancel", out.Message{}},
		{"escape-cli scans ignore", out.Message{}},
		{"escape-cli scans problems", Page[v3.ScanSummarizedWithProblems2]{}},
		{"escape-cli profiles problems", Page[v3.ProfileScanProblemsRow]{}},
		{"escape-cli problems", problemsCommandOutput},
	}

	for _, tt := range tests {
		schemas, ok := registry[tt.path]
		if !ok {
			t.Fatalf("expected %q in CommandSchemaRegistry", tt.path)
		}

		if !reflect.DeepEqual(schemas.Output, tt.wantOutput) {
			t.Fatalf("expected %q output %T, got %T", tt.path, tt.wantOutput, schemas.Output)
		}
	}
}

func TestBuildMCPToolSpecsExposesScanCancelIgnoreAndProblemTools(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	tests := []struct {
		name           string
		command        []string
		positional     []string
		required       []string
		flagProperties []string
	}{
		{
			name:       "scans_cancel",
			command:    []string{"scans", "cancel"},
			positional: []string{"scan_id"},
			required:   []string{"scan_id"},
		},
		{
			name:       "scans_ignore",
			command:    []string{"scans", "ignore"},
			positional: []string{"scan_id"},
			required:   []string{"scan_id"},
		},
		{
			name:           "scans_problems",
			command:        []string{"scans", "problems"},
			flagProperties: []string{"all_kinds"},
		},
		{
			name:           "profiles_problems",
			command:        []string{"profiles", "problems"},
			flagProperties: []string{"asset_id", "search"},
		},
		{
			name:           "problems",
			command:        []string{"problems"},
			flagProperties: []string{"search"},
		},
	}

	for _, tt := range tests {
		spec := specByName(t, specs, tt.name)

		if !slices.Equal(spec.Command, tt.command) {
			t.Fatalf("%s: expected command %v, got %v", tt.name, tt.command, spec.Command)
		}

		if !slices.Equal(spec.PositionalArgs, tt.positional) {
			t.Fatalf("%s: expected positional args %v, got %v", tt.name, tt.positional, spec.PositionalArgs)
		}

		input := decodeMCPToolSchema(t, tt.name, spec.Tool.RawInputSchema)
		if !slices.Equal(input.Required, tt.required) {
			t.Fatalf("%s: expected required properties %v, got %v", tt.name, tt.required, input.Required)
		}

		for _, arg := range tt.positional {
			property, ok := input.Properties[arg]
			if !ok {
				t.Fatalf("%s: expected positional property %q in %v", tt.name, arg, input.Properties)
			}

			if property["type"] != "string" {
				t.Fatalf("%s: expected positional %q to be string, got %v", tt.name, arg, property["type"])
			}
		}

		for _, property := range tt.flagProperties {
			if _, ok := input.Properties[property]; !ok {
				t.Fatalf("%s: expected flag property %q in %v", tt.name, property, input.Properties)
			}
		}
	}
}

func TestBuildMCPToolSpecsScanCancelIgnoreOutputSchema(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	for _, name := range []string{"scans_cancel", "scans_ignore"} {
		spec := specByName(t, specs, name)
		if len(spec.Tool.RawOutputSchema) == 0 {
			t.Fatalf("%s: expected advertised output schema for message-only command", name)
		}

		output := decodeMCPToolSchema(t, name, spec.Tool.RawOutputSchema)
		if output.Type != "object" {
			t.Fatalf("%s: expected object output schema, got %q", name, output.Type)
		}

		if !slices.Contains(output.Required, "msg") {
			t.Fatalf("%s: expected required property msg, got %v", name, output.Required)
		}
	}

	// Array outputs are wrapped as {"items":[...]} and advertised as that object.
	// Root problems is oneOf the summary page and the --all detail page.
	for _, name := range []string{"scans_problems", "profiles_problems"} {
		spec := specByName(t, specs, name)
		if len(spec.Tool.RawOutputSchema) == 0 {
			t.Fatalf("%s: expected wrapped items output schema", name)
		}

		output := decodeMCPToolSchema(t, name, spec.Tool.RawOutputSchema)
		if output.Type != "object" {
			t.Fatalf("%s: expected object output schema, got %q", name, output.Type)
		}

		items, ok := output.Properties["items"]
		if !ok || items["type"] != "array" {
			t.Fatalf("%s: expected properties.items array, got %#v", name, output.Properties["items"])
		}
	}

	problems := specByName(t, specs, "problems")
	var problemsOutput map[string]any
	if err := json.Unmarshal(problems.Tool.RawOutputSchema, &problemsOutput); err != nil {
		t.Fatalf("problems output schema: %v", err)
	}

	props, _ := problemsOutput["properties"].(map[string]any)
	items, _ := props["items"].(map[string]any)
	oneOf, _ := items["oneOf"].([]any)
	if len(oneOf) != 2 {
		t.Fatalf("problems items oneOf = %#v, want summary and detail", items["oneOf"])
	}
}

func TestProfileProblemsFiltersAreIndependentOfList(t *testing.T) {
	for _, name := range []string{"asset-id", "search", "domain", "issue-id", "tag-id", "initiator", "kind", "risk"} {
		list := profilesListCmd.Flags().Lookup(name)
		problems := profileProblemsCmd.Flags().Lookup(name)
		if list == nil || problems == nil {
			t.Fatalf("expected %s on both profiles list and profiles problems", name)
		}

		if list.Value == problems.Value {
			t.Fatalf("profiles problems %s shares the list command flag value", name)
		}
	}

	if err := profileProblemsCmd.ParseFlags([]string{"--asset-id", "asset-1", "--search", "prod"}); err != nil {
		t.Fatalf("profiles problems rejected its own filters: %v", err)
	}

	t.Cleanup(func() {
		_ = profileProblemsCmd.Flags().Set("search", "")
		// StringSlice Set appends. Replace the value so later tests see the default.
		if value, ok := profileProblemsCmd.Flags().Lookup("asset-id").Value.(interface{ Replace([]string) error }); ok {
			_ = value.Replace(nil)
		}
	})
}

// readOnlyToolPaths are the read-only command paths newly registered for
// PRD-2169; they were missing from the MCP tool catalog allowlist.
var readOnlyToolPaths = []string{
	"escape-cli stats",
	"escape-cli issues funnel",
	"escape-cli issues trends",
	"escape-cli assets list-activities",
	"escape-cli profiles get-schema",
}

func TestCommandSchemaRegistryIncludesReadOnlyTools(t *testing.T) {
	registry := CommandSchemaRegistry()

	for _, path := range readOnlyToolPaths {
		schemas, ok := registry[path]
		if !ok {
			t.Fatalf("expected %q in CommandSchemaRegistry", path)
		}

		if schemas.Output == nil {
			t.Fatalf("expected %q to declare an Output type", path)
		}
	}
}

func TestBuildMCPToolSpecsIncludesReadOnlyTools(t *testing.T) {
	specs := mcpSpecs(t)

	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}

	for _, want := range []string{
		"stats",
		"issues_funnel",
		"issues_trends",
		"assets_list_activities",
		"profiles_get_schema",
	} {
		if !slices.Contains(names, want) {
			t.Fatalf("expected MCP tool %q, got %v", want, names)
		}
	}
}

func TestReadOnlyToolSpecsAdvertiseOutputSchemas(t *testing.T) {
	capabilities := BuildCommandCapabilities(rootCmd, CommandSchemaRegistry())

	for _, path := range readOnlyToolPaths {
		found := false
		for _, capability := range capabilities {
			if capability.Path != path {
				continue
			}

			found = true
			if !capability.HasOutSchema {
				t.Fatalf("expected %q to carry an output schema", path)
			}
		}

		if !found {
			t.Fatalf("expected %q in the built capabilities", path)
		}
	}

	// MCP outputSchema requires a top-level object, so every read-only tool
	// advertises one: object commands their DTO schema directly, array
	// commands the {"items":[...]} wrapper the executor returns.
	specs := mcpSpecs(t)
	for _, spec := range specs {
		if !slices.Contains(readOnlyToolPaths, spec.Path) {
			continue
		}

		if len(spec.Tool.RawOutputSchema) == 0 {
			t.Fatalf("tool %q must advertise an output schema", spec.Name)
		}

		output := decodeMCPToolSchema(t, spec.Name, spec.Tool.RawOutputSchema)
		if output.Type != "object" {
			t.Fatalf("tool %q output schema type = %q, want object", spec.Name, output.Type)
		}
	}

	for _, name := range []string{"issues_funnel", "issues_trends", "assets_list_activities"} {
		spec := specByName(t, specs, name)
		output := decodeMCPToolSchema(t, name, spec.Tool.RawOutputSchema)
		items, ok := output.Properties["items"]
		if !ok || items["type"] != "array" {
			t.Fatalf("%s: expected properties.items array, got %#v", name, output.Properties["items"])
		}
	}

	for _, name := range []string{"stats", "profiles_get_schema"} {
		spec := specByName(t, specs, name)
		output := decodeMCPToolSchema(t, name, spec.Tool.RawOutputSchema)
		if _, wrapped := output.Properties["items"]; wrapped {
			t.Fatalf("%s: object output must not be items-wrapped: %s", name, spec.Tool.RawOutputSchema)
		}
	}
}

func TestProfilesGetSchemaToolOmitsFileFlag(t *testing.T) {
	spec := specByName(t, mcpSpecs(t), "profiles_get_schema")

	input := decodeMCPToolSchema(t, "profiles_get_schema", spec.Tool.RawInputSchema)
	if _, exists := input.Properties["file"]; exists {
		t.Fatal(`profiles_get_schema must not expose "file": -f/--file writes to arbitrary paths on the MCP server host`)
	}

	for _, binding := range spec.FlagBindings {
		if binding.FlagName == "file" {
			t.Fatalf("profiles_get_schema must not bind the file flag: %+v", binding)
		}
	}

	// Metadata flags stay. The path flags are hidden on every command.
	for _, want := range []string{"schema_id", "timeout"} {
		if _, exists := input.Properties[want]; !exists {
			t.Fatalf("expected profiles_get_schema to keep the %q flag, got %v", want, sortedKeys(input.Properties))
		}
	}
}

func TestMCPToolsOmitFilesystemFlags(t *testing.T) {
	blocked := map[string]struct{}{
		"file":             {},
		"out":              {},
		"output-file-type": {},
	}
	for _, spec := range mcpSpecs(t) {
		for _, binding := range spec.FlagBindings {
			if _, ok := blocked[binding.FlagName]; ok {
				t.Errorf("%s binds filesystem flag %q", spec.Name, binding.FlagName)
			}
		}
	}

	uploadSchema, _, err := rootCmd.Find([]string{"profiles", "upload-schema"})
	if err != nil {
		t.Fatalf("find profiles upload-schema: %v", err)
	}

	if !skipMCPFlag(uploadSchema, "file") {
		t.Fatal("profiles upload-schema --file would read an arbitrary path if the command is registered")
	}
}

func sortedKeys(m map[string]map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}
