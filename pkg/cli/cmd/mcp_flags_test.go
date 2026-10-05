package cmd

import (
	"encoding/json"
	"testing"

	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/spf13/cobra"
)

// MCP property name -> CLI flag name.
type mcpFlag struct {
	property string
	flag     string
}

func TestMCPScanToolsExposeOwnFlags(t *testing.T) {
	specs := mcpSpecs(t)

	list := mcpSpecByName(t, specs, "scans_list")
	assertMCPFlags(t, list, []mcpFlag{
		{property: "profile_id", flag: "profile-id"},
		{property: "status", flag: "status"},
		{property: "after", flag: "after"},
		{property: "before", flag: "before"},
		{property: "kind", flag: "kind"},
		{property: "project_id", flag: "project-id"},
		{property: "asset_id", flag: "asset-id"},
		{property: "initiator", flag: "initiator"},
		{property: "ignored", flag: "ignored"},
		{property: "sort_by", flag: "sort-by"},
		{property: "sort_direction", flag: "sort-direction"},
	})
	assertMCPFlagAbsent(t, list, "verbose", "output", "input_schema", "help")

	start := mcpSpecByName(t, specs, "scans_start")
	assertMCPFlags(t, start, []mcpFlag{
		{property: "override", flag: "override"},
		{property: "additional_properties", flag: "additional-properties"},
		{property: "commit_hash", flag: "commit-hash"},
		{property: "commit_branch", flag: "commit-branch"},
		{property: "commit_link", flag: "commit-link"},
		{property: "commit_author", flag: "commit-author"},
	})
	// --watch blocks until the scan finishes. --fail-on-severity is only valid
	// with --watch, so both stay off this tool.
	assertMCPFlagAbsent(t, start, "watch", "fail_on_severity", "verbose", "output", "input_schema", "help")

	// jobs get --watch blocks until the job finishes. The CLI flag stays;
	// the MCP catalog hides it and tells the model to poll jobs_get.
	jobsGet := mcpSpecByName(t, specs, "jobs_get")
	assertMCPFlagAbsent(t, jobsGet, "watch")
}

// scans problems registers its filters with PersistentFlags. The catalog reads
// LocalFlags, which includes a command's own persistent flags, so the tool
// still advertises every filter next to the page flags.
func TestMCPScansProblemsExposesFilters(t *testing.T) {
	problems := mcpSpecByName(t, mcpSpecs(t), "scans_problems")
	assertMCPFlags(t, problems, []mcpFlag{
		{property: "profile_id", flag: "profile-id"},
		{property: "project_id", flag: "project-id"},
		{property: "asset_id", flag: "asset-id"},
		{property: "after", flag: "after"},
		{property: "before", flag: "before"},
		{property: "ignored", flag: "ignored"},
		{property: "initiator", flag: "initiator"},
		{property: "kind", flag: "kind"},
		{property: "status", flag: "status"},
		{property: "sort_by", flag: "sort-by"},
		{property: "sort_direction", flag: "sort-direction"},
		{property: "size", flag: "size"},
		{property: "cursor", flag: "cursor"},
	})
	assertMCPFlagAbsent(t, problems, "verbose", "output", "input_schema", "help")
}

// Leaf scan commands have no subcommands, so their options must be local flags.
// PersistentFlags are invisible to Flags() until parse time, which is what
// dropped them from the MCP schema.
func TestScanLeafOptionsAreLocalFlags(t *testing.T) {
	list := []string{
		"profile-id", "status", "after", "before", "kind", "project-id",
		"asset-id", "initiator", "ignored", "sort-by", "sort-direction",
	}
	assertLocalFlag(t, scansListCmd, list, map[string]string{
		"profile-id": "p",
		"asset-id":   "a",
		"initiator":  "i",
		"kind":       "k",
		"status":     "s",
	})

	start := []string{
		"override", "additional-properties", "commit-hash", "commit-branch",
		"commit-link", "commit-author", "watch", "fail-on-severity",
	}
	assertLocalFlag(t, scanStartCmd, start, map[string]string{
		"watch":    "w",
		"override": "c",
	})
}

// Shorthand is read off the flag set. ParseFlags on the shared commands would
// leave pflag Changed set for every later test in the process.
func TestScanShortFlagsStillBind(t *testing.T) {
	shorts := map[*cobra.Command]map[string]string{
		scansListCmd: {
			"profile-id": "p",
			"asset-id":   "a",
			"initiator":  "i",
			"kind":       "k",
			"status":     "s",
		},
		scanStartCmd: {
			"watch":    "w",
			"override": "c",
		},
	}
	for command, flags := range shorts {
		for name, want := range flags {
			flag := command.Flags().Lookup(name)
			if flag == nil {
				t.Errorf("%s missing %s", command.CommandPath(), name)
				continue
			}

			if flag.Shorthand != want {
				t.Errorf("%s %s shorthand = %q, want %q", command.CommandPath(), name, flag.Shorthand, want)
			}
		}
	}
}

func TestSkipMCPFlagReadsAnnotation(t *testing.T) {
	for _, command := range []*cobra.Command{
		scanStartCmd,
		jobsGetCmd,
		jobsTriggerExportCmd,
		authenticationsStartCmd,
		authenticationsGetCmd,
	} {
		flag := command.Flags().Lookup("watch")
		if flag == nil {
			t.Fatalf("%s missing --watch", command.CommandPath())
		}

		if !flagMarkedMCPSkip(flag) {
			t.Fatalf("%s --watch annotations = %#v, want mcp=skip", command.CommandPath(), flag.Annotations)
		}

		if !skipMCPFlag(command, "watch") {
			t.Fatalf("%s --watch must be hidden from MCP", command.CommandPath())
		}
	}

	if !flagMarkedMCPSkip(scanStartCmd.Flags().Lookup("fail-on-severity")) {
		t.Fatal("scans start --fail-on-severity must be annotated mcp=skip")
	}

	if !skipMCPFlag(scanStartCmd, "fail-on-severity") {
		t.Fatal("scans start --fail-on-severity must be hidden from MCP")
	}

	if !skipMCPFlag(scanWatchCmd, "fail-on-severity") {
		t.Fatal("scans watch --fail-on-severity must be hidden from MCP")
	}

	// watch is hidden on every command. A forgotten annotation must not expose it.
	leaf := &cobra.Command{Use: "widget", RunE: func(*cobra.Command, []string) error { return nil }}
	leaf.Flags().Bool("watch", false, "blocks until completion")
	leaf.Flags().Bool("follow", false, "also blocks")
	markMCPSkip(leaf.Flags(), "follow")
	if !skipMCPFlag(leaf, "watch") {
		t.Fatal("watch must be hidden even without the mcp=skip annotation")
	}

	if !skipMCPFlag(leaf, "follow") {
		t.Fatal("mcp=skip must hide a flag on a command the catalog does not name")
	}
}

// A command that still uses PersistentFlags must show those flags on its MCP
// tool, and must not inherit the root verbose/output/input-schema flags.
func TestBuildMCPToolIncludesOwnPersistentFlags(t *testing.T) {
	root := &cobra.Command{Use: "escape-cli"}
	root.PersistentFlags().Bool("verbose", false, "verbose output")
	root.PersistentFlags().String("output", "pretty", "output format")
	root.PersistentFlags().Bool("input-schema", false, "print JSON Schema for stdin")

	leaf := &cobra.Command{Use: "widget", RunE: func(*cobra.Command, []string) error { return nil }}
	leaf.PersistentFlags().String("profile-id", "", "filter by profile")
	leaf.Flags().Bool("all-kinds", false, "include every kind")
	root.AddCommand(leaf)

	tool, bindings, _, _, err := buildMCPTool(CommandCapability{
		Path:  "escape-cli widget",
		Use:   "widget",
		Short: "widget",
	}, leaf, "widget", false)
	if err != nil {
		t.Fatalf("buildMCPTool: %v", err)
	}

	spec := climcp.ToolSpec{Name: "widget", Tool: tool, FlagBindings: bindings}
	assertMCPFlags(t, spec, []mcpFlag{
		{property: "profile_id", flag: "profile-id"},
		{property: "all_kinds", flag: "all-kinds"},
	})
	assertMCPFlagAbsent(t, spec, "verbose", "output", "input_schema")
}

func TestCapabilitiesHasFlagsIncludesOwnPersistentFlags(t *testing.T) {
	root := &cobra.Command{Use: "escape-cli"}
	root.PersistentFlags().Bool("verbose", false, "verbose output")

	leaf := &cobra.Command{Use: "widget", RunE: func(*cobra.Command, []string) error { return nil }}
	leaf.PersistentFlags().String("only-persistent", "", "declared persistent on a leaf")
	root.AddCommand(leaf)

	bare := &cobra.Command{Use: "bare", RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(bare)

	caps := BuildCommandCapabilities(root, nil)
	widget, ok := capabilityByPath(caps, "escape-cli widget")
	if !ok {
		t.Fatal("missing widget capability")
	}

	if !widget.HasFlags {
		t.Fatal("widget only declares a persistent flag, HasFlags is false")
	}

	inherited, ok := capabilityByPath(caps, "escape-cli bare")
	if !ok {
		t.Fatal("missing bare capability")
	}

	if inherited.HasFlags {
		t.Fatal("bare only inherits root flags, HasFlags is true")
	}
}

func assertLocalFlag(t *testing.T, command *cobra.Command, names []string, shorthands map[string]string) {
	t.Helper()
	for _, name := range names {
		flag := command.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("%s missing local flag %s", command.CommandPath(), name)
			continue
		}

		if command.PersistentFlags().Lookup(name) != nil {
			t.Errorf("%s flag %s is still persistent", command.CommandPath(), name)
		}

		if want, ok := shorthands[name]; ok && flag.Shorthand != want {
			t.Errorf("%s flag %s shorthand = %q, want %q", command.CommandPath(), name, flag.Shorthand, want)
		}
	}
}

func mcpSpecs(t *testing.T) []climcp.ToolSpec {
	t.Helper()
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	return specs
}

func mcpSpecByName(t *testing.T, specs []climcp.ToolSpec, name string) climcp.ToolSpec {
	t.Helper()
	for _, spec := range specs {
		if spec.Name == name {
			return spec
		}
	}

	t.Fatalf("missing MCP tool %s", name)

	return climcp.ToolSpec{}
}

func assertMCPFlags(t *testing.T, spec climcp.ToolSpec, wants []mcpFlag) {
	t.Helper()
	props := schemaProperties(t, spec)
	bound := map[string]string{}
	for _, binding := range spec.FlagBindings {
		bound[binding.Property] = binding.FlagName
	}

	for _, want := range wants {
		if _, ok := props[want.property]; !ok {
			t.Errorf("%s schema missing %s", spec.Name, want.property)
		}

		if bound[want.property] != want.flag {
			t.Errorf("%s binding %s = %q, want flag %s", spec.Name, want.property, bound[want.property], want.flag)
		}
	}
}

func assertMCPFlagAbsent(t *testing.T, spec climcp.ToolSpec, names ...string) {
	t.Helper()
	props := schemaProperties(t, spec)
	bound := map[string]struct{}{}
	for _, binding := range spec.FlagBindings {
		bound[binding.Property] = struct{}{}
	}

	for _, name := range names {
		if _, ok := props[name]; ok {
			t.Errorf("%s schema unexpectedly contains %s", spec.Name, name)
		}

		if _, ok := bound[name]; ok {
			t.Errorf("%s bindings unexpectedly contain %s", spec.Name, name)
		}
	}
}

func schemaProperties(t *testing.T, spec climcp.ToolSpec) map[string]any {
	t.Helper()
	var schema struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
		t.Fatalf("%s schema: %v", spec.Name, err)
	}

	return schema.Properties
}

func capabilityByPath(caps []CommandCapability, path string) (CommandCapability, bool) {
	for _, cap := range caps {
		if cap.Path == path {
			return cap, true
		}
	}

	return CommandCapability{}, false
}
