// Package schema builds JSON Schema documents from Go types for CLI output
// and the hosted MCP server.
//
// Generated OpenAPI models need rules reflect does not know about. time.Time
// is a date-time string. Pointers and OpenAPI NullableT wrappers are JSON
// Schema null unions (draft 2020-12 has no nullable keyword). anyOf wrappers
// are flattened so Go variant names never appear as properties. Enum types
// publish the Allowed*EnumValues lists registered in enums.go.
package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
)

//go:generate go run ./genenums

const jsonSchemaDraft = "https://json-schema.org/draft/2020-12/schema"

// JSONSchema is one JSON Schema node (draft 2020-12).
//
// Type is a JSON string ("object") or a list of strings (["string","null"])
// for a nullable primitive. Nullable objects, arrays, enums, and unions use
// AnyOf with a {"type":"null"} branch. Both forms are valid JSON Schema;
// the old "nullable" keyword is not.
type JSONSchema struct {
	Schema      string                 `json:"$schema,omitempty"`
	Type        any                    `json:"type,omitempty"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	OneOf       []*JSONSchema          `json:"oneOf,omitempty"`
	AnyOf       []*JSONSchema          `json:"anyOf,omitempty"`
	Format      string                 `json:"format,omitempty"`
	// Ref points at a $defs entry. Large enums are hoisted so a schema that
	// uses the same enum many times does not repeat the value list.
	Ref string `json:"$ref,omitempty"`
	// Defs holds hoisted schemas for this document. Only the root sets it.
	Defs map[string]*JSONSchema `json:"$defs,omitempty"`
	// AdditionalProperties is a pointer so false is emitted. nil means the
	// keyword is omitted, which JSON Schema treats as allowing extra fields.
	AdditionalProperties *bool `json:"additionalProperties,omitempty"`
}

// largeEnumLimit is where inlining an enum at every use dominates
// escape_get_tool_spec. Smaller enums stay on the field so it reads alone.
const largeEnumLimit = 20

var (
	timeType        = reflect.TypeOf(time.Time{})
	jsonUnmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

	// enumValuesByType is filled from enums.go. Reflect cannot see the
	// package-level Allowed*EnumValues slices in the generated client.
	enumValuesByType = map[reflect.Type][]string{}
)

// schemaProvider is implemented by values whose advertised JSON Schema is not
// the reflection of their Go type. CLI inputs that are a union of generated
// models (asset create, integration create) use this so MCP sees every variant.
type schemaProvider interface {
	JSONSchema() *JSONSchema
}

// Generate creates a JSON Schema from a Go type. Values that implement
// schemaProvider supply the schema themselves.
func Generate(v any) *JSONSchema {
	if provider, ok := v.(schemaProvider); ok {
		if schema := provider.JSONSchema(); schema != nil {
			return schema
		}
	}

	g := &genState{
		stack: map[reflect.Type]struct{}{},
		cache: map[reflect.Type]*JSONSchema{},
	}
	schema, _ := g.build(reflect.TypeOf(v))
	if schema == nil {
		schema = &JSONSchema{Type: "object"}
	}

	// Clone so the draft URI stays on this root and not on a cached node
	// that may also appear nested.
	cloned := *schema
	cloned.Schema = jsonSchemaDraft
	hoistLargeEnums(&cloned)

	return &cloned
}

// hoistLargeEnums moves enums longer than largeEnumLimit into root $defs and
// replaces each use with $ref. The value list stays available once. Nested
// anyOf (nullable enums, union bodies) is left in place.
func hoistLargeEnums(root *JSONSchema) {
	defs := map[string]*JSONSchema{}
	names := map[string]string{}
	var walk func(*JSONSchema)
	walk = func(node *JSONSchema) {
		if node == nil || node.Ref != "" {
			return
		}

		if len(node.Enum) > largeEnumLimit {
			key := strings.Join(node.Enum, "\n")
			name, ok := names[key]
			if !ok {
				sum := sha256.Sum256([]byte(key))
				name = "enum_" + hex.EncodeToString(sum[:6])
				names[key] = name
				defs[name] = &JSONSchema{Type: node.Type, Enum: node.Enum, Format: node.Format}
			}

			node.Ref = "#/$defs/" + name
			node.Enum = nil
			node.Type = nil
			node.Format = ""
		}

		for _, child := range node.Properties {
			walk(child)
		}

		walk(node.Items)
		for _, child := range node.AnyOf {
			walk(child)
		}

		for _, child := range node.OneOf {
			walk(child)
		}
	}
	walk(root)
	if len(defs) > 0 {
		root.Defs = defs
	}
}

// genState walks a type once per Generate call. stack is the structs currently
// being expanded; cache stores finished nodes so a DAG of models stays linear.
type genState struct {
	stack map[reflect.Type]struct{}
	cache map[reflect.Type]*JSONSchema
}

// build returns the schema for t. The bool is true only when t itself is
// already on the stack (a cycle edge), not when a nested field is cyclic.
// Callers use that bit to avoid caching a placeholder as if it were the
// finished schema.
func (g *genState) build(t reflect.Type) (*JSONSchema, bool) {
	if t == nil {
		return &JSONSchema{Type: "object"}, false
	}

	if t.Kind() == reflect.Ptr {
		inner, cyclic := g.build(t.Elem())
		return withNull(inner), cyclic
	}

	if schema, ok := g.cache[t]; ok {
		return schema, false
	}

	if t == timeType {
		schema := &JSONSchema{Type: "string", Format: "date-time"}
		g.cache[t] = schema

		return schema, false
	}

	if inner, ok := nullableInner(t); ok {
		value, cyclic := g.build(inner)
		schema := withNull(value)
		if !cyclic {
			g.cache[t] = schema
		}

		return schema, cyclic
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		items, cyclic := g.build(t.Elem())
		schema := &JSONSchema{Type: "array", Items: items}
		// A slice of the struct currently on the stack must not be cached:
		// its items are the cycle placeholder, not the finished element schema.
		if !cyclic {
			g.cache[t] = schema
		}

		return schema, cyclic
	case reflect.Struct:
		if _, busy := g.stack[t]; busy {
			return &JSONSchema{Type: "object"}, true
		}

		g.stack[t] = struct{}{}
		schema := g.structSchema(t)
		delete(g.stack, t)
		g.cache[t] = schema

		return schema, false
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map, reflect.Pointer, reflect.String, reflect.UnsafePointer:
		// Pointers are unwrapped above. Everything else is a scalar or has no
		// JSON shape; scalarSchema decides which.
	}

	schema := scalarSchema(t)
	g.cache[t] = schema

	return schema, false
}

func (g *genState) schema(t reflect.Type) *JSONSchema {
	schema, _ := g.build(t)
	return schema
}

func (g *genState) structSchema(t reflect.Type) *JSONSchema {
	if isUnionWrapper(t) {
		variants := make([]*JSONSchema, 0, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() || field.Type.Kind() != reflect.Pointer {
				continue
			}

			// The pointer is the generator's "which variant is set" slot,
			// not a JSON null. Schema the element.
			variants = append(variants, g.schema(field.Type.Elem()))
		}

		return &JSONSchema{AnyOf: variants}
	}

	schema := &JSONSchema{
		Type:       "object",
		Properties: map[string]*JSONSchema{},
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		tagParts := strings.Split(jsonTag, ",")
		fieldName := tagParts[0]
		if fieldName == "" {
			fieldName = field.Name
		}

		if fieldName == "AdditionalProperties" || field.Name == "AdditionalProperties" {
			continue
		}

		optional := false
		for _, part := range tagParts[1:] {
			if part == "omitempty" {
				optional = true
				break
			}
		}

		prop := g.schema(field.Type)
		if desc := field.Tag.Get("description"); desc != "" {
			// prop may be cached and shared. Clone before attaching a
			// description that belongs only to this field.
			cloned := *prop
			cloned.Description = desc
			prop = &cloned
		}

		schema.Properties[fieldName] = prop
		if !optional && field.Type.Kind() != reflect.Ptr {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	return schema
}

func scalarSchema(t reflect.Type) *JSONSchema {
	schema := &JSONSchema{}
	switch t.Kind() {
	case reflect.String:
		schema.Type = "string"
		schema.Enum = enumValues(t)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
		schema.Enum = enumValues(t)
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice, reflect.Struct, reflect.UnsafePointer:
		// Maps, interfaces, and channels have no useful JSON shape here.
		schema.Type = "object"
	}

	return schema
}

// nullableInner reports the inner type of an OpenAPI Generator NullableT:
// `struct { value *T; isSet bool }`. Those two fields are a presence flag,
// not JSON properties. The wire value is T or null.
func nullableInner(t reflect.Type) (reflect.Type, bool) {
	if t.Kind() != reflect.Struct || t.NumField() != 2 {
		return nil, false
	}

	value, hasValue := t.FieldByName("value")
	isSet, hasSet := t.FieldByName("isSet")
	if !hasValue || !hasSet || value.IsExported() || isSet.IsExported() {
		return nil, false
	}

	if value.Type.Kind() != reflect.Pointer || isSet.Type.Kind() != reflect.Bool {
		return nil, false
	}

	return value.Type.Elem(), true
}

// minUnionVariants is the smallest anyOf wrapper. One exported pointer field
// is an ordinary struct, not a union.
const minUnionVariants = 2

// isUnionWrapper reports an OpenAPI Generator anyOf struct. Every exported
// field is an untagged pointer, and UnmarshalJSON selects one of them. The
// field names (CreateNotifyWorkflowActionUsingSlack, ArrayOfString) exist
// only in Go; the JSON value is the variant itself.
func isUnionWrapper(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	if !t.Implements(jsonUnmarshaler) && !reflect.PointerTo(t).Implements(jsonUnmarshaler) {
		return false
	}

	variants := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous || field.Tag.Get("json") != "" || field.Type.Kind() != reflect.Pointer {
			return false
		}

		variants++
	}

	return variants >= minUnionVariants
}

// withNull adds null to a schema without the non-standard nullable keyword.
// Primitives (and primitives with a format, such as date-time) become a type
// union. Pure unions gain a null branch. Anything with structure becomes
// anyOf [schema, null], because a type union would not express it and an
// enum list would reject null unless null were added as a value.
func withNull(inner *JSONSchema) *JSONSchema {
	if inner == nil {
		return &JSONSchema{Type: "null"}
	}

	if pureAnyOf(inner) {
		if anyOfHasNull(inner) {
			return inner
		}

		cloned := *inner
		cloned.AnyOf = make([]*JSONSchema, len(inner.AnyOf), len(inner.AnyOf)+1)
		copy(cloned.AnyOf, inner.AnyOf)
		cloned.AnyOf = append(cloned.AnyOf, &JSONSchema{Type: "null"})

		return &cloned
	}

	if name, ok := simpleType(inner); ok {
		cloned := *inner
		cloned.Type = []string{name, "null"}

		return &cloned
	}

	return &JSONSchema{AnyOf: []*JSONSchema{inner, {Type: "null"}}}
}

func simpleType(schema *JSONSchema) (string, bool) {
	if schema == nil {
		return "", false
	}

	name, ok := schema.Type.(string)
	if !ok || name == "" {
		return "", false
	}

	if len(schema.Enum) > 0 || len(schema.Properties) > 0 || schema.Items != nil ||
		len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 || len(schema.Required) > 0 {
		return "", false
	}

	return name, true
}

func pureAnyOf(schema *JSONSchema) bool {
	return schema != nil &&
		len(schema.AnyOf) > 0 &&
		schema.Type == nil &&
		schema.Description == "" &&
		len(schema.Properties) == 0 &&
		schema.Items == nil &&
		len(schema.Required) == 0 &&
		len(schema.Enum) == 0 &&
		len(schema.OneOf) == 0 &&
		schema.Format == "" &&
		schema.Schema == ""
}

func anyOfHasNull(schema *JSONSchema) bool {
	for _, branch := range schema.AnyOf {
		if branch != nil && branch.Type == "null" && len(branch.AnyOf) == 0 && len(branch.Enum) == 0 {
			return true
		}
	}

	return false
}

func enumValues(t reflect.Type) []string {
	src := enumValuesByType[t]
	if len(src) == 0 {
		return nil
	}

	return slices.Clone(src)
}

func registerEnumLists(lists ...any) {
	for _, list := range lists {
		value := reflect.ValueOf(list)
		if value.Kind() != reflect.Slice || value.Type().Elem().Kind() != reflect.String {
			continue
		}

		values := make([]string, value.Len())
		for i := 0; i < value.Len(); i++ {
			values[i] = value.Index(i).String()
		}

		enumValuesByType[value.Type().Elem()] = values
	}
}

// Print outputs the JSON Schema for the given type.
func Print(v any) error {
	schema := Generate(v)
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	fmt.Println(string(output))

	return nil
}

// CommandSchema holds schema information for a CLI command.
type CommandSchema struct {
	Command     string      `json:"command"`
	Description string      `json:"description"`
	Output      *JSONSchema `json:"output"`
}

// PrintCommandSchema outputs schema with command metadata.
func PrintCommandSchema(command, description string, v any) error {
	cmdSchema := CommandSchema{
		Command:     command,
		Description: description,
		Output:      Generate(v),
	}
	output, err := json.MarshalIndent(cmdSchema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal command schema: %w", err)
	}

	fmt.Println(string(output))

	return nil
}
