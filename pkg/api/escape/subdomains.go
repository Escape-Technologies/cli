package escape

import (
	"context"
	"errors"
	"fmt"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

// ListSubdomains lists subdomains with pagination
func ListSubdomains(ctx context.Context, count *int, after *string) ([]v2.ListSubdomains200ResponseDataInner, string, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, "", fmt.Errorf("unable to init client: %w", err)
	}
	req := client.SubdomainsAPI.ListSubdomains(ctx)
	if count != nil {
		req = req.Count(*count)
	}
	if after != nil {
		req = req.After(*after)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, "", fmt.Errorf("unable to get subdomains: %w", err)
	}
	if data == nil {
		return nil, "", errors.New("unable to get subdomains: no data received")
	}
	return data.Data, data.NextCursor, nil
}
