package cmd

import (
	"context"
	"net/http"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestContinuousScansStartPrintsDeclaredDocument(t *testing.T) {
	const profileID = "00000000-0000-0000-0000-0000000000cf"

	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/continuous-scans" {
			http.NotFound(w, r)
			return
		}

		writeJSON(t, w, map[string]any{
			"id":             "33333333-3333-3333-3333-333333333333",
			"status":         "STARTING",
			"createdAt":      "2026-01-01T00:00:00Z",
			"updatedAt":      "2026-01-01T00:00:00Z",
			"duration":       0,
			"progressRatio":  0,
			"initiator":      "CI",
			"kind":           "CONTINUOUS_PENTEST",
			"profileId":      profileID,
			"organizationId": "00000000-0000-0000-0000-000000000002",
			"links":          map[string]any{"scanIssues": "https://example.com/issues"},
		})
	})

	continuousScansStartCmd.SetContext(context.Background())
	stdout, stderr, err := captureJSONCommand(t, func() error {
		return continuousScansStartCmd.RunE(continuousScansStartCmd, []string{profileID})
	})
	if err != nil {
		t.Fatalf("start: %v\nstderr: %s", err, stderr)
	}

	var result v3.StartScan200Response
	mustUnmarshalOne(t, stdout, &result)
	if result.GetId() != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("id = %q", result.GetId())
	}

	if !strings.EqualFold(string(result.GetKind()), "CONTINUOUS_PENTEST") {
		t.Fatalf("kind = %q", result.GetKind())
	}
}
