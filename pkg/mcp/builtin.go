package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// GetToolSpecToolName is the built-in tool that returns the full input schema
// of a compressed/stubbed tool on demand. Intent-aware tools/list serves
// minimal stubs for non-selected tools; the LLM uses this tool to zoom in
// when it decides to call one of them.
const GetToolSpecToolName = "escape_get_tool_spec"

// listCapabilitiesToolName is the built-in tool that ranks the CLI catalog.
const listCapabilitiesToolName = "list_capabilities"

// BuiltinToolNames are the helper tools RegisterBuiltinTools adds beside the
// CLI catalog. Guidance may name these; it may not invent others.
func BuiltinToolNames() []string {
	return []string{GetToolSpecToolName, listCapabilitiesToolName}
}

// defaultCapabilitiesLimit caps the number of tools listed by the built-in
// list_capabilities tool when the caller omits or passes a non-positive limit.
const defaultCapabilitiesLimit = 25

// RegisterBuiltinTools registers the MCP-native helper tools (e.g. the tool
// discovery endpoint) that ship with the embedded server.
func RegisterBuiltinTools(server *mcpserver.MCPServer, specs []ToolSpec) {
	registerGetToolSpec(server, specs)
	tool := mcpgo.NewTool(
		listCapabilitiesToolName,
		mcpgo.WithDescription("List the available Escape CLI-backed MCP tools."),
		mcpgo.WithString("objective", mcpgo.Description("Optional search intent used to rank relevant tools.")),
		mcpgo.WithNumber("limit", mcpgo.Description("Maximum number of tools to return.")),
		mcpgo.WithReadOnlyHintAnnotation(true),
		mcpgo.WithDestructiveHintAnnotation(false),
		mcpgo.WithIdempotentHintAnnotation(true),
		mcpgo.WithOpenWorldHintAnnotation(false),
	)

	server.AddTool(tool, buildListCapabilitiesHandler(specs))
}

func buildListCapabilitiesHandler(specs []ToolSpec) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := AuthFromContext(ctx); err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		objective := strings.TrimSpace(strings.ToLower(request.GetString("objective", "")))
		limit := int(request.GetFloat("limit", defaultCapabilitiesLimit))
		if limit <= 0 {
			limit = defaultCapabilitiesLimit
		}

		items := make([]map[string]any, 0, len(specs))
		for _, spec := range specs {
			score := capabilityScore(spec, objective)
			items = append(items, map[string]any{
				"name":        spec.Name,
				"path":        spec.Path,
				"description": spec.Description,
				"score":       score,
			})
		}

		slices.SortFunc(items, func(left, right map[string]any) int {
			leftScore := left["score"].(int)
			rightScore := right["score"].(int)
			if leftScore != rightScore {
				return rightScore - leftScore
			}

			return strings.Compare(left["name"].(string), right["name"].(string))
		})

		if limit < len(items) {
			items = items[:limit]
		}

		lines := make([]string, 0, len(items)+1)
		lines = append(lines, fmt.Sprintf("Available tools: %d", len(items)))
		for _, item := range items {
			lines = append(lines, fmt.Sprintf("- %s: %s", item["name"], item["description"]))
		}

		return mcpgo.NewToolResultStructured(wrapStructuredPayload(items), strings.Join(lines, "\n")), nil
	}
}

func registerGetToolSpec(server *mcpserver.MCPServer, specs []ToolSpec) {
	tool := mcpgo.NewTool(
		GetToolSpecToolName,
		mcpgo.WithDescription(
			"Fetch the full input schema for a tool by name. Useful after tools/list returned a compact stub for the tool you want to call.",
		),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Exact tool name returned by tools/list.")),
		mcpgo.WithReadOnlyHintAnnotation(true),
		mcpgo.WithDestructiveHintAnnotation(false),
		mcpgo.WithIdempotentHintAnnotation(true),
		mcpgo.WithOpenWorldHintAnnotation(false),
	)

	byName := make(map[string]ToolSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}

	server.AddTool(tool, func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := AuthFromContext(ctx); err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		name := strings.TrimSpace(request.GetString("name", ""))
		if name == "" {
			return mcpgo.NewToolResultError("missing required argument \"name\""), nil
		}

		spec, ok := byName[name]
		if !ok {
			return mcpgo.NewToolResultError(fmt.Sprintf("unknown tool %q", name)), nil
		}

		// spec.Tool carries the original full RawInputSchema built by buildMCPTool.
		result, err := toolSpecResult(spec)
		if err != nil {
			return mcpgo.NewToolResultError(fmt.Sprintf("spec for %q: %v", name, err)), nil
		}

		return result, nil
	})
}

// toolSpecResult is the escape_get_tool_spec payload. The schema is one compact
// JSON document in the text content. Structured content would repeat that
// document, and indentation roughly doubles a body that is already large.
func toolSpecResult(spec ToolSpec) (*mcpgo.CallToolResult, error) {
	var schema any
	if len(spec.Tool.RawInputSchema) > 0 {
		if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
			return nil, fmt.Errorf("decode schema: %w", err)
		}
	} else {
		schema = map[string]any{}
	}

	payload := map[string]any{
		"name":        spec.Name,
		"description": spec.Description,
		"inputSchema": schema,
	}
	text, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode tool spec: %w", err)
	}

	return mcpgo.NewToolResultText(string(text)), nil
}

// capabilitySynonyms bridges user vocabulary and the CLI command corpus
// (name + path + Short) used by list_capabilities ranking. Plain substring
// matching misses cross-concept terms: "findings" never appears in an
// issues_* descriptor. Keys are stemmed singular terms; values are stemmed
// corpus terms. Keep this small — it is a ranking aid, not a thesaurus.
var capabilitySynonyms = map[string][]string{
	"finding":       {"issue"},
	"vulnerability": {"issue", "severity"},
	"bug":           {"issue"},
	"domain":        {"asset"},
	"host":          {"asset"},
	"app":           {"profile"},
	"application":   {"profile"},
	"pentest":       {"scan"},
	"pentesting":    {"pentest", "scan"},
	"exploit":       {"pentest", "scan"},
	"exploitation":  {"pentest", "scan"},
	"target":        {"asset", "profile"},
	"remediation":   {"issue"},
	"ticket":        {"integration"},
	"inventory":     {"asset"},
}

func capabilityScore(spec ToolSpec, objective string) int {
	if objective == "" {
		return 0
	}

	corpus := strings.ToLower(spec.Name + " " + spec.Path + " " + spec.Description)
	score := 0
	for _, token := range strings.Fields(objective) {
		if capabilityTokenMatches(corpus, token) {
			score++
		}
	}

	return score
}

// capabilityTokenMatches reports whether the corpus mentions the token
// directly or under a synonym of its (stemmed) form.
func capabilityTokenMatches(corpus, token string) bool {
	forms := []string{token, StemToken(token)}
	for _, form := range forms {
		if strings.Contains(corpus, form) {
			return true
		}
	}

	for _, form := range forms {
		for _, synonym := range capabilitySynonyms[form] {
			if strings.Contains(corpus, synonym) {
				return true
			}
		}
	}

	return false
}
