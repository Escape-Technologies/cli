package escape

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListRoles lists all roles in the organization
func ListRoles(ctx context.Context) ([]v3.ListRoles200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.RolesAPI.ListRoles(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// GetRole gets a role by ID
func GetRole(ctx context.Context, roleID string) (*v3.CreateRole200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.RolesAPI.GetRole(ctx, roleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// CreateRole creates a new role from raw JSON bytes
func CreateRole(ctx context.Context, body []byte) (*v3.CreateRole200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	var req v3.CreateRoleRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	data, _, err := client.RolesAPI.CreateRole(ctx).CreateRoleRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// UpdateRole updates a role by ID from raw JSON bytes
func UpdateRole(ctx context.Context, roleID string, body []byte) (*v3.CreateRole200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	var req v3.UpdateRoleRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	data, _, err := client.RolesAPI.UpdateRole(ctx, roleID).UpdateRoleRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// CreateRoleBindings creates role bindings associating users to roles
func CreateRoleBindings(ctx context.Context, roleID, userID string) ([]v3.CreateRoleBindings200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := v3.CreateRoleBindingsRequest{
		Bindings: []v3.CreateRoleBindingsRequestBindingsInner{
			{RoleId: roleID, UserId: userID},
		},
	}
	data, _, err := client.RolesAPI.CreateRoleBindings(ctx).CreateRoleBindingsRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// DeleteRoleBinding deletes a role binding by ID
func DeleteRoleBinding(ctx context.Context, bindingID string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, _, err = client.RolesAPI.DeleteRoleBinding(ctx, bindingID).Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}
