package mcp

import (
	"encoding/json"
	"fmt"

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
func BuildStubTool(spec ToolSpec) (mcpgo.Tool, error) {
	properties := map[string]any{}
	required := []string{}

	for _, name := range spec.PositionalArgs {
		properties[name] = map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("Positional argument %s.", name),
		}
		required = append(required, name)
	}

	for _, binding := range spec.FlagBindings {
		properties[binding.Property] = flagStubSchema(binding.Kind)
	}

	if spec.BodyProperty != "" {
		properties[spec.BodyProperty] = stubBodySchemaFor(spec.Name)
	}

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

	stub := mcpgo.NewToolWithRawSchema(spec.Name, stubDescription, rawSchema)
	// Preserve safety annotations so the LLM trusts the stub the same way it
	// would the full tool. Without this the stub serializes Annotations as
	// `{}` and reasoning models become reluctant to call it.
	stub.Annotations = spec.Tool.Annotations
	return stub, nil
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
