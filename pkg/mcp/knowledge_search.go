package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultDocsSiteURL        = "https://docs.escape.tech/"
	defaultDocsSearchIndexURL = "https://docs.escape.tech/search/search_index.json"
	defaultDocsIndexTTL       = 5 * time.Minute
	maxDocsSnippetChars       = 220
	maxDocsResultsPerQuery    = 8

	// docsHTTPTimeout caps both the index fetch and any single search call.
	// Generous because the docs index is ~MB-sized JSON pulled cold once.
	docsHTTPTimeout = 15 * time.Second

	// docsErrorBodyPeekBytes bounds how much of an HTTP error body we surface
	// in the wrapped error to avoid leaking large/sensitive payloads.
	docsErrorBodyPeekBytes = 256

	// termCoverageWeight is the multiplier applied to the
	// (matched / total query terms) ratio. Tuned alongside the per-field
	// weights below to stay parity with the TS reference implementation.
	termCoverageWeight = 3
)

// DocsSearchIndex caches and queries the docs.escape.tech search index built
// by Docusaurus.
type DocsSearchIndex struct {
	docsSiteURL    string
	searchIndexURL string
	ttl            time.Duration
	httpClient     *http.Client

	mu          sync.Mutex
	cache       []indexedDoc
	cacheUntil  time.Time
	inFlight    chan struct{}
	inFlightErr error
	inFlightRes []indexedDoc
}

// KnowledgeSearchResult is what the tool returns per match.
type KnowledgeSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type rawSearchIndexDoc struct {
	Location string `json:"location"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}

type indexedDoc struct {
	location           string
	url                string
	title              string
	text               string
	normalizedTitle    string
	normalizedText     string
	normalizedLocation string
}

// NewDocsSearchIndex builds a docs search index with sensible defaults. Any
// option can be overridden by passing DocsSearchIndexOptions.
func NewDocsSearchIndex(options DocsSearchIndexOptions) *DocsSearchIndex {
	index := &DocsSearchIndex{
		// Trailing slashes in the docs site URL break url.ResolveReference
		// because the joined location ends up at the docs site root rather
		// than the relative path. Strip them defensively.
		docsSiteURL:    strings.TrimRight(firstNonEmpty(options.DocsSiteURL, defaultDocsSiteURL), "/"),
		searchIndexURL: firstNonEmpty(options.SearchIndexURL, defaultDocsSearchIndexURL),
		ttl:            firstNonZero(options.TTL, defaultDocsIndexTTL),
		httpClient:     options.HTTPClient,
	}
	if index.httpClient == nil {
		index.httpClient = &http.Client{Timeout: docsHTTPTimeout}
	}
	return index
}

// DocsSearchIndexOptions configures the docs search index.
type DocsSearchIndexOptions struct {
	DocsSiteURL    string
	SearchIndexURL string
	TTL            time.Duration
	HTTPClient     *http.Client
}

// Search returns the top N documentation matches for the given query.
func (d *DocsSearchIndex) Search(ctx context.Context, queryInput string, limit int) ([]KnowledgeSearchResult, error) {
	if limit < 1 {
		limit = 1
	}
	if limit > maxDocsResultsPerQuery {
		limit = maxDocsResultsPerQuery
	}

	spec := toDocsQuery(queryInput)
	if spec.normalizedQuery == "" {
		return nil, nil
	}

	docs, err := d.loadDocs(ctx)
	if err != nil {
		return nil, err
	}

	ranked := make([]scoredDoc, 0, len(docs))
	for _, doc := range docs {
		s := scoreDoc(doc, spec)
		if s > 0 {
			ranked = append(ranked, scoredDoc{doc: doc, score: s})
		}
	}

	// Sort: higher score first; ties prefer non-anchor links, shorter paths,
	// and stable URL order.
	sortScoredDocs(ranked)

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	out := make([]KnowledgeSearchResult, 0, len(ranked))
	for _, entry := range ranked {
		out = append(out, KnowledgeSearchResult{
			Title:   entry.doc.title,
			URL:     entry.doc.url,
			Snippet: snippetFor(entry.doc),
		})
	}
	return out, nil
}

type docsQuerySpec struct {
	normalizedQuery string
	terms           []string
}

var docsQueryStopWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "documentation": {}, "doc": {}, "docs": {},
	"for": {}, "from": {}, "give": {}, "guide": {}, "how": {}, "i": {}, "in": {},
	"is": {}, "link": {}, "links": {}, "me": {}, "of": {}, "on": {}, "please": {},
	"show": {}, "the": {}, "to": {}, "url": {}, "urls": {},
}

func toDocsQuery(query string) docsQuerySpec {
	normalized := normalizeForSearch(query)
	raw := strings.Fields(normalized)
	stemmed := make([]string, 0, len(raw))
	for _, token := range raw {
		if token == "" {
			continue
		}
		stemmed = append(stemmed, StemToken(token))
	}
	filtered := make([]string, 0, len(stemmed))
	for _, token := range stemmed {
		if _, isStop := docsQueryStopWords[token]; isStop {
			continue
		}
		filtered = append(filtered, token)
	}
	if len(filtered) == 0 {
		filtered = stemmed
	}
	// Deduplicate while preserving order.
	seen := make(map[string]struct{}, len(filtered))
	unique := make([]string, 0, len(filtered))
	for _, token := range filtered {
		if _, dup := seen[token]; dup {
			continue
		}
		seen[token] = struct{}{}
		unique = append(unique, token)
	}
	return docsQuerySpec{
		normalizedQuery: strings.Join(unique, " "),
		terms:           unique,
	}
}

func scoreDoc(doc indexedDoc, query docsQuerySpec) float64 {
	if query.normalizedQuery == "" || len(query.terms) == 0 {
		return 0
	}

	var score float64

	if doc.normalizedTitle == query.normalizedQuery ||
		doc.normalizedTitle == query.normalizedQuery+"s" {
		score += 4
	} else if strings.HasPrefix(doc.normalizedTitle, query.normalizedQuery) {
		score += 2
	}

	if strings.Contains(doc.normalizedTitle, query.normalizedQuery) {
		score += 8
	}
	if strings.Contains(doc.normalizedLocation, query.normalizedQuery) {
		score += 6
	}
	if strings.Contains(doc.normalizedText, query.normalizedQuery) {
		score += 4
	}

	matched := 0
	for _, term := range query.terms {
		did := false
		if strings.Contains(doc.normalizedTitle, term) {
			score += 2.5
			did = true
		}
		if strings.Contains(doc.normalizedLocation, term) {
			score += 1.5
			did = true
		}
		if strings.Contains(doc.normalizedText, term) {
			score++
			did = true
		}
		if did {
			matched++
		}
	}
	score += (float64(matched) / float64(len(query.terms))) * termCoverageWeight

	if strings.Contains(doc.location, "#") {
		score -= 0.35
	}
	return score
}

func snippetFor(doc indexedDoc) string {
	if doc.text != "" {
		return truncateRunes(doc.text, maxDocsSnippetChars)
	}
	return truncateRunes(doc.title, maxDocsSnippetChars)
}

func (d *DocsSearchIndex) loadDocs(ctx context.Context) ([]indexedDoc, error) {
	d.mu.Lock()
	if len(d.cache) > 0 && time.Now().Before(d.cacheUntil) {
		defer d.mu.Unlock()
		return d.cache, nil
	}
	if d.inFlight != nil {
		ch := d.inFlight
		d.mu.Unlock()
		select {
		case <-ch:
			d.mu.Lock()
			defer d.mu.Unlock()
			if d.inFlightErr != nil {
				return nil, d.inFlightErr
			}
			return d.inFlightRes, nil
		case <-ctx.Done():
			return nil, fmt.Errorf("docs index wait: %w", ctx.Err())
		}
	}

	d.inFlight = make(chan struct{})
	d.mu.Unlock()

	docs, err := d.fetchDocs(ctx)

	d.mu.Lock()
	d.inFlightErr = err
	d.inFlightRes = docs
	close(d.inFlight)
	d.inFlight = nil
	if err == nil {
		d.cache = docs
		d.cacheUntil = time.Now().Add(d.ttl)
	}
	d.mu.Unlock()
	return docs, err
}

func (d *DocsSearchIndex) fetchDocs(ctx context.Context) ([]indexedDoc, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.searchIndexURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new docs index request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docs index fetch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, docsErrorBodyPeekBytes))
		return nil, fmt.Errorf("docs index http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Docs []rawSearchIndexDoc `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode docs index: %w", err)
	}

	docs := make([]indexedDoc, 0, len(payload.Docs))
	for _, raw := range payload.Docs {
		if strings.TrimSpace(raw.Location) == "" {
			continue
		}
		docs = append(docs, toIndexedDoc(raw, d.docsSiteURL))
	}
	return docs, nil
}

func toIndexedDoc(raw rawSearchIndexDoc, docsSiteURL string) indexedDoc {
	cleanText := collapseWhitespace(raw.Text)
	cleanTitle := collapseWhitespace(raw.Title)
	if cleanTitle == "" {
		cleanTitle = raw.Location
	}

	absolute := raw.Location
	if base, err := url.Parse(docsSiteURL); err == nil {
		if ref, err := url.Parse(raw.Location); err == nil {
			absolute = base.ResolveReference(ref).String()
		}
	}

	return indexedDoc{
		location:           raw.Location,
		url:                absolute,
		title:              cleanTitle,
		text:               cleanText,
		normalizedTitle:    normalizeForSearch(cleanTitle),
		normalizedText:     normalizeForSearch(cleanText),
		normalizedLocation: normalizeForSearch(raw.Location),
	}
}

func normalizeForSearch(value string) string {
	var out strings.Builder
	out.Grow(len(value))
	lastWasSpace := false
	for _, r := range strings.ToLower(value) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			out.WriteRune(r)
			lastWasSpace = false
		default:
			if !lastWasSpace {
				out.WriteRune(' ')
				lastWasSpace = true
			}
		}
	}
	return strings.TrimSpace(out.String())
}

func collapseWhitespace(value string) string {
	fields := strings.Fields(value)
	return strings.Join(fields, " ")
}

func firstNonEmpty(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func firstNonZero(primary, fallback time.Duration) time.Duration {
	if primary > 0 {
		return primary
	}
	return fallback
}

func truncateRunes(value string, maxChars int) string {
	runes := []rune(value)
	if len(runes) <= maxChars {
		return value
	}
	return string(runes[:maxChars]) + "..."
}

type scoredDoc struct {
	doc   indexedDoc
	score float64
}

func sortScoredDocs(entries []scoredDoc) {
	// sort.Slice would be nicer but we need a stable tie-break, so keep an
	// explicit insertion-style sort: the TS version uses the same multi-level
	// comparator, and n here is small (≤ 200).
	for i := 1; i < len(entries); i++ {
		j := i
		for j > 0 && compareScoredDocs(entries[j-1], entries[j]) > 0 {
			entries[j-1], entries[j] = entries[j], entries[j-1]
			j--
		}
	}
}

func compareScoredDocs(left, right scoredDoc) int {
	if left.score != right.score {
		if right.score > left.score {
			return 1
		}
		return -1
	}
	leftAnchor := 0
	rightAnchor := 0
	if strings.Contains(left.doc.location, "#") {
		leftAnchor = 1
	}
	if strings.Contains(right.doc.location, "#") {
		rightAnchor = 1
	}
	if leftAnchor != rightAnchor {
		return leftAnchor - rightAnchor
	}
	if len(left.doc.location) != len(right.doc.location) {
		return len(left.doc.location) - len(right.doc.location)
	}
	return strings.Compare(left.doc.url, right.doc.url)
}

// ErrDocsIndexUnavailable is returned when the docs index can't be fetched.
var ErrDocsIndexUnavailable = errors.New("docs index unavailable")
