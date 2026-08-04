package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	defaultReasoningListLimit    = 200
	maxReasoningListLimit        = 500
	maxReasoningHydrateWorkers   = 10
)

var reasoningHydrateLimit int
var reasoningListLimit int
var reasoningAgentID string
var reasoningSearch string

// ScanReasoningLogs bundles AI agent reasoning and action events for a scan so
// MCP/CLI callers can inspect how an assessment unfolded in one tool call.
type ScanReasoningLogs struct {
	ScanID          string                   `json:"scanId"`
	Summaries       []v3.EventSummarized     `json:"summaries"`
	ListTruncated   bool                     `json:"listTruncated,omitempty"`
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

Fetches AGENT_REASONING and AGENT_ACTION events for a scan. Summaries are always
returned (up to --list-limit). Full descriptions and attachments are hydrated
for the first --hydrate-limit events only — AI pentest scans can emit thousands
of reasoning lines, so both caps keep responses bounded.

OUTPUT SHAPE (JSON):
  {
    "scanId": "<scan-id>",
    "summaries": [<EventSummarized>, ...],
    "listTruncated": <bool>,           // true when more events exist beyond list-limit
    "events": [<EventDetailed>, ...],  // hydrated subset
    "eventsTruncated": <bool>,         // true when more summaries exist than hydrated
    "eventErrors": [{"eventId": "...", "error": "..."}, ...]  // omitted on full success
  }

Per-event hydration failures do NOT cancel siblings and do NOT fail the
command — they surface in 'eventErrors'. The command only fails if listing
events cannot be completed.`,
	Example: `  # Fetch reasoning logs for an AI pentest scan
  escape-cli scans reasoning <scan-id> -o json

  # Summaries only (no per-event GET fan-out)
  escape-cli scans reasoning <scan-id> --hydrate-limit 0 -o json

  # Fetch reasoning logs for one agent
  escape-cli scans reasoning <scan-id> --agent-id <agent-id> -o json

  # Search within agent reasoning logs
  escape-cli scans reasoning <scan-id> --agent-id <agent-id> --search "catalog/product" -o json`,
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

		result, err := fetchScanReasoningLogs(
			cmd.Context(),
			args[0],
			reasoningAgentID,
			reasoningSearch,
			reasoningListLimit,
			reasoningHydrateLimit,
		)
		if err != nil {
			return err
		}

		out.Print(result, "")
		return nil
	},
}

func fetchScanReasoningLogs(
	ctx context.Context,
	scanID string,
	agentID string,
	search string,
	listLimit int,
	hydrateLimit int,
) (ScanReasoningLogs, error) {
	if listLimit < 0 {
		return ScanReasoningLogs{}, errors.New("list-limit must be >= 0")
	}
	if listLimit > maxReasoningListLimit {
		return ScanReasoningLogs{}, fmt.Errorf("list-limit must be <= %d", maxReasoningListLimit)
	}
	if hydrateLimit < 0 {
		return ScanReasoningLogs{}, errors.New("hydrate-limit must be >= 0")
	}
	if hydrateLimit > maxReasoningHydrateLimit {
		return ScanReasoningLogs{}, fmt.Errorf("hydrate-limit must be <= %d", maxReasoningHydrateLimit)
	}

	if listLimit == 0 {
		listLimit = defaultReasoningListLimit
	}

	summaries, listTruncated, err := listReasoningEvents(ctx, scanID, agentID, search, listLimit)
	if err != nil {
		return ScanReasoningLogs{}, err
	}

	result := ScanReasoningLogs{
		ScanID:          scanID,
		Summaries:       summaries,
		ListTruncated:   listTruncated,
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

// ponytail: caps API pagination — assessments can have 10k+ reasoning events;
// listing them all times out MCP (30s default) and staging public API.
func listReasoningEvents(
	ctx context.Context,
	scanID string,
	agentID string,
	search string,
	limit int,
) ([]v3.EventSummarized, bool, error) {
	if strings.TrimSpace(agentID) != "" {
		return listAgentReasoningEvents(ctx, scanID, agentID, search, limit)
	}

	filters := &escape.ListEventsFilters{
		ScanIDs: []string{scanID},
		Stages:  reasoningStages,
		Search:  search,
	}

	var summaries []v3.EventSummarized
	next := ""
	listTruncated := false
	for {
		events, cursor, err := escape.ListEvents(ctx, next, filters)
		if err != nil {
			return nil, false, fmt.Errorf("unable to list reasoning events: %w", err)
		}
		summaries = append(summaries, events...)
		if len(summaries) >= limit {
			summaries = summaries[:limit]
			listTruncated = cursor != nil && *cursor != ""
			break
		}
		if cursor == nil || *cursor == "" {
			break
		}
		next = *cursor
	}
	return summaries, listTruncated, nil
}

func listAgentReasoningEvents(
	ctx context.Context,
	scanID string,
	agentID string,
	search string,
	limit int,
) ([]v3.EventSummarized, bool, error) {
	filters := &escape.ListScanAgentLogsFilters{
		Search:        search,
		Stages:        reasoningStages,
		SortDirection: "desc",
	}

	var summaries []v3.EventSummarized
	next := ""
	listTruncated := false
	for {
		logs, cursor, err := escape.ListScanAgentLogs(ctx, scanID, agentID, next, 100, filters)
		if err != nil {
			return nil, false, fmt.Errorf("unable to list agent reasoning logs: %w", err)
		}
		for _, log := range logs {
			summary := v3.EventSummarized{
				Id:        log.ID,
				CreatedAt: log.CreatedAt,
				Title:     log.Title,
				Stage:     v3.ENUMPROPERTIESEVENTSITEMSPROPERTIESSTAGE(log.Stage),
				Level:     v3.ENUMPROPERTIESEVENTSITEMSPROPERTIESLEVEL(log.Level),
			}
			summaries = append(summaries, summary)
		}
		if len(summaries) >= limit {
			summaries = summaries[:limit]
			listTruncated = cursor != nil && *cursor != ""
			break
		}
		if cursor == nil || *cursor == "" {
			break
		}
		next = *cursor
	}
	return summaries, listTruncated, nil
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
