package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListPermissions returns every permission a role can be granted. The returned
// ids are the values `roles create` and `roles update` accept.
func ListPermissions(ctx context.Context) ([]v3.PermissionCatalogEntry, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.RolesAPI.ListPermissions(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
