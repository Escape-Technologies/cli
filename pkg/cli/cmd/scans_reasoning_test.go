package cmd

import (
	"testing"
)

func TestFetchScanReasoningLogsRejectsInvalidHydrateLimit(t *testing.T) {
	t.Parallel()

	_, err := fetchScanReasoningLogs(t.Context(), "scan-1", -1)
	if err == nil {
		t.Fatal("expected negative hydrate-limit error")
	}

	_, err = fetchScanReasoningLogs(t.Context(), "scan-1", maxReasoningHydrateLimit+1)
	if err == nil {
		t.Fatal("expected hydrate-limit cap error")
	}
}

func TestReasoningStagesIncludeAgentReasoningAndAction(t *testing.T) {
	t.Parallel()

	if len(reasoningStages) != 2 {
		t.Fatalf("expected 2 reasoning stages, got %d", len(reasoningStages))
	}
	if reasoningStages[0] != "AGENT_REASONING" {
		t.Fatalf("expected AGENT_REASONING first, got %q", reasoningStages[0])
	}
	if reasoningStages[1] != "AGENT_ACTION" {
		t.Fatalf("expected AGENT_ACTION second, got %q", reasoningStages[1])
	}
}
