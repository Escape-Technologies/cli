package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMethodLabelBoundsUnexpectedMethods(t *testing.T) {
	known := map[string]bool{
		http.MethodGet:     true,
		http.MethodPost:    true,
		http.MethodDelete:  true,
		http.MethodOptions: true,
		http.MethodPut:     true,
		http.MethodPatch:   true,
		http.MethodHead:    true,
	}
	for method := range known {
		if got := methodLabel(method); got != method {
			t.Fatalf("method %q should pass through, got %q", method, got)
		}
	}

	for _, method := range []string{"FOO1", "FOO2", "CONNECT", "TRACE"} {
		if got := methodLabel(method); got != "other" {
			t.Fatalf("unexpected method %q should collapse to other, got %q", method, got)
		}
	}
}

func resetSessions(t *testing.T) {
	t.Helper()
	activeSessionTracker.mu.Lock()
	activeSessionTracker.last = map[string]time.Time{}
	activeSessionTracker.mu.Unlock()
}

func TestWrapWithMetricsCountsSessionsOnlyForAuthenticatedRequests(t *testing.T) {
	resetSessions(t)

	handler := wrapWithMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	unauthenticated := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
	handler.ServeHTTP(httptest.NewRecorder(), unauthenticated)
	if count := activeSessionTracker.Count(); count != 0 {
		t.Fatalf("unauthenticated request must not count as a session, got %v", count)
	}

	authenticated := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
	authenticated.Header.Set("X-ESCAPE-API-KEY", "smoke-key")
	handler.ServeHTTP(httptest.NewRecorder(), authenticated)
	if count := activeSessionTracker.Count(); count != 1 {
		t.Fatalf("authenticated request must count as one session, got %v", count)
	}
}

func TestWrapWithMetricsSkipsSessionOnUnauthorized(t *testing.T) {
	resetSessions(t)

	handler := wrapWithMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
	request.Header.Set("X-ESCAPE-API-KEY", "revoked-key")
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if count := activeSessionTracker.Count(); count != 0 {
		t.Fatalf("401 response must not count as a session, got %v", count)
	}
}

func TestSessionTrackerCountExpiresWithoutTraffic(t *testing.T) {
	resetSessions(t)

	activeSessionTracker.touch("live-key")
	activeSessionTracker.mu.Lock()
	activeSessionTracker.last["stale-key"] = time.Now().Add(-2 * sessionWindow)
	activeSessionTracker.mu.Unlock()

	if count := activeSessionTracker.Count(); count != 1 {
		t.Fatalf("expired fingerprint must not count, got %v", count)
	}

	if _, exists := activeSessionTracker.last["stale-key"]; exists {
		t.Fatal("Count must prune expired fingerprints")
	}
}
