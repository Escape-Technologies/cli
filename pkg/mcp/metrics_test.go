package mcp

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Escape-Technologies/cli/pkg/metrics"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
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

func TestToolErrorTextPrefersErrThenContent(t *testing.T) {
	errText := toolErrorText(errors.New("Insufficient permissions"), nil)
	if errText != "Insufficient permissions" {
		t.Fatalf("err must win over result content, got %q", errText)
	}

	content := mcpgo.NewToolResultError("boom: issue not found")
	if text := toolErrorText(nil, content); text != "boom: issue not found" {
		t.Fatalf("IsError result text must be extracted, got %q", text)
	}

	if text := toolErrorText(nil, nil); text != "" {
		t.Fatalf("no err and no result must stay empty, got %q", text)
	}

	long := strings.Repeat("x", maxLoggedToolError+100)
	if got := toolErrorText(errors.New(long), nil); len(got) != maxLoggedToolError {
		t.Fatalf("error text must be capped at %d, got %d", maxLoggedToolError, len(got))
	}
}

func TestRegisterToolCountsAndLogsError(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	server := mcpserver.NewMCPServer("test", "0.0.0", mcpserver.WithToolCapabilities(false))
	registerTool(server, mcpgo.NewTool("error_probe"), func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return mcpgo.NewToolResultError("boom: Insufficient permissions"), nil
	})

	registration := server.GetTool("error_probe")
	if registration == nil {
		t.Fatal("tool must be registered")
	}

	result, err := registration.Handler(t.Context(), mcpgo.CallToolRequest{})
	if err != nil {
		t.Fatalf("IsError results must not surface as Go errors: %v", err)
	}

	if result == nil || !result.IsError {
		t.Fatal("handler error result must carry IsError")
	}

	var rendered bytes.Buffer
	metrics.WriteText(&rendered)
	if !strings.Contains(rendered.String(), `,tool="error_probe",result="error"}`) {
		t.Fatalf("error call must be counted, got:\n%s", rendered.String())
	}

	if !strings.Contains(logs.String(), `tool=error_probe`) || !strings.Contains(logs.String(), `boom: Insufficient permissions`) {
		t.Fatalf("failed tool call must be logged with tool and error, got:\n%s", logs.String())
	}
}

func TestRegisterToolCountsOkWithoutLog(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	server := mcpserver.NewMCPServer("test", "0.0.0", mcpserver.WithToolCapabilities(false))
	registerTool(server, mcpgo.NewTool("ok_probe"), func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		return mcpgo.NewToolResultText("fine"), nil
	})

	registration := server.GetTool("ok_probe")
	if registration == nil {
		t.Fatal("tool must be registered")
	}

	if _, err := registration.Handler(t.Context(), mcpgo.CallToolRequest{}); err != nil {
		t.Fatalf("ok call must not error: %v", err)
	}

	var rendered bytes.Buffer
	metrics.WriteText(&rendered)
	if !strings.Contains(rendered.String(), `,tool="ok_probe",result="ok"}`) {
		t.Fatalf("ok call must be counted, got:\n%s", rendered.String())
	}

	if strings.Contains(logs.String(), `ok_probe`) {
		t.Fatalf("ok call must not log, got:\n%s", logs.String())
	}
}
