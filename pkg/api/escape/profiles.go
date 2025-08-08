package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListProfiles lists all profiles
func ListProfiles(ctx context.Context, next string) ([]v3.ProfileSummarized, string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, "", fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ProfilesAPI.ListProfiles(ctx)
	if next != "" {
		req = req.Cursor(next)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, "", fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}
