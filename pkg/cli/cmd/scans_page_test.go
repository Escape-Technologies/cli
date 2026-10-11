package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const (
	scanRowA   = `{"id":"scan-a","status":"FINISHED","createdAt":"2026-01-01T00:00:00Z","duration":1,"progressRatio":1,"initiator":"MANUAL","kind":"BLST_REST","links":{"scanIssues":"https://app.escape.tech/a"}}`
	scanRowB   = `{"id":"scan-b","status":"FINISHED","createdAt":"2026-01-01T00:00:00Z","duration":1,"progressRatio":1,"initiator":"MANUAL","kind":"BLST_REST","links":{"scanIssues":"https://app.escape.tech/b"}}`
	agentRowA  = `{"id":"agent-a","title":"Coverage","status":"FINISHED"}`
	agentRowB  = `{"id":"agent-b","title":"Exploit","status":"RUNNING"}`
	targetRowA = `{"id":"target-a","createdAt":"2026-01-01T00:00:00Z"}`
	targetRowB = `{"id":"target-b","createdAt":"2026-01-01T00:00:00Z"}`
)

func TestScansListPageForwardsCursorAndReturnsNext(t *testing.T) {
	seen := servePagedJSON(t, "/v3/scans", scanRowA, "page-3", 9)
	setChangedFlag(t, scansListCmd, "size", "20")
	setChangedFlag(t, scansListCmd, "cursor", "page-2")
	previousLimit := scanListLimit
	scanListLimit = 1
	t.Cleanup(func() { scanListLimit = previousLimit })

	stdout := runCmdJSON(t, scansListCmd, nil)
	assertOnePageRequest(t, seen, "page-2", "20")
	assertPageJSON(t, stdout, "scan-a", "page-3", 9)
}

func TestScansListDefaultWalksEveryPage(t *testing.T) {
	resetFlag(t, scansListCmd, "size")
	resetFlag(t, scansListCmd, "cursor")
	previousLimit := scanListLimit
	scanListLimit = 0
	t.Cleanup(func() { scanListLimit = previousLimit })
	previousOutput := rootCmdOutputStr
	rootCmdOutputStr = "json"
	t.Cleanup(func() { rootCmdOutputStr = previousOutput })

	seen := serveTwoPages(t, "/v3/scans", scanRowA, scanRowB)
	stdout := runCmdJSON(t, scansListCmd, nil)
	assertFullWalk(t, seen, stdout, "scan-a", "scan-b")
}

func TestScansAgentsPageForwardsCursorAndReturnsNext(t *testing.T) {
	seen := servePagedJSON(t, "/v3/scans/scan-1/agents", agentRowA, "page-3", 4)
	setChangedFlag(t, scansAgentsCmd, "size", "20")
	setChangedFlag(t, scansAgentsCmd, "cursor", "page-2")

	stdout := runCmdJSON(t, scansAgentsCmd, []string{"scan-1"})
	assertOnePageRequest(t, seen, "page-2", "20")
	assertPageJSON(t, stdout, "agent-a", "page-3", 4)
}

func TestScansAgentsDefaultWalksEveryPage(t *testing.T) {
	resetFlag(t, scansAgentsCmd, "size")
	resetFlag(t, scansAgentsCmd, "cursor")
	seen := serveTwoPages(t, "/v3/scans/scan-1/agents", agentRowA, agentRowB)
	stdout := runCmdJSON(t, scansAgentsCmd, []string{"scan-1"})
	assertFullWalk(t, seen, stdout, "agent-a", "agent-b")
	if !strings.Contains(stdout, `"scanId"`) {
		t.Fatalf("default agents output = %s", stdout)
	}
}

func TestScansTargetsPageForwardsCursorAndReturnsNext(t *testing.T) {
	seen := servePagedJSON(t, "/v3/scans/scan-1/targets", targetRowA, "page-3", 6)
	setChangedFlag(t, scanTargetsCmd, "size", "20")
	setChangedFlag(t, scanTargetsCmd, "cursor", "page-2")

	stdout := runCmdJSON(t, scanTargetsCmd, []string{"scan-1"})
	assertOnePageRequest(t, seen, "page-2", "20")
	assertPageJSON(t, stdout, "target-a", "page-3", 6)
}

func TestScansTargetsDefaultWalksEveryPage(t *testing.T) {
	resetFlag(t, scanTargetsCmd, "size")
	resetFlag(t, scanTargetsCmd, "cursor")
	seen := serveTwoPages(t, "/v3/scans/scan-1/targets", targetRowA, targetRowB)
	stdout := runCmdJSON(t, scanTargetsCmd, []string{"scan-1"})
	assertFullWalk(t, seen, stdout, "target-a", "target-b")
}

func TestScanListingsRejectSizeZeroBeforeCallingTheAPI(t *testing.T) {
	// --size is the page size for scans list and scans agents, so 0 is outside
	// 1..100. scans targets is not in this list: without --cursor, --size 0
	// still means "no total cap" and must walk every page.
	commands := []struct {
		command *cobra.Command
		args    []string
	}{
		{command: scansListCmd, args: nil},
		{command: scansAgentsCmd, args: []string{"scan-1"}},
	}
	for _, item := range commands {
		called := false
		serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
			called = true
			http.NotFound(w, r)
		})
		setChangedFlag(t, item.command, "size", "0")
		resetFlag(t, item.command, "cursor")
		item.command.SetContext(t.Context())
		err := item.command.RunE(item.command, item.args)
		if err == nil || !strings.Contains(err.Error(), "--size must be between 1 and 100") {
			t.Errorf("%s err = %v", item.command.CommandPath(), err)
		}

		if called {
			t.Errorf("%s called the API with --size 0", item.command.CommandPath())
		}
	}
}

func servePagedJSON(t *testing.T, path, row, next string, total int) *[]url.Values {
	t.Helper()
	seen := []url.Values{}
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Errorf("path = %s, want %s", r.URL.Path, path)
			http.NotFound(w, r)

			return
		}

		seen = append(seen, r.URL.Query())
		writeRawJSON(w, pageDocument(row, next, total))
	})

	return &seen
}

func serveTwoPages(t *testing.T, path, first, second string) *[]url.Values {
	t.Helper()
	seen := []url.Values{}
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Errorf("path = %s, want %s", r.URL.Path, path)
			http.NotFound(w, r)

			return
		}

		query := r.URL.Query()
		seen = append(seen, query)
		switch query.Get("cursor") {
		case "":
			writeRawJSON(w, pageDocument(first, "page-2", 2))
		case "page-2":
			writeRawJSON(w, pageDocument(second, "", 2))
		default:
			t.Errorf("cursor = %q", query.Get("cursor"))
			http.NotFound(w, r)
		}
	})

	return &seen
}

func pageDocument(row, next string, total int) string {
	cursor := "null"
	if next != "" {
		cursor = `"` + next + `"`
	}

	return fmt.Sprintf(`{"data":[%s],"nextCursor":%s,"totalCount":%d}`, row, cursor, total)
}

func writeRawJSON(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(body))
}

func assertOnePageRequest(t *testing.T, seen *[]url.Values, cursor, size string) {
	t.Helper()
	if len(*seen) != 1 {
		t.Fatalf("requests = %d, want one page", len(*seen))
	}

	if (*seen)[0].Get("cursor") != cursor || (*seen)[0].Get("size") != size {
		t.Fatalf("query = %v", (*seen)[0])
	}
}

func assertPageJSON(t *testing.T, stdout, id, next string, total int) {
	t.Helper()
	if !strings.Contains(stdout, `"nextCursor":"`+next+`"`) || !strings.Contains(stdout, id) {
		t.Fatalf("page = %s", stdout)
	}

	if !strings.Contains(stdout, fmt.Sprintf(`"totalCount":%d`, total)) && !strings.Contains(stdout, fmt.Sprintf(`"totalCount": %d`, total)) {
		t.Fatalf("page = %s", stdout)
	}
}

func TestScansTargetsSizeWithoutCursorCapsTheArray(t *testing.T) {
	resetFlag(t, scanTargetsCmd, "cursor")
	setChangedFlag(t, scanTargetsCmd, "size", "1")
	seen := serveTwoPages(t, "/v3/scans/scan-1/targets", targetRowA, targetRowB)
	stdout := runCmdJSON(t, scanTargetsCmd, []string{"scan-1"})
	if len(*seen) != 1 || (*seen)[0].Get("size") != "1" {
		t.Fatalf("queries = %#v", *seen)
	}

	if !strings.Contains(stdout, "target-a") || strings.Contains(stdout, "target-b") || strings.Contains(stdout, "nextCursor") {
		t.Fatalf("capped listing = %s", stdout)
	}
}

func TestScansTargetsSizeZeroWalksEveryPage(t *testing.T) {
	resetFlag(t, scanTargetsCmd, "cursor")
	setChangedFlag(t, scanTargetsCmd, "size", "0")
	seen := serveTwoPages(t, "/v3/scans/scan-1/targets", targetRowA, targetRowB)
	stdout := runCmdJSON(t, scanTargetsCmd, []string{"scan-1"})
	assertFullWalk(t, seen, stdout, "target-a", "target-b")
}

func TestScansTargetsEmptyCursorIsTheFirstPage(t *testing.T) {
	resetFlag(t, scanTargetsCmd, "size")
	setChangedFlag(t, scanTargetsCmd, "cursor", "")
	seen := servePagedJSON(t, "/v3/scans/scan-1/targets", targetRowA, "page-2", 2)
	stdout := runCmdJSON(t, scanTargetsCmd, []string{"scan-1"})
	if len(*seen) != 1 || (*seen)[0].Get("cursor") != "" {
		t.Fatalf("queries = %#v", *seen)
	}

	assertPageJSON(t, stdout, "target-a", "page-2", 2)
}

func TestScansTargetsNegativeSizeIsRejected(t *testing.T) {
	resetFlag(t, scanTargetsCmd, "cursor")
	setChangedFlag(t, scanTargetsCmd, "size", "-1")
	called := false
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		http.NotFound(w, r)
	})
	scanTargetsCmd.SetContext(t.Context())
	err := scanTargetsCmd.RunE(scanTargetsCmd, []string{"scan-1"})
	if err == nil || !strings.Contains(err.Error(), "--size must be greater than or equal to 0") {
		t.Fatalf("err = %v", err)
	}

	if called {
		t.Fatal("negative --size called the API")
	}
}

func TestScansListLimitCapsWithoutChangingTheArrayShape(t *testing.T) {
	resetFlag(t, scansListCmd, "size")
	resetFlag(t, scansListCmd, "cursor")
	previousLimit := scanListLimit
	scanListLimit = 1
	t.Cleanup(func() { scanListLimit = previousLimit })
	seen := serveTwoPages(t, "/v3/scans", scanRowA, scanRowB)
	stdout := runCmdJSON(t, scansListCmd, nil)
	if len(*seen) != 1 {
		t.Fatalf("requests = %d, want the cap to stop after the first page", len(*seen))
	}

	if !strings.Contains(stdout, "scan-a") || strings.Contains(stdout, "scan-b") || strings.Contains(stdout, "nextCursor") {
		t.Fatalf("limited listing = %s", stdout)
	}
}

func TestScansAgentsFiltersKeepTheScanAgentsShape(t *testing.T) {
	resetFlag(t, scansAgentsCmd, "size")
	resetFlag(t, scansAgentsCmd, "cursor")
	setChangedFlag(t, scansAgentsCmd, "search", "Coverage")
	setChangedFlag(t, scansAgentsCmd, "roots-only", "true")
	seen := serveTwoPages(t, "/v3/scans/scan-1/agents", agentRowA, agentRowB)
	stdout := runCmdJSON(t, scansAgentsCmd, []string{"scan-1"})
	assertFullWalk(t, seen, stdout, "agent-a", "agent-b")
	if !strings.Contains(stdout, `"scanId"`) {
		t.Fatalf("filtered agents output = %s", stdout)
	}
}

func TestScanListingSchemasMatchTheMode(t *testing.T) {
	resetFlag(t, scansListCmd, "size")
	resetFlag(t, scansListCmd, "cursor")
	resetFlag(t, scanTargetsCmd, "size")
	resetFlag(t, scanTargetsCmd, "cursor")
	resetFlag(t, scansAgentsCmd, "size")
	resetFlag(t, scansAgentsCmd, "cursor")

	assertSchema := func(command *cobra.Command, args []string, want []string, reject []string) {
		t.Helper()
		stdout, stderr, err := captureOutput(t, "schema", func() error {
			command.SetContext(t.Context())

			return command.RunE(command, args)
		})
		if err != nil {
			t.Fatalf("%s schema: %v\n%s", command.CommandPath(), err, stderr)
		}

		for _, field := range want {
			if !strings.Contains(stdout, field) {
				t.Errorf("%s schema missing %s: %s", command.CommandPath(), field, stdout)
			}
		}

		for _, field := range reject {
			if strings.Contains(stdout, field) {
				t.Errorf("%s schema contains %s: %s", command.CommandPath(), field, stdout)
			}
		}
	}

	assertSchema(scansListCmd, nil, []string{`"type": "array"`}, []string{`"nextCursor"`})
	assertSchema(scanTargetsCmd, []string{"scan-1"}, []string{`"type": "array"`}, []string{`"nextCursor"`})
	assertSchema(scansAgentsCmd, []string{"scan-1"}, []string{`"scanId"`, `"agents"`}, []string{`"nextCursor"`})

	setChangedFlag(t, scansListCmd, "size", "20")
	setChangedFlag(t, scanTargetsCmd, "cursor", "")
	setChangedFlag(t, scansAgentsCmd, "cursor", "page-2")
	assertSchema(scansListCmd, nil, []string{`"nextCursor"`, `"totalCount"`, `"items"`}, nil)
	assertSchema(scanTargetsCmd, []string{"scan-1"}, []string{`"nextCursor"`, `"totalCount"`, `"items"`}, nil)
	assertSchema(scansAgentsCmd, []string{"scan-1"}, []string{`"nextCursor"`, `"totalCount"`, `"items"`}, nil)

	// --size without --cursor stays the array schema for scans targets.
	resetFlag(t, scanTargetsCmd, "cursor")
	setChangedFlag(t, scanTargetsCmd, "size", "20")
	assertSchema(scanTargetsCmd, []string{"scan-1"}, []string{`"type": "array"`}, []string{`"nextCursor"`})
}

func assertFullWalk(t *testing.T, seen *[]url.Values, stdout, firstID, secondID string) {
	t.Helper()
	if len(*seen) != 2 || (*seen)[0].Get("cursor") != "" || (*seen)[1].Get("cursor") != "page-2" {
		t.Fatalf("cursors = %#v", *seen)
	}

	if !strings.Contains(stdout, firstID) || !strings.Contains(stdout, secondID) || strings.Contains(stdout, "nextCursor") {
		t.Fatalf("default listing = %s", stdout)
	}
}
