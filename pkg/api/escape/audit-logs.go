package escape

import (
	"context"
	"errors"
	"fmt"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

// ListAuditLogs lists audit logs with pagination
func ListAuditLogs(ctx context.Context, count *int, after *string) ([]v2.ListAuditLogs200ResponseDataInner, string, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, "", fmt.Errorf("unable to init client: %w", err)
	}
	req := client.AuditAPI.ListAuditLogs(ctx)
	if count != nil {
		req = req.Count(*count)
	}
	if after != nil {
		req = req.After(*after)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, "", fmt.Errorf("unable to get audit logs: %w", err)
	}
	if data == nil {
		return nil, "", errors.New("unable to get audit logs: no data received")
	}
	return data.Data, data.NextCursor, nil
}
