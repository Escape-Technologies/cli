package mcp

import (
	"encoding/json"
	"fmt"
	"sort"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

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

	for _, binding := range spec.FlagBindings {
		properties[binding.Property] = flagStubSchema(binding.Kind)
	}

	if spec.BodyProperty != "" {
		properties[spec.BodyProperty] = stubBodySchemaFor(spec.Name)
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

	rawSchema, err := json.Marshal(map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	})
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

func flagStubSchema(kind string) map[string]any {
	switch kind {
	case "bool":
		return map[string]any{"type": "boolean"}
	case "int":
		return map[string]any{"type": "integer"}
	case "stringSlice":
		return map[string]any{
			"type":  "array",
			"items": map[string]any{"type": "string"},
		}
	default:
		return map[string]any{"type": "string"}
	}
}
