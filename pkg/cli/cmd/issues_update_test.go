package cmd

import (
	"encoding/json"
	"testing"
)

func marshalToMap(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal %s: %v", raw, err)
	}

	return payload
}

// saveIssueUpdateFlags snapshots the globals read by buildUpdateIssueRequest and restores them.
func saveIssueUpdateFlags(t *testing.T) {
	t.Helper()
	prevStatus := issueUpdateStatusStr
	prevReason := issueUpdateReason
	prevSeverity := issueUpdateSeverity
	prevReset := issueResetSeverity
	t.Cleanup(func() {
		issueUpdateStatusStr = prevStatus
		issueUpdateReason = prevReason
		issueUpdateSeverity = prevSeverity
		issueResetSeverity = prevReset
	})
}

func TestBuildUpdateIssueRequestResetSeverityCarriesReason(t *testing.T) {
	saveIssueUpdateFlags(t)

	issueUpdateStatusStr = ""
	issueUpdateSeverity = ""
	issueUpdateReason = "scanner severity restored after manual override"
	issueResetSeverity = true

	body, err := buildUpdateIssueRequest()
	if err != nil {
		t.Fatalf("buildUpdateIssueRequest: %v", err)
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

func TestBuildUpdateIssueRequestResetSeverityWithoutReasonIsNull(t *testing.T) {
	saveIssueUpdateFlags(t)

	issueUpdateStatusStr = ""
	issueUpdateSeverity = ""
	issueUpdateReason = ""
	issueResetSeverity = true

	body, err := buildUpdateIssueRequest()
	if err != nil {
		t.Fatalf("buildUpdateIssueRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	value, present := payload["severity"]
	if !present || value != nil {
		t.Fatalf("severity = %#v, want null", payload["severity"])
	}
}

func TestIssueUpdateRejectsReasonWithoutStatusOrSeverity(t *testing.T) {
	saveIssueUpdateFlags(t)

	issueUpdateStatusStr = ""
	issueUpdateSeverity = ""
	issueUpdateReason = "because"
	issueResetSeverity = false

	_, err := buildUpdateIssueRequest()
	if err == nil || err.Error() != "--reason requires --status, --severity, or --reset-severity" {
		t.Fatalf("expected --reason requires --status, --severity, or --reset-severity error, got %v", err)
	}
}

func TestBuildUpdateIssueRequestSeverityCarriesReason(t *testing.T) {
	saveIssueUpdateFlags(t)

	issueUpdateStatusStr = ""
	issueUpdateSeverity = "HIGH"
	issueUpdateReason = "CVSS reassessed after CVE lookup"
	issueResetSeverity = false

	body, err := buildUpdateIssueRequest()
	if err != nil {
		t.Fatalf("buildUpdateIssueRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	severity, ok := payload["severity"].(map[string]any)
	if !ok {
		t.Fatalf("expected severity object {value, reason}, got %v", payload["severity"])
	}

	if severity["value"] != "HIGH" {
		t.Errorf("severity value = %v, want HIGH", severity["value"])
	}

	if severity["reason"] != "CVSS reassessed after CVE lookup" {
		t.Errorf("severity reason = %v, want the given reason", severity["reason"])
	}

	if _, hasStatus := payload["status"]; hasStatus {
		t.Errorf("unexpected status in payload: %v", payload["status"])
	}
}

func TestBuildUpdateIssueRequestSeverityWithoutReasonKeepsPlainEnum(t *testing.T) {
	saveIssueUpdateFlags(t)

	issueUpdateStatusStr = ""
	issueUpdateSeverity = "LOW"
	issueUpdateReason = ""
	issueResetSeverity = false

	body, err := buildUpdateIssueRequest()
	if err != nil {
		t.Fatalf("buildUpdateIssueRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	if payload["severity"] != "LOW" {
		t.Fatalf("expected plain severity enum, got %v", payload["severity"])
	}
}

func TestBuildUpdateIssueRequestReasonAppliesToStatusAndSeverity(t *testing.T) {
	saveIssueUpdateFlags(t)

	issueUpdateStatusStr = "IGNORED"
	issueUpdateSeverity = "CRITICAL"
	issueUpdateReason = "accepted risk per security review"
	issueResetSeverity = false

	body, err := buildUpdateIssueRequest()
	if err != nil {
		t.Fatalf("buildUpdateIssueRequest: %v", err)
	}

	payload := marshalToMap(t, body)
	status, ok := payload["status"].(map[string]any)
	if !ok {
		t.Fatalf("expected status object {value, reason}, got %v", payload["status"])
	}

	if status["value"] != "IGNORED" || status["reason"] != "accepted risk per security review" {
		t.Errorf("unexpected status payload: %v", status)
	}

	severity, ok := payload["severity"].(map[string]any)
	if !ok {
		t.Fatalf("expected severity object {value, reason}, got %v", payload["severity"])
	}

	if severity["value"] != "CRITICAL" || severity["reason"] != "accepted risk per security review" {
		t.Errorf("unexpected severity payload: %v", severity)
	}
}

func TestBuildUpdateIssueRequestRejectsResetWithSeverity(t *testing.T) {
	saveIssueUpdateFlags(t)

	issueUpdateStatusStr = ""
	issueUpdateSeverity = "HIGH"
	issueUpdateReason = ""
	issueResetSeverity = true

	_, err := buildUpdateIssueRequest()
	if err == nil || err.Error() != "--reset-severity and --severity are mutually exclusive" {
		t.Fatalf("expected mutual exclusion error, got %v", err)
	}
}
