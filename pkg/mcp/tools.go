package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// defaultToolExecutionTimeout bounds the lifetime of a single CLI subprocess
// spawned to satisfy an MCP tool call. Hung children are killed when the
// per-request context deadline elapses.
const defaultToolExecutionTimeout = 30 * time.Second

// stringSliceFlagArgsPerValue is the number of CLI args emitted per value for
// a repeated string flag (flag name + value), e.g. `--status RUNNING`.
const stringSliceFlagArgsPerValue = 2

// FlagBinding maps an MCP tool property to a CLI flag on the underlying Cobra
// command, including the type the value must be rendered as.
type FlagBinding struct {
	Property string
	FlagName string
	Kind     string
}

// ToolSpec fully describes one CLI-backed MCP tool: its MCP-level metadata
// plus the subprocess invocation plan (command path, positional + flag
// bindings, and optional stdin body property).
type ToolSpec struct {
	Name           string
	Path           string
	Description    string
	Tool           mcpgo.Tool
	Command        []string
	PositionalArgs []string
	FlagBindings   []FlagBinding
	BodyProperty   string
	// AllowExtraArgs opts the tool into reading a free-form `args` string array
	// from the request payload and forwarding it verbatim to the subprocess.
	// Off by default so the Cobra-to-MCP mapping remains an explicit allowlist.
	AllowExtraArgs bool
}

// CommandExecutionOptions carries the shared runtime inputs the tool handlers
// need that are not part of a specific tool spec.
type CommandExecutionOptions struct {
	PublicAPIURL string
}

// RegisterCommandTools walks the supplied specs and registers one MCP handler
// per tool on the server. Each handler spawns a bounded CLI subprocess.
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

		execCtx, cancel := context.WithTimeout(ctx, defaultToolExecutionTimeout)
		defer cancel()

		result, err := ExecuteCLICommand(execCtx, ExecutionOptions{
			Command:        append(append([]string{}, spec.Command...), commandArgs...),
			DisplayCommand: append([]string{}, spec.Command...),
			Body:           body,
			Auth:           auth,
			PublicAPIURL:   options.PublicAPIURL,
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
	commandArgs := make([]string, 0, len(spec.PositionalArgs)+len(spec.FlagBindings)*stringSliceFlagArgsPerValue)

	for _, property := range spec.PositionalArgs {
		value, ok := rawArgs[property]
		if !ok {
			return nil, nil, fmt.Errorf("missing required argument %q", property)
		}

		text, err := stringifyCLIValue(value)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for %q: %w", property, err)
		}
		if strings.HasPrefix(text, "-") {
			return nil, nil, fmt.Errorf("positional %q must not start with '-'", property)
		}

		commandArgs = append(commandArgs, text)
	}

	if spec.AllowExtraArgs {
		if extraArgs, ok := rawArgs["args"]; ok {
			args, err := stringifyCLIArray(extraArgs)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid args value: %w", err)
			}
			for _, arg := range args {
				if arg == "--output" || strings.HasPrefix(arg, "--output=") {
					return nil, nil, errors.New(`args must not override the injected "--output" flag`)
				}
			}
			commandArgs = append(commandArgs, args...)
		}
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
			return nil, errors.New("expected boolean")
		}
		return []string{fmt.Sprintf("--%s=%t", binding.FlagName, booleanValue)}, nil
	case "stringSlice":
		values, err := stringifyCLIArray(value)
		if err != nil {
			return nil, err
		}

		args := make([]string, 0, len(values)*stringSliceFlagArgsPerValue)
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
