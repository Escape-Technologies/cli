package mcp

import (
	_ "embed"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"unicode"
)

//go:embed platform_routes.json
var embeddedPlatformRoutesJSON []byte

// PlatformRoute is a single entry of the curated platform-URL catalog.
// Ported from services/mcp-server/src/generated/platform-routes.generated.ts.
type PlatformRoute struct {
	Path   string   `json:"path"`
	Tokens []string `json:"tokens"`
}

// PlatformLink is what the selector returns.
type PlatformLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// PlatformLinkSelector picks relevant platform deep links for a query.
type PlatformLinkSelector struct {
	baseURL string
	catalog []PlatformRoute
}

var platformLinkStopWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "by": {},
	"dashboard": {}, "doc": {}, "docs": {}, "documentation": {}, "edit": {},
	"for": {}, "from": {}, "how": {}, "in": {}, "inside": {}, "is": {},
	"link": {}, "links": {}, "of": {}, "on": {}, "or": {}, "platform": {},
	"please": {}, "show": {}, "the": {}, "to": {}, "url": {}, "urls": {},
	"what": {}, "when": {}, "where": {}, "with": {}, "why": {},
}

// NewPlatformLinkSelector builds a selector using the embedded catalog and
// the given base URL (e.g. https://app.escape.tech).
func NewPlatformLinkSelector(baseURL string) (*PlatformLinkSelector, error) {
	return NewPlatformLinkSelectorWithCatalog(baseURL, nil)
}

// NewPlatformLinkSelectorWithCatalog lets callers override the catalog
// (mainly for tests). A nil/empty catalog falls back to the embedded one.
func NewPlatformLinkSelectorWithCatalog(baseURL string, catalog []PlatformRoute) (*PlatformLinkSelector, error) {
	if len(catalog) == 0 {
		if err := json.Unmarshal(embeddedPlatformRoutesJSON, &catalog); err != nil {
			return nil, err
		}
	}
	return &PlatformLinkSelector{
		baseURL: strings.TrimRight(baseURL, "/"),
		catalog: catalog,
	}, nil
}

// Select returns up to `limit` platform links that match the query context.
// Ported from platform-links.ts (scoreRoute + createPlatformLinkSelector).
func (s *PlatformLinkSelector) Select(queryContext string, limit int) []PlatformLink {
	tokens := tokenizePlatformQuery(queryContext)
	if len(tokens) == 0 {
		return nil
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 5 {
		limit = 5
	}
	normalizedQuery := normalizeForSearch(queryContext)

	type scoredRoute struct {
		route PlatformRoute
		score float64
	}
	scored := make([]scoredRoute, 0, len(s.catalog))
	for _, route := range s.catalog {
		score := scoreRoute(route, tokens, normalizedQuery)
		if score < 1.5 {
			continue
		}
		scored = append(scored, scoredRoute{route: route, score: score})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].route.Path < scored[j].route.Path
	})
	if len(scored) > limit {
		scored = scored[:limit]
	}

	out := make([]PlatformLink, 0, len(scored))
	for _, entry := range scored {
		resolved := entry.route.Path
		if base, err := url.Parse(s.baseURL); err == nil {
			if ref, err := url.Parse(entry.route.Path); err == nil {
				resolved = base.ResolveReference(ref).String()
			}
		}
		out = append(out, PlatformLink{
			Label: labelFromRoutePath(entry.route.Path),
			URL:   resolved,
		})
	}
	return out
}

func tokenizePlatformQuery(value string) []string {
	normalized := normalizeForSearch(value)
	if normalized == "" {
		return nil
	}
	seen := make(map[string]struct{})
	tokens := make([]string, 0)
	for _, token := range strings.Fields(normalized) {
		stemmed := StemToken(token)
		if stemmed == "" {
			continue
		}
		if _, stop := platformLinkStopWords[stemmed]; stop {
			continue
		}
		if _, dup := seen[stemmed]; dup {
			continue
		}
		seen[stemmed] = struct{}{}
		tokens = append(tokens, stemmed)
	}
	return tokens
}

func scoreRoute(route PlatformRoute, queryTokens []string, normalizedQuery string) float64 {
	if route.Path == "/" {
		return 0
	}
	routeTokens := make(map[string]struct{}, len(route.Tokens))
	for _, token := range route.Tokens {
		stemmed := StemToken(token)
		if stemmed == "" {
			continue
		}
		routeTokens[stemmed] = struct{}{}
	}
	if len(routeTokens) == 0 {
		return 0
	}

	matched := 0
	for _, token := range queryTokens {
		if _, ok := routeTokens[token]; ok {
			matched++
		}
	}
	if matched == 0 {
		return 0
	}

	score := float64(matched) * 2
	score += float64(matched) / float64(len(queryTokens))

	normalizedRoutePath := normalizeForSearch(route.Path)
	if normalizedQuery != "" && strings.Contains(normalizedRoutePath, normalizedQuery) {
		score += 2
	}

	depth := 0
	for _, segment := range strings.Split(route.Path, "/") {
		if segment != "" {
			depth++
		}
	}
	if depth > 2 {
		score -= float64(depth-2) * 0.15
	}
	return score
}

func labelFromRoutePath(path string) string {
	segments := strings.FieldsFunc(path, func(r rune) bool { return r == '/' })
	if len(segments) == 0 {
		return "Platform"
	}
	last := segments[len(segments)-1]
	parts := strings.Split(last, "-")
	titled := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		titled = append(titled, string(runes))
	}
	if len(titled) == 0 {
		return "Platform"
	}
	return strings.Join(titled, " ")
}
