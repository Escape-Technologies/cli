package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	publicAPIToolName = "public_api_answer_question"

	// publicAPIDefaultBaseURL is the production root for the v3 public API.
	// Used when neither ServerOptions.PublicAPIURL nor servers[0].url from the
	// fetched spec are usable.
	publicAPIDefaultBaseURL = "https://public.escape.tech/v3"

	publicAPIDefaultLimit = 1
	publicAPIMaxLimit     = 3
)

// PublicAPIOptions configures the public-API question tool. Zero values fall
// back to environment variables and then production defaults.
type PublicAPIOptions struct {
	SpecURL      string
	PublicAPIURL string
	TTL          time.Duration
}

// RegisterPublicAPITools registers the public_api_answer_question tool on the
// given server.
func RegisterPublicAPITools(server *mcpserver.MCPServer, opts PublicAPIOptions) error {
	specURL := firstNonEmpty(opts.SpecURL, firstNonEmpty(os.Getenv(publicAPIOpenAPIEnv), defaultPublicAPIOpenAPIURL))
	index := NewOpenAPISearchIndex(OpenAPISearchIndexOptions{SpecURL: specURL, TTL: opts.TTL})

	tool := mcpgo.NewTool(
		publicAPIToolName,
		mcpgo.WithDescription(
			"Answer questions about the Escape public REST API. Returns the best-matching "+
				"operation(s) from the live OpenAPI spec with a copy-pasteable cURL example. "+
				"Use for any question about how to call the public API, list endpoints, params, "+
				"or auth.",
		),
		mcpgo.WithString(
			"question",
			mcpgo.Required(),
			mcpgo.Description("Free-form question about the public API (e.g. 'how do I list scans from the last 3 days?')."),
		),
		mcpgo.WithNumber(
			"limit",
			mcpgo.Description("Maximum number of operations to return (1-3, default 1)."),
		),
		mcpgo.WithReadOnlyHintAnnotation(true),
		mcpgo.WithDestructiveHintAnnotation(false),
		mcpgo.WithIdempotentHintAnnotation(true),
		mcpgo.WithOpenWorldHintAnnotation(true),
	)

	server.AddTool(tool, buildPublicAPIHandler(index, opts.PublicAPIURL))
	return nil
}

func buildPublicAPIHandler(index *OpenAPISearchIndex, configuredBaseURL string) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := AuthFromContext(ctx); err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		question := strings.TrimSpace(request.GetString("question", ""))
		if question == "" {
			return mcpgo.NewToolResultError(
				`Invalid input. Expected {"question":"...","limit"?: number}.`,
			), nil
		}

		limit := int(request.GetFloat("limit", float64(publicAPIDefaultLimit)))
		if limit < 1 {
			limit = publicAPIDefaultLimit
		}
		if limit > publicAPIMaxLimit {
			limit = publicAPIMaxLimit
		}

		matches, servers, err := index.Search(ctx, question, limit)
		if err != nil {
			return formatPublicAPIFallback(question, err), nil
		}
		if len(matches) == 0 {
			return formatPublicAPIFallback(question, nil), nil
		}

		baseURL := resolveBaseURL(configuredBaseURL, servers)
		return formatPublicAPIResult(question, matches, baseURL), nil
	}
}

// resolveBaseURL prefers the operator-configured PublicAPIURL (so staging/dev
// MCP deployments hit the right environment), then falls back to the OpenAPI
// document's first server, then the production default.
func resolveBaseURL(configured string, servers []string) string {
	if configured = strings.TrimSpace(configured); configured != "" {
		// PublicAPIURL points at the API root (e.g. https://public.escape.tech).
		// The OpenAPI spec is mounted under /v3 today, so suffix accordingly.
		base := strings.TrimRight(configured, "/")
		if strings.HasSuffix(base, "/v3") {
			return base
		}
		return base + "/v3"
	}
	if len(servers) > 0 && strings.TrimSpace(servers[0]) != "" {
		return strings.TrimRight(servers[0], "/")
	}
	return publicAPIDefaultBaseURL
}

func formatPublicAPIResult(question string, matches []indexedOperation, baseURL string) *mcpgo.CallToolResult {
	lines := []string{fmt.Sprintf("Question: %s", question), ""}

	apiMatches := make([]map[string]any, 0, len(matches))
	for i, op := range matches {
		curl := RenderCurl(op, baseURL)
		header := fmt.Sprintf("## %s %s", op.Method, op.Path)
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, header)
		if op.Summary != "" {
			lines = append(lines, op.Summary)
		}
		if op.Description != "" {
			lines = append(lines, "", op.Description)
		}
		lines = append(lines, "", "```bash", curl, "```")

		apiMatches = append(apiMatches, map[string]any{
			"operationId": op.OperationID,
			"method":      op.Method,
			"path":        op.Path,
			"summary":     op.Summary,
			"description": op.Description,
			"tags":        op.Tags,
			"parameters":  op.Parameters,
			"requestBody": op.RequestBody,
			"curl":        curl,
		})
	}

	payload := map[string]any{
		"question": question,
		"baseUrl":  baseURL,
		"matches":  apiMatches,
	}
	return mcpgo.NewToolResultStructured(payload, strings.Join(lines, "\n"))
}

func formatPublicAPIFallback(question string, fetchErr error) *mcpgo.CallToolResult {
	lead := "I could not match your question to a public API operation."
	if fetchErr != nil {
		lead = "The OpenAPI spec is temporarily unreachable; please try again in a moment."
	}
	lines := []string{
		fmt.Sprintf("Question: %s", question),
		"",
		lead,
		"Browse the full reference: https://public.escape.tech/v3/",
	}
	payload := map[string]any{
		"question":        question,
		"matches":         []map[string]any{},
		"fallback":        true,
		"specUnavailable": fetchErr != nil,
	}
	return mcpgo.NewToolResultStructured(payload, strings.Join(lines, "\n"))
}
