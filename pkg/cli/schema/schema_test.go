package schema

import (
	"testing"
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
	Name    string  `json:"name"`
	Deleted *bool   `json:"deleted,omitempty"`
	Count   *int    `json:"count"`
	Inner   *simpleStruct `json:"inner,omitempty"`
}

type tagVariations struct {
	Explicit    string `json:"explicit_name"`
	Omitempty   string `json:"omit_field,omitempty"`
	Skipped     string `json:"-"`
	NoTag       string
	Described   string `json:"described" description:"A helpful description"`
	unexported  string //nolint:unused
}

type withAdditionalProps struct {
	Name                 string         `json:"name"`
	AdditionalProperties map[string]any `json:"AdditionalProperties,omitempty"`
}

type withSlice struct {
	Items []string       `json:"items"`
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

	// Pointer fields should be nullable
	deleted := s.Properties["deleted"]
	if deleted == nil {
		t.Fatal("expected deleted property")
	}
	if !deleted.Nullable {
		t.Error("pointer field should be nullable")
	}
	if deleted.Type != "boolean" {
		t.Errorf("expected boolean type for *bool, got %q", deleted.Type)
	}

	// Pointer to struct
	inner := s.Properties["inner"]
	if inner == nil {
		t.Fatal("expected inner property")
	}
	if !inner.Nullable {
		t.Error("pointer to struct should be nullable")
	}
	if inner.Type != "object" {
		t.Errorf("expected object type for *struct, got %q", inner.Type)
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

	if !s.Nullable {
		t.Error("pointer to struct should be nullable")
	}
	if s.Type != "object" {
		t.Errorf("expected type object, got %q", s.Type)
	}
	if s.Properties["name"] == nil {
		t.Error("expected name property through pointer")
	}
}

func TestGenerateInterface(t *testing.T) {
	t.Parallel()

	var v any
	s := Generate(&v)

	if s.Type != "object" {
		t.Errorf("expected type object for interface, got %q", s.Type)
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

// helper
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
