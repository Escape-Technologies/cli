package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListAuditLogs lists audit logs
func ListAuditLogs(ctx context.Context, dateFrom string, dateTo string, actionType string, next string) ([]v3.AuditLogSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.AuditAPI.ListAuditLogs(ctx).SortDirection("desc").StartTime(dateFrom).EndTime(dateTo).Action(actionType)
	if next != "" {
		req = req.Cursor(next)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}
