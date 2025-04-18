package escape

import (
	"context"
	"fmt"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

// ListScans lists all scans for an application
func ListScans(ctx context.Context, applicationID, next string) ([]v2.ListScans200ResponseDataInner, string, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, "", fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ApplicationsAPI.ListScans(ctx, applicationID)
	if next != "" {
		req.After(next)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, "", fmt.Errorf("unable to list scans: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetScanIssues returns issues found in a scan
func GetScanIssues(ctx context.Context, scanID string) ([]v2.ListIssues200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ScansAPI.ListIssues(ctx, scanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get scan: %w", err)
	}
	return data, nil
}
