package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func TestStemToken_MatchesTSBehaviour(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in, want string
	}{
		{"Issues", "issue"},
		{"Issue", "issue"},
		{"POLICIES", "policy"},
		{"processes", "process"},
		{"Scans", "scan"},
		{"categories", "category"},
		{"is", "is"},
		{"", ""},
		// "API" with punctuation survives and becomes "api".
		{"API!", "api"},
	}
	for _, tc := range cases {
		got := StemToken(tc.in)
		if got != tc.want {
			t.Errorf("StemToken(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDetectLinkIntent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		query    string
		target   LinkTarget
		explicit bool
	}{
		{"what is a private location?", LinkTargetBoth, true}, // knowledge prefix + knowledge hint → defaults to 'both'
		{"link to the docs", LinkTargetDocs, true},
		{"show me the dashboard", LinkTargetPlatform, false},
		{"give me the platform URL and docs URL", LinkTargetBoth, true},
		{"list my scans", LinkTargetNone, false}, // action prefix
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			got := DetectLinkIntent(tc.query)
			if got.Target != tc.target {
				t.Errorf("target = %q, want %q", got.Target, tc.target)
			}

			if got.ExplicitLinkRequest != tc.explicit {
				t.Errorf("explicit = %v, want %v", got.ExplicitLinkRequest, tc.explicit)
			}
		})
	}
}

func TestBuildDocsQuery_StripsNoise(t *testing.T) {
	t.Parallel()

	got := BuildDocsQuery("Please show me the link to the documentation on SSO")
	// "please", "show", "me", "the", "link", "to", "documentation" all in QUERY_NOISE.
	if !strings.Contains(got, "sso") {
		t.Errorf("expected 'sso' in %q", got)
	}

	if strings.Contains(got, "please") {
		t.Errorf("expected noise words stripped from %q", got)
	}
}

func TestPlatformLinkSelector_MatchesRelevantRoutes(t *testing.T) {
	t.Parallel()

	selector, err := NewPlatformLinkSelector("https://app.escape.tech")
	if err != nil {
		t.Fatalf("NewPlatformLinkSelector: %v", err)
	}

	got := selector.Select("show me the issues dashboard", 3)
	if len(got) == 0 {
		t.Fatalf("expected at least one link for 'issues' query")
	}

	// At least one result should include "issues" in the URL path.
	found := false
	for _, link := range got {
		if strings.Contains(link.URL, "issue") {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected an issues-related link, got %+v", got)
	}
}

func TestDocsSearchIndex_SearchScoresMatches(t *testing.T) {
	t.Parallel()

	index := newDocsIndexFixture(t, []rawSearchIndexDoc{
		{Location: "documentation/platform/private-location/", Title: "Private Locations", Text: "<p>A private location is a <code>self-hosted</code> scanner&#39;s tunnel.</p>"},
		{Location: "documentation/api-reference/", Title: "API reference", Text: "<p>Endpoints and payloads.</p>"},
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := index.Search(ctx, "private location", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected at least one result")
	}

	if !strings.Contains(strings.ToLower(results[0].Title), "private") {
		t.Errorf("expected top result to be about 'private', got %q", results[0].Title)
	}

	if !strings.HasPrefix(results[0].URL, "https://docs.escape.tech/") {
		t.Errorf("expected absolute URL, got %q", results[0].URL)
	}

	if want := "A private location is a self-hosted scanner's tunnel."; results[0].Snippet != want {
		t.Errorf("snippet = %q, want HTML stripped to %q", results[0].Snippet, want)
	}
}

// newDocsIndexFixture serves the given docs entries in the Zensical
// search.json shape, plus Markdown exports keyed by URL path, through a fake
// transport so production URLs stay intact. Other paths answer with an HTML
// 404 like the real site.
func newDocsIndexFixture(t *testing.T, docs []rawSearchIndexDoc, markdown map[string]string) *DocsSearchIndex {
	t.Helper()

	body, err := json.Marshal(map[string]any{"config": map[string]any{}, "items": docs})
	if err != nil {
		t.Fatalf("marshal docs index: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/search.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, ok := markdown[r.URL.Path]
		if !ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("<html>Not found</html>"))

			return
		}

		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, _ = w.Write([]byte(content))
	})

	return NewDocsSearchIndex(DocsSearchIndexOptions{
		DocsSiteURL:    "https://docs.escape.tech/",
		SearchIndexURL: "https://docs.escape.tech/search.json",
		TTL:            time.Minute,
		HTTPClient:     &http.Client{Transport: handlerTransport{mux}},
	})
}

// handlerTransport answers every request with an in-process handler.
type handlerTransport struct{ handler http.Handler }

func (h handlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, r)

	return rec.Result(), nil
}

func privateLocationDocs() []rawSearchIndexDoc {
	return []rawSearchIndexDoc{
		{Location: "documentation/private-location/", Title: "Private Location", Text: "<p>Intro about private locations.</p>"},
		{Location: "documentation/private-location/#setup", Title: "Setup", Text: "<p>Setup steps for a private location.</p>"},
		{Location: "documentation/other/", Title: "Other", Text: "<p>Unrelated page.</p>"},
	}
}

func TestHTMLToText_IgnoresAttributes(t *testing.T) {
	t.Parallel()

	got := collapseWhitespace(htmlToText(`<ul><li><code title="a > b">hello</code></li><li>x &amp; y</li></ul>`))
	if got != "hello x & y" {
		t.Errorf("htmlToText = %q, want %q", got, "hello x & y")
	}
}

func TestDocsSearchIndex_GetPagePrefersMarkdown(t *testing.T) {
	t.Parallel()

	markdown := "# Private Location\n\n```yaml\nlocation: private\n```"
	index := newDocsIndexFixture(t, privateLocationDocs(), map[string]string{
		"/documentation/private-location/index.md": markdown,
	})

	page, err := index.GetPage(context.Background(), "/documentation/private-location/#setup")
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}

	if page.Text != markdown {
		t.Errorf("text = %q, want the Markdown export %q", page.Text, markdown)
	}

	if page.URL != "https://docs.escape.tech/documentation/private-location/" {
		t.Errorf("url = %q", page.URL)
	}
}

func TestDocsSearchIndex_GetPageAggregatesSections(t *testing.T) {
	t.Parallel()

	index := newDocsIndexFixture(t, privateLocationDocs(), nil)
	ctx := context.Background()

	page, err := index.GetPage(ctx, "https://docs.escape.tech/documentation/private-location/")
	if err != nil {
		t.Fatalf("GetPage by absolute URL: %v", err)
	}

	if page.Title != "Private Location" {
		t.Errorf("title = %q, want %q", page.Title, "Private Location")
	}

	if page.URL != "https://docs.escape.tech/documentation/private-location/" {
		t.Errorf("url = %q", page.URL)
	}

	if !strings.Contains(page.Text, "Intro about private locations.") ||
		!strings.Contains(page.Text, "Setup steps for a private location.") {
		t.Errorf("expected page to aggregate intro + section text, got %q", page.Text)
	}

	if strings.Contains(page.Text, "Unrelated") {
		t.Errorf("expected other pages to stay out, got %q", page.Text)
	}

	if page.Truncated {
		t.Errorf("expected no truncation for a short page")
	}

	byPath, err := index.GetPage(ctx, "/documentation/private-location#setup")
	if err != nil {
		t.Fatalf("GetPage by path with anchor: %v", err)
	}

	if byPath.Text != page.Text {
		t.Errorf("path lookup should match URL lookup: %q vs %q", byPath.Text, page.Text)
	}
}

func TestDocsSearchIndex_GetPageUnknownURL(t *testing.T) {
	t.Parallel()

	index := newDocsIndexFixture(t, privateLocationDocs(), nil)

	for _, input := range []string{"documentation/does-not-exist/", "https://docs.escape.tech/nope/", "   "} {
		if _, err := index.GetPage(context.Background(), input); !errors.Is(err, ErrDocsPageNotFound) {
			t.Errorf("GetPage(%q) err = %v, want ErrDocsPageNotFound", input, err)
		}
	}
}

func TestDocsSearchIndex_GetPageTruncates(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", maxDocsPageChars+100)
	index := newDocsIndexFixture(t, []rawSearchIndexDoc{
		{Location: "documentation/long-page/", Title: "Long Page", Text: long},
	}, nil)

	page, err := index.GetPage(context.Background(), "documentation/long-page/")
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}

	if !page.Truncated {
		t.Errorf("expected truncated=true for a %d-char page", len(long))
	}

	if len([]rune(page.Text)) != maxDocsPageChars {
		t.Errorf("text length = %d, want cap %d", len([]rune(page.Text)), maxDocsPageChars)
	}
}

func TestBuildKnowledgeGetPageHandler(t *testing.T) {
	t.Parallel()

	handler := buildKnowledgeGetPageHandler(newDocsIndexFixture(t, privateLocationDocs(), nil))

	// Auth gate.
	res, err := handler(context.Background(), newCallToolRequest(map[string]any{"url": "documentation/private-location/"}))
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}

	if !res.IsError {
		t.Fatalf("expected auth error, got: %+v", res)
	}

	// Missing url.
	res, _ = handler(newAuthedContext(), newCallToolRequest(map[string]any{"url": "  "}))
	if !res.IsError {
		t.Fatalf("expected validation error for blank url")
	}

	// Unknown page.
	res, _ = handler(newAuthedContext(), newCallToolRequest(map[string]any{"url": "documentation/nope/"}))
	if !res.IsError {
		t.Fatalf("expected not-found error for unknown page")
	}

	if !strings.Contains(textFromResult(t, res), "No documentation page found") {
		t.Errorf("unexpected not-found message: %s", textFromResult(t, res))
	}

	// Happy path returns full aggregated text + structured payload.
	res, _ = handler(newAuthedContext(), newCallToolRequest(map[string]any{
		"url": "https://docs.escape.tech/documentation/private-location/",
	}))
	if res.IsError {
		t.Fatalf("unexpected error result: %+v", res)
	}

	text := textFromResult(t, res)
	requireContains(t, text, "# Private Location")
	requireContains(t, text, "Intro about private locations.")
	requireContains(t, text, "Setup steps for a private location.")

	payload := structuredFromResult(t, res)
	if payload["title"] != "Private Location" {
		t.Errorf("payload title = %v", payload["title"])
	}

	if truncated, _ := payload["truncated"].(bool); truncated {
		t.Errorf("payload truncated should be false")
	}
}

func TestRegisterKnowledgeTools_AddsGetPageTool(t *testing.T) {
	t.Parallel()

	server := mcpserver.NewMCPServer("test", "0.0.0", mcpserver.WithToolCapabilities(false))
	if err := RegisterKnowledgeTools(server, KnowledgeOptions{}); err != nil {
		t.Fatalf("RegisterKnowledgeTools: %v", err)
	}

	tools := server.ListTools()
	registration, ok := tools[knowledgeGetPageToolName]
	if !ok {
		t.Fatalf("expected %q to be registered, got %v", knowledgeGetPageToolName, tools)
	}

	annotations := registration.Tool.Annotations
	if annotations.ReadOnlyHint == nil || !*annotations.ReadOnlyHint {
		t.Errorf("knowledge_get_page should carry readOnlyHint=true")
	}

	if annotations.DestructiveHint == nil || *annotations.DestructiveHint {
		t.Errorf("knowledge_get_page should carry destructiveHint=false")
	}

	if annotations.IdempotentHint == nil || !*annotations.IdempotentHint {
		t.Errorf("knowledge_get_page should carry idempotentHint=true")
	}
}

func TestFormatGeneralResult_ReferencesRealTools(t *testing.T) {
	t.Parallel()

	matches := []KnowledgeSearchResult{
		{Title: "Private Location", URL: "https://docs.escape.tech/documentation/private-location/", Snippet: "intro"},
	}
	res := formatGeneralResult("what is a private location", nil, matches)

	text := textFromResult(t, res)
	requireContains(t, text, "issues_list")
	requireContains(t, text, "list_capabilities")
	requireContains(t, text, "knowledge_get_page")
	requireMissing(t, text, "insights/actions")
}
