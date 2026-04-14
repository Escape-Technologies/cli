package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// GetMe returns the current authenticated user and organization context
func GetMe(ctx context.Context) (*v3.GetMe200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.UsersAPI.GetMe(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// ListUsers lists all users in the organization
func ListUsers(ctx context.Context) ([]v3.ListUsers200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.UsersAPI.ListUsers(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// GetUser gets a user by ID
func GetUser(ctx context.Context, userID string) (*v3.GetUser200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.UsersAPI.GetUser(ctx, userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// InviteUsers invites one or more users by email address
func InviteUsers(ctx context.Context, emails []string) ([]v3.ListUsers200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.UsersAPI.InviteUser(ctx).InviteUserRequest(v3.InviteUserRequest{
		Emails: emails,
	}).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}
