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

		server := climcp.NewServer(climcp.ServerOptions{
			Version:      version.GetVersion().DisplayVersion(),
			Port:         mcpServePort,
			PublicAPIURL: resolveMCPPublicAPIURL(),
			Tools:        toolSpecs,
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

	properties["args"] = map[string]any{
		"type":        "array",
		"description": "Optional extra positional CLI arguments appended after the named ones.",
		"items": map[string]any{
			"type": "string",
		},
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

	options := []mcpgo.ToolOption{
		mcpgo.WithDescription(capability.Short),
		mcpgo.WithRawInputSchema(inputSchema),
	}
	if capability.OutputSchema != nil {
		outputSchema, err := json.Marshal(capability.OutputSchema)
		if err != nil {
			return mcpgo.Tool{}, nil, nil, "", fmt.Errorf("failed to marshal output schema for %q: %w", capability.Path, err)
		}
		options = append(options, mcpgo.WithRawOutputSchema(outputSchema))
	}

	return mcpgo.NewTool(buildMCPToolName(capability.Path), options...), flagBindings, positionalArgs, bodyProperty, nil
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
		name == "version"
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
