package mcp

import "strings"

// minStemmableLength: tokens of three characters or fewer skip stemming
// entirely (no morphology to strip cleanly).
const minStemmableLength = 3

// StemToken applies a lightweight English stemmer — NOT a full Porter stemmer —
// designed to match what the docs search index expects.
func StemToken(value string) string {
	clean := strings.Builder{}
	clean.Grow(len(value))
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			clean.WriteRune(r)
		}
	}
	out := clean.String()

	if len(out) <= minStemmableLength {
		return out
	}

	if strings.HasSuffix(out, "ies") && len(out) > 4 {
		return out[:len(out)-3] + "y"
	}

	if strings.HasSuffix(out, "sses") && len(out) > 5 {
		return out[:len(out)-2]
	}

	if strings.HasSuffix(out, "es") && len(out) > 4 &&
		!strings.HasSuffix(out, "ses") &&
		!strings.HasSuffix(out, "les") &&
		!strings.HasSuffix(out, "ues") {
		return out[:len(out)-2]
	}

	if strings.HasSuffix(out, "s") && !strings.HasSuffix(out, "ss") {
		return out[:len(out)-1]
	}

	return out
}
