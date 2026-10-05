package escape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

const watchTestScanID = "00000000-0000-0000-0000-000000000010"

func shortenWatchCadence(t *testing.T) {
	t.Helper()
	previousInterval := watchScanInterval
	previousTries := watchScanMaxTries
	watchScanInterval = time.Millisecond
	watchScanMaxTries = 2
	t.Cleanup(func() {
		watchScanInterval = previousInterval
		watchScanMaxTries = previousTries
	})
}

func startWatchServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
}

// scanDocument is a full StartScan200Response body: the client rejects
// bodies missing required properties. progressRatio stays 0 so unchanged
// progress never emits an event, only terminal statuses do.
func scanDocument(status string) string {
	return `{
		"id":"` + watchTestScanID + `",
		"status":"` + status + `",
		"createdAt":"2026-01-01T00:00:00Z",
		"updatedAt":"2026-01-01T00:01:00Z",
		"duration":1,
		"progressRatio":0,
		"initiator":"MANUAL",
		"kind":"BLST_REST",
		"profileId":"00000000-0000-0000-0000-000000000001",
		"organizationId":"00000000-0000-0000-0000-000000000002",
		"links":{"scanIssues":"https://example.com/issues"}
	}`
}

func watchScanEvents(t *testing.T, handler http.HandlerFunc) []*v3.StartScan200Response {
	t.Helper()
	shortenWatchCadence(t)
	startWatchServer(t, handler)
	ch, err := WatchScan(context.Background(), watchTestScanID)
	if err != nil {
		t.Fatalf("watch scan: %v", err)
	}

	var events []*v3.StartScan200Response
	for event := range ch {
		events = append(events, event)
	}

	return events
}

func TestWatchScanKeepsPollingUntilATerminalStatus(t *testing.T) {
	statuses := []string{"PENDING", "PENDING", "FINISHED"}
	events := watchScanEvents(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/scans/"+watchTestScanID) {
			http.NotFound(w, r)
			return
		}

		status := statuses[0]
		if len(statuses) > 1 {
			statuses = statuses[1:]
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(scanDocument(status)))
	})
	if len(events) != 1 {
		t.Fatalf("events = %v, want only the terminal status", events)
	}

	if events[0].Status != "FINISHED" {
		t.Fatalf("status = %q, want FINISHED", events[0].Status)
	}
}

func TestWatchScanClosesAnEmptyStreamWhenPollingKeepsFailing(t *testing.T) {
	events := watchScanEvents(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream down", http.StatusInternalServerError)
	})
	if len(events) != 0 {
		t.Fatalf("events = %v, want an empty stream", events)
	}
}
