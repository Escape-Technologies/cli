package mcp

import "strings"

// LinkTarget describes which surfaces the link-intent detector thinks the
// question is about.
type LinkTarget string

// LinkTarget* enumerate the surfaces the link-intent detector can route the
// caller to.
const (
	LinkTargetDocs     LinkTarget = "docs"
	LinkTargetPlatform LinkTarget = "platform"
	LinkTargetBoth     LinkTarget = "both"
	LinkTargetNone     LinkTarget = "none"
)

// LinkIntent is the detector's verdict.
type LinkIntent struct {
	ExplicitLinkRequest bool
	Target              LinkTarget
}

var (
	linkHints            = []string{"link", "url", "href"}
	docsHints            = []string{"doc", "docs", "documentation", "guide", "kb", "readme"}
	platformHints        = []string{"platform", "app", "dashboard", "console", "ui", "in the platform", "inside the platform", "edit", "configure", "settings"}
	knowledgeIntentHints = []string{"configure", "configuration", "setup", "set up", "troubleshoot", "troubleshooting", "private location", "sso", "firewall"}
	knowledgePrefixes    = []string{"what is ", "what are ", "how to ", "how do ", "how can ", "explain ", "define ", "help "}
	actionPrefixes       = []string{"show me ", "list ", "start ", "run ", "create ", "update ", "delete ", "add ", "top ", "how many ", "count ", "when was "}

	queryNoise = map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "for": {}, "from": {}, "give": {},
		"here": {}, "how": {}, "i": {}, "in": {}, "inside": {}, "is": {},
		"link": {}, "links": {}, "me": {}, "of": {}, "on": {}, "platform": {},
		"please": {}, "send": {}, "show": {}, "the": {}, "to": {}, "url": {},
		"urls": {}, "documentation": {}, "documentations": {}, "docs": {}, "doc": {},
	}
)

func tokenizeNormalized(value string) []string {
	normalized := normalizeForSearch(value)
	if normalized == "" {
		return nil
	}
	return strings.Fields(normalized)
}

func containsTokenSequence(tokens, sequence []string) bool {
	if len(sequence) == 0 || len(tokens) < len(sequence) {
		return false
	}
	for start := 0; start+len(sequence) <= len(tokens); start++ {
		matches := true
		for offset, expected := range sequence {
			if tokens[start+offset] != expected {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func containsHint(text string, hints []string) bool {
	normalized := normalizeForSearch(text)
	if normalized == "" {
		return false
	}
	tokens := strings.Fields(normalized)
	tokenSet := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = struct{}{}
	}
	for _, hint := range hints {
		hintNormalized := normalizeForSearch(hint)
		if hintNormalized == "" {
			continue
		}
		hintTokens := strings.Fields(hintNormalized)
		if len(hintTokens) == 1 {
			if _, ok := tokenSet[hintTokens[0]]; ok {
				return true
			}
			continue
		}
		if containsTokenSequence(tokens, hintTokens) {
			return true
		}
	}
	return false
}

func startsWithAny(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func isKnowledgeQuestion(query string) bool {
	normalized := normalizeForSearch(query)
	if normalized == "" || startsWithAny(normalized, actionPrefixes) {
		return false
	}
	if startsWithAny(normalized, knowledgePrefixes) {
		return true
	}
	return containsHint(normalized, knowledgeIntentHints)
}

// DetectLinkIntent classifies a user question by the surface they want a link
// to (docs / platform / both / none) and whether they asked explicitly.
func DetectLinkIntent(query string) LinkIntent {
	knowledge := isKnowledgeQuestion(query)
	explicit := containsHint(query, linkHints) || knowledge
	wantsDocs := containsHint(query, docsHints)
	wantsPlatform := containsHint(query, platformHints)

	switch {
	case wantsDocs && wantsPlatform:
		return LinkIntent{ExplicitLinkRequest: explicit, Target: LinkTargetBoth}
	case wantsDocs:
		return LinkIntent{ExplicitLinkRequest: explicit, Target: LinkTargetDocs}
	case wantsPlatform:
		target := LinkTargetPlatform
		if knowledge {
			target = LinkTargetBoth
		}
		return LinkIntent{ExplicitLinkRequest: explicit, Target: target}
	case explicit:
		return LinkIntent{ExplicitLinkRequest: explicit, Target: LinkTargetBoth}
	}
	return LinkIntent{ExplicitLinkRequest: explicit, Target: LinkTargetNone}
}

// BuildDocsQuery drops noise words from a user query so the docs search index
// can match meaningful terms.
func BuildDocsQuery(query string) string {
	tokens := tokenizeNormalized(query)
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, noise := queryNoise[token]; noise {
			continue
		}
		filtered = append(filtered, token)
	}
	if len(filtered) == 0 {
		return normalizeForSearch(query)
	}
	return strings.Join(filtered, " ")
}
