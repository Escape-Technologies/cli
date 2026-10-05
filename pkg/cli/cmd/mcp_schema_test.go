package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestMCPToolDescriptionUsesFirstParagraph(t *testing.T) {
	t.Parallel()

	command := &cobra.Command{
		Short: "List things",
		Long:  "List Things - A Short Headline\n\nFirst paragraph explains formats.\n\nSecond paragraph should be dropped.",
	}
	got := mcpToolDescription(command)
	want := "List things\nFirst paragraph explains formats."
	if got != want {
		t.Fatalf("description = %q", got)
	}

	command.Long = "First paragraph explains formats\nand stays because it wraps.\n\nSecond paragraph should be dropped."
	want = "List things\nFirst paragraph explains formats and stays because it wraps."
	if got := mcpToolDescription(command); got != want {
		t.Fatalf("multiline description = %q", got)
	}

	command.Long = "Tiny headline"
	if got := mcpToolDescription(command); got != "List things" {
		t.Fatalf("headline-only description = %q", got)
	}

	command.Long = "Same text"
	command.Short = "Same text"
	if got := mcpToolDescription(command); got != "Same text" {
		t.Fatalf("duplicate long = %q", got)
	}
}

func TestMCPToolDescriptionCapsLength(t *testing.T) {
	t.Parallel()

	command := &cobra.Command{
		Short: "Short",
		Long:  strings.Repeat("word ", 200),
	}
	got := mcpToolDescription(command)
	if len([]rune(got)) > maxMCPToolDescriptionRunes {
		t.Fatalf("description length %d", len([]rune(got)))
	}

	if !strings.HasSuffix(got, "...") {
		t.Fatalf("description = %q", got)
	}

	if strings.Count(got, "word") < 10 {
		t.Fatalf("description dropped the paragraph: %q", got)
	}
}

func TestEnumFromUsage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		usage string
		want  string
	}{
		{"filter by status: [OPEN RESOLVED]", "OPEN,RESOLVED"},
		{"sort direction: asc, desc", ""},
		{"Filter by location type: PRIVATE or ESCAPE (case-insensitive)", ""},
		{
			"output format: pretty (human-readable tables), json (machine-readable), yaml (configuration files), schema (JSON Schema for AI agents)",
			"",
		},
		{"sort field: LAST_SEEN, FIRST_SEEN, SEVERITY, STATUS", ""},
		{"free-text search across issue names and descriptions", ""},
		{`cron schedule (e.g., "0 22 * * *")`, ""},
		{"time bucket interval (e.g. '1 day', '1 week')", ""},
		{"sort field (e.g., LAST_SEEN, CREATED_AT)", ""},
		{"verbose output: -v (debug), -vv (trace), -vvv (http/raw debug)", ""},
	}
	for _, tc := range cases {
		t.Run(tc.usage, func(t *testing.T) {
			t.Parallel()
			got := strings.Join(enumFromUsage(tc.usage), ",")
			if got != tc.want {
				t.Fatalf("enum = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSchemaForFlagSetsEnum(t *testing.T) {
	t.Parallel()

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	var status []string
	var direction string
	flags.StringSliceVar(&status, "status", nil, "filter by status: [OPEN RESOLVED]")
	flags.StringVar(&direction, "sort-direction", "", "sort direction: asc, desc")

	statusSchema, kind := schemaForFlag(flags.Lookup("status"))
	if kind != "stringSlice" {
		t.Fatalf("kind = %s", kind)
	}

	items, _ := statusSchema["items"].(map[string]any)
	if strings.Join(stringEnum(t, items["enum"]), ",") != "OPEN,RESOLVED" {
		t.Fatalf("status schema = %#v", statusSchema)
	}

	directionSchema, kind := schemaForFlag(flags.Lookup("sort-direction"))
	if kind != "string" {
		t.Fatalf("kind = %s", kind)
	}

	if _, ok := directionSchema["enum"]; ok {
		t.Fatalf("colon sample became a closed enum: %#v", directionSchema)
	}
}

func TestCompactCatalogKeepsEnumsDescriptionsAndLongHelp(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("build specs: %v", err)
	}

	byName := map[string]climcp.ToolSpec{}
	for _, spec := range specs {
		byName[spec.Name] = spec
	}

	issues := byName["issues_list"]
	assets := byName["assets_list"]
	if issues.Name == "" || assets.Name == "" {
		t.Fatalf("missing tools: issues=%q assets=%q", issues.Name, assets.Name)
	}

	if !strings.Contains(issues.Description, "List security issues with powerful filtering") {
		t.Fatalf("description missing short: %q", issues.Description)
	}

	if strings.Contains(issues.Description, "Query Your Vulnerability Database") {
		t.Fatalf("description kept the headline: %q", issues.Description)
	}

	if !strings.Contains(issues.Description, "List and filter security issues") {
		t.Fatalf("description missing the paragraph after the headline: %q", issues.Description)
	}

	if strings.Contains(issues.Description, "FILTER OPTIONS") {
		t.Fatalf("description kept later paragraphs: %q", issues.Description)
	}

	if len([]rune(issues.Description)) > maxMCPToolDescriptionRunes {
		t.Fatalf("description length %d", len([]rune(issues.Description)))
	}

	issuesStub, err := climcp.BuildStubTool(issues)
	if err != nil {
		t.Fatal(err)
	}

	statusName := flagProperty(t, issues, "status")
	fullStatus := schemaProperty(t, issues.Tool.RawInputSchema, statusName)
	stubStatus := schemaProperty(t, issuesStub.RawInputSchema, statusName)
	wantStatus := enumStringList(v3.AllowedENUMPROPERTIESFILTERPROPERTIESSTATUSITEMSEnumValues)
	if strings.Join(enumFromProperty(t, fullStatus), ",") != wantStatus {
		t.Fatalf("full status enum = %#v", enumFromProperty(t, fullStatus))
	}

	if strings.Join(enumFromProperty(t, stubStatus), ",") != wantStatus {
		t.Fatalf("stub status enum = %#v", enumFromProperty(t, stubStatus))
	}

	statusDesc, _ := stubStatus["description"].(string)
	if statusDesc == "" || strings.Contains(statusDesc, "\n") || len([]rune(statusDesc)) > 120 {
		t.Fatalf("stub status description = %q", statusDesc)
	}

	directionName := flagProperty(t, issues, "sort-direction")
	direction := schemaProperty(t, issuesStub.RawInputSchema, directionName)
	if propertyHasEnum(direction) {
		t.Fatalf("stub sort-direction is a sample, not a closed enum: %#v", direction)
	}

	if direction["description"] != "sort direction: asc, desc" {
		t.Fatalf("stub sort-direction description = %#v", direction["description"])
	}

	assetsStub, err := climcp.BuildStubTool(assets)
	if err != nil {
		t.Fatal(err)
	}

	typesName := flagProperty(t, assets, "types")
	fullTypes := schemaProperty(t, assets.Tool.RawInputSchema, typesName)
	stubTypes := schemaProperty(t, assetsStub.RawInputSchema, typesName)
	fullDesc, _ := fullTypes["description"].(string)
	stubDesc, _ := stubTypes["description"].(string)
	if len([]rune(fullDesc)) <= 120 {
		t.Fatalf("expected a long types description, got %q", fullDesc)
	}

	if stubDesc == "" || strings.Contains(stubDesc, "\n") || len([]rune(stubDesc)) > 120 {
		t.Fatalf("stub types description = %q (len %d)", stubDesc, len([]rune(stubDesc)))
	}

	wantTypes := enumStringList(v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues)
	if strings.Join(enumFromProperty(t, stubTypes), ",") != wantTypes {
		t.Fatalf("stub types enum len %d, want %d", len(enumFromProperty(t, stubTypes)), len(v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	}

	if !strings.Contains(wantTypes, "FRONTEND_REACT") || !strings.Contains(wantTypes, "REST_GIN") {
		t.Fatalf("framework enum = %s", wantTypes)
	}
}

func TestGeneratedFlagEnumsAreClosedAndSamplesAreNot(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("build specs: %v", err)
	}

	byName := map[string]climcp.ToolSpec{}
	for _, spec := range specs {
		byName[spec.Name] = spec
	}

	scans := byName["scans_list"]
	kind := schemaProperty(t, scans.Tool.RawInputSchema, flagProperty(t, scans, "kind"))
	wantKinds := enumStringList(v3.AllowedENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMSEnumValues)
	if strings.Join(enumFromProperty(t, kind), ",") != wantKinds {
		t.Fatalf("scans_list kind enum len %d, want %d", len(enumFromProperty(t, kind)), len(v3.AllowedENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMSEnumValues))
	}

	if !strings.Contains(wantKinds, "ASM_") || !strings.Contains(wantKinds, "RETEST") || len(enumFromProperty(t, kind)) <= 4 {
		t.Fatalf("kind enum = %s", wantKinds)
	}

	initiator := schemaProperty(t, scans.Tool.RawInputSchema, flagProperty(t, scans, "initiator"))
	gotInitiator := strings.Join(enumFromProperty(t, initiator), ",")
	wantInitiator := enumStringList(v3.AllowedENUMPROPERTIESINITIATOREnumValues)
	if gotInitiator != wantInitiator {
		t.Fatalf("initiator enum = %s, want %s", gotInitiator, wantInitiator)
	}

	if strings.Contains(gotInitiator, "API") {
		t.Fatalf("initiator enum includes API, which is not a generated value: %s", gotInitiator)
	}

	status := schemaProperty(t, scans.Tool.RawInputSchema, flagProperty(t, scans, "status"))
	gotStatus := strings.Join(enumFromProperty(t, status), ",")
	if gotStatus != enumStringList(v3.AllowedENUMPROPERTIESSTATUSEnumValues) {
		t.Fatalf("status enum = %s", gotStatus)
	}

	if !strings.Contains(gotStatus, "COMPLETED") || !strings.Contains(gotStatus, "PENDING") {
		t.Fatalf("status enum = %s", gotStatus)
	}

	direction := schemaProperty(t, scans.Tool.RawInputSchema, flagProperty(t, scans, "sort-direction"))
	if propertyHasEnum(direction) {
		t.Fatalf("scans_list sort-direction enum = %#v", direction)
	}

	locations := byName["locations_list"]
	locType := schemaProperty(t, locations.Tool.RawInputSchema, flagProperty(t, locations, "type"))
	if propertyHasEnum(locType) {
		t.Fatalf("locations_list type enum = %#v", locType)
	}

	retests := byName["retests_list"]
	sortBy := schemaProperty(t, retests.Tool.RawInputSchema, flagProperty(t, retests, "sort-by"))
	if propertyHasEnum(sortBy) {
		t.Fatalf("retests_list sort_by enum = %#v", sortBy)
	}

	start := byName["retests_start"]
	scanner := schemaProperty(t, start.Tool.RawInputSchema, flagProperty(t, start, "scanner-kind"))
	if strings.Join(enumFromProperty(t, scanner), ",") != wantKinds {
		t.Fatalf("retests_start scanner-kind enum len %d, want %d", len(enumFromProperty(t, scanner)), len(v3.AllowedENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMSEnumValues))
	}
}

func TestHoistedEnumRefsResolveAtDocumentRoot(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("build specs: %v", err)
	}

	for _, spec := range specs {
		assertSchemaRefsResolve(t, spec.Name+" input", spec.Tool.RawInputSchema)
		assertSchemaRefsResolve(t, spec.Name+" output", spec.Tool.RawOutputSchema)
	}
}

func assertSchemaRefsResolve(t *testing.T, name string, raw []byte) {
	t.Helper()
	if len(raw) == 0 {
		return
	}

	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("%s: %v", name, err)
	}

	root, _ := doc.(map[string]any)
	defs, _ := root["$defs"].(map[string]any)
	var walk func(any, bool)
	walk = func(node any, isRoot bool) {
		switch typed := node.(type) {
		case map[string]any:
			if _, ok := typed["$defs"]; ok && !isRoot {
				t.Errorf("%s: $defs is not at the document root", name)
			}

			if ref, _ := typed["$ref"].(string); strings.HasPrefix(ref, "#/$defs/") {
				key := strings.TrimPrefix(ref, "#/$defs/")
				if _, ok := defs[key]; !ok {
					t.Errorf("%s: unresolved %s", name, ref)
				}
			}

			for key, child := range typed {
				if key == "$defs" {
					continue
				}

				walk(child, false)
			}
		case []any:
			for _, child := range typed {
				walk(child, false)
			}
		}
	}
	walk(doc, true)
}

func propertyHasEnum(prop map[string]any) bool {
	if _, ok := prop["enum"]; ok {
		return true
	}

	items, _ := prop["items"].(map[string]any)
	if items == nil {
		return false
	}

	_, ok := items["enum"]

	return ok
}

func flagProperty(t *testing.T, spec climcp.ToolSpec, flagName string) string {
	t.Helper()
	for _, binding := range spec.FlagBindings {
		if binding.FlagName == flagName {
			return binding.Property
		}
	}

	t.Fatalf("missing flag %s", flagName)

	return ""
}

func schemaProperty(t *testing.T, raw []byte, name string) map[string]any {
	t.Helper()
	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}

	prop := schema.Properties[name]
	if prop == nil {
		t.Fatalf("missing property %s", name)
	}

	return prop
}

func enumFromProperty(t *testing.T, prop map[string]any) []string {
	t.Helper()
	if items, ok := prop["items"].(map[string]any); ok {
		if _, exists := items["enum"]; exists {
			return stringEnum(t, items["enum"])
		}
	}

	return stringEnum(t, prop["enum"])
}

func stringEnum(t *testing.T, value any) []string {
	t.Helper()
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, len(typed))
		for i, item := range typed {
			text, ok := item.(string)
			if !ok {
				t.Fatalf("enum value %T", item)
			}

			out[i] = text
		}

		return out
	default:
		t.Fatalf("enum type %T", value)
		return nil
	}
}

func enumStringList[T ~string](values []T) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = string(value)
	}

	return strings.Join(parts, ",")
}
