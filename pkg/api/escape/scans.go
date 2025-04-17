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
	data, resp, err := req.Execute()
	defer resp.Body.Close() //nolint:errcheck
	if err != nil {
		return nil, "", fmt.Errorf("unable to list scans: %w", err)
	}
	return data.Data, data.NextCursor, nil
}
