package cmd

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
)

func TestMergeJSONDeepMergesObjectsAndReplacesArrays(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"mode": "read_only",
		"authentication": map[string]any{
			"users": []any{map[string]any{"name": "alice"}},
		},
		"frontend_dast": map[string]any{
			"hotstart": []any{"https://old.example"},
			"scope":    "app",
		},
	}
	patch := map[string]any{
		"authentication": map[string]any{
			"headers": []any{map[string]any{"name": "X-Test"}},
		},
		"frontend_dast": map[string]any{
			"hotstart": []any{"https://new.example"},
		},
		"network": map[string]any{"proxy": true},
	}

	got := mergeJSON(base, patch)
	want := map[string]any{
		"mode": "read_only",
		"authentication": map[string]any{
			"users":   []any{map[string]any{"name": "alice"}},
			"headers": []any{map[string]any{"name": "X-Test"}},
		},
		"frontend_dast": map[string]any{
			"hotstart": []any{"https://new.example"},
			"scope":    "app",
		},
		"network": map[string]any{"proxy": true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeJSON() = %#v, want %#v", got, want)
	}
}

func TestMergeJSONReplacesScalarsNullAndNonObjects(t *testing.T) {
	t.Parallel()

	base := map[string]any{
		"mode":    "read_only",
		"network": map[string]any{"proxy": true},
		"kept":    "yes",
	}

	cleared := mergeJSON(base, map[string]any{"mode": nil}).(map[string]any)
	if _, ok := cleared["mode"]; !ok || cleared["mode"] != nil {
		t.Fatalf("null patch should set mode to null, got %#v", cleared["mode"])
	}

	if cleared["kept"] != "yes" {
		t.Fatalf("omitted key was dropped: %#v", cleared)
	}

	replaced := mergeJSON(base, map[string]any{"network": "direct"}).(map[string]any)
	if replaced["network"] != "direct" {
		t.Fatalf("scalar should replace object, got %#v", replaced["network"])
	}

	unchanged := mergeJSON(base, map[string]any{})
	if !reflect.DeepEqual(unchanged, base) {
		t.Fatalf("empty object patch changed base: %#v", unchanged)
	}

	if got := mergeJSON(base, []any{"nope"}); !reflect.DeepEqual(got, []any{"nope"}) {
		t.Fatalf("non-object patch should replace base, got %#v", got)
	}
}

func TestMergeProfileConfigurationKeepsUnpatchedSections(t *testing.T) {
	t.Parallel()

	mode := v3.ENUMPROPERTIESCONFIGURATIONPROPERTIESMODE_READ_ONLY
	current := v3.GetProfile200ResponseConfiguration{Mode: &mode}
	patch := []byte(`{
		"configuration": {"frontend_dast": {"hotstart": ["https://app.example.com/#/accounts"]}},
		"note": "keep"
	}`)

	raw, err := mergeProfileConfiguration(current, patch)
	if err != nil {
		t.Fatalf("mergeProfileConfiguration: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal merged body: %v", err)
	}

	if got["note"] != "keep" {
		t.Fatalf("top-level patch key dropped: %#v", got)
	}

	cfg, ok := got["configuration"].(map[string]any)
	if !ok {
		t.Fatalf("configuration = %#v", got["configuration"])
	}

	if cfg["mode"] != "read_only" {
		t.Fatalf("mode = %#v, want read_only", cfg["mode"])
	}

	frontend, ok := cfg["frontend_dast"].(map[string]any)
	if !ok {
		t.Fatalf("frontend_dast = %#v", cfg["frontend_dast"])
	}

	hotstart, ok := frontend["hotstart"].([]any)
	if !ok || len(hotstart) != 1 || hotstart[0] != "https://app.example.com/#/accounts" {
		t.Fatalf("hotstart = %#v", frontend["hotstart"])
	}
}

func TestMergeProfileConfigurationRejectsPartialShape(t *testing.T) {
	t.Parallel()

	current := v3.GetProfile200ResponseConfiguration{}
	cases := []struct {
		name  string
		patch string
		want  string
	}{
		{name: "null", patch: "null", want: `"configuration" object`},
		{name: "missing", patch: `{"mode":"read_only"}`, want: `"configuration" object`},
		{name: "array", patch: `{"configuration":[]}`, want: "JSON object"},
		{name: "invalid", patch: `{`, want: "invalid JSON"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := mergeProfileConfiguration(current, []byte(tc.patch))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestProfileUpdateConfigurationMergeFlagDefaultsFalse(t *testing.T) {
	flag := profileUpdateConfigurationCmd.Flags().Lookup("merge")
	if flag == nil {
		t.Fatal("missing --merge flag")
	}

	if flag.DefValue != "false" {
		t.Fatalf("CLI --merge default = %q, want false", flag.DefValue)
	}
}

func TestProfilesUpdateConfigurationMCPDefaultsMerge(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	var found bool
	for _, spec := range specs {
		if spec.Name != "profiles_update_configuration" {
			continue
		}

		found = true
		var bindingFound bool
		for _, binding := range spec.FlagBindings {
			if binding.FlagName != "merge" {
				continue
			}

			bindingFound = true
			if binding.Default != true {
				t.Fatalf("MCP merge default = %#v, want true", binding.Default)
			}
		}

		if !bindingFound {
			t.Fatal("profiles_update_configuration is missing the merge flag binding")
		}

		var schema map[string]any
		if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
			t.Fatalf("schema: %v", err)
		}

		properties, _ := schema["properties"].(map[string]any)
		merge, _ := properties["merge"].(map[string]any)
		if merge["default"] != true {
			t.Fatalf("schema merge default = %#v", merge)
		}

		desc, _ := merge["description"].(string)
		if !strings.Contains(desc, "MCP default: true") {
			t.Fatalf("merge description = %q", desc)
		}

		fromFlag, kind := schemaForFlag(profileUpdateConfigurationCmd.Flags().Lookup("merge"))
		if kind != "bool" || fromFlag["default"] != true {
			t.Fatalf("schemaForFlag merge = %#v (%s)", fromFlag, kind)
		}

		stub, err := climcp.BuildStubTool(spec)
		if err != nil {
			t.Fatalf("stub: %v", err)
		}

		var stubSchema map[string]any
		if err := json.Unmarshal(stub.RawInputSchema, &stubSchema); err != nil {
			t.Fatalf("stub schema: %v", err)
		}

		stubProps, _ := stubSchema["properties"].(map[string]any)
		stubMerge, _ := stubProps["merge"].(map[string]any)
		if stubMerge["default"] != true {
			t.Fatalf("stub merge default = %#v", stubMerge)
		}

		stubDesc, _ := stubMerge["description"].(string)
		if !strings.Contains(stubDesc, "MCP default: true") {
			t.Fatalf("stub merge description = %q", stubDesc)
		}
	}

	if !found {
		t.Fatal("profiles_update_configuration tool was not built")
	}
}
