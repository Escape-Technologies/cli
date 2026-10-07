package cmd

import (
	"context"
	"net/http"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

// TestRegressionTestsCommandsPrintDeclaredDocument exercises every leaf of the
// regression-tests group and checks each prints its declared response.
func TestRegressionTestsCommandsPrintDeclaredDocument(t *testing.T) {
	const (
		testID = "00000000-0000-0000-0000-0000000000d0"
		runID  = "00000000-0000-0000-0000-0000000000d1"
	)

	runBody := map[string]any{
		"id":                    runID,
		"status":                "CLARIFICATION_REQUIRED",
		"clarificationQuestion": "Which endpoint?",
		"validationId":          "00000000-0000-0000-0000-0000000000d2",
		"createdAt":             "2026-01-01T00:00:00Z",
		"updatedAt":             "2026-01-01T00:00:00Z",
	}
	detailBody := map[string]any{
		"id":                testID,
		"name":              "Checkout",
		"additionalContext": "Focus on checkout",
		"inputFilename":     "report.pdf",
		"inputFormat":       "PDF",
		"status":            "RUNNING",
		"signedUrl":         "https://example.s3.amazonaws.com/report.pdf",
		"createdAt":         "2026-01-01T00:00:00Z",
		"updatedAt":         "2026-01-01T00:00:00Z",
		"runs":              []map[string]any{runBody},
		"hasMoreRuns":       false,
	}
	summaryBody := map[string]any{
		"id":            testID,
		"name":          "Checkout",
		"inputFilename": "report.pdf",
		"inputFormat":   "PDF",
		"status":        "RUNNING",
		"createdAt":     "2026-01-01T00:00:00Z",
		"updatedAt":     "2026-01-01T00:00:00Z",
	}

	serveRegressionTestsAPI(t, testID, runID, summaryBody, detailBody, runBody)

	detail := `{"name":"Checkout","additionalContext":"Focus on checkout","inputFilename":"report.pdf","temporaryObjectKey":"00000000-0000-0000-0000-0000000000d3"}`
	tests := []struct {
		name  string
		cmd   *cobra.Command
		args  []string
		stdin string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "list",
			cmd:  regressionTestsListCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var summaries []v3.RegressionTestSummary
				mustUnmarshalOne(t, stdout, &summaries)
				if len(summaries) != 1 || summaries[0].GetId() != testID {
					t.Fatalf("list = %#v", summaries)
				}
			},
		},
		{
			name: "get",
			cmd:  regressionTestsGetCmd,
			args: []string{testID},
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var detail v3.CreateRegressionTest200Response
				mustUnmarshalOne(t, stdout, &detail)
				if detail.GetId() != testID || len(detail.GetRuns()) != 1 {
					t.Fatalf("get = %#v", detail)
				}
			},
		},
		{
			name:  "create",
			cmd:   regressionTestsCreateCmd,
			stdin: detail,
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var detail v3.CreateRegressionTest200Response
				mustUnmarshalOne(t, stdout, &detail)
				if detail.GetId() != testID {
					t.Fatalf("create = %#v", detail)
				}
			},
		},
		{
			name:  "update",
			cmd:   regressionTestsUpdateCmd,
			args:  []string{testID},
			stdin: `{"name":"Checkout v2"}`,
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var detail v3.CreateRegressionTest200Response
				mustUnmarshalOne(t, stdout, &detail)
				if detail.GetId() != testID {
					t.Fatalf("update = %#v", detail)
				}
			},
		},
		{
			name: "delete",
			cmd:  regressionTestsDeleteCmd,
			args: []string{testID},
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var deleted v3.DeleteProfile200Response
				mustUnmarshalOne(t, stdout, &deleted)
				if deleted.GetMessage() == "" {
					t.Fatalf("delete = %#v", deleted)
				}
			},
		},
		{
			name: "run",
			cmd:  regressionTestsRunCmd,
			args: []string{testID},
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var detail v3.CreateRegressionTest200Response
				mustUnmarshalOne(t, stdout, &detail)
				if detail.GetId() != testID {
					t.Fatalf("run = %#v", detail)
				}
			},
		},
		{
			name: "stop",
			cmd:  regressionTestsStopCmd,
			args: []string{testID},
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var detail v3.CreateRegressionTest200Response
				mustUnmarshalOne(t, stdout, &detail)
				if detail.GetId() != testID {
					t.Fatalf("stop = %#v", detail)
				}
			},
		},
		{
			name: "history",
			cmd:  regressionTestsHistoryCmd,
			args: []string{testID},
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var runs []v3.RegressionTestRun
				mustUnmarshalOne(t, stdout, &runs)
				if len(runs) != 1 || runs[0].GetId() != runID {
					t.Fatalf("history = %#v", runs)
				}
			},
		},
		{
			name:  "answer",
			cmd:   regressionTestsAnswerCmd,
			args:  []string{testID, runID},
			stdin: `{"content":"Use the v2 checkout API"}`,
			check: func(t *testing.T, stdout string) {
				t.Helper()
				var run v3.ReplyRegressionTestClarification200Response
				mustUnmarshalOne(t, stdout, &run)
				if run.GetId() != runID {
					t.Fatalf("answer = %#v", run)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.stdin != "" {
				tt.cmd.SetIn(strings.NewReader(tt.stdin))
			}

			tt.cmd.SetContext(context.Background())

			stdout, stderr, err := captureJSONCommand(t, func() error {
				return tt.cmd.RunE(tt.cmd, tt.args)
			})
			if err != nil {
				t.Fatalf("%s: %v\nstderr: %s", tt.name, err, stderr)
			}

			tt.check(t, stdout)
		})
	}
}

// serveRegressionTestsAPI serves the regression-tests endpoints exercised by
// TestRegressionTestsCommandsPrintDeclaredDocument.
func serveRegressionTestsAPI(
	t *testing.T,
	testID string,
	runID string,
	summaryBody map[string]any,
	detailBody map[string]any,
	runBody map[string]any,
) {
	t.Helper()
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case r.Method == http.MethodGet && path == "/v3/regression-tests":
			writeJSON(t, w, map[string]any{"data": []map[string]any{summaryBody}, "nextCursor": nil, "totalCount": 1})
		case r.Method == http.MethodPost && path == "/v3/regression-tests":
			writeJSON(t, w, detailBody)
		case r.Method == http.MethodPut && path == "/v3/regression-tests/"+testID:
			writeJSON(t, w, detailBody)
		case r.Method == http.MethodDelete && path == "/v3/regression-tests/"+testID:
			writeJSON(t, w, map[string]any{"message": "deleted"})
		case r.Method == http.MethodPost && path == "/v3/regression-tests/"+testID+"/run":
			writeJSON(t, w, detailBody)
		case r.Method == http.MethodPost && path == "/v3/regression-tests/"+testID+"/stop":
			writeJSON(t, w, detailBody)
		case r.Method == http.MethodGet && path == "/v3/regression-tests/"+testID+"/history":
			writeJSON(t, w, map[string]any{"data": []map[string]any{runBody}, "nextCursor": nil, "totalCount": 1})
		case r.Method == http.MethodPost && path == "/v3/regression-tests/"+testID+"/runs/"+runID+"/clarifications":
			writeJSON(t, w, runBody)
		case r.Method == http.MethodGet && path == "/v3/regression-tests/"+testID:
			writeJSON(t, w, detailBody)
		default:
			http.NotFound(w, r)
		}
	})
}
