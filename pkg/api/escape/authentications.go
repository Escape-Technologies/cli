package escape

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func StartAuthentication(ctx context.Context, body []byte) (*v3.StartAuthentication200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var req v3.StartAuthenticationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	data, _, err := client.ProfilesAPI.StartAuthentication(ctx).StartAuthenticationRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

func GetAuthentication(ctx context.Context, authenticationID string) (*v3.GetAuthentication200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.ProfilesAPI.GetAuthentication(ctx, authenticationID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}
