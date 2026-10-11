package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// maxRolePageSize is the public API page cap. Draining pages at the cap keeps
// the number of requests minimal for a large organization.
const maxRolePageSize = 100

// cursorResponseHeader carries the next-page cursor of a bare-array listing.
const cursorResponseHeader = "X-Next-Cursor"

// drainCursorPages follows the X-Next-Cursor header of a bare-array listing
// until the API stops returning one. fetch receives the cursor to start from
// and returns the page plus the next cursor from the response header. Roles
// and custom rules are bare arrays: a single request silently capped the CLI
// listing at the default page size, and the header is the only continuation.
func drainCursorPages[T any](fetch func(cursor string) ([]T, *http.Response, error)) ([]T, error) {
	all := []T{}
	seen := map[string]struct{}{}
	cursor := ""
	for {
		page, httpRes, err := fetch(cursor)
		if err != nil {
			return nil, err
		}

		all = append(all, page...)

		next := ""
		if httpRes != nil {
			next = httpRes.Header.Get(cursorResponseHeader)
		}

		if next == "" {
			return all, nil
		}

		// A server that returns the same cursor forever would otherwise loop,
		// and returning the rows so far would pass a truncated list off as complete.
		if _, dup := seen[next]; dup {
			return nil, fmt.Errorf("pagination did not advance: cursor %q was returned twice", next)
		}

		seen[next] = struct{}{}
		cursor = next
	}
}

// ListRoles lists every role in the organization, draining every page the
// X-Next-Cursor header announces.
func ListRoles(ctx context.Context) ([]v3.ListRoles200ResponseInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	roles, err := drainCursorPages(func(cursor string) ([]v3.ListRoles200ResponseInner, *http.Response, error) {
		req := client.RolesAPI.ListRoles(ctx).Size(maxRolePageSize)
		if cursor != "" {
			req = req.Cursor(cursor)
		}

		data, httpRes, err := req.Execute()
		if err != nil {
			return nil, nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
		}

		return data, httpRes, nil
	})
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRole gets a role by ID
func GetRole(ctx context.Context, roleID string) (*v3.CreateRole200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.RolesAPI.GetRole(ctx, roleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// DeleteRoleBinding deletes a role binding by ID.
func DeleteRoleBinding(ctx context.Context, bindingID string) (*v3.DeleteCustomRule200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.RolesAPI.DeleteRoleBinding(ctx, bindingID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// DeleteRole deletes a role by ID.
func DeleteRole(ctx context.Context, roleID string) error {
	if err := rawRequest(ctx, http.MethodDelete, rawPath("roles", roleID), nil, nil); err != nil {
		return fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return nil
}
