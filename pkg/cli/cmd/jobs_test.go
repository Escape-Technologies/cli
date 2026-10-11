package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

func jobWithStatus(status string) *v3.GetJob200Response {
	return v3.NewGetJob200Response(
		"00000000-0000-0000-0000-000000000001",
		v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS(status),
		v3.ENUMPROPERTIESDATAITEMSPROPERTIESKIND("EXPORT"),
		"2026-10-10T10:00:00Z",
		0,
		[]string{},
		[]v3.JobArtefact{},
	)
}

func newWatchTestCommand() *cobra.Command {
	command := &cobra.Command{Use: "jobs get"}
	command.SetContext(context.Background())

	return command
}

// watchJobFunc builds a status source that answers calls in order and fails the test
// when the watch polls more often than the fixture has statuses.
func watchJobFunc(t *testing.T, statuses ...string) jobStatusFn {
	t.Helper()

	calls := 0

	return func(_ context.Context, _ string) (*v3.GetJob200Response, error) {
		if calls >= len(statuses) {
			t.Fatalf("watch polled %d times, fixture has %d statuses", calls, len(statuses))
		}

		status := statuses[calls]
		calls++

		return jobWithStatus(status), nil
	}
}

// TestJobsListPrintsDeclaredDocument checks the new list command walks the
// jobs page and prints the summary table the API returns.
func TestJobsListPrintsDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/jobs" {
			http.NotFound(w, r)

			return
		}

		if got := r.URL.Query().Get("statuses"); got != "FAILED" {
			t.Errorf("statuses = %q, want FAILED", got)
		}

		if got := r.URL.Query().Get("sortDirection"); got != "desc" {
			t.Errorf("sortDirection = %q, want desc (the CLI lists most recent first)", got)
		}

		writeJSON(t, w, map[string]any{"data": []map[string]any{{
			"id":        "00000000-0000-0000-0000-000000000001",
			"status":    "FAILED",
			"kind":      "SAAS_EXPORT_TEMPLATE",
			"name":      "Weekly report",
			"createdAt": "2026-10-10T10:00:00Z",
			"error":     map[string]any{"message": "scan not found"},
			"retries":   0,
			"retriedAt": []string{},
		}}, "totalCount": 1})
	})

	jobsListStatuses = []string{"FAILED"}
	t.Cleanup(func() { jobsListStatuses = nil })

	jobsListCmd.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return jobsListCmd.RunE(jobsListCmd, nil)
	})
	if err != nil {
		t.Fatalf("list: %v\nstderr: %s", err, stderr)
	}

	var jobs []v3.JobSummary
	mustUnmarshalOne(t, stdout, &jobs)
	if len(jobs) != 1 || jobs[0].GetStatus() != "FAILED" {
		t.Fatalf("jobs = %#v", jobs)
	}

	errorMessage, ok := jobs[0].GetErrorOk()
	if !ok || errorMessage.GetMessage() != "scan not found" {
		t.Fatalf("error = %#v", jobs[0].GetError())
	}
}

func TestRunJobWatchEndsOnCompleted(t *testing.T) {
	previousStream, previousTimeout, previousInterval := jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval
	t.Cleanup(func() {
		jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval = previousStream, previousTimeout, previousInterval
	})
	jobsWatchStream = false
	jobsWatchTimeout = time.Minute
	defaultJobWatchPollInterval = time.Millisecond

	command := newWatchTestCommand()
	getJob := watchJobFunc(t, "PENDING", "RUNNING", "COMPLETED")

	output := captureStdout(t, func() {
		if err := runJobWatch(command, "job-1", getJob); err != nil {
			t.Errorf("watch ended with error: %v", err)
		}
	})

	if count := strings.Count(output, "COMPLETED"); count != 1 {
		t.Fatalf("expected one final state, got %d:\n%s", count, output)
	}

	if strings.Contains(output, "PENDING") {
		t.Errorf("watch printed a non-final state:\n%s", output)
	}
}

func TestRunJobWatchFailsOnAFailedJob(t *testing.T) {
	previousStream, previousTimeout, previousInterval := jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval
	t.Cleanup(func() {
		jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval = previousStream, previousTimeout, previousInterval
	})
	jobsWatchStream = false
	jobsWatchTimeout = time.Minute
	defaultJobWatchPollInterval = time.Millisecond

	command := newWatchTestCommand()
	getJob := watchJobFunc(t, "FAILED")

	err := runJobWatch(command, "job-1", getJob)
	if err == nil || !strings.Contains(err.Error(), "FAILED") {
		t.Fatalf("expected an error naming the status, got %v", err)
	}
}

func TestRunJobWatchStopsOnTimeout(t *testing.T) {
	previousStream, previousTimeout, previousInterval := jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval
	t.Cleanup(func() {
		jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval = previousStream, previousTimeout, previousInterval
	})
	jobsWatchStream = false
	jobsWatchTimeout = 10 * time.Millisecond
	defaultJobWatchPollInterval = time.Millisecond

	command := newWatchTestCommand()

	// A job that never ends: only the deadline stops the watch.
	stuck := func(_ context.Context, _ string) (*v3.GetJob200Response, error) {
		return jobWithStatus("RUNNING"), nil
	}

	err := runJobWatch(command, "job-1", stuck)
	if err == nil || !strings.Contains(err.Error(), "still RUNNING") {
		t.Fatalf("expected a timeout error carrying the last status, got %v", err)
	}
}

func TestRunJobWatchStreamKeepsEveryPoll(t *testing.T) {
	previousStream, previousTimeout, previousInterval := jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval
	t.Cleanup(func() {
		jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval = previousStream, previousTimeout, previousInterval
	})
	jobsWatchStream = true
	jobsWatchTimeout = time.Minute
	defaultJobWatchPollInterval = time.Millisecond

	command := newWatchTestCommand()
	getJob := watchJobFunc(t, "PENDING", "COMPLETED")

	output := captureStdout(t, func() {
		if err := runJobWatch(command, "job-1", getJob); err != nil {
			t.Errorf("watch ended with error: %v", err)
		}
	})

	for _, status := range []string{"PENDING", "COMPLETED"} {
		if !strings.Contains(output, status) {
			t.Errorf("stream mode did not print %s:\n%s", status, output)
		}
	}
}

func TestRunJobWatchRejectsANegativeTimeout(t *testing.T) {
	previousTimeout := jobsWatchTimeout
	t.Cleanup(func() { jobsWatchTimeout = previousTimeout })
	jobsWatchTimeout = -time.Second

	called := false
	getJob := func(_ context.Context, _ string) (*v3.GetJob200Response, error) {
		called = true
		return jobWithStatus("RUNNING"), nil
	}

	err := runJobWatch(newWatchTestCommand(), "job-1", getJob)
	if err == nil || !strings.Contains(err.Error(), "invalid --timeout") {
		t.Fatalf("expected a timeout validation error, got %v", err)
	}

	if called {
		t.Error("a negative timeout must fail before any poll")
	}
}

func TestRunJobWatchReportsCancellationNotTimeout(t *testing.T) {
	previousStream, previousTimeout, previousInterval := jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval
	t.Cleanup(func() {
		jobsWatchStream, jobsWatchTimeout, defaultJobWatchPollInterval = previousStream, previousTimeout, previousInterval
	})
	jobsWatchStream = false
	jobsWatchTimeout = time.Hour
	defaultJobWatchPollInterval = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	command := newWatchTestCommand()
	command.SetContext(ctx)

	running := func(_ context.Context, _ string) (*v3.GetJob200Response, error) {
		cancel()
		return jobWithStatus("RUNNING"), nil
	}

	err := runJobWatch(command, "job-1", running)
	if err == nil || !strings.Contains(err.Error(), "canceled") || strings.Contains(err.Error(), "after") {
		t.Fatalf("expected a cancellation error, got %v", err)
	}
}

// TestTriggerExportForwardsScopeFlags pins the export request the CLI builds:
// the scope flags used to be missing, so a caller could only export the whole
// organization.
func TestTriggerExportForwardsScopeFlags(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/jobs" {
			http.NotFound(w, r)

			return
		}

		data, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(data, &body); err != nil {
			t.Errorf("body is not JSON: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"00000000-0000-0000-0000-000000000001","status":"PENDING"}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	command := jobsTriggerExportCmd
	previousBlocks, previousScanID, previousName := jobsExportBlocks, jobsExportScanID, jobsExportName
	previousAsset, previousProfile, previousIssue := jobsExportAssetWhere, jobsExportProfileWhere, jobsExportIssueWhere
	previousDry, previousNoNotify := jobsExportDryRun, jobsExportNoNotify
	t.Cleanup(func() {
		jobsExportBlocks, jobsExportScanID, jobsExportName = previousBlocks, previousScanID, previousName
		jobsExportAssetWhere, jobsExportProfileWhere, jobsExportIssueWhere = previousAsset, previousProfile, previousIssue
		jobsExportDryRun, jobsExportNoNotify = previousDry, previousNoNotify
	})
	jobsExportBlocks = []string{"ISSUE_LIST"}
	jobsExportScanID = "00000000-0000-0000-0000-000000000002"
	jobsExportName = "Weekly digest"
	jobsExportAssetWhere = `{"statuses":["MONITORED"]}`
	jobsExportProfileWhere = `{"name":"prod"}`
	jobsExportIssueWhere = `{"severity":"HIGH"}`
	jobsExportDryRun = true
	jobsExportNoNotify = true

	command.SetIn(strings.NewReader(""))
	command.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return command.RunE(command, nil)
	})
	if err != nil {
		t.Fatalf("trigger export: %v\nstderr: %s", err, stderr)
	}

	if jobID := strings.TrimSpace(stdout); !strings.Contains(jobID, "00000000-0000-0000-0000-000000000001") {
		t.Fatalf("stdout = %s, want the created job id", jobID)
	}

	blocks, _ := body["blocks"].([]any)
	if len(blocks) != 1 {
		t.Fatalf("blocks = %#v", body["blocks"])
	}

	if block, _ := blocks[0].(map[string]any); block["kind"] != "ISSUE_LIST" {
		t.Fatalf("block = %#v", blocks[0])
	}

	for name, want := range map[string]any{
		"scanId":       "00000000-0000-0000-0000-000000000002",
		"name":         "Weekly digest",
		"dry":          true,
		"notify":       false,
		"assetWhere":   map[string]any{"statuses": []any{"MONITORED"}},
		"profileWhere": map[string]any{"name": "prod"},
		"issueWhere":   map[string]any{"severity": "HIGH"},
	} {
		got, ok := body[name]
		if !ok {
			t.Fatalf("export request is missing %s: %#v", name, body)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s = %#v, want %#v", name, got, want)
		}
	}
}

func TestParseExportWhereFlags(t *testing.T) {
	previousAsset, previousProfile, previousIssue := jobsExportAssetWhere, jobsExportProfileWhere, jobsExportIssueWhere
	t.Cleanup(func() {
		jobsExportAssetWhere, jobsExportProfileWhere, jobsExportIssueWhere = previousAsset, previousProfile, previousIssue
	})

	jobsExportAssetWhere = `{"statuses":["MONITORED"]}`
	jobsExportProfileWhere = "not-json"
	jobsExportIssueWhere = ""

	_, err := parseExportWhereFlags(jobsTriggerExportCmd)
	if err == nil || !strings.Contains(err.Error(), "profile-where") {
		t.Fatalf("err = %v, want a profile-where JSON error", err)
	}

	jobsExportProfileWhere = ""
	if _, err := parseExportWhereFlags(jobsTriggerExportCmd); err != nil {
		t.Fatalf("valid flags: %v", err)
	}

	jobsExportAssetWhere = `["not","an","object"]`
	if _, err := parseExportWhereFlags(jobsTriggerExportCmd); err == nil || !strings.Contains(err.Error(), "asset-where") {
		t.Fatalf("err = %v, want an asset-where object error", err)
	}

	// An explicitly blank flag (an empty shell expansion) must not widen the export.
	jobsExportAssetWhere = ""
	if err := jobsTriggerExportCmd.Flags().Set("asset-where", ""); err != nil {
		t.Fatalf("set flag: %v", err)
	}

	t.Cleanup(func() { jobsTriggerExportCmd.Flags().Lookup("asset-where").Changed = false })

	if _, err := parseExportWhereFlags(jobsTriggerExportCmd); err == nil || !strings.Contains(err.Error(), "must not be blank") {
		t.Fatalf("err = %v, want a blank asset-where error", err)
	}
}
