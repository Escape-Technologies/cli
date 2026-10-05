package cmd

import (
	"encoding/json"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// TestBuildMCPToolSpecsGeneratesAuditAndEventGetTools guards the two MCP tool
// generation bugs: the registry must expose the `audit list` leaf (parents are
// skipped so `escape-cli audit` produced no tool) and `events get` must
// advertise its required event ID positional argument.
func TestBuildMCPToolSpecsGeneratesAuditAndEventGetTools(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	requiredByName := make(map[string][]string, len(specs))
	for _, spec := range specs {
		var schema struct {
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
			t.Fatalf("unmarshal input schema for tool %q: %v", spec.Name, err)
		}

		requiredByName[spec.Name] = schema.Required
	}

	auditRequired, ok := requiredByName["audit_list"]
	if !ok {
		t.Fatalf("expected MCP tool %q to be generated, got %v", "audit_list", requiredByName)
	}

	if len(auditRequired) != 0 {
		t.Errorf("audit_list should require no positional arguments, got %v", auditRequired)
	}

	eventsGetRequired, ok := requiredByName["events_get"]
	if !ok {
		t.Fatalf("expected MCP tool %q to be generated, got %v", "events_get", requiredByName)
	}

	if !slices.Contains(eventsGetRequired, "event_id") {
		t.Errorf("events_get should require the event_id positional argument, got %v", eventsGetRequired)
	}
}

// maxProbedPositionalArgs bounds the positional-arg counts probed against a
// command's Args validator when deriving its required arg count.
const maxProbedPositionalArgs = 3

// requiredPositionalArgs reports how many positional arguments the command's
// Args validator accepts, when it accepts exactly one count (e.g. NoArgs or
// ExactArgs(1)). ok is false when the command has no Args validator or accepts
// more than one count, so callers can skip commands without a fixed arity.
// Validators may print help on failure; output is discarded while probing.
func requiredPositionalArgs(t *testing.T, command *cobra.Command) (count int, ok bool) {
	t.Helper()

	if command.Args == nil {
		return 0, false
	}

	previousOut := command.OutOrStdout()
	previousErr := command.ErrOrStderr()
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	defer func() {
		command.SetOut(previousOut)
		command.SetErr(previousErr)
	}()

	accepted := -1
	for n := 0; n <= maxProbedPositionalArgs; n++ {
		args := make([]string, n)
		for i := range args {
			args[i] = "arg"
		}

		if command.Args(command, args) != nil {
			continue
		}

		if accepted >= 0 {
			return 0, false
		}

		accepted = n
	}

	if accepted < 0 {
		return 0, false
	}

	return accepted, true
}

// TestCommandSchemaRegistryKeysAreLeafCommands guards against registering a
// parent command (one with subcommands): the MCP catalog builder skips
// commands with subcommands, so such a key silently produces no tool.
func TestCommandSchemaRegistryKeysAreLeafCommands(t *testing.T) {
	registry := CommandSchemaRegistry()
	commands := indexCommands(rootCmd)

	for path := range registry {
		command := commands[path]
		if command == nil {
			t.Errorf("registry key %q does not match any available cobra command", path)
			continue
		}

		if command.HasAvailableSubCommands() {
			t.Errorf("registry key %q points to a command with subcommands; it would never be generated as an MCP tool", path)
		}
	}
}

// TestMCPToolRequiredPropertiesMatchPositionalArgs ensures that when a
// registered command requires exactly N positional arguments plus any flags
// marked required with cobra.MarkFlagRequired, its MCP tool input schema
// exposes exactly those properties as required. Otherwise
// additionalProperties:false blocks the caller from passing the IDs and the
// spawned subprocess can never satisfy the command's Args validation.
func TestMCPToolRequiredPropertiesMatchPositionalArgs(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	commands := indexCommands(rootCmd)
	for _, spec := range specs {
		command := commands[spec.Path]
		if command == nil {
			t.Errorf("no cobra command for tool %q (%s)", spec.Name, spec.Path)
			continue
		}

		want, ok := requiredPositionalArgs(t, command)
		if !ok {
			continue
		}

		var schema struct {
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
			t.Fatalf("unmarshal input schema for tool %q: %v", spec.Name, err)
		}

		expected := expectedRequiredProperties(command, spec)
		if !slices.Equal(schema.Required, expected) {
			t.Errorf("tool %q (%s): command requires %d positional argument(s), but schema declares required %v, want %v",
				spec.Name, spec.Path, want, schema.Required, expected)
		}
	}
}

// expectedRequiredProperties mirrors buildMCPTool's required list: the
// command's positional arguments first, then its non-skipped flags marked
// required with cobra.MarkFlagRequired, in VisitAll order. Destructive tools
// append confirm last. That property is the MCP gate, not a CLI flag.
func expectedRequiredProperties(command *cobra.Command, spec climcp.ToolSpec) []string {
	expected := make([]string, 0, len(spec.PositionalArgs)+1)
	expected = append(expected, spec.PositionalArgs...)

	properties := make(map[string]struct{}, len(spec.PositionalArgs))
	for _, name := range spec.PositionalArgs {
		properties[name] = struct{}{}
	}

	command.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if skipMCPFlag(command, flag.Name) {
			return
		}

		property := normalizePropertyName(flag.Name)
		if spec.Destructive && property == climcp.ConfirmProperty {
			return
		}

		if _, exists := properties[property]; exists {
			property = "flag_" + property
		}

		properties[property] = struct{}{}
		if isRequiredFlag(flag) || flagMarkedMCPRequired(flag) {
			expected = append(expected, property)
		}
	})
	if spec.Destructive {
		expected = append(expected, climcp.ConfirmProperty)
	}

	return expected
}

func TestMCPListToolsDefaultToOnePage(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	byName := make(map[string]climcp.ToolSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}

	pageSize := strconv.Itoa(climcp.DefaultListPageSize)
	for _, name := range []string{
		"issues_list",
		"assets_list",
		"events_list",
		"profiles_list",
		"projects_list",
		"integrations_list",
		"locations_list",
		"workflows_list",
		"scans_issues",
		"problems",
		"profiles_problems",
		"scans_problems",
		"retests_list",
	} {
		spec, ok := byName[name]
		if !ok {
			t.Fatalf("missing tool %s", name)
		}

		if spec.DefaultArgs["size"] != climcp.DefaultListPageSize {
			t.Fatalf("%s default size = %#v", name, spec.DefaultArgs["size"])
		}

		if spec.Description != spec.Tool.Description {
			t.Fatalf("%s tool description = %q, spec = %q", name, spec.Tool.Description, spec.Description)
		}

		if !strings.Contains(spec.Description, "nextCursor") || !strings.Contains(spec.Description, pageSize) {
			t.Fatalf("%s description = %q", name, spec.Description)
		}

		output := decodeMCPToolSchema(t, name, spec.Tool.RawOutputSchema)
		if output.Properties["nextCursor"]["type"] != "string" {
			t.Fatalf("%s nextCursor = %#v", name, output.Properties["nextCursor"])
		}

		if output.Properties["totalCount"]["type"] != "integer" {
			t.Fatalf("%s totalCount = %#v", name, output.Properties["totalCount"])
		}
	}

	for _, name := range []string{"issues_get", "scans_targets", "retests_get"} {
		spec, ok := byName[name]
		if !ok {
			t.Fatalf("missing tool %s", name)
		}

		if spec.DefaultArgs != nil {
			t.Fatalf("%s should not inject default args, got %#v", name, spec.DefaultArgs)
		}
	}

	commands := indexCommands(rootCmd)
	for _, spec := range specs {
		command := commands[spec.Path]
		if spec.DefaultArgs != nil && !pagedListCommand(command) {
			t.Fatalf("%s has DefaultArgs %#v without %s", spec.Path, spec.DefaultArgs, pagedListAnnotation)
		}
	}
}

func TestPagedCommandsDocumentPageShape(t *testing.T) {
	registry := CommandSchemaRegistry()
	for _, path := range []string{
		"escape-cli issues list",
		"escape-cli assets list",
		"escape-cli events list",
		"escape-cli profiles list",
		"escape-cli profiles problems",
		"escape-cli projects list",
		"escape-cli integrations list",
		"escape-cli locations list",
		"escape-cli workflows list",
		"escape-cli scans issues",
		"escape-cli scans problems",
		"escape-cli problems",
		"escape-cli retests list",
	} {
		schemas, ok := registry[path]
		if !ok {
			t.Fatalf("missing %s", path)
		}

		doc := schemaFor(schemas.Output)
		if path == "escape-cli problems" {
			assertProblemsDocumentsBothShapes(t, doc)
			continue
		}

		if doc.Type != "object" {
			t.Fatalf("%s type = %#v, want the page object", path, doc.Type)
		}

		for _, key := range []string{"items", "nextCursor", "totalCount"} {
			if doc.Properties[key] == nil {
				t.Fatalf("%s schema missing %s", path, key)
			}
		}

		if doc.Properties["items"].Type != "array" {
			t.Fatalf("%s items = %#v", path, doc.Properties["items"].Type)
		}
	}
}

func assertProblemsDocumentsBothShapes(t *testing.T, doc *clischema.JSONSchema) {
	t.Helper()
	if doc.Type != "object" {
		t.Fatalf("problems type = %#v, want the page object", doc.Type)
	}

	items := doc.Properties["items"]
	if items == nil || len(items.OneOf) != 2 {
		t.Fatalf("problems items = %#v, want oneOf summary and detail arrays", items)
	}

	var sawSummary, sawDetail bool
	for _, branch := range items.OneOf {
		if branch.Type != "array" || branch.Items == nil || branch.Items.Properties == nil {
			t.Fatalf("problems items branch = %#v", branch)
		}

		if branch.Items.Properties["ProblemCount"] != nil {
			sawSummary = true
		}

		if branch.Items.Properties["Code"] != nil && branch.Items.Properties["Message"] != nil {
			sawDetail = true
		}
	}

	if !sawSummary || !sawDetail {
		t.Fatalf("problems items oneOf summary=%v detail=%v", sawSummary, sawDetail)
	}
}

func TestCommandSchemaRegistryIncludesNonDestructiveWriteTools(t *testing.T) {
	registry := CommandSchemaRegistry()

	// These commands take their input from flags (no stdin JSON body), so the
	// registry entries must declare the response type they print as Output and
	// no Input schema.
	wantOutputs := map[string]any{
		"escape-cli tags create":    v3.CreateTag200Response{},
		"escape-cli tags update":    v3.CreateTag200Response{},
		"escape-cli assets update":  v3.UpdateAsset200Response{},
		"escape-cli assets comment": v3.CreateAssetComment200Response{},
		"escape-cli issues notify":  v3.NotifyIssueOwners200Response{},
		"escape-cli asm trigger":    v3.TriggerAsmScans200Response{},
		"escape-cli roles bind":     []v3.CreateRoleBindings200ResponseInner{},
		"escape-cli roles unbind":   v3.DeleteCustomRule200Response{},
	}

	for path, wantOutput := range wantOutputs {
		schemas, ok := registry[path]
		if !ok {
			t.Fatalf("expected %q in CommandSchemaRegistry", path)
		}

		if got := reflect.TypeOf(schemas.Output); got != reflect.TypeOf(wantOutput) {
			t.Fatalf("expected output type %v for %q, got %v", reflect.TypeOf(wantOutput), path, got)
		}

		if schemas.Input != nil {
			t.Fatalf("did not expect an input schema for flag-based command %q", path)
		}
	}
}

func TestBuildMCPToolSpecsIncludesNonDestructiveWriteTools(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	tests := []struct {
		tool       string
		path       string
		required   []string
		properties []string
	}{
		{tool: "tags_create", path: "escape-cli tags create", properties: []string{"color", "name"}},
		{tool: "tags_update", path: "escape-cli tags update", required: []string{"tag_id"}, properties: []string{"color", "name", "tag_id"}},
		{tool: "assets_update", path: "escape-cli assets update", required: []string{"asset_id"}, properties: []string{"asset_id", "description", "framework", "name", "owners", "project_id", "status", "tag_ids"}},
		{tool: "assets_comment", path: "escape-cli assets comment", required: []string{"asset_id"}, properties: []string{"asset_id", "message"}},
		{tool: "issues_notify", path: "escape-cli issues notify", required: []string{"issue_id"}, properties: []string{"issue_id", "scan_id"}},
		{tool: "asm_trigger", path: "escape-cli asm trigger", required: []string{"asset_id"}, properties: []string{"asset_id"}},
		{tool: "roles_bind", path: "escape-cli roles bind", properties: []string{"role_id", "user_id"}},
		{tool: "roles_unbind", path: "escape-cli roles unbind", required: []string{"binding_id", "confirm"}, properties: []string{"binding_id", "confirm"}},
	}

	for _, tt := range tests {
		spec := findToolSpecByName(specs, tt.tool)
		if spec == nil {
			t.Fatalf("expected MCP tool %q for %q", tt.tool, tt.path)
		}

		if spec.Path != tt.path {
			t.Fatalf("expected tool %q to map to path %q, got %q", tt.tool, tt.path, spec.Path)
		}

		var inputSchema struct {
			Properties map[string]any `json:"properties"`
			Required   []string       `json:"required"`
		}
		if err := json.Unmarshal(spec.Tool.RawInputSchema, &inputSchema); err != nil {
			t.Fatalf("unmarshal input schema for %q: %v", tt.tool, err)
		}

		for _, property := range tt.properties {
			if _, ok := inputSchema.Properties[property]; !ok {
				t.Fatalf("expected tool %q to expose property %q", tt.tool, property)
			}
		}

		for _, want := range tt.required {
			if !slices.Contains(inputSchema.Required, want) {
				t.Fatalf("expected tool %q to require %q, got %v", tt.tool, want, inputSchema.Required)
			}
		}
	}
}

func TestAsmTriggerMCPRequiresAnAssetFilter(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	spec := findToolSpecByName(specs, "asm_trigger")
	if spec == nil {
		t.Fatal("expected asm_trigger")
	}

	if spec.Destructive {
		t.Fatal("a filtered ASM trigger does not need the confirm gate")
	}

	var required bool
	for _, binding := range spec.FlagBindings {
		if binding.FlagName == "asset-id" && binding.Property == "asset_id" && binding.Required {
			required = true
		}
	}

	if !required {
		t.Fatalf("asset_id binding = %#v, want Required", spec.FlagBindings)
	}

	schema := mcpInputSchema(t, *spec)
	got, _ := schema["required"].([]any)
	if !slices.Contains(got, any("asset_id")) {
		t.Fatalf("required = %#v, want asset_id", got)
	}

	props, _ := schema["properties"].(map[string]any)
	if _, exists := props[climcp.ConfirmProperty]; exists {
		t.Fatal("asm_trigger schema includes confirm")
	}

	flag := asmTriggerCmd.Flags().Lookup("asset-id")
	if isRequiredFlag(flag) {
		t.Fatal("CLI --asset-id stays optional")
	}

	stub, err := climcp.BuildStubTool(*spec)
	if err != nil {
		t.Fatalf("BuildStubTool: %v", err)
	}

	stubSchema := decodeMCPToolSchema(t, "asm_trigger stub", stub.RawInputSchema)
	if !slices.Contains(stubSchema.Required, "asset_id") {
		t.Fatalf("stub required = %#v, want asset_id", stubSchema.Required)
	}
}

func TestAssetsUpdateMCPToolBindsProjectAndNameFlags(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	spec := findToolSpecByName(specs, "assets_update")
	if spec == nil {
		t.Fatal("expected MCP tool \"assets_update\"")
	}

	for _, want := range []climcp.FlagBinding{
		{Property: "project_id", FlagName: "project-id", Kind: "stringSlice"},
		{Property: "name", FlagName: "name", Kind: "string"},
	} {
		if !slices.ContainsFunc(spec.FlagBindings, func(binding climcp.FlagBinding) bool {
			return binding == want
		}) {
			t.Fatalf("expected flag binding %+v, got %v", want, spec.FlagBindings)
		}
	}

	var inputSchema struct {
		Properties map[string]struct {
			Type string `json:"type"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(spec.Tool.RawInputSchema, &inputSchema); err != nil {
		t.Fatalf("unmarshal input schema: %v", err)
	}

	if got := inputSchema.Properties["project_id"].Type; got != "array" {
		t.Fatalf("expected project_id property to be an array, got %q", got)
	}

	if got := inputSchema.Properties["name"].Type; got != "string" {
		t.Fatalf("expected name property to be a string, got %q", got)
	}
}

func findToolSpecByName(specs []climcp.ToolSpec, name string) *climcp.ToolSpec {
	for i := range specs {
		if specs[i].Name == name {
			return &specs[i]
		}
	}

	return nil
}

func TestCommandSchemaRegistryIncludesIssueTriggerWorkflow(t *testing.T) {
	registry := CommandSchemaRegistry()

	if _, ok := registry["escape-cli issues trigger-workflow"]; !ok {
		t.Fatalf("expected %q in CommandSchemaRegistry", "escape-cli issues trigger-workflow")
	}
}

func TestBuildMCPToolSpecsIssueTriggerWorkflow(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	spec := findToolSpecByName(specs, "issues_trigger_workflow")
	if spec == nil {
		t.Fatalf("expected MCP tool %q in %v", "issues_trigger_workflow", specs)
	}

	if spec.Path != "escape-cli issues trigger-workflow" {
		t.Errorf("unexpected tool path %q", spec.Path)
	}

	if !slices.Contains(spec.PositionalArgs, "issue_id") {
		t.Errorf("expected positional args to include issue_id, got %v", spec.PositionalArgs)
	}

	if !strings.Contains(spec.Description, "workflows_list") {
		t.Errorf("expected tool description to point the model at workflows_list, got %q", spec.Description)
	}

	var bound bool
	for _, binding := range spec.FlagBindings {
		if binding.Property == "workflow_id" && binding.FlagName == "workflow-id" && binding.Kind == "string" {
			bound = true
			break
		}
	}

	if !bound {
		t.Errorf("expected a flag binding mapping workflow_id to --workflow-id, got %v", spec.FlagBindings)
	}

	var inputSchema struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(spec.Tool.RawInputSchema, &inputSchema); err != nil {
		t.Fatalf("unmarshal tool input schema: %v", err)
	}

	for _, want := range []string{"issue_id", "workflow_id"} {
		if _, ok := inputSchema.Properties[want]; !ok {
			t.Errorf("expected input schema property %q, got %v", want, inputSchema.Properties)
		}

		if !slices.Contains(inputSchema.Required, want) {
			t.Errorf("expected input schema to require %q, got %v", want, inputSchema.Required)
		}
	}
}
