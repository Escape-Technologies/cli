package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/env"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/Escape-Technologies/cli/pkg/version"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const defaultMCPServePort = 8080

var mcpServePort int
var mcpServePublicAPIURL string

var mcpCmd = &cobra.Command{
	Use:    "mcp",
	Short:  "MCP server commands",
	Hidden: true,
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the embedded MCP server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		toolSpecs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
		if err != nil {
			return fmt.Errorf("failed to build MCP tool catalog: %w", err)
		}

		mode := climcp.ModeFromEnv(climcp.IntentModeCompactOnly)
		var classifier climcp.Classifier
		if mode == climcp.IntentModeOn {
			classifier, err = climcp.NewClassifierFromEnv()
			if err != nil {
				return fmt.Errorf("failed to configure MCP classifier: %w", err)
			}
		}

		server := climcp.NewServer(climcp.ServerOptions{
			Version:      version.GetVersion().DisplayVersion(),
			Port:         mcpServePort,
			PublicAPIURL: resolveMCPPublicAPIURL(),
			Tools:        toolSpecs,
			IntentMode:   mode,
			Classifier:   classifier,
		})

		return server.Serve(cmd.Context())
	},
}

func init() {
	mcpServeCmd.Flags().IntVar(&mcpServePort, "port", defaultMCPServePort, "port to listen on")
	mcpServeCmd.Flags().StringVar(
		&mcpServePublicAPIURL,
		"public-api-url",
		"",
		"public API base URL (defaults to ESCAPE_API_URL)",
	)
	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(mcpCmd)
}

func buildMCPToolSpecs(
	root *cobra.Command,
	registry map[string]CommandSchemas,
) ([]climcp.ToolSpec, error) {
	capabilities := BuildCommandCapabilities(root, registry)
	commands := indexCommands(root)
	toolSpecs := make([]climcp.ToolSpec, 0, len(capabilities))

	for _, capability := range capabilities {
		if capability.HasSub || skipMCPCommand(capability.Path) {
			continue
		}
		if _, allowed := registry[capability.Path]; !allowed {
			continue
		}

		command := commands[capability.Path]
		if command == nil {
			return nil, fmt.Errorf("missing cobra command for %q", capability.Path)
		}

		tool, flagBindings, positionalArgs, bodyProperty, err := buildMCPTool(capability, command)
		if err != nil {
			return nil, err
		}

		toolSpecs = append(toolSpecs, climcp.ToolSpec{
			Name:           tool.Name,
			Path:           capability.Path,
			Description:    capability.Short,
			Tool:           tool,
			Command:        strings.Fields(capability.Path)[1:],
			PositionalArgs: positionalArgs,
			FlagBindings:   flagBindings,
			BodyProperty:   bodyProperty,
		})
	}

	return toolSpecs, nil
}

func buildMCPTool(
	capability CommandCapability,
	command *cobra.Command,
) (mcpgo.Tool, []climcp.FlagBinding, []string, string, error) {
	properties := map[string]any{}
	required := []string{}
	flagBindings := make([]climcp.FlagBinding, 0)
	bodyProperty := ""

	positionalArgs := parsePositionalArgs(capability.Use, capability.HasInSchema)
	for _, name := range positionalArgs {
		properties[name] = map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("Positional argument %s.", name),
		}
		required = append(required, name)
	}

	command.Flags().VisitAll(func(flag *pflag.Flag) {
		if skipMCPFlag(flag.Name) {
			return
		}

		property := normalizePropertyName(flag.Name)
		if _, exists := properties[property]; exists {
			property = "flag_" + property
		}

		schema, kind := schemaForFlag(flag)
		properties[property] = schema
		flagBindings = append(flagBindings, climcp.FlagBinding{
			Property: property,
			FlagName: flag.Name,
			Kind:     kind,
		})
	})

	if capability.InputSchema != nil {
		bodyProperty = "body"
		properties[bodyProperty] = capability.InputSchema
	}

	inputSchema, err := json.Marshal(map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	})
	if err != nil {
		return mcpgo.Tool{}, nil, nil, "", fmt.Errorf("failed to marshal input schema for %q: %w", capability.Path, err)
	}

	// Use NewToolWithRawSchema so the default structured InputSchema (Type:"object")
	// is not set alongside RawInputSchema — the library refuses to marshal a tool
	// that has both set (errToolSchemaConflict).
	tool := mcpgo.NewToolWithRawSchema(buildMCPToolName(capability.Path), capability.Short, inputSchema)
	// Set explicit safety annotations so reasoning models (KIMIK2THINKING etc.)
	// know which tools are safe to call without hesitation. Without explicit
	// hints the LLM may decline to call CLI tools and produce empty answers.
	tool.Annotations = annotationsForCapability(capability)
	// MCP's tool outputSchema must have top-level type "object" (the client
	// validates this via Zod). Several CLI commands (`* list`) return arrays,
	// so skip declaring the schema in those cases — the payload is still
	// returned via structuredContent/text, just without an advertised shape.
	if capability.OutputSchema != nil && capability.OutputSchema.Type == "object" {
		outputSchema, err := json.Marshal(capability.OutputSchema)
		if err != nil {
			return mcpgo.Tool{}, nil, nil, "", fmt.Errorf("failed to marshal output schema for %q: %w", capability.Path, err)
		}
		tool.RawOutputSchema = outputSchema
	}

	return tool, flagBindings, positionalArgs, bodyProperty, nil
}

// annotationsForCapability sets MCP per-tool hints based on the CLI command
// name. mark3labs/mcp-go's NewToolWithRawSchema leaves Annotations as the zero
// value, which serializes to `{}` — clients then assume "unknown" and some
// reasoning models become reluctant to call. Be explicit instead.
func annotationsForCapability(capability CommandCapability) mcpgo.ToolAnnotation {
	leaf := lastSegment(capability.Path)
	switch {
	case strings.HasPrefix(leaf, "list") || leaf == "list" ||
		strings.HasPrefix(leaf, "get") || leaf == "get" ||
		strings.HasPrefix(leaf, "search") || strings.HasPrefix(leaf, "show") ||
		strings.HasPrefix(leaf, "describe") || strings.HasPrefix(leaf, "fetch") ||
		strings.HasPrefix(leaf, "test") || strings.HasPrefix(leaf, "view") ||
		strings.HasPrefix(leaf, "tail") || strings.HasPrefix(leaf, "logs"):
		return mcpgo.ToolAnnotation{
			ReadOnlyHint:    mcpgo.ToBoolPtr(true),
			DestructiveHint: mcpgo.ToBoolPtr(false),
			IdempotentHint:  mcpgo.ToBoolPtr(true),
			OpenWorldHint:   mcpgo.ToBoolPtr(true),
		}
	case strings.HasPrefix(leaf, "delete") || strings.HasPrefix(leaf, "remove") ||
		strings.HasPrefix(leaf, "destroy") || strings.HasPrefix(leaf, "purge") ||
		strings.HasPrefix(leaf, "cancel"):
		return mcpgo.ToolAnnotation{
			ReadOnlyHint:    mcpgo.ToBoolPtr(false),
			DestructiveHint: mcpgo.ToBoolPtr(true),
			IdempotentHint:  mcpgo.ToBoolPtr(true),
			OpenWorldHint:   mcpgo.ToBoolPtr(true),
		}
	default:
		// create / update / start / etc. — non-destructive but mutating.
		return mcpgo.ToolAnnotation{
			ReadOnlyHint:    mcpgo.ToBoolPtr(false),
			DestructiveHint: mcpgo.ToBoolPtr(false),
			IdempotentHint:  mcpgo.ToBoolPtr(false),
			OpenWorldHint:   mcpgo.ToBoolPtr(true),
		}
	}
}

func lastSegment(commandPath string) string {
	fields := strings.Fields(commandPath)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

func indexCommands(root *cobra.Command) map[string]*cobra.Command {
	commands := map[string]*cobra.Command{}

	var walk func(command *cobra.Command)
	walk = func(command *cobra.Command) {
		if !command.IsAvailableCommand() || command.Hidden {
			return
		}
		commands[command.CommandPath()] = command
		for _, child := range command.Commands() {
			walk(child)
		}
	}

	walk(root)
	return commands
}

func skipMCPCommand(path string) bool {
	name := strings.TrimPrefix(path, "escape-cli ")
	return name == "help" ||
		name == "help-all" ||
		name == "completion" ||
		name == "capabilities" ||
		name == "version" ||
		name == "mcp"
}

func skipMCPFlag(name string) bool {
	return name == "verbose" || name == "output" || name == "input-schema" || name == "help"
}

func buildMCPToolName(path string) string {
	name := strings.TrimPrefix(path, "escape-cli ")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func parsePositionalArgs(use string, hasBody bool) []string {
	parts := strings.Fields(use)
	if len(parts) <= 1 {
		return nil
	}

	names := make([]string, 0, len(parts)-1)
	for _, part := range parts[1:] {
		candidate := strings.Trim(part, "<>[]")
		candidate = strings.TrimSuffix(candidate, "...")
		if candidate == "" || candidate == "flags" {
			continue
		}
		if hasBody && strings.Contains(candidate, ".json") {
			continue
		}

		names = append(names, normalizePropertyName(candidate))
	}

	return names
}

func normalizePropertyName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, ".", "_")
	return value
}

func schemaForFlag(flag *pflag.Flag) (map[string]any, string) {
	description := flag.Usage
	switch flag.Value.Type() {
	case "bool":
		return map[string]any{
			"type":        "boolean",
			"description": description,
		}, "bool"
	case "int", "int64":
		return map[string]any{
			"type":        "integer",
			"description": description,
		}, "int"
	case "stringSlice", "stringArray":
		return map[string]any{
			"type":        "array",
			"description": description,
			"items": map[string]any{
				"type": "string",
			},
		}, "stringSlice"
	default:
		return map[string]any{
			"type":        "string",
			"description": description,
		}, "string"
	}
}

func resolveMCPPublicAPIURL() string {
	if mcpServePublicAPIURL != "" {
		return mcpServePublicAPIURL
	}

	apiURL, err := env.GetAPIURL()
	if err != nil {
		return ""
	}

	return apiURL.String()
}
