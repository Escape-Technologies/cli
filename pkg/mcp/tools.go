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

// ReasoningToolExecutionTimeout extends the default MCP subprocess budget for
// scans reasoning and scans coverage, which paginate large assessments.
const ReasoningToolExecutionTimeout = 2 * time.Minute //nolint:mnd

// DefaultListPageSize is the page size injected when an MCP list tool is called
// without size or cursor. The CLI still fetches every page unless those flags are set.
const DefaultListPageSize = 50

// ListPageToolDescription extends a command summary with the one-page contract
// agents need: the JSON shape, the default page size, and how to continue.
func ListPageToolDescription(short string) string {
	short = strings.TrimSpace(short)
	if short != "" && !strings.HasSuffix(short, ".") {
		short += "."
	}

	return fmt.Sprintf(
		"%s Returns one page as {items, nextCursor, totalCount}. If size and cursor are omitted, size defaults to %d. Pass cursor from nextCursor for the next page, or a smaller size or narrower filters when the page is too large.",
		short,
		DefaultListPageSize,
	)
}

// stringSliceFlagArgsPerValue is the number of CLI args emitted per value for
// a repeated string flag (flag name + value), e.g. `--status RUNNING`.
const stringSliceFlagArgsPerValue = 2

// FlagBinding maps an MCP tool property to a CLI flag on the underlying Cobra
// command, including the type the value must be rendered as.
//
// Default, when non-nil, is sent when the caller omits the property. A nil
// Default leaves the flag off the command line so the CLI's own default
// applies. buildMCPTool copies a JSON Schema default onto this field, so an
// omitted argument follows the schema (mcp-default) rather than the CLI flag.
//
// Required means an omitted or empty value is refused before the subprocess
// starts. The CLI flag can stay optional. The catalog sets this when omitting
// the flag would widen the operation to every resource.
type FlagBinding struct {
	Property string
	FlagName string
	Kind     string
	Default  any
	Required bool
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
	// ExecutionTimeout overrides defaultToolExecutionTimeout when set.
	ExecutionTimeout time.Duration
	// DefaultArgs are flag values used when the caller omitted them.
	// List tools set size to DefaultListPageSize. That default is not applied
	// when the caller already passed size or cursor: either one means they chose the page.
	DefaultArgs map[string]any
	// ForcedArgs are appended after caller arguments. Cobra keeps the last
	// value of a repeated flag, so this is how the catalog caps emails wait
	// under the tool budget no matter what the model sends.
	ForcedArgs []string
	// Destructive tools delete or bulk-mutate resources. The handler refuses
	// the call unless the request sets ConfirmProperty to true. That property
	// is MCP-only and is never forwarded as a CLI flag.
	Destructive bool
}

// EmailsWaitTimeout is the longest `emails wait` an MCP tool call will run.
// It stays below defaultToolExecutionTimeout so the CLI can return a pending
// result before the subprocess is killed.
const EmailsWaitTimeout = 25 * time.Second //nolint:mnd

// ConfirmProperty is the MCP argument that authorizes a destructive tool call.
// Callers must set it to true only after a person has confirmed the action.
const ConfirmProperty = "confirm"

// ConfirmPropertyDescription is the schema text shown for ConfirmProperty.
const ConfirmPropertyDescription = "Set to true only after the user explicitly confirmed this destructive action"

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
		rawArgs := request.GetArguments()
		if err := guardDestructive(spec, rawArgs); err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		auth, err := AuthFromContext(ctx)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		commandArgs, body, err := buildCommandArgs(spec, rawArgs)
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		timeout := defaultToolExecutionTimeout
		if spec.ExecutionTimeout > 0 {
			timeout = spec.ExecutionTimeout
		}

		execCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		result, err := ExecuteCLICommand(execCtx, ExecutionOptions{
			Command:        append(append([]string{}, spec.Command...), commandArgs...),
			DisplayCommand: append([]string{}, spec.Command...),
			Body:           body,
			Auth:           auth,
			PublicAPIURL:   options.PublicAPIURL,
			Timeout:        timeout,
		})
		// A deadline already names the poll tool. A truncated buffer on the
		// killed child must not replace that instruction with a retry hint.
		var timedOut *executionTimeoutError
		if errors.As(err, &timedOut) {
			return mcpgo.NewToolResultError(timedOut.Error()), nil
		}

		// A capped stdout is not JSON. Surface that before the exit-code error,
		// which would otherwise forward the cut bytes as text.
		if result != nil && result.StdoutTruncated {
			return truncatedToolResult(result, err), nil
		}

		if err != nil {
			return mcpgo.NewToolResultError(commandFailureText(err, result)), nil
		}

		if result.Payload != nil {
			return mcpgo.NewToolResultStructured(wrapStructuredPayload(result.Payload), result.Stdout), nil
		}

		return mcpgo.NewToolResultText(result.Stdout), nil
	}
}

// wrapStructuredPayload guarantees the value handed to the MCP client as
// structuredContent is a JSON object. The MCP TypeScript SDK validates
// structuredContent with a Zod object schema, so returning a top-level array
// (e.g. from `* list` commands) trips clients with "Expected object, received
// array". Wrap arrays in {items: [...]} and primitives in {value: ...} while
// passing real objects through untouched.
func wrapStructuredPayload(payload any) any {
	switch typed := payload.(type) {
	case map[string]any:
		return typed
	case []any:
		return map[string]any{"items": typed}
	default:
		return map[string]any{"value": typed}
	}
}

// guardDestructive refuses a destructive tool call until confirm is exactly true.
// The error names the resource and the arguments that select it so the caller
// can show that to the user before retrying.
func guardDestructive(spec ToolSpec, rawArgs map[string]any) error {
	if !spec.Destructive {
		return nil
	}

	if confirmed, ok := rawArgs[ConfirmProperty].(bool); ok && confirmed {
		return nil
	}

	return fmt.Errorf(
		"refusing destructive tool %q: %s. %s",
		spec.Name,
		destructiveEffect(spec, rawArgs),
		ConfirmPropertyDescription,
	)
}

func destructiveEffect(spec ToolSpec, rawArgs map[string]any) string {
	action := destructiveAction(spec.Command)
	target := destructiveTarget(spec, rawArgs)
	if target == "" {
		return action + " (no target arguments were provided)"
	}

	if strings.Contains(action, "matching") {
		return action + " " + target
	}

	return action + " identified by " + target
}

func destructiveAction(command []string) string {
	if len(command) == 0 {
		return "this will permanently delete the resource"
	}

	switch command[len(command)-1] {
	case "bulk-update":
		return "this will update every asset matching"
	case "bulk-delete":
		return "this will delete every asset matching"
	case "delete":
		return "this will permanently delete the " + singularResource(command[0])
	default:
		return "this will permanently change the resource"
	}
}

func singularResource(name string) string {
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.TrimSuffix(name, "s")
	if name == "" {
		return "resource"
	}

	return name
}

func destructiveTarget(spec ToolSpec, rawArgs map[string]any) string {
	parts := make([]string, 0, len(spec.PositionalArgs)+len(spec.FlagBindings))
	seen := make(map[string]struct{}, len(spec.PositionalArgs)+len(spec.FlagBindings))
	add := func(name string, value any) {
		if name == ConfirmProperty {
			return
		}

		if _, ok := seen[name]; ok {
			return
		}

		rendered, ok := renderDestructiveArg(value)
		if !ok {
			return
		}

		seen[name] = struct{}{}
		parts = append(parts, name+"="+rendered)
	}
	for _, name := range spec.PositionalArgs {
		if value, ok := rawArgs[name]; ok {
			add(name, value)
		}
	}

	for _, binding := range spec.FlagBindings {
		if value, ok := rawArgs[binding.Property]; ok {
			add(binding.Property, value)
		}
	}

	return strings.Join(parts, ", ")
}

func renderDestructiveArg(value any) (string, bool) {
	switch value.(type) {
	case []any, []string:
		values, err := stringifyCLIArray(value)
		if err != nil || len(values) == 0 {
			return "", false
		}

		return strings.Join(values, ","), true
	default:
		text, err := stringifyCLIValue(value)
		if err != nil || text == "" {
			return "", false
		}

		return text, true
	}
}

func buildCommandArgs(spec ToolSpec, rawArgs map[string]any) ([]string, any, error) {
	rawArgs = applyDefaultArgs(spec, rawArgs)
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
		// ConfirmProperty is the MCP gate, not a CLI flag, even if a binding
		// was registered under that name by mistake.
		if binding.Property == ConfirmProperty || binding.FlagName == ConfirmProperty {
			continue
		}

		value, ok := rawArgs[binding.Property]
		if !ok {
			if binding.Default == nil {
				if binding.Required {
					return nil, nil, fmt.Errorf("missing required argument %q", binding.Property)
				}

				continue
			}

			value = binding.Default
		}

		if binding.Required && requiredArgEmpty(value) {
			return nil, nil, fmt.Errorf("%q must not be empty", binding.Property)
		}

		flagArgs, err := buildFlagArgs(binding, value)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid value for flag %q: %w", binding.Property, err)
		}

		commandArgs = append(commandArgs, flagArgs...)
	}

	// Last so a catalog cap overrides a caller-supplied flag of the same name.
	commandArgs = append(commandArgs, spec.ForcedArgs...)

	var body any
	if spec.BodyProperty != "" {
		body = rawArgs[spec.BodyProperty]
	}

	return commandArgs, body, nil
}

// applyDefaultArgs fills omitted flag values. A size default is skipped when
// cursor is already set, because that cursor selects the page on its own.
// The caller's map is left unchanged.
func applyDefaultArgs(spec ToolSpec, rawArgs map[string]any) map[string]any {
	if len(spec.DefaultArgs) == 0 {
		return rawArgs
	}

	args := make(map[string]any, len(rawArgs)+len(spec.DefaultArgs))
	for key, value := range rawArgs {
		args[key] = value
	}

	_, hasCursor := args["cursor"]
	for key, value := range spec.DefaultArgs {
		if _, present := args[key]; present {
			continue
		}

		if key == "size" && hasCursor {
			continue
		}

		args[key] = value
	}

	return args
}

func truncatedToolResult(result *ExecutionResult, runErr error) *mcpgo.CallToolResult {
	payload := map[string]any{
		"truncated": true,
		"error":     result.Stdout,
	}
	if existing, ok := result.Payload.(map[string]any); ok {
		for key, value := range existing {
			payload[key] = value
		}
	}

	text := result.Stdout
	// A failed command explains itself on stderr. Without it the caller only
	// sees the retry hint, even when the cause was an auth or validation error.
	if runErr != nil {
		payload["exitError"] = runErr.Error()
		if result.Stderr != "" {
			payload["stderr"] = result.Stderr
		}

		text = commandFailureText(runErr, result)
	}

	toolResult := mcpgo.NewToolResultStructured(payload, text)
	toolResult.IsError = true

	return toolResult
}

// requiredArgEmpty is true when a required flag would not select any resource.
func requiredArgEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return typed == ""
	case []string:
		return len(typed) == 0
	case []any:
		return len(typed) == 0
	default:
		return false
	}
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
	// A deadline already names the poll tool. Appending "signal: killed" makes
	// models retry the same call instead of polling.
	var timedOut *executionTimeoutError
	if errors.As(err, &timedOut) {
		return timedOut.Error()
	}

	lines := []string{err.Error()}

	if result != nil && result.Stderr != "" {
		lines = append(lines, result.Stderr)
	}

	if result != nil && result.Stdout != "" {
		lines = append(lines, result.Stdout)
	}

	return strings.Join(lines, "\n\n")
}
