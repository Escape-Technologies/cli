package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// stubFlagDescriptionRunes caps a flag description copied into a compact
// stub. The full usage string often repeats the enum list; the enum itself
// is copied separately so truncation cannot drop the allowed values.
const stubFlagDescriptionRunes = 120

// stubBodySchemaFor builds a compact placeholder body schema. The LLM can
// either populate it best-effort or call escape_get_tool_spec to fetch the
// original shape before the real tool call.
func stubBodySchemaFor(toolName string) map[string]any {
	return map[string]any{
		"type": "object",
		"description": fmt.Sprintf(
			"Request body for %s. Use escape_get_tool_spec to retrieve the full input schema.",
			toolName,
		),
		"additionalProperties": true,
	}
}

// BuildStubTool rebuilds a ToolSpec's mcp Tool with the heavy `body` property
// replaced by a small placeholder. Positional args and flag bindings are
// preserved verbatim because they are already compact and useful hints. The
// executor still uses the original spec (positional args, flags, body) at
// call time — this only affects what is advertised on tools/list.
//
// `required` is reconstructed from the original RawInputSchema so any tool
// whose full schema marks `body` (or another non-positional flag) as required
// stays required in the stub. Otherwise compact_only mode would let the model
// emit calls that the executor immediately rejects.
func BuildStubTool(spec ToolSpec) (mcpgo.Tool, error) {
	properties := map[string]any{}
	requiredSet := map[string]struct{}{}

	for _, name := range spec.PositionalArgs {
		properties[name] = map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("Positional argument %s.", name),
		}
		requiredSet[name] = struct{}{}
	}

	originalProps := originalProperties(spec)
	for _, binding := range spec.FlagBindings {
		properties[binding.Property] = flagStubSchema(binding.Kind, originalProps[binding.Property])
		if binding.Required {
			requiredSet[binding.Property] = struct{}{}
		}
	}

	if spec.BodyProperty != "" {
		properties[spec.BodyProperty] = stubBodySchemaFor(spec.Name)
	}

	// confirm is not a flag binding: the handler reads it and drops it before
	// the CLI runs. Compact stubs are the default tools/list shape, so the
	// gate has to stay visible and required there too.
	if spec.Destructive {
		properties[ConfirmProperty] = map[string]any{
			"type":        "boolean",
			"description": ConfirmPropertyDescription,
		}
		requiredSet[ConfirmProperty] = struct{}{}
	}

	// Carry forward any required properties from the original schema that
	// still exist in the stub. Names not in the stub are dropped silently
	// so we never advertise a required property the executor can't bind.
	for _, name := range originalRequired(spec) {
		if _, exists := properties[name]; exists {
			requiredSet[name] = struct{}{}
		}
	}

	required := make([]string, 0, len(requiredSet))
	for name := range requiredSet {
		required = append(required, name)
	}

	// Stable ordering keeps the marshalled stub byte-deterministic so
	// tests/snapshots don't flake on map iteration order.
	sort.Strings(required)

	stubSchema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
	rawSchema, err := json.Marshal(stubSchema)
	if err != nil {
		return mcpgo.Tool{}, fmt.Errorf("marshal stub schema for %q: %w", spec.Name, err)
	}

	stubDescription := spec.Description
	if spec.BodyProperty != "" {
		stubDescription += " (compact stub — call escape_get_tool_spec for the full input schema)"
	}

	return mcpgo.NewToolWithRawSchema(spec.Name, stubDescription, rawSchema), nil
}

// originalRequired pulls the `required` array out of the spec's full
// RawInputSchema, returning nil when the schema is missing or malformed.
// We never propagate errors: the worst-case is "stub forgot a required
// field", which is a regression of the pre-fix behaviour, not a new bug.
func originalRequired(spec ToolSpec) []string {
	if len(spec.Tool.RawInputSchema) == 0 {
		return nil
	}

	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
		return nil
	}

	return schema.Required
}

// originalProperties reads the full tool's property schemas so a stub can
// keep the flag description and enum that flagStubSchema would otherwise drop.
func originalProperties(spec ToolSpec) map[string]map[string]any {
	if len(spec.Tool.RawInputSchema) == 0 {
		return nil
	}

	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
		return nil
	}

	return schema.Properties
}

// flagStubSchema keeps the compact flag type, and copies a one-line
// description, enum values, and default from the full schema. Long usage
// strings are truncated so tools/list stays small. The enum and default are
// copied separately so the cut cannot drop allowed values or an MCP default
// such as profiles update-configuration --merge.
func flagStubSchema(kind string, original map[string]any) map[string]any {
	var schema map[string]any
	switch kind {
	case "bool":
		schema = map[string]any{"type": "boolean"}
	case "int":
		schema = map[string]any{"type": "integer"}
	case "stringSlice":
		schema = map[string]any{
			"type":  "array",
			"items": map[string]any{"type": "string"},
		}
	default:
		schema = map[string]any{"type": "string"}
	}

	if len(original) == 0 {
		return schema
	}

	if desc, _ := original["description"].(string); desc != "" {
		if compact := compactDescription(desc); compact != "" {
			schema["description"] = compact
		}
	}

	if value, ok := original["default"]; ok {
		schema["default"] = value
	}

	switch kind {
	case "stringSlice":
		items, _ := original["items"].(map[string]any)
		if items == nil {
			return schema
		}

		if enum, ok := items["enum"].([]any); ok && len(enum) > 0 {
			schema["items"].(map[string]any)["enum"] = enum
		}
	default:
		if enum, ok := original["enum"].([]any); ok && len(enum) > 0 {
			schema["enum"] = enum
		}
	}

	return schema
}

// compactDescription collapses a flag usage string to one line and trims it
// to stubFlagDescriptionRunes. Enum values are not recovered from the cut
// tail; flagStubSchema copies those from the property's enum keyword.
func compactDescription(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	return fitRunes(text, stubFlagDescriptionRunes)
}

// truncatedEllipsis is appended when fitRunes cuts a string. Reserving its
// length keeps the marker inside the rune budget.
const truncatedEllipsis = "..."

// fitRunes keeps text within maxRunes, including the ASCII ellipsis. It
// prefers a word boundary in the last 40 runes so an enum token is not split.
func fitRunes(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}

	if maxRunes <= len(truncatedEllipsis) {
		return string(runes[:maxRunes])
	}

	cut := maxRunes - len(truncatedEllipsis)
	for i := cut; i > cut-40 && i > 0; i-- {
		if runes[i] == ' ' {
			cut = i
			break
		}
	}

	if cut <= 0 {
		cut = maxRunes - len(truncatedEllipsis)
	}

	return strings.TrimRight(string(runes[:cut]), " ") + truncatedEllipsis
}
