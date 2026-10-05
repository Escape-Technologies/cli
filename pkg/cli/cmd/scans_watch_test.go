package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const scanID = "00000000-0000-0000-0000-000000000010"

func scanDocument(status string) string {
	return `{
		"id":"` + scanID + `",
		"status":"` + status + `",
		"createdAt":"2026-01-01T00:00:00Z",
		"updatedAt":"2026-01-01T00:01:00Z",
		"duration":1,
		"progressRatio":1,
		"initiator":"MANUAL",
		"kind":"BLST_REST",
		"profileId":"00000000-0000-0000-0000-000000000001",
		"organizationId":"00000000-0000-0000-0000-000000000002",
		"links":{"scanIssues":"https://example.com/issues"}
	}`
}

func TestStartWatchJSONPrintsOnlyTheFinalScan(t *testing.T) {
	var sawIssues bool
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/scans"):
			writeJSON(t, w, json.RawMessage(scanDocument("STARTING")))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/scans/"+scanID):
			writeJSON(t, w, json.RawMessage(scanDocument("FINISHED")))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/issues"):
			sawIssues = true
			writeJSON(t, w, map[string]any{"data": []any{}})
		default:
			http.NotFound(w, r)
		}
	})

	previous := scanStartCmdWatch
	if err := scanStartCmd.Flags().Set("watch", "true"); err != nil {
		t.Fatalf("set watch: %v", err)
	}

	t.Cleanup(func() {
		_ = scanStartCmd.Flags().Set("watch", "false")
		scanStartCmdWatch = previous
	})

	scanStartCmd.SetContext(context.Background())
	stdout, stderr, err := captureJSONCommand(t, func() error {
		return scanStartCmd.RunE(scanStartCmd, []string{"00000000-0000-0000-0000-000000000001"})
	})
	if err != nil {
		t.Fatalf("start --watch: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	if sawIssues {
		t.Fatal("scans start --watch -o json must not print the issue list")
	}

	dec := json.NewDecoder(strings.NewReader(stdout))
	var scan map[string]any
	if err := dec.Decode(&scan); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}

	if scan["status"] != "FINISHED" {
		t.Fatalf("status = %#v, want the finished scan", scan["status"])
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout has a second document (%v): %s", err, stdout)
	}
}

func TestWatchJSONPrintsTheIssueListAfterTheFinalScan(t *testing.T) {
	var sawIssues bool
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/issues") {
			sawIssues = true
			writeJSON(t, w, map[string]any{"data": []any{}})

			return
		}

		http.NotFound(w, r)
	})

	var status v3.StartScan200Response
	if err := json.Unmarshal([]byte(scanDocument("FINISHED")), &status); err != nil {
		t.Fatalf("scan fixture: %v", err)
	}

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return finishWatch(context.Background(), scanID, &status, watchJSONStatusAndIssues)
	})
	if err != nil {
		t.Fatalf("watch json: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	if !sawIssues {
		t.Fatal("scans watch -o json did not request the issue list")
	}

	dec := json.NewDecoder(strings.NewReader(stdout))
	var scan map[string]any
	if err := dec.Decode(&scan); err != nil {
		t.Fatalf("final scan: %v\n%s", err, stdout)
	}

	if scan["status"] != "FINISHED" {
		t.Fatalf("status = %#v", scan["status"])
	}

	var issues []any
	if err := dec.Decode(&issues); err != nil {
		t.Fatalf("issue list: %v\n%s", err, stdout)
	}

	if len(issues) != 0 {
		t.Fatalf("issues = %#v", issues)
	}

	if err := dec.Decode(&scan); err != io.EOF {
		t.Fatalf("stdout has a third document (%v): %s", err, stdout)
	}
}

func TestWatchJSONSkipsIssuesWhenTheScanFailed(t *testing.T) {
	assertWatchTerminal(t, "json", "FAILED", watchJSONStatusAndIssues, fmt.Sprintf("scan %s failed", scanID))
}

func TestWatchJSONCanceledScanReturnsOneDocument(t *testing.T) {
	assertWatchTerminal(t, "json", "CANCELED", watchJSONStatusAndIssues, fmt.Sprintf("scan %s was canceled", scanID))
}

func TestStartWatchJSONFailedScanReturnsOneDocument(t *testing.T) {
	assertWatchTerminal(t, "json", "FAILED", watchJSONStatus, fmt.Sprintf("scan %s failed", scanID))
}

func TestStartWatchJSONCanceledScanReturnsOneDocument(t *testing.T) {
	assertWatchTerminal(t, "json", "CANCELED", watchJSONStatus, fmt.Sprintf("scan %s was canceled", scanID))
}

func TestWatchPrettyFailedScanReturnsError(t *testing.T) {
	assertWatchTerminal(t, "pretty", "FAILED", watchJSONStatusAndIssues, fmt.Sprintf("scan %s failed", scanID))
}

func TestWatchPrettyCanceledScanReturnsError(t *testing.T) {
	assertWatchTerminal(t, "pretty", "CANCELED", watchJSONSilent, fmt.Sprintf("scan %s was canceled", scanID))
}

func TestWatchPrettyFinishedReturnsNil(t *testing.T) {
	sawIssues := serveScanIssues(t, nil)
	status := mustScanStatus(t, "FINISHED")
	stdout, stderr, err := captureOutput(t, "pretty", func() error {
		return finishWatch(context.Background(), scanID, status, watchJSONStatusAndIssues)
	})
	if err != nil {
		t.Fatalf("finished pretty watch: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	if !*sawIssues {
		t.Fatal("finished pretty watch did not fetch issues")
	}

	if !strings.Contains(stdout, "Scan completed") {
		t.Fatalf("stdout = %q", stdout)
	}

	if !strings.Contains(stdout, "SEVERITY") {
		t.Fatalf("pretty watch did not print the issue table: %s", stdout)
	}
}

// watchStream builds an already-closed WatchScan stream with the given
// events, so followWatch can be tested without the real polling loop.
func watchStream(t *testing.T, events ...*v3.StartScan200Response) chan *v3.StartScan200Response {
	t.Helper()
	ch := make(chan *v3.StartScan200Response, len(events))
	for _, event := range events {
		ch <- event
	}

	close(ch)

	return ch
}

func TestWatchScanFailsWhenTheStreamEndsOnRunning(t *testing.T) {
	// The API stopped answering mid-watch: WatchScan gives up and closes the
	// channel while the scan is still running. That must fail, not exit 0.
	sawIssues := serveScanIssues(t, nil)
	stream := watchStream(t,
		mustScanStatus(t, "STARTING"),
		mustScanStatus(t, "RUNNING"),
	)
	stdout, stderr, err := captureOutput(t, "json", func() error {
		return followWatch(context.Background(), scanID, stream, watchJSONStatusAndIssues)
	})
	if err == nil || err.Error() != fmt.Sprintf("scan %s did not finish (last status RUNNING)", scanID) {
		t.Fatalf("err = %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	if *sawIssues {
		t.Fatal("unfinished watch fetched issues")
	}

	assertSingleScanDocument(t, stdout, "RUNNING")
}

func TestWatchScanFollowsPendingRunningToFinished(t *testing.T) {
	sawIssues := serveScanIssues(t, nil)
	stream := watchStream(t,
		mustScanStatus(t, "PENDING"),
		mustScanStatus(t, "RUNNING"),
		mustScanStatus(t, "FINISHED"),
	)
	stdout, stderr, err := captureOutput(t, "pretty", func() error {
		return followWatch(context.Background(), scanID, stream, watchJSONStatusAndIssues)
	})
	if err != nil {
		t.Fatalf("pending to finished watch: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	if !*sawIssues {
		t.Fatal("finished watch did not fetch issues")
	}

	for _, want := range []string{"PENDING", "RUNNING", "Scan completed", "SEVERITY"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout %q does not contain %q", stdout, want)
		}
	}
}

// mainError wraps a command error the way cmd.Execute and cli.Run do before
// main.go hands it to out.PrintError. Tests use it to print through the real
// error path instead of an invented message.
func mainError(command string, err error) error {
	return fmt.Errorf("failed to execute command: %w", fmt.Errorf("command %s failed: %w", command, err))
}

func TestWatchJSONFailedScanPrintsTheErrorOnStderr(t *testing.T) {
	status := mustScanStatus(t, "FAILED")
	stdout, stderr, err := captureOutput(t, "json", func() error {
		watchErr := finishWatch(context.Background(), scanID, status, watchJSONStatusAndIssues)
		if watchErr != nil {
			// main does this after Execute returns. stdout must stay one document.
			out.PrintError(mainError(scanWatchCmd.Name(), watchErr))
		}

		return watchErr
	})
	if err == nil || err.Error() != fmt.Sprintf("scan %s failed", scanID) {
		t.Fatalf("err = %v", err)
	}

	for _, want := range []string{"failed to execute command", "command watch failed", fmt.Sprintf("scan %s failed", scanID)} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr %q does not contain %q", stderr, want)
		}
	}

	assertSingleScanDocument(t, stdout, "FAILED")
}

func TestFailOnSeverityTriggersAtTheThreshold(t *testing.T) {
	pages := [][]map[string]any{
		{issueDocument("00000000-0000-0000-0000-0000000000a1", "MEDIUM", "OPEN")},
		{
			issueDocument("00000000-0000-0000-0000-0000000000a2", "HIGH", "OPEN"),
			issueDocument("00000000-0000-0000-0000-0000000000a3", "HIGH", "MANUAL_REVIEW"),
		},
	}
	sawIssues := servePagedIssues(t, pages)
	setFailOnSeverity(t, "high")
	if err := validateScanFailOnSeverity(true); err != nil {
		t.Fatalf("normalize severity: %v", err)
	}

	if scanFailOnSeverity != "HIGH" {
		t.Fatalf("severity = %q, want HIGH", scanFailOnSeverity)
	}

	status := mustScanStatus(t, "FINISHED")
	stdout, stderr, err := captureOutput(t, "pretty", func() error {
		return finishWatch(context.Background(), scanID, status, watchJSONStatusAndIssues)
	})
	if err == nil || err.Error() != "2 issue(s) at or above HIGH" {
		t.Fatalf("err = %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	if !*sawIssues {
		t.Fatal("severity gate did not fetch issues")
	}

	if !strings.Contains(stdout, "00000000-0000-0000-0000-0000000000a2") {
		t.Fatalf("pretty table missing the high issue: %s", stdout)
	}
}

func TestFailOnSeverityDoesNotTriggerBelowTheThreshold(t *testing.T) {
	pages := [][]map[string]any{{
		issueDocument("00000000-0000-0000-0000-0000000000b1", "MEDIUM", "OPEN"),
		issueDocument("00000000-0000-0000-0000-0000000000b2", "CRITICAL", "IGNORED"),
		issueDocument("00000000-0000-0000-0000-0000000000b3", "CRITICAL", "RESOLVED"),
		issueDocument("00000000-0000-0000-0000-0000000000b4", "HIGH", "FALSE_POSITIVE"),
	}}
	servePagedIssues(t, pages)
	setFailOnSeverity(t, "HIGH")

	status := mustScanStatus(t, "FINISHED")
	stdout, stderr, err := captureOutput(t, "json", func() error {
		return finishWatch(context.Background(), scanID, status, watchJSONStatusAndIssues)
	})
	if err != nil {
		t.Fatalf("below threshold: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}

	dec := json.NewDecoder(strings.NewReader(stdout))
	var scan map[string]any
	if err := dec.Decode(&scan); err != nil {
		t.Fatalf("final scan: %v\n%s", err, stdout)
	}

	var issues []any
	if err := dec.Decode(&issues); err != nil {
		t.Fatalf("issue list: %v\n%s", err, stdout)
	}

	if len(issues) != 4 {
		t.Fatalf("issues = %#v", issues)
	}

	if err := dec.Decode(&scan); err != io.EOF {
		t.Fatalf("stdout has a third document (%v): %s", err, stdout)
	}
}

func TestStartWatchJSONFailOnSeverityKeepsOneDocument(t *testing.T) {
	pages := [][]map[string]any{{
		issueDocument("00000000-0000-0000-0000-0000000000c1", "CRITICAL", "OPEN"),
	}}
	servePagedIssues(t, pages)
	setFailOnSeverity(t, "CRITICAL")
	status := mustScanStatus(t, "FINISHED")
	stdout, stderr, err := captureOutput(t, "json", func() error {
		watchErr := finishWatch(context.Background(), scanID, status, watchJSONStatus)
		if watchErr != nil {
			out.PrintError(watchErr)
		}

		return watchErr
	})
	if err == nil || err.Error() != "1 issue(s) at or above CRITICAL" {
		t.Fatalf("err = %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stderr, "1 issue(s) at or above CRITICAL") {
		t.Fatalf("stderr = %q", stderr)
	}

	assertSingleScanDocument(t, stdout, "FINISHED")
}

func TestFailOnSeverityRejectsAnUnknownLevel(t *testing.T) {
	setCommandFailOnSeverity(t, scanWatchCmd, "severe")
	scanWatchCmd.SetContext(context.Background())
	err := scanWatchCmd.RunE(scanWatchCmd, []string{scanID})
	if err == nil || !strings.Contains(err.Error(), "invalid severity") {
		t.Fatalf("err = %v", err)
	}
}

func TestStartFailOnSeverityRequiresWatch(t *testing.T) {
	called := false
	serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		http.Error(w, "unexpected", http.StatusInternalServerError)
	})
	previousWatch := scanStartCmdWatch
	if err := scanStartCmd.Flags().Set("watch", "false"); err != nil {
		t.Fatalf("clear watch: %v", err)
	}

	t.Cleanup(func() {
		_ = scanStartCmd.Flags().Set("watch", "false")
		scanStartCmdWatch = previousWatch
	})
	setCommandFailOnSeverity(t, scanStartCmd, "HIGH")
	scanStartCmd.SetContext(context.Background())
	err := scanStartCmd.RunE(scanStartCmd, []string{"00000000-0000-0000-0000-000000000001"})
	if err == nil || err.Error() != "--fail-on-severity requires --watch" {
		t.Fatalf("err = %v", err)
	}

	if called {
		t.Fatal("start without --watch called the API")
	}
}

func TestStartWatchRejectsAnUnknownSeverityBeforeStarting(t *testing.T) {
	called := false
	serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		http.Error(w, "unexpected", http.StatusInternalServerError)
	})
	previousWatch := scanStartCmdWatch
	if err := scanStartCmd.Flags().Set("watch", "true"); err != nil {
		t.Fatalf("set watch: %v", err)
	}

	t.Cleanup(func() {
		_ = scanStartCmd.Flags().Set("watch", "false")
		scanStartCmdWatch = previousWatch
	})
	setCommandFailOnSeverity(t, scanStartCmd, "nope")
	scanStartCmd.SetContext(context.Background())
	err := scanStartCmd.RunE(scanStartCmd, []string{"00000000-0000-0000-0000-000000000001"})
	if err == nil || !strings.Contains(err.Error(), "invalid severity") {
		t.Fatalf("err = %v", err)
	}

	if called {
		t.Fatal("invalid severity started a scan")
	}
}

func TestIssueSeverityOrderCoversGeneratedEnum(t *testing.T) {
	for _, level := range v3.AllowedENUMPROPERTIESFILTERPROPERTIESSEVERITIESITEMSEnumValues {
		if _, ok := severityRank(level); !ok {
			t.Errorf("severity %s is missing from issueSeverityOrder", level)
		}
	}
}

func assertWatchTerminal(t *testing.T, mode, scanStatus string, jsonResult watchJSONResult, wantErr string) {
	t.Helper()
	sawIssues := serveScanIssues(t, nil)
	status := mustScanStatus(t, scanStatus)
	stdout, stderr, err := captureOutput(t, mode, func() error {
		return finishWatch(context.Background(), scanID, status, jsonResult)
	})
	if err == nil || err.Error() != wantErr {
		t.Fatalf("err = %v, want %s\nstderr: %s\nstdout: %s", err, wantErr, stderr, stdout)
	}

	if *sawIssues {
		t.Fatal("failed or canceled watch fetched issues")
	}

	if mode == "json" {
		assertSingleScanDocument(t, stdout, scanStatus)
		return
	}

	wantLine := "Scan failed"
	if scanStatus == "CANCELED" {
		wantLine = "Scan canceled"
	}

	if !strings.Contains(stdout, wantLine) {
		t.Fatalf("stdout = %q, want %s", stdout, wantLine)
	}
}

func assertSingleScanDocument(t *testing.T, stdout, scanStatus string) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	var scan map[string]any
	if err := dec.Decode(&scan); err != nil {
		t.Fatalf("final scan: %v\n%s", err, stdout)
	}

	if scan["status"] != scanStatus {
		t.Fatalf("status = %#v, want %s", scan["status"], scanStatus)
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout has a second document (%v): %s", err, stdout)
	}
}

func mustScanStatus(t *testing.T, scanStatus string) *v3.StartScan200Response {
	t.Helper()
	var status v3.StartScan200Response
	if err := json.Unmarshal([]byte(scanDocument(scanStatus)), &status); err != nil {
		t.Fatalf("scan fixture: %v", err)
	}

	return &status
}

func serveScanIssues(t *testing.T, issues []map[string]any) *bool {
	t.Helper()
	if issues == nil {
		issues = []map[string]any{}
	}

	return servePagedIssues(t, [][]map[string]any{issues})
}

func servePagedIssues(t *testing.T, pages [][]map[string]any) *bool {
	t.Helper()
	saw := false
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/issues") {
			saw = true
			page := 0
			if r.URL.Query().Get("cursor") != "" {
				page = 1
			}

			if page >= len(pages) {
				t.Errorf("unexpected issues page %d", page)
				writeJSON(t, w, map[string]any{"data": []any{}})

				return
			}

			body := map[string]any{"data": pages[page], "totalCount": 0}
			if page == 0 && len(pages) > 1 {
				body["nextCursor"] = "page-2"
			}

			writeJSON(t, w, body)

			return
		}

		http.NotFound(w, r)
	})

	return &saw
}

func issueDocument(id, severity, status string) map[string]any {
	return map[string]any{
		"id":             id,
		"name":           "issue",
		"fullName":       "issue full",
		"category":       "INJECTION",
		"severity":       severity,
		"manualSeverity": false,
		"status":         status,
		"context":        "",
		"risks":          []any{},
		"alertUid":       "alert",
		"createdAt":      "2026-01-01T00:00:00Z",
		"asset": map[string]any{
			"id":         "00000000-0000-0000-0000-000000000020",
			"class":      "API_SERVICE",
			"type":       "REST",
			"name":       "api",
			"createdAt":  "2026-01-01T00:00:00Z",
			"lastSeenAt": "2026-01-01T00:00:00Z",
			"status":     "MONITORED",
			"tags":       []any{},
			"risks":      []any{},
			"projectIds": []any{},
			"links":      map[string]any{"assetOverview": "https://example.com/asset"},
		},
		"links": map[string]any{"issueOverview": "https://example.com/issue/" + id},
	}
}

func setFailOnSeverity(t *testing.T, value string) {
	t.Helper()
	previous := scanFailOnSeverity
	scanFailOnSeverity = value
	t.Cleanup(func() { scanFailOnSeverity = previous })
}

func setCommandFailOnSeverity(t *testing.T, command *cobra.Command, value string) {
	t.Helper()
	previous := scanFailOnSeverity
	if err := command.Flags().Set("fail-on-severity", value); err != nil {
		t.Fatalf("set fail-on-severity: %v", err)
	}

	t.Cleanup(func() {
		_ = command.Flags().Set("fail-on-severity", "")
		scanFailOnSeverity = previous
	})
}
