package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// Test types used across tests
type simpleStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email,omitempty"`
}

type nestedStruct struct {
	ID      int          `json:"id"`
	Profile simpleStruct `json:"profile"`
}

type pointerFieldStruct struct {
	Name    string        `json:"name"`
	Deleted *bool         `json:"deleted,omitempty"`
	Count   *int          `json:"count"`
	Inner   *simpleStruct `json:"inner,omitempty"`
}

type tagVariations struct {
	Explicit   string `json:"explicit_name"`
	Omitempty  string `json:"omit_field,omitempty"`
	Skipped    string `json:"-"`
	NoTag      string
	Described  string `json:"described" description:"A helpful description"`
	unexported string //nolint:unused
}

type withAdditionalProps struct {
	Name                 string         `json:"name"`
	AdditionalProperties map[string]any `json:"AdditionalProperties,omitempty"`
}

type withSlice struct {
	Items  []string       `json:"items"`
	Nested []simpleStruct `json:"nested"`
}

type withMap struct {
	Metadata map[string]string `json:"metadata"`
}

type EmbeddedBase struct {
	ID string `json:"id"`
}

type withEmbedded struct {
	EmbeddedBase
	Name string `json:"name"`
}

func TestGeneratePrimitiveTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    any
		wantType string
	}{
		{"string", "", "string"},
		{"int", 0, "integer"},
		{"int8", int8(0), "integer"},
		{"int16", int16(0), "integer"},
		{"int32", int32(0), "integer"},
		{"int64", int64(0), "integer"},
		{"uint", uint(0), "integer"},
		{"uint8", uint8(0), "integer"},
		{"uint16", uint16(0), "integer"},
		{"uint32", uint32(0), "integer"},
		{"uint64", uint64(0), "integer"},
		{"float32", float32(0), "number"},
		{"float64", float64(0), "number"},
		{"bool", false, "boolean"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := Generate(tc.input)
			if s.Type != tc.wantType {
				t.Errorf("expected type %q, got %q", tc.wantType, s.Type)
			}

			if s.Schema != "https://json-schema.org/draft/2020-12/schema" {
				t.Errorf("expected $schema on root, got %q", s.Schema)
			}
		})
	}
}

func TestGenerateStruct(t *testing.T) {
	t.Parallel()

	s := Generate(simpleStruct{})

	if s.Type != "object" {
		t.Fatalf("expected type object, got %q", s.Type)
	}

	if len(s.Properties) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(s.Properties))
	}

	if s.Properties["name"] == nil || s.Properties["name"].Type != "string" {
		t.Error("expected name property with type string")
	}

	if s.Properties["age"] == nil || s.Properties["age"].Type != "integer" {
		t.Error("expected age property with type integer")
	}

	if s.Properties["email"] == nil || s.Properties["email"].Type != "string" {
		t.Error("expected email property with type string")
	}
}

func TestGenerateStructRequiredFields(t *testing.T) {
	t.Parallel()

	s := Generate(simpleStruct{})

	if !contains(s.Required, "name") {
		t.Error("expected name to be required")
	}

	if !contains(s.Required, "age") {
		t.Error("expected age to be required")
	}

	if contains(s.Required, "email") {
		t.Error("email has omitempty, should not be required")
	}
}

func TestGenerateStructJSONTags(t *testing.T) {
	t.Parallel()

	s := Generate(tagVariations{})

	// Explicit name from tag
	if s.Properties["explicit_name"] == nil {
		t.Error("expected property with explicit json tag name")
	}

	// omitempty field present but not required
	if s.Properties["omit_field"] == nil {
		t.Error("expected omit_field property")
	}

	if contains(s.Required, "omit_field") {
		t.Error("omit_field should not be required")
	}

	// json:"-" should be skipped
	if s.Properties["Skipped"] != nil || s.Properties["-"] != nil {
		t.Error("json:\"-\" field should be excluded")
	}

	// No tag: falls back to field name
	if s.Properties["NoTag"] == nil {
		t.Error("expected field with no json tag to use Go field name")
	}

	// unexported fields should be skipped
	if s.Properties["unexported"] != nil {
		t.Error("unexported field should be excluded")
	}
}

func TestGenerateStructDescription(t *testing.T) {
	t.Parallel()

	s := Generate(tagVariations{})

	prop := s.Properties["described"]
	if prop == nil {
		t.Fatal("expected described property")
	}

	if prop.Description != "A helpful description" {
		t.Errorf("expected description tag value, got %q", prop.Description)
	}
}

func TestGenerateAdditionalPropertiesSkipped(t *testing.T) {
	t.Parallel()

	s := Generate(withAdditionalProps{})

	if s.Properties["AdditionalProperties"] != nil {
		t.Error("AdditionalProperties field should be excluded")
	}

	if s.Properties["name"] == nil {
		t.Error("name property should still be present")
	}
}

func TestGenerateNestedStruct(t *testing.T) {
	t.Parallel()

	s := Generate(nestedStruct{})

	profile := s.Properties["profile"]
	if profile == nil {
		t.Fatal("expected profile property")
	}

	if profile.Type != "object" {
		t.Errorf("expected nested type object, got %q", profile.Type)
	}

	if profile.Properties["name"] == nil {
		t.Error("expected nested name property")
	}

	// Nested schemas should not have $schema
	if profile.Schema != "" {
		t.Error("nested schema should not have $schema")
	}
}

func TestGeneratePointerFields(t *testing.T) {
	t.Parallel()

	s := Generate(pointerFieldStruct{})

	// Pointer fields allow JSON null without the non-standard nullable keyword.
	deleted := s.Properties["deleted"]
	if deleted == nil {
		t.Fatal("expected deleted property")
	}

	if !allowsNull(deleted) {
		t.Error("pointer field should allow null")
	}

	if !hasType(deleted, "boolean") {
		t.Errorf("expected boolean type for *bool, got %#v", deleted.Type)
	}

	// Pointer to struct
	inner := s.Properties["inner"]
	if inner == nil {
		t.Fatal("expected inner property")
	}

	if !allowsNull(inner) {
		t.Error("pointer to struct should allow null")
	}

	objectBranch := inner
	if len(inner.AnyOf) > 0 {
		objectBranch = inner.AnyOf[0]
	}

	if objectBranch.Type != "object" {
		t.Errorf("expected object type for *struct, got %#v", objectBranch.Type)
	}

	if objectBranch.Properties["name"] == nil {
		t.Error("expected name property inside nullable struct")
	}

	// Pointer fields should not be required (regardless of omitempty)
	if contains(s.Required, "deleted") {
		t.Error("pointer field with omitempty should not be required")
	}

	if contains(s.Required, "count") {
		t.Error("pointer field should not be required")
	}
}

func TestGenerateSlice(t *testing.T) {
	t.Parallel()

	s := Generate([]string{})

	if s.Type != "array" {
		t.Fatalf("expected type array, got %q", s.Type)
	}

	if s.Items == nil {
		t.Fatal("expected items schema")
	}

	if s.Items.Type != "string" {
		t.Errorf("expected items type string, got %q", s.Items.Type)
	}
}

func TestGenerateSliceOfStructs(t *testing.T) {
	t.Parallel()

	s := Generate([]simpleStruct{})

	if s.Type != "array" {
		t.Fatalf("expected type array, got %q", s.Type)
	}

	if s.Items == nil {
		t.Fatal("expected items schema")
	}

	if s.Items.Type != "object" {
		t.Errorf("expected items type object, got %q", s.Items.Type)
	}

	if s.Items.Properties["name"] == nil {
		t.Error("expected name property in items schema")
	}
}

func TestGenerateStructWithSliceField(t *testing.T) {
	t.Parallel()

	s := Generate(withSlice{})

	items := s.Properties["items"]
	if items == nil || items.Type != "array" {
		t.Fatal("expected items field with type array")
	}

	if items.Items == nil || items.Items.Type != "string" {
		t.Error("expected string items in items array")
	}

	nested := s.Properties["nested"]
	if nested == nil || nested.Type != "array" {
		t.Fatal("expected nested field with type array")
	}

	if nested.Items == nil || nested.Items.Type != "object" {
		t.Error("expected object items in nested array")
	}
}

func TestGenerateMap(t *testing.T) {
	t.Parallel()

	s := Generate(map[string]string{})

	if s.Type != "object" {
		t.Errorf("expected type object for map, got %q", s.Type)
	}
}

func TestGenerateStructWithMapField(t *testing.T) {
	t.Parallel()

	s := Generate(withMap{})

	metadata := s.Properties["metadata"]
	if metadata == nil {
		t.Fatal("expected metadata property")
	}

	if metadata.Type != "object" {
		t.Errorf("expected object type for map field, got %q", metadata.Type)
	}
}

func TestGenerateEmbeddedStruct(t *testing.T) {
	t.Parallel()

	s := Generate(withEmbedded{})

	if s.Type != "object" {
		t.Fatalf("expected type object, got %q", s.Type)
	}

	if s.Properties["name"] == nil {
		t.Error("expected name property")
	}

	// Embedded struct is treated as a nested object property (not promoted)
	embedded := s.Properties["EmbeddedBase"]
	if embedded == nil {
		t.Fatal("expected EmbeddedBase property")
	}

	if embedded.Type != "object" {
		t.Errorf("expected embedded field type object, got %q", embedded.Type)
	}

	if embedded.Properties["id"] == nil {
		t.Error("expected id property inside embedded struct")
	}
}

func TestGeneratePointerToStruct(t *testing.T) {
	t.Parallel()

	s := Generate(&simpleStruct{})

	if !allowsNull(s) {
		t.Error("pointer to struct should allow null")
	}

	if s.Schema == "" {
		t.Error("root schema should have $schema")
	}

	objectBranch := s
	if len(s.AnyOf) > 0 {
		objectBranch = s.AnyOf[0]
	}

	if objectBranch.Type != "object" {
		t.Errorf("expected type object, got %#v", objectBranch.Type)
	}

	if objectBranch.Properties["name"] == nil {
		t.Error("expected name property through pointer")
	}
}

func TestGenerateInterface(t *testing.T) {
	t.Parallel()

	var v any
	s := Generate(&v)

	if !hasType(s, "object") {
		t.Errorf("expected type object for interface, got %#v", s.Type)
	}

	if !allowsNull(s) {
		t.Error("pointer to interface should allow null")
	}
}

func TestSchemaFieldNotOnNested(t *testing.T) {
	t.Parallel()

	s := Generate(nestedStruct{})

	// Root should have $schema
	if s.Schema == "" {
		t.Error("root schema should have $schema")
	}

	// Nested properties should not
	for name, prop := range s.Properties {
		if prop.Schema != "" {
			t.Errorf("nested property %q should not have $schema", name)
		}
	}
}

type customSchema struct{}

func (customSchema) JSONSchema() *JSONSchema {
	return &JSONSchema{Type: "object", Description: "custom", OneOf: []*JSONSchema{{Type: "string"}}}
}

func TestGenerateSchemaProvider(t *testing.T) {
	t.Parallel()

	s := Generate(customSchema{})
	if s.Description != "custom" {
		t.Fatalf("expected provider schema, got %#v", s)
	}

	if len(s.OneOf) != 1 || s.OneOf[0].Type != "string" {
		t.Fatalf("expected oneOf to round-trip, got %#v", s.OneOf)
	}
}

type datedSample struct {
	Created time.Time  `json:"created"`
	Deleted *time.Time `json:"deleted,omitempty"`
}

type enumSample struct {
	Severity v3.ENUMPROPERTIESFILTERPROPERTIESSEVERITIESITEMS  `json:"severity"`
	Note     *v3.ENUMPROPERTIESFILTERPROPERTIESSEVERITIESITEMS `json:"note,omitempty"`
}

type nullableSample struct {
	value *simpleStruct
	isSet bool
}

type recNode struct {
	Name  string   `json:"name"`
	Child *recNode `json:"child,omitempty"`
}

func TestGenerateDateTimeSnapshot(t *testing.T) {
	t.Parallel()

	got := jsonOf(t, datedSample{})
	want := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "created": {
      "type": "string",
      "format": "date-time"
    },
    "deleted": {
      "type": [
        "string",
        "null"
      ],
      "format": "date-time"
    }
  },
  "required": [
    "created"
  ]
}`
	if got != want {
		t.Fatalf("date schema mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestGenerateEnumSnapshot(t *testing.T) {
	t.Parallel()

	got := jsonOf(t, enumSample{})
	want := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "note": {
      "anyOf": [
        {
          "type": "string",
          "enum": [
            "CRITICAL",
            "HIGH",
            "INFO",
            "LOW",
            "MEDIUM"
          ]
        },
        {
          "type": "null"
        }
      ]
    },
    "severity": {
      "type": "string",
      "enum": [
        "CRITICAL",
        "HIGH",
        "INFO",
        "LOW",
        "MEDIUM"
      ]
    }
  },
  "required": [
    "severity"
  ]
}`
	if got != want {
		t.Fatalf("enum schema mismatch\n got: %s\nwant: %s", got, want)
	}

	if strings.Contains(got, "nullable") {
		t.Fatalf("schema still uses nullable: %s", got)
	}
}

func TestGenerateUnionSnapshot(t *testing.T) {
	t.Parallel()

	got := jsonOf(t, v3.GetIssueTrendsApplicationIdsParameter{})
	want := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "anyOf": [
    {
      "type": "array",
      "items": {
        "type": "string"
      }
    },
    {
      "type": "string"
    }
  ]
}`
	if got != want {
		t.Fatalf("union schema mismatch\n got: %s\nwant: %s", got, want)
	}

	if strings.Contains(got, "ArrayOfString") {
		t.Fatalf("union schema kept the Go wrapper name: %s", got)
	}
}

type largeEnumSample struct {
	Kind v3.ENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMS  `json:"kind" description:"scanner kind"`
	Also v3.ENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMS  `json:"also"`
	Note *v3.ENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMS `json:"note,omitempty"`
}

func TestGenerateHoistsLargeEnums(t *testing.T) {
	t.Parallel()

	raw := jsonOf(t, largeEnumSample{})
	if strings.Count(raw, "ASM_APK") != 1 {
		t.Fatalf("ASM_APK count = %d, want 1\n%s", strings.Count(raw, "ASM_APK"), raw)
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}

	defs, _ := doc["$defs"].(map[string]any)
	if len(defs) != 1 {
		t.Fatalf("$defs = %#v", defs)
	}

	props := doc["properties"].(map[string]any)
	kind := props["kind"].(map[string]any)
	also := props["also"].(map[string]any)
	if kind["$ref"] == nil || kind["$ref"] != also["$ref"] {
		t.Fatalf("kind and also should share one $ref: %#v %#v", kind, also)
	}

	if _, ok := kind["enum"]; ok {
		t.Fatalf("hoisted field kept an inline enum: %#v", kind)
	}

	if kind["description"] != "scanner kind" {
		t.Fatalf("description = %#v", kind["description"])
	}

	note := props["note"].(map[string]any)
	branches, _ := note["anyOf"].([]any)
	if len(branches) != 2 {
		t.Fatalf("nullable enum = %#v", note)
	}

	refBranch := branches[0].(map[string]any)
	nullBranch := branches[1].(map[string]any)
	if refBranch["$ref"] != kind["$ref"] || nullBranch["type"] != "null" {
		t.Fatalf("nullable branches = %#v", branches)
	}
}

func TestHoistLargeEnumsThreshold(t *testing.T) {
	t.Parallel()

	small := make([]string, largeEnumLimit)
	large := make([]string, largeEnumLimit+1)
	for i := range large {
		large[i] = strings.Repeat("V", i+1)
	}

	copy(small, large)
	root := &JSONSchema{
		Type: "object",
		Properties: map[string]*JSONSchema{
			"small": {Type: "string", Enum: small},
			"large": {Type: "string", Enum: append([]string(nil), large...), Description: "kept"},
			"again": {Type: "string", Enum: append([]string(nil), large...)},
		},
	}
	hoistLargeEnums(root)
	if root.Properties["small"].Ref != "" || len(root.Properties["small"].Enum) != largeEnumLimit {
		t.Fatalf("small enum = %#v", root.Properties["small"])
	}

	if len(root.Properties["large"].Enum) != 0 || root.Properties["large"].Ref == "" {
		t.Fatalf("large enum = %#v", root.Properties["large"])
	}

	if root.Properties["large"].Ref != root.Properties["again"].Ref {
		t.Fatalf("refs = %s %s", root.Properties["large"].Ref, root.Properties["again"].Ref)
	}

	if root.Properties["large"].Description != "kept" {
		t.Fatalf("description = %q", root.Properties["large"].Description)
	}

	if len(root.Defs) != 1 {
		t.Fatalf("$defs = %#v", root.Defs)
	}
}

func TestGenerateWorkflowActionUnionDropsWrapperNames(t *testing.T) {
	t.Parallel()

	schema := Generate(v3.CreateWorkflowRequestActionsInner{})
	if len(schema.AnyOf) < 2 {
		t.Fatalf("expected anyOf variants, got %#v", schema.AnyOf)
	}

	raw := jsonOf(t, v3.CreateWorkflowRequestActionsInner{})
	for _, wrapper := range []string{
		"CreateNotifyWorkflowActionUsingSlack",
		"ArrayOfString",
		"nullable",
	} {
		if strings.Contains(raw, wrapper) {
			t.Fatalf("schema contains %q", wrapper)
		}
	}

	var slack *JSONSchema
	for _, variant := range schema.AnyOf {
		kind := variant.Properties["integrationKind"]
		if kind == nil {
			continue
		}

		if contains(kind.Enum, "SLACK_WEBHOOK") {
			slack = variant
			break
		}
	}

	if slack == nil {
		t.Fatal("expected a slack action variant with integrationKind SLACK_WEBHOOK")
	}

	if !contains(slack.Properties["type"].Enum, "NOTIFY") {
		t.Fatalf("slack action type enum = %#v", slack.Properties["type"].Enum)
	}

	params := slack.Properties["parameters"]
	if params == nil || params.Properties["url"] == nil {
		t.Fatal("expected slack parameters.url")
	}
}

func TestGenerateNullableWrapper(t *testing.T) {
	t.Parallel()

	sample := nullableSample{}
	if sample.value != nil || sample.isSet {
		t.Fatal("nullable fixture must start unset")
	}

	schema := Generate(sample)
	if !allowsNull(schema) {
		t.Fatal("NullableT wrapper should allow null")
	}

	branch := schema
	if len(schema.AnyOf) > 0 {
		branch = schema.AnyOf[0]
	}

	if branch.Properties["name"] == nil {
		t.Fatalf("expected inner struct properties, got %s", jsonOf(t, nullableSample{}))
	}

	if strings.Contains(jsonOf(t, nullableSample{}), "\"value\"") {
		t.Fatal("nullable wrapper leaked the Go value field")
	}
}

func TestGenerateRecursiveStructTerminates(t *testing.T) {
	t.Parallel()

	schema := Generate(recNode{})
	child := schema.Properties["child"]
	if child == nil || !allowsNull(child) {
		t.Fatalf("recursive child = %#v", child)
	}

	if schema.Properties["name"] == nil || schema.Properties["name"].Type != "string" {
		t.Fatal("expected name string on recursive node")
	}
}

func TestGeneratedEnumListsAreRegistered(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	v3Dir := filepath.Join(filepath.Dir(file), "..", "..", "api", "v3")
	enumsSrc, err := os.ReadFile(filepath.Join(filepath.Dir(file), "enums.go"))
	if err != nil {
		t.Fatal(err)
	}

	pat := regexp.MustCompile(`var\s+(Allowed[A-Za-z0-9]+EnumValues)\s*=`)
	var missing []string
	seen := map[string]struct{}{}
	err = filepath.WalkDir(v3Dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		for _, match := range pat.FindAllSubmatch(body, -1) {
			name := string(match[1])
			if _, dup := seen[name]; dup {
				continue
			}

			seen[name] = struct{}{}
			if !strings.Contains(string(enumsSrc), "v3."+name+",") {
				missing = append(missing, name)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(seen) == 0 {
		t.Fatal("found no Allowed*EnumValues slices")
	}

	if len(missing) > 0 {
		t.Fatalf("enums.go is missing %d generated enum lists, including %v", len(missing), missing[:min(8, len(missing))])
	}

	// No count check against enumValuesByType: it is keyed by element type,
	// so two slices sharing a type register once. The missing check above
	// already proves every slice is listed.
}

func jsonOf(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.MarshalIndent(Generate(v), "", "  ")
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	return string(raw)
}

func hasType(schema *JSONSchema, want string) bool {
	switch typed := schema.Type.(type) {
	case string:
		return typed == want
	case []string:
		return contains(typed, want)
	default:
		return false
	}
}

func allowsNull(schema *JSONSchema) bool {
	if schema == nil {
		return false
	}

	if hasType(schema, "null") {
		return true
	}

	for _, branch := range schema.AnyOf {
		if branch != nil && branch.Type == "null" {
			return true
		}
	}

	return false
}

// helper
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
