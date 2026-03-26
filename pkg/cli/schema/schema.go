// Package schema provides JSON Schema generation for CLI output types
package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// JSONSchema represents a JSON Schema document
type JSONSchema struct {
	Schema      string                 `json:"$schema,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Nullable    bool                   `json:"nullable,omitempty"`
}

// Generate creates a JSON Schema from a Go type
func Generate(v any) *JSONSchema {
	t := reflect.TypeOf(v)
	return generateSchema(t, true)
}

func generateSchema(t reflect.Type, root bool) *JSONSchema {
	schema := &JSONSchema{}

	if root {
		schema.Schema = "https://json-schema.org/draft/2020-12/schema"
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		inner := generateSchema(t.Elem(), false)
		inner.Nullable = true
		return inner
	}

	// Handle slices/arrays
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		schema.Type = "array"
		schema.Items = generateSchema(t.Elem(), false)
		return schema
	}

	// Handle maps
	if t.Kind() == reflect.Map {
		schema.Type = "object"
		return schema
	}

	// Handle basic types
	switch t.Kind() {
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Struct:
		schema.Type = "object"
		schema.Properties = make(map[string]*JSONSchema)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Get JSON tag
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			// Parse JSON tag
			tagParts := strings.Split(jsonTag, ",")
			fieldName := tagParts[0]
			if fieldName == "" {
				fieldName = field.Name
			}

			// Skip AdditionalProperties field (internal use)
			if fieldName == "AdditionalProperties" || field.Name == "AdditionalProperties" {
				continue
			}

			isOptional := false
			for _, part := range tagParts[1:] {
				if part == "omitempty" {
					isOptional = true
					break
				}
			}

			propSchema := generateSchema(field.Type, false)

			// Add description from struct tag if available
			if desc := field.Tag.Get("description"); desc != "" {
				propSchema.Description = desc
			}

			schema.Properties[fieldName] = propSchema

			if !isOptional && field.Type.Kind() != reflect.Ptr {
				schema.Required = append(schema.Required, fieldName)
			}
		}
	case reflect.Interface:
		// For interface{}, we can't determine the type
		schema.Type = "object"
	}

	return schema
}

// Print outputs the JSON Schema for the given type
func Print(v any) error {
	schema := Generate(v)
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// CommandSchema holds schema information for a CLI command
type CommandSchema struct {
	Command     string      `json:"command"`
	Description string      `json:"description"`
	Output      *JSONSchema `json:"output"`
}

// PrintCommandSchema outputs schema with command metadata
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
