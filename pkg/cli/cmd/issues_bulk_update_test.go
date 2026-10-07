package cmd

import (
	"strings"
	"testing"
)

// saveBulkUpdateFlags snapshots the globals read by buildBulkUpdateIssuesRequest and restores them.
func saveBulkUpdateFlags(t *testing.T) {
	t.Helper()
	prevStatus := bulkIssueStatus
	prevReason := bulkIssueReason
	prevSetSeverity := setSeverity
	prevResetSeverity := resetSeverity
	prevIDs := bulkIssueIDs
	prevAssetIDs := bulkIssueAssetIDs
	prevSeverities := bulkIssueSeverities
	prevProfileIDs := bulkIssueProfileIDs
	prevTagIDs := bulkIssueTagIDs
	prevScannerKinds := bulkIssueScannerKinds
	prevAll := bulkIssueAll
	prevDryRun := bulkIssueDryRun
	t.Cleanup(func() {
		bulkIssueStatus = prevStatus
		bulkIssueReason = prevReason
		setSeverity = prevSetSeverity
		resetSeverity = prevResetSeverity
		bulkIssueIDs = prevIDs
		bulkIssueAssetIDs = prevAssetIDs
		bulkIssueSeverities = prevSeverities
		bulkIssueProfileIDs = prevProfileIDs
		bulkIssueTagIDs = prevTagIDs
		bulkIssueScannerKinds = prevScannerKinds
		bulkIssueAll = prevAll
		bulkIssueDryRun = prevDryRun
	})

	bulkIssueStatus = ""
	bulkIssueReason = ""
	setSeverity = ""
	resetSeverity = false
	bulkIssueIDs = nil
	bulkIssueAssetIDs = nil
	bulkIssueSeverities = nil
	bulkIssueProfileIDs = nil
	bulkIssueTagIDs = nil
	bulkIssueScannerKinds = nil
	bulkIssueAll = false
	bulkIssueDryRun = false
}

func TestIssueBulkUpdateRequiresASelection(t *testing.T) {
	saveBulkUpdateFlags(t)
	bulkIssueStatus = "IGNORED"

	err := issueBulkUpdateCmd.RunE(issueBulkUpdateCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--issue-id") {
		t.Fatalf("expected a filter requirement, got %v", err)
	}
}

func TestIssueBulkUpdateAllAloneIsAccepted(t *testing.T) {
	saveBulkUpdateFlags(t)
	bulkIssueStatus = "IGNORED"
	bulkIssueAll = true

	if err := requireIssueBulkSelection(); err != nil {
		t.Fatalf("expected --all without a filter to pass, got %v", err)
	}
}

func TestIssueBulkUpdateAllWithFilterIsRejected(t *testing.T) {
	saveBulkUpdateFlags(t)
	bulkIssueStatus = "IGNORED"
	bulkIssueAll = true
	bulkIssueIDs = []string{"00000000-0000-0000-0000-000000000001"}

	err := requireIssueBulkSelection()
	if err == nil || !strings.Contains(err.Error(), "--all cannot be combined with a filter") {
		t.Fatalf("expected a --all mutual exclusion error, got %v", err)
	}
}

func TestBuildBulkUpdateIssuesRequestSetsAllAndDryRun(t *testing.T) {
	saveBulkUpdateFlags(t)
	bulkIssueStatus = "IGNORED"
	bulkIssueAll = true
	bulkIssueDryRun = true

	body, err := buildBulkUpdateIssuesRequest()
	if err != nil {
		t.Fatalf("buildBulkUpdateIssuesRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	if payload["all"] != true {
		t.Errorf("all = %#v, want true", payload["all"])
	}

	if payload["dryRun"] != true {
		t.Errorf("dryRun = %#v, want true", payload["dryRun"])
	}

	if _, hasWhere := payload["where"]; hasWhere {
		t.Errorf("unexpected where payload with --all: %v", payload["where"])
	}
}

func TestBuildBulkUpdateIssuesRequestDryRunFalseByDefault(t *testing.T) {
	saveBulkUpdateFlags(t)
	bulkIssueStatus = "IGNORED"
	bulkIssueIDs = []string{"00000000-0000-0000-0000-000000000001"}

	body, err := buildBulkUpdateIssuesRequest()
	if err != nil {
		t.Fatalf("buildBulkUpdateIssuesRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	if _, hasAll := payload["all"]; hasAll {
		t.Errorf("unexpected all payload without --all: %v", payload["all"])
	}

	if _, hasDryRun := payload["dryRun"]; hasDryRun {
		t.Errorf("unexpected dryRun payload without --dry-run: %v", payload["dryRun"])
	}
}

func TestBuildBulkUpdateIssuesRequestSeverityCarriesReason(t *testing.T) {
	saveBulkUpdateFlags(t)

	setSeverity = "MEDIUM"
	bulkIssueReason = "re-scored after false-positive triage"

	body, err := buildBulkUpdateIssuesRequest()
	if err != nil {
		t.Fatalf("buildBulkUpdateIssuesRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	severity, ok := payload["severity"].(map[string]any)
	if !ok {
		t.Fatalf("expected severity object {value, reason}, got %v", payload["severity"])
	}

	if severity["value"] != "MEDIUM" {
		t.Errorf("severity value = %v, want MEDIUM", severity["value"])
	}

	if severity["reason"] != "re-scored after false-positive triage" {
		t.Errorf("severity reason = %v, want the given reason", severity["reason"])
	}

	if _, hasStatus := payload["status"]; hasStatus {
		t.Errorf("unexpected status in payload: %v", payload["status"])
	}
}

func TestBuildBulkUpdateIssuesRequestReasonAppliesToStatusAndSeverity(t *testing.T) {
	saveBulkUpdateFlags(t)

	bulkIssueStatus = "IGNORED"
	setSeverity = "LOW"
	bulkIssueReason = "quarterly triage"

	body, err := buildBulkUpdateIssuesRequest()
	if err != nil {
		t.Fatalf("buildBulkUpdateIssuesRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	status, ok := payload["status"].(map[string]any)
	if !ok {
		t.Fatalf("expected status object {value, reason}, got %v", payload["status"])
	}

	if status["value"] != "IGNORED" || status["reason"] != "quarterly triage" {
		t.Errorf("unexpected status payload: %v", status)
	}

	severity, ok := payload["severity"].(map[string]any)
	if !ok {
		t.Fatalf("expected severity object {value, reason}, got %v", payload["severity"])
	}

	if severity["value"] != "LOW" || severity["reason"] != "quarterly triage" {
		t.Errorf("unexpected severity payload: %v", severity)
	}
}

func TestBuildBulkUpdateIssuesRequestSeverityWithoutReason(t *testing.T) {
	saveBulkUpdateFlags(t)

	setSeverity = "HIGH"

	body, err := buildBulkUpdateIssuesRequest()
	if err != nil {
		t.Fatalf("buildBulkUpdateIssuesRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	severity, ok := payload["severity"].(map[string]any)
	if !ok {
		t.Fatalf("expected severity object, got %v", payload["severity"])
	}

	if severity["value"] != "HIGH" {
		t.Errorf("severity value = %v, want HIGH", severity["value"])
	}

	if _, hasReason := severity["reason"]; hasReason {
		t.Errorf("unexpected reason in severity payload: %v", severity["reason"])
	}
}

func TestBuildBulkUpdateIssuesRequestResetSeverityCarriesReason(t *testing.T) {
	saveBulkUpdateFlags(t)

	resetSeverity = true
	bulkIssueReason = "scanner severity restored after manual override"

	body, err := buildBulkUpdateIssuesRequest()
	if err != nil {
		t.Fatalf("buildBulkUpdateIssuesRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	severity, ok := payload["severity"].(map[string]any)
	if !ok {
		t.Fatalf("expected severity object {value, reason}, got %v", payload["severity"])
	}

	value, present := severity["value"]
	if !present || value != nil {
		t.Fatalf("severity value = %#v, want null", severity["value"])
	}

	if severity["reason"] != "scanner severity restored after manual override" {
		t.Errorf("severity reason = %v, want the given reason", severity["reason"])
	}
}

func TestBuildBulkUpdateIssuesRequestResetSeverityWithoutReasonIsNull(t *testing.T) {
	saveBulkUpdateFlags(t)

	resetSeverity = true

	body, err := buildBulkUpdateIssuesRequest()
	if err != nil {
		t.Fatalf("buildBulkUpdateIssuesRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	value, present := payload["severity"]
	if !present || value != nil {
		t.Fatalf("severity = %#v, want null", payload["severity"])
	}
}

func TestBuildBulkUpdateIssuesRequestRejectsReasonWithoutStatusOrSeverity(t *testing.T) {
	saveBulkUpdateFlags(t)

	bulkIssueReason = "because"

	_, err := buildBulkUpdateIssuesRequest()
	if err == nil || err.Error() != "--reason requires --status, --set-severity, or --reset-severity" {
		t.Fatalf("expected --reason requires --status, --set-severity, or --reset-severity error, got %v", err)
	}
}
