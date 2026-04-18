package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	knowledgeToolName        = "knowledge_answer_question"
	defaultPlatformBaseURL   = "https://app.escape.tech"
	knowledgePlatformBaseEnv = "PLATFORM_BASE_URL"
	knowledgeDocsIndexEnv    = "DOCS_SEARCH_INDEX_URL"
	knowledgeDocsSiteEnv     = "DOCS_SITE_URL"
	knowledgeDefaultLimit    = 5
	knowledgeMaxLimit        = 8
)

// KnowledgeOptions configures the answer_question tool. Zero-value fields
// fall back to the production defaults (docs.escape.tech, app.escape.tech).
type KnowledgeOptions struct {
	DocsSiteURL     string
	DocsIndexURL    string
	PlatformBaseURL string
	DocsTTL         int64
}

// RegisterKnowledgeTools registers the knowledge_answer_question tool on the
// given server. Port of createKnowledgeTools from the TS mcp-server.
func RegisterKnowledgeTools(server *mcpserver.MCPServer, opts KnowledgeOptions) error {
	docsSite := firstNonEmpty(opts.DocsSiteURL, firstNonEmpty(os.Getenv(knowledgeDocsSiteEnv), defaultDocsSiteURL))
	docsIndex := firstNonEmpty(opts.DocsIndexURL, firstNonEmpty(os.Getenv(knowledgeDocsIndexEnv), defaultDocsSearchIndexURL))
	platformBase := firstNonEmpty(opts.PlatformBaseURL, firstNonEmpty(os.Getenv(knowledgePlatformBaseEnv), defaultPlatformBaseURL))

	index := NewDocsSearchIndex(DocsSearchIndexOptions{
		DocsSiteURL:    docsSite,
		SearchIndexURL: docsIndex,
	})

	selector, err := NewPlatformLinkSelector(platformBase)
	if err != nil {
		return fmt.Errorf("build platform link selector: %w", err)
	}

	tool := mcpgo.NewTool(
		knowledgeToolName,
		mcpgo.WithDescription(
			"Answer Escape product/documentation questions and return authoritative docs/platform links.",
		),
		mcpgo.WithString(
			"question",
			mcpgo.Required(),
			mcpgo.Description("Customer or internal question to answer."),
		),
		mcpgo.WithString(
			"topic",
			mcpgo.Description("Optional category hint (for example: sso, scan, asm, mcp)."),
		),
		mcpgo.WithNumber(
			"limit",
			mcpgo.Description("Maximum number of documentation matches to return."),
		),
	)

	server.AddTool(tool, buildKnowledgeHandler(index, selector, docsSite, platformBase))
	return nil
}

func buildKnowledgeHandler(
	index *docsSearchIndex,
	selector *PlatformLinkSelector,
	docsSite string,
	platformBase string,
) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := AuthFromContext(ctx); err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		question := strings.TrimSpace(request.GetString("question", ""))
		if question == "" {
			return mcpgo.NewToolResultError(
				`Invalid input. Expected {"question":"...","topic"?: "...","limit"?: number}.`,
			), nil
		}
		topic := strings.TrimSpace(request.GetString("topic", ""))
		limit := int(request.GetFloat("limit", float64(knowledgeDefaultLimit)))
		if limit < 1 {
			limit = knowledgeDefaultLimit
		}
		if limit > knowledgeMaxLimit {
			limit = knowledgeMaxLimit
		}

		query := question
		if topic != "" {
			query = topic + " " + question
		}
		docsQuery := BuildDocsQuery(query)
		intent := DetectLinkIntent(query)

		loadDocs := func() []KnowledgeSearchResult {
			if docsQuery == "" {
				return nil
			}
			matches, err := index.Search(ctx, docsQuery, limit)
			if err != nil {
				return nil
			}
			return matches
		}

		if intent.ExplicitLinkRequest && intent.Target != LinkTargetNone {
			var docsMatches []KnowledgeSearchResult
			if intent.Target == LinkTargetDocs || intent.Target == LinkTargetBoth {
				docsMatches = loadDocs()
			}

			platformContext := query
			for _, match := range docsMatches {
				platformContext += " " + match.Title + " " + match.Snippet
			}

			var platformLinks []PlatformLink
			if intent.Target == LinkTargetPlatform || intent.Target == LinkTargetBoth {
				platformLinks = selector.Select(platformContext, 3)
			}

			return formatDirectedLinkResult(
				question, intent, docsMatches, platformLinks, docsSite, docsQuery, platformBase,
			), nil
		}

		matches := loadDocs()
		if len(matches) == 0 {
			return formatFallbackResult(
				question, selector.Select(query, 3), docsSite, docsQuery, platformBase,
			), nil
		}

		platformContext := query
		for _, match := range matches {
			platformContext += " " + match.Title + " " + match.Snippet
		}
		platformLinks := selector.Select(platformContext, 3)
		return formatGeneralResult(question, platformLinks, matches), nil
	}
}

func formatDirectedLinkResult(
	question string,
	intent LinkIntent,
	docsMatches []KnowledgeSearchResult,
	platformLinks []PlatformLink,
	docsSite string,
	docsQuery string,
	platformBase string,
) *mcpgo.CallToolResult {
	lines := []string{fmt.Sprintf("Question: %s", question)}

	if intent.Target == LinkTargetDocs || intent.Target == LinkTargetBoth {
		appendDocumentationLinks(&lines, docsMatches, docsSite, docsQuery)
	}
	if intent.Target == LinkTargetPlatform || intent.Target == LinkTargetBoth {
		appendPlatformLinks(&lines, platformLinks, platformBase)
	}

	payload := map[string]any{
		"question":      question,
		"docsMatches":   docsMatches,
		"platformLinks": platformLinks,
	}
	return mcpgo.NewToolResultStructured(payload, strings.Join(lines, "\n"))
}

func formatFallbackResult(
	question string,
	platformLinks []PlatformLink,
	docsSite string,
	docsQuery string,
	platformBase string,
) *mcpgo.CallToolResult {
	lines := []string{
		fmt.Sprintf("Question: %s", question),
		"",
		"I could not find strong documentation matches right now.",
		"Try rephrasing with product terms (for example: scans, profiles, asm, api).",
		fmt.Sprintf("Docs home: [Documentation home](%s)", docsHomeURL(docsSite)),
	}
	if docsQuery != "" {
		lines = append(lines, fmt.Sprintf("Docs search: [Documentation search](%s)", docsSearchURL(docsSite, docsQuery)))
	}
	appendPlatformLinks(&lines, platformLinks, platformBase)

	payload := map[string]any{
		"question":      question,
		"docsMatches":   []KnowledgeSearchResult{},
		"platformLinks": platformLinks,
		"fallback":      true,
	}
	return mcpgo.NewToolResultStructured(payload, strings.Join(lines, "\n"))
}

func formatGeneralResult(
	question string,
	platformLinks []PlatformLink,
	matches []KnowledgeSearchResult,
) *mcpgo.CallToolResult {
	lines := []string{
		fmt.Sprintf("Question: %s", question),
		"",
		fmt.Sprintf("Top documentation matches (%d):", len(matches)),
	}
	for i, match := range matches {
		lines = append(lines, fmt.Sprintf("%d. [%s](%s)", i+1, match.Title, match.URL))
		lines = append(lines, "   "+match.Snippet)
	}
	if len(platformLinks) > 0 {
		lines = append(lines, "", "Relevant platform links:")
		for _, link := range platformLinks {
			lines = append(lines, fmt.Sprintf("- [%s](%s)", link.Label, link.URL))
		}
	}
	lines = append(lines, "", "If this requires tenant-specific data, use insights/actions tools and combine with these docs.")

	payload := map[string]any{
		"question":      question,
		"docsMatches":   matches,
		"platformLinks": platformLinks,
	}
	return mcpgo.NewToolResultStructured(payload, strings.Join(lines, "\n"))
}

func appendDocumentationLinks(lines *[]string, matches []KnowledgeSearchResult, docsSite, docsQuery string) {
	*lines = append(*lines, "", "Documentation links:")
	if len(matches) > 0 {
		for i, match := range matches {
			*lines = append(*lines, fmt.Sprintf("%d. [%s](%s)", i+1, match.Title, match.URL))
		}
		return
	}
	*lines = append(*lines, fmt.Sprintf("1. [Documentation home](%s)", docsHomeURL(docsSite)))
	if docsQuery != "" {
		*lines = append(*lines, fmt.Sprintf("2. [Documentation search](%s)", docsSearchURL(docsSite, docsQuery)))
	}
}

func appendPlatformLinks(lines *[]string, platformLinks []PlatformLink, platformBase string) {
	*lines = append(*lines, "", "Platform links:")
	if len(platformLinks) > 0 {
		for i, link := range platformLinks {
			*lines = append(*lines, fmt.Sprintf("%d. [%s](%s)", i+1, link.Label, link.URL))
		}
		return
	}
	*lines = append(*lines, fmt.Sprintf("1. [Escape platform](%s)", ensureTrailingSlash(platformBase)))
}

func docsHomeURL(docsSite string) string {
	if base, err := url.Parse(docsSite); err == nil {
		return base.ResolveReference(&url.URL{Path: "/documentation/"}).String()
	}
	return strings.TrimRight(docsSite, "/") + "/documentation/"
}

func docsSearchURL(docsSite, query string) string {
	base, err := url.Parse(docsSite)
	if err != nil {
		return strings.TrimRight(docsSite, "/") + "/documentation/?q=" + url.QueryEscape(query)
	}
	ref := &url.URL{Path: "/documentation/", RawQuery: "q=" + url.QueryEscape(query)}
	return base.ResolveReference(ref).String()
}

func ensureTrailingSlash(value string) string {
	if strings.HasSuffix(value, "/") {
		return value
	}
	return value + "/"
}

// Encode helper kept here so we don't pull the json package for formatting.
// Defer to encoding/json if we ever need more complex formatting.
var _ = json.Marshal
