package mcp

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Escape-Technologies/cli/pkg/metrics"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// Tool call budgets reach 15 minutes (SchemaAssetCreateToolExecutionTimeout),
// so the bucket set extends far past the default request buckets.
var toolCallBuckets = []float64{
	0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300, 900,
}

// sessionWindow is how long a credential fingerprint keeps counting as an
// active session after its last request. Stateless HTTP means there is no
// server-side session to poll, so activity recency is the honest proxy.
const sessionWindow = 5 * time.Minute

// mcpRoute is the only route these metrics observe; hardcoding it bounds the
// path label instead of echoing arbitrary request paths.
const mcpRoute = "/mcp"

var (
	httpRequestsTotal = metrics.NewCounter(
		"http_requests_total",
		"method", "path", "status",
	)
	httpRequestLatency = metrics.NewHistogram(
		"http_request_latency_seconds",
		nil,
		"method", "path", "status",
	)
	toolCallsTotal = metrics.NewCounter(
		"mcp_tool_calls_total",
		"tool", "result",
	)
	toolCallLatency = metrics.NewHistogram(
		"mcp_tool_call_latency_seconds",
		toolCallBuckets,
	)
	authFailuresTotal = metrics.NewCounter(
		"mcp_auth_failures_total",
		"reason",
	)
	inflightRequests = metrics.NewGauge("mcp_inflight_requests")
)

// wrapWithMetrics records HTTP request count, latency, in-flight concurrency,
// and active sessions around the MCP handler.
func wrapWithMetrics(next http.Handler) http.Handler {
	// Emit the in-flight gauge at zero before the first request so an
	// idle-but-healthy server still reports a series. Active sessions need no
	// zero-init: the computed gauge always renders a series.
	inflightRequests.Set(0)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		inflightRequests.Inc()
		defer inflightRequests.Dec()

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, req)

		// Touch sessions only after the handler ran: the auth gate above
		// rejects unauthenticated requests with 401, and counting those
		// fingerprints would let an unauthenticated caller inflate
		// mcp_active_sessions with random keys.
		if recorder.status != http.StatusUnauthorized {
			if fingerprint := credentialFingerprint(req); fingerprint != "" {
				activeSessionTracker.touch(fingerprint)
			}
		}

		method := methodLabel(req.Method)
		status := strconv.Itoa(recorder.status)
		httpRequestsTotal.Inc(method, mcpRoute, status)
		httpRequestLatency.Observe(time.Since(start).Seconds(), method, mcpRoute, status)
	})
}

// mcp_active_sessions recomputes the active-session count at every render, so
// the gauge decays once traffic stops instead of freezing at its last touched
// value. NewGaugeFunc registers the family as a side effect.
func init() {
	metrics.NewGaugeFunc("mcp_active_sessions", activeSessionTracker.Count)
}

// methodLabel collapses unexpected HTTP methods into one label value. Go
// accepts any HTTP token as a method, and wrapWithMetrics runs outside auth,
// so echoing req.Method would let a client create unbounded label values.
func methodLabel(method string) string {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions, http.MethodPut, http.MethodPatch, http.MethodHead:
		return method
	default:
		return "other"
	}
}

// sessionTracker counts credential fingerprints seen inside sessionWindow.
// The map is capped at maxSessions so even authenticated callers cannot grow
// the tracker without bound. ponytail: if active credentials ever grow past
// the cap, replace the cap with a periodic sweep instead of raising it
// silently.
type sessionTracker struct {
	mu   sync.Mutex
	last map[string]time.Time
}

// maxSessions bounds how many fingerprints the tracker holds at once.
const maxSessions = 10_000

// activeSessionTracker is the process-wide session state. One MCP server runs
// per process, so a singleton keeps the computed gauge in sync with the
// middleware that touches it.
var activeSessionTracker = newSessionTracker()

func newSessionTracker() *sessionTracker {
	return &sessionTracker{last: map[string]time.Time{}}
}

func (t *sessionTracker) touch(fingerprint string) {
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.last[fingerprint]; !exists && len(t.last) >= maxSessions {
		// At the cap, evict expired entries once instead of scanning on
		// every touch. A full tracker of live sessions drops new ones.
		for id, seen := range t.last {
			if now.Sub(seen) >= sessionWindow {
				delete(t.last, id)
			}
		}

		if len(t.last) >= maxSessions {
			return
		}
	}

	t.last[fingerprint] = now
}

// Count prunes expired fingerprints and returns the active count. It runs at
// metric render time, so the gauge expires sessions without requiring more
// traffic.
func (t *sessionTracker) Count() float64 {
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()
	for id, seen := range t.last {
		if now.Sub(seen) >= sessionWindow {
			delete(t.last, id)
		}
	}

	return float64(len(t.last))
}

// registerTool adds one tool to the server with call count and latency
// attached. Every tool registration goes through here — built-in, knowledge,
// public-API, and CLI-backed — so one definition covers the whole catalog.
func registerTool(server *mcpserver.MCPServer, tool mcpgo.Tool, handler mcpserver.ToolHandlerFunc) {
	toolName := tool.Name
	server.AddTool(tool, func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		start := time.Now()

		result, err := handler(ctx, request)

		outcome := "ok"
		if err != nil || (result != nil && result.IsError) {
			outcome = "error"
		}

		toolCallsTotal.Inc(toolName, outcome)
		toolCallLatency.Observe(time.Since(start).Seconds())

		return result, err
	})
}

// recordAuthFailure counts a request rejected at the auth gate or inside a
// tool handler because no usable credential was presented.
func recordAuthFailure(reason string) {
	authFailuresTotal.Inc(reason)
}
