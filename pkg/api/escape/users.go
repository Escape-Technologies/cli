package escape

import (
	"context"
	"fmt"
	"strings"

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

// ListUsers lists all users in the organization.
// Search is applied client-side because the public API does not expose a search filter.
func ListUsers(ctx context.Context, search string) ([]v3.ListUsers200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.UsersAPI.ListUsers(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}

	term := strings.TrimSpace(strings.ToLower(search))
	if term == "" {
		return data, nil
	}

	filtered := make([]v3.ListUsers200ResponseInner, 0, len(data))
	for _, user := range data {
		if strings.Contains(strings.ToLower(user.GetEmail()), term) ||
			strings.Contains(strings.ToLower(fmt.Sprint(user.AdditionalProperties["name"])), term) {
			filtered = append(filtered, user)
		}
	}
	return filtered, nil
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

// InviteUsers invites one or more users by email address.
func InviteUsers(ctx context.Context, emails []string, roleID string) ([]v3.ListUsers200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := v3.InviteUserRequest{
		Emails: emails,
	}
	if roleID != "" {
		req.Bindings = []v3.InviteUserRequestBindingsInner{
			{RoleId: roleID},
		}
	}
	data, _, err := client.UsersAPI.InviteUser(ctx).InviteUserRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}
