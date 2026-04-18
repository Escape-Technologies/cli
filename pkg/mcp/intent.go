package mcp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// IntentMode controls how the tools/list response is shaped. See the
// MCP_INTENT_FILTER env var for the runtime switch.
type IntentMode string

const (
	// IntentModeOn runs the classifier when a chat-context header is present.
	// Selected tools ship with full schemas; the rest ship as compact stubs.
	IntentModeOn IntentMode = "on"
	// IntentModeCompactOnly always ships compact stubs for every command-backed
	// tool. The classifier is never consulted. Useful when the classifier is
	// misconfigured but we still need to fit under size limits.
	IntentModeCompactOnly IntentMode = "compact_only"
	// IntentModeOff disables the middleware and returns the raw library
	// response (every tool with its full schema). Emergency rollback only.
	IntentModeOff IntentMode = "off"

	chatContextHeader = "X-Escape-Chat-Context"
	intentModeEnv     = "MCP_INTENT_FILTER"

	// maxRequestBodyBytes is the soft limit applied to incoming MCP HTTP
	// requests. 10 MiB is generous for a JSON-RPC envelope (largest expected
	// payload is a tools/list reply we proxy back); anything beyond is
	// rejected before we attempt to decode it.
	maxRequestBodyBytes = 10 << 20

	// maxToolDescriptionRunes bounds how long a tool description we ship to
	// the classifier digest. Counted in runes so non-ASCII descriptions are
	// not split mid-character. Tuned to keep the total digest under typical
	// LLM context budgets when shipping ~50 tools.
	maxToolDescriptionRunes       = 160
	toolDescriptionEllipsisCutoff = 157
)

// IntentOptions configures the HTTP interceptor.
type IntentOptions struct {
	Mode       IntentMode
	Classifier Classifier
	Specs      []ToolSpec
}

// NewIntentMiddleware wraps next so that tools/list responses are rewritten
// according to Mode. All other requests pass through unchanged.
func NewIntentMiddleware(next http.Handler, opts IntentOptions) http.Handler {
	if opts.Mode == IntentModeOff {
		return next
	}

	byName := indexSpecs(opts.Specs)
	specNames := make(map[string]struct{}, len(opts.Specs))
	for _, s := range opts.Specs {
		specNames[s.Name] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
		if err != nil {
			// Restore body so the downstream handler can surface the real error.
			r.Body = io.NopCloser(bytes.NewReader(nil))
			next.ServeHTTP(w, r)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		var peek struct {
			Method string `json:"method"`
		}
		if jerr := json.Unmarshal(body, &peek); jerr != nil || peek.Method != "tools/list" {
			next.ServeHTTP(w, r)
			return
		}

		chatCtx, hasCtx := parseChatContext(r.Header.Get(chatContextHeader))

		var selected []string
		var classifierErr error
		var classifierDuration time.Duration
		if opts.Mode == IntentModeOn && opts.Classifier != nil && hasCtx {
			start := time.Now()
			names, rankErr := opts.Classifier.Rank(r.Context(), chatCtx, buildDigest(opts.Specs))
			classifierDuration = time.Since(start)
			classifierErr = rankErr
			if rankErr == nil {
				selected = filterKnown(names, specNames)
			}
		}

		// Let the library build the canonical tools/list response so that
		// built-in tools (list_capabilities, escape_get_tool_spec) are
		// included. We replace the command-backed entries in the response.
		recorder := newResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if recorder.status != http.StatusOK || !isJSON(recorder.contentType()) {
			recorder.flushTo(w)
			return
		}

		rewritten, fullCount, stubCount, bytesOut, rerr := rewriteToolsListResponse(
			recorder.buffer.Bytes(), byName, selected,
		)
		if rerr != nil {
			log.Printf("WARN intent tools/list rewrite failed: %v", rerr)
			recorder.flushTo(w)
			return
		}

		log.Printf(
			"INFO intent tools/list mode=%s has_ctx=%t classifier_ms=%d classifier_err=%v full=%d stub=%d bytes=%d",
			opts.Mode, hasCtx, classifierDuration.Milliseconds(),
			classifierErrString(classifierErr), fullCount, stubCount, bytesOut,
		)

		for key, values := range recorder.Header() {
			if strings.EqualFold(key, "Content-Length") {
				continue
			}
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.Header().Set("Content-Length", strconv.Itoa(bytesOut))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(rewritten)
	})
}

// ModeFromEnv resolves MCP_INTENT_FILTER, defaulting to fallback when unset
// or invalid.
func ModeFromEnv(fallback IntentMode) IntentMode {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(intentModeEnv)))
	switch raw {
	case string(IntentModeOn), string(IntentModeCompactOnly), string(IntentModeOff):
		return IntentMode(raw)
	}
	return fallback
}

// parseChatContext decodes the X-Escape-Chat-Context header. The api base64-
// encodes the JSON payload because HTTP headers must be ISO-8859-1 and chat
// messages routinely contain characters beyond 0xFF (emoji, CJK, etc.). We
// accept a raw JSON value too so older api builds or manual curl tests don't
// break.
func parseChatContext(raw string) (ChatContext, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ChatContext{}, false
	}

	payload := []byte(raw)
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) > 0 {
		payload = decoded
	}

	var parsed ChatContext
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return ChatContext{}, false
	}
	if strings.TrimSpace(parsed.Current) == "" && len(parsed.History) == 0 {
		return ChatContext{}, false
	}
	return parsed, true
}

func buildDigest(specs []ToolSpec) []ToolDigest {
	digest := make([]ToolDigest, 0, len(specs))
	for _, s := range specs {
		desc := s.Description
		// Truncate by runes — byte-slicing can split a multi-byte UTF-8
		// codepoint mid-character and corrupt the description shipped to
		// the classifier (descriptions can contain emoji or accented chars).
		runes := []rune(desc)
		if len(runes) > maxToolDescriptionRunes {
			desc = string(runes[:toolDescriptionEllipsisCutoff]) + "..."
		}
		digest = append(digest, ToolDigest{Name: s.Name, Description: desc})
	}
	return digest
}

func indexSpecs(specs []ToolSpec) map[string]ToolSpec {
	m := make(map[string]ToolSpec, len(specs))
	for _, s := range specs {
		m[s.Name] = s
	}
	return m
}

func filterKnown(names []string, known map[string]struct{}) []string {
	out := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if _, ok := known[n]; !ok {
			continue
		}
		if _, dup := seen[n]; dup {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

// rewriteToolsListResponse rewrites the library's JSON-RPC response so that
// command-backed tool specs are replaced by compact stubs (with full schemas
// retained for the `selected` subset). Built-in tools (anything not in
// byName) are preserved untouched.
func rewriteToolsListResponse(
	body []byte,
	byName map[string]ToolSpec,
	selected []string,
) (rewritten []byte, fullCount, stubCount, bytesOut int, err error) {
	// Tolerate the library returning either a plain JSON body or an SSE
	// stream with a single `data:` line. When WithDisableStreaming(true) is
	// set it's plain JSON, but be defensive.
	trimmed := bytes.TrimSpace(body)
	if bytes.HasPrefix(trimmed, []byte("event:")) || bytes.HasPrefix(trimmed, []byte("data:")) {
		if inner, ok := extractFirstDataFrame(trimmed); ok {
			trimmed = inner
		} else {
			return nil, 0, 0, 0, errors.New("could not extract SSE data frame")
		}
	}

	var envelope struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  struct {
			Tools      json.RawMessage `json:"tools"`
			NextCursor string          `json:"nextCursor,omitempty"`
		} `json:"result"`
		Error json.RawMessage `json:"error,omitempty"`
	}
	if jerr := json.Unmarshal(trimmed, &envelope); jerr != nil {
		return nil, 0, 0, 0, fmt.Errorf("unmarshal tools/list envelope: %w", jerr)
	}
	if len(envelope.Error) > 0 {
		// Don't rewrite error responses.
		return nil, 0, 0, 0, errors.New("response carries error; skipping rewrite")
	}

	var originalTools []json.RawMessage
	if err := json.Unmarshal(envelope.Result.Tools, &originalTools); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("unmarshal tools array: %w", err)
	}

	selectedSet := make(map[string]struct{}, len(selected))
	for _, n := range selected {
		selectedSet[n] = struct{}{}
	}

	rewrittenTools := make([]any, 0, len(originalTools))
	for _, raw := range originalTools {
		var header struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &header); err != nil {
			// Unknown shape — preserve verbatim.
			rewrittenTools = append(rewrittenTools, json.RawMessage(raw))
			continue
		}
		spec, isCommandBacked := byName[header.Name]
		if !isCommandBacked {
			rewrittenTools = append(rewrittenTools, json.RawMessage(raw))
			continue
		}
		if _, picked := selectedSet[header.Name]; picked {
			rewrittenTools = append(rewrittenTools, json.RawMessage(raw))
			fullCount++
			continue
		}
		stub, err := BuildStubTool(spec)
		if err != nil {
			rewrittenTools = append(rewrittenTools, json.RawMessage(raw))
			fullCount++
			continue
		}
		rewrittenTools = append(rewrittenTools, stub)
		stubCount++
	}

	// Build response envelope, preserving id and jsonrpc fields.
	result := map[string]any{"tools": rewrittenTools}
	if envelope.Result.NextCursor != "" {
		result["nextCursor"] = envelope.Result.NextCursor
	}
	out := map[string]any{
		"jsonrpc": envelope.JSONRPC,
		"result":  result,
	}
	if len(envelope.ID) > 0 {
		out["id"] = envelope.ID
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("marshal rewritten tools/list: %w", err)
	}
	return encoded, fullCount, stubCount, len(encoded), nil
}

// extractFirstDataFrame pulls the payload from the first `data:` line of an
// SSE-formatted body.
func extractFirstDataFrame(body []byte) ([]byte, bool) {
	for _, line := range bytes.Split(body, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("data:")) {
			return bytes.TrimSpace(line[len("data:"):]), true
		}
	}
	return nil, false
}

func isJSON(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "application/json")
}

func classifierErrString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

// responseRecorder buffers the downstream handler's response so we can
// rewrite the body before sending it to the real client.
type responseRecorder struct {
	parent http.ResponseWriter
	header http.Header
	buffer *bytes.Buffer
	status int
	done   bool
}

func newResponseRecorder(parent http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		parent: parent,
		header: make(http.Header),
		buffer: &bytes.Buffer{},
	}
}

func (r *responseRecorder) Header() http.Header { return r.header }

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.buffer.Write(b)
	if err != nil {
		return n, fmt.Errorf("buffer response: %w", err)
	}
	return n, nil
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.done {
		return
	}
	r.status = status
	r.done = true
}

func (r *responseRecorder) contentType() string {
	return r.header.Get("Content-Type")
}

// flushTo writes the buffered response to the real writer unchanged.
func (r *responseRecorder) flushTo(w http.ResponseWriter) {
	for key, values := range r.header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if r.status == 0 {
		r.status = http.StatusOK
	}
	w.WriteHeader(r.status)
	_, _ = io.Copy(w, r.buffer)
}
