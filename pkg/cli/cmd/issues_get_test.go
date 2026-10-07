package cmd

import (
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
)

// saveIssueGetOutput forces pretty output for the duration of a test and
// restores the previous mode.
func saveIssueGetOutput(t *testing.T) {
	t.Helper()
	prev := rootCmdOutputStr
	t.Cleanup(func() {
		rootCmdOutputStr = prev
		_ = out.SetOutput("pretty")
	})

	rootCmdOutputStr = ""
	if err := out.SetOutput("pretty"); err != nil {
		t.Fatalf("set pretty output: %v", err)
	}
}

// sampleIssueDetails builds an issue carrying every AI pentest detail field
// that `issues get` renders in pretty mode.
func sampleIssueDetails() *v3.GetIssue200Response {
	return &v3.GetIssue200Response{
		Id: "00000000-0000-0000-0000-000000000001",
		ReproductionSteps: []v3.IssueReproductionStep{
			{Order: 0, Title: "Send the replay", Content: "curl -X POST /admin"},
		},
		Exploits: []string{"' OR 1=1 --"},
		AttackChain: []v3.IssueAttackChainStep{
			{
				Order:       0,
				Stage:       v3.ENUMPROPERTIESATTACKCHAINITEMSPROPERTIESSTAGE("INITIAL_ACCESS"),
				Title:       "Reach the admin route",
				Description: "The route is exposed without auth",
				Content:     "GET /admin",
			},
		},
		AiFalsePositive: v3.GetIssue200ResponseAiFalsePositive{
			IsFalsePositive:  true,
			Reasoning:        v3.PtrString("The endpoint is internal only"),
			ReasoningSummary: v3.PtrString("Internal endpoint"),
		},
		Ticket: &v3.IssueTicket{
			ExternalId:  "SEC-42",
			ExternalUrl: "https://jira.example.com/browse/SEC-42",
			CreatedAt:   "2026-01-01T00:00:00Z",
		},
		Impact:    v3.PtrString("Full database read"),
		FixPrompt: "Escape the SQL parameter in the admin query.",
	}
}

func TestPrintIssueDetailsRendersEverySection(t *testing.T) {
	saveIssueGetOutput(t)

	stdout := captureStdout(t, func() {
		printIssueDetails(sampleIssueDetails())
	})

	for _, want := range []string{
		"REPRODUCTION STEPS",
		"Send the replay",
		"curl -X POST /admin",
		"EXPLOITS",
		"' OR 1=1 --",
		"ATTACK CHAIN",
		"INITIAL_ACCESS",
		"Reach the admin route",
		"AI FALSE POSITIVE",
		"The endpoint is internal only",
		"TICKET",
		"SEC-42",
		"https://jira.example.com/browse/SEC-42",
		"IMPACT",
		"Full database read",
		"FIX PROMPT",
		"Escape the SQL parameter in the admin query.",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("output missing %q\n%s", want, stdout)
		}
	}
}

func TestPrintIssueDetailsSkipsEmptySections(t *testing.T) {
	saveIssueGetOutput(t)

	stdout := captureStdout(t, func() {
		printIssueDetails(&v3.GetIssue200Response{Id: "00000000-0000-0000-0000-000000000001"})
	})

	for _, absent := range []string{"REPRODUCTION STEPS", "EXPLOITS", "ATTACK CHAIN", "TICKET", "IMPACT"} {
		if strings.Contains(stdout, absent) {
			t.Errorf("output should not contain %q when the field is empty\n%s", absent, stdout)
		}
	}

	// The AI verdict and fix prompt always render so the absence is explicit.
	if !strings.Contains(stdout, "AI FALSE POSITIVE") {
		t.Errorf("expected the AI FALSE POSITIVE section\n%s", stdout)
	}

	if !strings.Contains(stdout, "FIX PROMPT") {
		t.Errorf("expected the FIX PROMPT section\n%s", stdout)
	}
}

func TestPrintIssueDetailsIsPrettyOnly(t *testing.T) {
	saveIssueGetOutput(t)
	rootCmdOutputStr = "json"
	t.Cleanup(func() { rootCmdOutputStr = "" })

	stdout := captureStdout(t, func() {
		printIssueDetails(sampleIssueDetails())
	})

	if stdout != "" {
		t.Fatalf("expected no pretty sections in json mode, got %q", stdout)
	}
}
