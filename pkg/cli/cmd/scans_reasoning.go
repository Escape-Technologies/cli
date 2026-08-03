package cmd

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	defaultReasoningHydrateLimit = 50
	maxReasoningHydrateLimit     = 200
	maxReasoningHydrateWorkers   = 10
)

var reasoningHydrateLimit int

// ScanReasoningLogs bundles AI agent reasoning and action events for a scan so
// MCP/CLI callers can inspect how an assessment unfolded in one tool call.
type ScanReasoningLogs struct {
	ScanID          string                   `json:"scanId"`
	Events          []v3.GetEvent200Response `json:"events"`
	EventsTruncated bool                     `json:"eventsTruncated,omitempty"`
	EventErrors     []IssueEventHydrateError `json:"eventErrors,omitempty"`
}

var reasoningStages = []string{
	string(v3.ENUMPROPERTIESEVENTSITEMSPROPERTIESSTAGE_AGENT_REASONING),
	string(v3.ENUMPROPERTIESEVENTSITEMSPROPERTIESSTAGE_AGENT_ACTION),
}

var scansReasoningCmd = &cobra.Command{
	Use:     "reasoning scan-id",
	Aliases: []string{"reasoning-logs", "agent-reasoning"},
	Short:   "List AI agent reasoning logs for a scan",
	Long: `List AI Agent Reasoning Logs - Inspect How an Assessment Unfolded

Fetches AGENT_REASONING and AGENT_ACTION events for a scan and hydrates each
event's description and attachments so an AI agent can answer questions such as
which scenarios were tested, how many vulnerabilities were found, and how a
specific finding was discovered.

OUTPUT SHAPE (JSON):
  {
    "scanId": "<scan-id>",
    "events": [<EventDetailed>, ...],
    "eventsTruncated": <bool>,          // true when more events exist than hydrated
    "eventErrors": [{"eventId": "...", "error": "..."}, ...]  // omitted on full success
  }

Per-event hydration failures do NOT cancel siblings and do NOT fail the
command — they surface in 'eventErrors'. The command only fails if listing
events cannot be completed.`,
	Example: `  # Fetch reasoning logs for an AI pentest scan
  escape-cli scans reasoning <scan-id> -o json

  # Hydrate more events (default limit is 50)
  escape-cli scans reasoning <scan-id> --hydrate-limit 100 -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(ScanReasoningLogs{}) {
			return nil
		}

		result, err := fetchScanReasoningLogs(cmd.Context(), args[0], reasoningHydrateLimit)
		if err != nil {
			return err
		}

		out.Print(result, "")
		return nil
	},
}

func fetchScanReasoningLogs(ctx context.Context, scanID string, hydrateLimit int) (ScanReasoningLogs, error) {
	if hydrateLimit < 0 {
		return ScanReasoningLogs{}, errors.New("hydrate-limit must be >= 0")
	}
	if hydrateLimit > maxReasoningHydrateLimit {
		return ScanReasoningLogs{}, fmt.Errorf("hydrate-limit must be <= %d", maxReasoningHydrateLimit)
	}

	summaries, err := listAllReasoningEvents(ctx, scanID)
	if err != nil {
		return ScanReasoningLogs{}, err
	}

	result := ScanReasoningLogs{
		ScanID:          scanID,
		Events:          []v3.GetEvent200Response{},
		EventsTruncated: hydrateLimit > 0 && len(summaries) > hydrateLimit,
	}

	if hydrateLimit == 0 || len(summaries) == 0 {
		return result, nil
	}

	toHydrate := summaries
	if len(toHydrate) > hydrateLimit {
		toHydrate = toHydrate[:hydrateLimit]
	}

	events, eventErrors := hydrateReasoningEvents(ctx, toHydrate)
	result.Events = events
	if len(eventErrors) > 0 {
		result.EventErrors = eventErrors
	}
	return result, nil
}

func listAllReasoningEvents(ctx context.Context, scanID string) ([]v3.EventSummarized, error) {
	filters := &escape.ListEventsFilters{
		ScanIDs: []string{scanID},
		Stages:  reasoningStages,
	}

	var allEvents []v3.EventSummarized
	next := ""
	for {
		events, cursor, err := escape.ListEvents(ctx, next, filters)
		if err != nil {
			return nil, fmt.Errorf("unable to list reasoning events: %w", err)
		}
		allEvents = append(allEvents, events...)
		if cursor == nil || *cursor == "" {
			break
		}
		next = *cursor
	}
	return allEvents, nil
}

func hydrateReasoningEvents(
	ctx context.Context,
	summaries []v3.EventSummarized,
) ([]v3.GetEvent200Response, []IssueEventHydrateError) {
	events := make([]v3.GetEvent200Response, len(summaries))
	eventErrors := make([]IssueEventHydrateError, len(summaries))
	eventOK := make([]bool, len(summaries))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxReasoningHydrateWorkers)
	var mu sync.Mutex
	for index, summary := range summaries {
		g.Go(func() error {
			event, err := escape.GetEvent(gctx, summary.GetId())
			mu.Lock()
			defer mu.Unlock()
			if err != nil || event == nil {
				msg := "nil response"
				if err != nil {
					msg = err.Error()
				}
				eventErrors[index] = IssueEventHydrateError{EventID: summary.GetId(), Error: msg}
				return nil
			}
			events[index] = *event
			eventOK[index] = true
			return nil
		})
	}
	_ = g.Wait()

	hydratedEvents := make([]v3.GetEvent200Response, 0, len(events))
	hydratedErrors := make([]IssueEventHydrateError, 0)
	for index := range summaries {
		if eventOK[index] {
			hydratedEvents = append(hydratedEvents, events[index])
			continue
		}
		if eventErrors[index].EventID != "" {
			hydratedErrors = append(hydratedErrors, eventErrors[index])
		}
	}
	return hydratedEvents, hydratedErrors
}
