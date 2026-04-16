package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type FlagBinding struct {
	Property string
	FlagName string
	Kind     string
}

type ToolSpec struct {
	Name           string
	Path           string
	Description    string
	Tool           mcpgo.Tool
	Command        []string
	PositionalArgs []string
	FlagBindings   []FlagBinding
	BodyProperty   string
}

type CommandExecutionOptions struct {
	PublicAPIURL string
}

func RegisterCommandTools(
	server *mcpserver.MCPServer,
	specs []ToolSpec,
	options CommandExecutionOptions,
) {
	for _, spec := range specs {
		server.AddTool(spec.Tool, buildToolHandler(spec, options))
	}
}

func buildToolHandler(
	spec ToolSpec,
	options CommandExecutionOptions,
) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		auth, err := AuthFromContext(ctx)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		commandArgs, body, err := buildCommandArgs(spec, request.GetArguments())
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		result, err := ExecuteCLICommand(ctx, ExecutionOptions{
			Command:      append(append([]string{}, spec.Command...), commandArgs...),
			Body:         body,
			Auth:         auth,
			PublicAPIURL: options.PublicAPIURL,
		})
		if err != nil {
			return mcpgo.NewToolResultError(commandFailureText(err, result)), nil
		}

		if result.Payload != nil {
			return mcpgo.NewToolResultStructured(result.Payload, result.Stdout), nil
		}

		return mcpgo.NewToolResultText(result.Stdout), nil
	}
}

func buildCommandArgs(spec ToolSpec, rawArgs map[string]any) ([]string, any, error) {
	commandArgs := make([]string, 0, len(spec.PositionalArgs)+len(spec.FlagBindings)*2)

	for _, property := range spec.PositionalArgs {
		value, ok := rawArgs[property]
		if !ok {
			return nil, nil, fmt.Errorf("missing required argument %q", property)
		}

		text, err := stringifyCLIValue(value)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for %q: %w", property, err)
		}

		commandArgs = append(commandArgs, text)
	}

	if extraArgs, ok := rawArgs["args"]; ok {
		args, err := stringifyCLIArray(extraArgs)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid args value: %w", err)
		}
		commandArgs = append(commandArgs, args...)
	}

	for _, binding := range spec.FlagBindings {
		value, ok := rawArgs[binding.Property]
		if !ok {
			continue
		}

		flagArgs, err := buildFlagArgs(binding, value)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for flag %q: %w", binding.Property, err)
		}
		commandArgs = append(commandArgs, flagArgs...)
	}

	var body any
	if spec.BodyProperty != "" {
		body = rawArgs[spec.BodyProperty]
	}

	return commandArgs, body, nil
}

func buildFlagArgs(binding FlagBinding, value any) ([]string, error) {
	switch binding.Kind {
	case "bool":
		booleanValue, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("expected boolean")
		}
		return []string{fmt.Sprintf("--%s=%t", binding.FlagName, booleanValue)}, nil
	case "stringSlice":
		values, err := stringifyCLIArray(value)
		if err != nil {
			return nil, err
		}

		args := make([]string, 0, len(values)*2)
		for _, item := range values {
			args = append(args, "--"+binding.FlagName, item)
		}
		return args, nil
	default:
		text, err := stringifyCLIValue(value)
		if err != nil {
			return nil, err
		}
		return []string{"--" + binding.FlagName, text}, nil
	}
}

func stringifyCLIArray(value any) ([]string, error) {
	switch typed := value.(type) {
	case []string:
		return typed, nil
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			text, err := stringifyCLIValue(item)
			if err != nil {
				return nil, err
			}
			values = append(values, text)
		}
		return values, nil
	default:
		text, err := stringifyCLIValue(value)
		if err != nil {
			return nil, err
		}
		return []string{text}, nil
	}
}

func stringifyCLIValue(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10), nil
		}
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case float32:
		value64 := float64(typed)
		if value64 == float64(int64(value64)) {
			return strconv.FormatInt(int64(value64), 10), nil
		}
		return strconv.FormatFloat(value64, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(typed), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case json.Number:
		return typed.String(), nil
	default:
		return "", fmt.Errorf("unsupported value type %T", value)
	}
}

func commandFailureText(err error, result *ExecutionResult) string {
	lines := []string{err.Error()}

	if result != nil && result.Stderr != "" {
		lines = append(lines, result.Stderr)
	}

	if result != nil && result.Stdout != "" {
		lines = append(lines, result.Stdout)
	}

	return strings.Join(lines, "\n\n")
}
