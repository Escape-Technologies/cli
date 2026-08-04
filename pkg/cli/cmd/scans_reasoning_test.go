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

func TestReasoningListTruncated(t *testing.T) {
	t.Parallel()

	empty := ""
	more := "next-page"

	cases := []struct {
		name        string
		accumulated int
		limit       int
		cursor      *string
		want        bool
	}{
		{
			name:        "final page exceeds remaining limit without next cursor",
			accumulated: 210,
			limit:       200,
			cursor:      &empty,
			want:        true,
		},
		{
			name:        "exact limit with no next cursor",
			accumulated: 200,
			limit:       200,
			cursor:      &empty,
			want:        false,
		},
		{
			name:        "exact limit with next cursor",
			accumulated: 200,
			limit:       200,
			cursor:      &more,
			want:        true,
		},
		{
			name:        "overflow with next cursor",
			accumulated: 250,
			limit:       200,
			cursor:      &more,
			want:        true,
		},
		{
			name:        "nil cursor",
			accumulated: 205,
			limit:       200,
			cursor:      nil,
			want:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := reasoningListTruncated(tc.accumulated, tc.limit, tc.cursor); got != tc.want {
				t.Fatalf("reasoningListTruncated(%d, %d, %v) = %v, want %v",
					tc.accumulated, tc.limit, tc.cursor, got, tc.want)
			}
		})
	}
}

func TestListReasoningEventsStopsAtLimit(t *testing.T) {
	t.Parallel()

	summaries := make([]v3.EventSummarized, 0, 250)
	for index := 0; index < 250; index++ {
		summaries = append(summaries, v3.EventSummarized{})
	}

	limit := 200
	if !reasoningListTruncated(len(summaries), limit, nil) {
		t.Fatal("expected listTruncated when source exceeds limit")
	}
	capped := summaries[:limit]
	if len(capped) != limit {
		t.Fatalf("expected %d summaries, got %d", limit, len(capped))
	}
}
