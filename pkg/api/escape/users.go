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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ListUsers lists users in the organization. Search is sent to the API, whose
// search matches the email only: filtering names client-side promised a match
// the endpoint never performs. deactivated is the query enum "true" or "false";
// an empty value lists every user.
func ListUsers(ctx context.Context, search, deactivated string) ([]v3.ListUsers200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.UsersAPI.ListUsers(ctx)
	if search != "" {
		req = req.Search(search)
	}

	if deactivated != "" {
		req = req.Deactivated(deactivated)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// DeactivateUser deactivates a user of the organization. The same route
// reactivates a user, but the CLI exposes the one operation the ticket asks
// for; the API refuses a self-deactivation.
func DeactivateUser(ctx context.Context, userID string) (*v3.ListUsers200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := v3.NewUpdateUserRequest(true)
	data, _, err := client.UsersAPI.UpdateUser(ctx, userID).UpdateUserRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
