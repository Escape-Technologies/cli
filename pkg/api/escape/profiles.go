package escape

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListProfiles lists all profiles
func ListProfiles(ctx context.Context, next string) ([]v3.ProfileSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ProfilesAPI.ListProfiles(ctx)
	if next != "" {
		req = req.Cursor(next)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetProfile gets a profile by ID
func GetProfile(ctx context.Context, profileID string) (*v3.ProfileDetailed, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ProfilesAPI.GetProfile(ctx, profileID)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

func CreateProfileRest(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastRestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.ProfilesAPI.CreateDastRestProfile(ctx)
	profile, _, err := req.CreateDastRestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return profile, nil
}

func CreateProfileWebapp(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastWebAppProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.ProfilesAPI.CreateDastWebAppProfile(ctx)
	profile, _, err := req.CreateDastWebAppProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return profile, nil
}

func CreateProfileGraphql(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.ApiCreateDastGraphqlProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.ProfilesAPI.CreateDastGraphqlProfile(ctx)
	profile, _, err := req.ApiService.CreateDastGraphqlProfileExecute(payload)
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return profile, nil
}