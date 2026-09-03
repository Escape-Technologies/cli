package cmd

import "testing"

func TestIssueUpdateRejectsReasonWithoutStatus(t *testing.T) {
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

	issueUpdateStatusStr = ""
	issueUpdateSeverity = "HIGH"
	issueUpdateReason = "because"
	issueResetSeverity = false

	err := issueUpdateStatusCmd.RunE(issueUpdateStatusCmd, []string{"00000000-0000-0000-0000-000000000001"})
	if err == nil || err.Error() != "--reason requires --status" {
		t.Fatalf("expected --reason requires --status error, got %v", err)
	}
}
