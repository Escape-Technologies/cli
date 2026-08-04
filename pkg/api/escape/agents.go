package escape

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// AgentSummarized is a lightweight AI pentest agent row.
type AgentSummarized struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	ParentID     *string  `json:"parentId"`
	HasChildren  bool     `json:"hasChildren"`
	Duration     *float64 `json:"duration"`
	IssuesCount  int      `json:"issuesCount"`
}

// AgentLogSummarized is a reasoning or action log entry for an agent.
type AgentLogSummarized struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"createdAt"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Stage       string `json:"stage"`
	Level       string `json:"level"`
}

type pageResponse[T any] struct {
	NextCursor *string `json:"nextCursor"`
	TotalCount int     `json:"totalCount"`
	Data       []T     `json:"data"`
}

// ListScanAgentsFilters holds optional filters for listing scan agents.
type ListScanAgentsFilters struct {
	Search      string
	EventSearch string
	RootsOnly   bool
	SortType    string
	SortDirection string
}

// ListScanAgents lists AI pentest agents for a scan.
func ListScanAgents(
	ctx context.Context,
	scanID string,
	next string,
	size int,
	filters *ListScanAgentsFilters,
) ([]AgentSummarized, *string, error) {
	if strings.TrimSpace(scanID) == "" {
		return nil, nil, fmt.Errorf("scanID is required")
	}

	query := url.Values{}
	if next != "" {
		query.Set("cursor", next)
	}
	if size > 0 {
		query.Set("size", fmt.Sprintf("%d", size))
	}
	if filters != nil {
		if filters.Search != "" {
			query.Set("search", filters.Search)
		}
		if filters.EventSearch != "" {
			query.Set("eventSearch", filters.EventSearch)
		}
		if filters.RootsOnly {
			query.Set("rootsOnly", "true")
		}
		if filters.SortType != "" {
			query.Set("sortType", filters.SortType)
		}
		if filters.SortDirection != "" {
			query.Set("sortDirection", filters.SortDirection)
		}
	}

	path := rawPath("scans", scanID, "agents")
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var page pageResponse[AgentSummarized]
	if err := rawRequest(ctx, "GET", path, nil, &page); err != nil {
		return nil, nil, fmt.Errorf("unable to list scan agents: %w", err)
	}
	return page.Data, page.NextCursor, nil
}

// ListScanAgentLogsFilters holds optional filters for listing agent logs.
type ListScanAgentLogsFilters struct {
	Search        string
	Stages        []string
	SortType      string
	SortDirection string
}

// ListScanAgentLogs lists reasoning logs for a scan agent.
func ListScanAgentLogs(
	ctx context.Context,
	scanID string,
	agentID string,
	next string,
	size int,
	filters *ListScanAgentLogsFilters,
) ([]AgentLogSummarized, *string, error) {
	if strings.TrimSpace(scanID) == "" {
		return nil, nil, fmt.Errorf("scanID is required")
	}
	if strings.TrimSpace(agentID) == "" {
		return nil, nil, fmt.Errorf("agentID is required")
	}

	query := url.Values{}
	if next != "" {
		query.Set("cursor", next)
	}
	if size > 0 {
		query.Set("size", fmt.Sprintf("%d", size))
	}
	if filters != nil {
		if filters.Search != "" {
			query.Set("search", filters.Search)
		}
		for _, stage := range filters.Stages {
			query.Add("stages", stage)
		}
		if filters.SortType != "" {
			query.Set("sortType", filters.SortType)
		}
		if filters.SortDirection != "" {
			query.Set("sortDirection", filters.SortDirection)
		}
	}

	path := rawPath("scans", scanID, "agents", agentID, "logs")
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var page pageResponse[AgentLogSummarized]
	if err := rawRequest(ctx, "GET", path, nil, &page); err != nil {
		return nil, nil, fmt.Errorf("unable to list scan agent logs: %w", err)
	}
	return page.Data, page.NextCursor, nil
}
