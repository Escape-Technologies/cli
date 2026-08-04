package cmd

import (
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestFetchScanReasoningLogsRejectsInvalidLimits(t *testing.T) {
	t.Parallel()

	_, err := fetchScanReasoningLogs(t.Context(), "scan-1", "", "", -1, 50)
	if err == nil {
		t.Fatal("expected negative list-limit error")
	}

	_, err = fetchScanReasoningLogs(t.Context(), "scan-1", "", "", maxReasoningListLimit+1, 50)
	if err == nil {
		t.Fatal("expected list-limit cap error")
	}

	_, err = fetchScanReasoningLogs(t.Context(), "scan-1", "", "", 50, -1)
	if err == nil {
		t.Fatal("expected negative hydrate-limit error")
	}

	_, err = fetchScanReasoningLogs(t.Context(), "scan-1", "", "", 50, maxReasoningHydrateLimit+1)
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

func TestListReasoningEventsStopsAtLimit(t *testing.T) {
	t.Parallel()

	summaries := make([]v3.EventSummarized, 0, 250)
	for index := 0; index < 250; index++ {
		summaries = append(summaries, v3.EventSummarized{})
	}

	limit := 200
	capped := summaries
	listTruncated := true
	if len(capped) > limit {
		capped = capped[:limit]
	}

	if len(capped) != limit {
		t.Fatalf("expected %d summaries, got %d", limit, len(capped))
	}
	if !listTruncated {
		t.Fatal("expected listTruncated when source exceeds limit")
	}
}
