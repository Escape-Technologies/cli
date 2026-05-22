package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

	mockIndex := struct {
		Docs []rawSearchIndexDoc `json:"docs"`
	}{
		Docs: []rawSearchIndexDoc{
			{Location: "documentation/private-location/", Title: "Private Location", Text: "A private location is a self-hosted scanner tunnel."},
			{Location: "documentation/api-reference/", Title: "API reference", Text: "Endpoints and payloads."},
		},
	}
	body, err := json.Marshal(mockIndex)
	if err != nil {
		t.Fatalf("marshal mock index: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	index := NewDocsSearchIndex(DocsSearchIndexOptions{
		DocsSiteURL:    "https://docs.escape.tech/",
		SearchIndexURL: srv.URL,
		TTL:            time.Minute,
	})

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
}
