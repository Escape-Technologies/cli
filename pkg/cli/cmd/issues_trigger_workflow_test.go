package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestIssueTriggerWorkflowRejectsMissingWorkflowID(t *testing.T) {
	prevWorkflowID := triggerWorkflowID
	t.Cleanup(func() { triggerWorkflowID = prevWorkflowID })

	triggerWorkflowID = ""
	err := issueTriggerWorkflowCmd.RunE(issueTriggerWorkflowCmd, []string{"00000000-0000-0000-0000-000000000001"})
	if err == nil || err.Error() != "--workflow-id is required" {
		t.Fatalf("expected --workflow-id is required error, got %v", err)
	}
}

func TestIssueTriggerWorkflowJSONMatchesDeclaredOutput(t *testing.T) {
	const issueID = "00000000-0000-0000-0000-000000000001"
	const workflowID = "00000000-0000-0000-0000-000000000002"

	var gotBody map[string]any
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/issues/"+issueID+"/manual-workflow" {
			http.NotFound(w, r)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}

		writeJSON(t, w, map[string]any{
			"workflow": map[string]any{
				"id":   workflowID,
				"name": "Export to Jira",
			},
		})
	})

	prevWorkflowID := triggerWorkflowID
	t.Cleanup(func() {
		triggerWorkflowID = prevWorkflowID
		flag := issueTriggerWorkflowCmd.Flags().Lookup("workflow-id")
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
	if err := issueTriggerWorkflowCmd.Flags().Set("workflow-id", workflowID); err != nil {
		t.Fatal(err)
	}

	issueTriggerWorkflowCmd.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return issueTriggerWorkflowCmd.RunE(issueTriggerWorkflowCmd, []string{issueID})
	})
	if err != nil {
		t.Fatalf("trigger: %v\nstderr: %s", err, stderr)
	}

	if gotBody["workflowId"] != workflowID {
		t.Fatalf("request body = %#v, want workflowId %s", gotBody, workflowID)
	}

	var result v3.TriggerIssueManualWorkflow200Response
	mustUnmarshalOne(t, stdout, &result)
	workflow := result.GetWorkflow()
	if workflow.Id != workflowID || workflow.Name != "Export to Jira" {
		t.Fatalf("result = %#v", result)
	}
}
