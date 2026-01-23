package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListAuditLogsFilters holds optional filters for listing audit logs
type ListAuditLogsFilters struct {
	DateFrom   string
	DateTo     string
	ActionType string
	Actor      string
	Search     string
}

// ListAuditLogs lists audit logs
func ListAuditLogs(ctx context.Context, next string, filters *ListAuditLogsFilters) ([]v3.AuditLogSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.AuditAPI.ListAuditLogs(ctx).SortDirection("desc")
	if next != "" {
		req = req.Cursor(next)
	}
	if filters != nil {
		if filters.DateFrom != "" {
			req = req.StartTime(filters.DateFrom)
		}
		if filters.DateTo != "" {
			req = req.EndTime(filters.DateTo)
		}
		if filters.ActionType != "" {
			req = req.Action(filters.ActionType)
		}
		if filters.Actor != "" {
			req = req.Actor(filters.Actor)
		}
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}
