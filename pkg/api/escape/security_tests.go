package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// SecurityTestsFilters narrows the organization security-test catalog. Empty
// fields are left out of the query so the server keeps its own default.
type SecurityTestsFilters struct {
	Search        string
	Enabled       string
	ScannerKinds  []string
	AssetTypes    []string
	SortType      string
	SortDirection string
}

// ListSecurityTests returns one page of the security-test catalog.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListSecurityTests(ctx context.Context, next string, size int, filters *SecurityTestsFilters) ([]v3.SecurityTest, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.SecurityTestsAPI.ListSecurityTests(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if filters != nil {
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}

		if filters.Enabled != "" {
			req = req.Enabled(filters.Enabled)
		}

		if len(filters.ScannerKinds) > 0 {
			req = req.ScannerKinds(filters.ScannerKinds)
		}

		if len(filters.AssetTypes) > 0 {
			req = req.AssetTypes(filters.AssetTypes)
		}

		if filters.SortType != "" {
			req = req.SortType(filters.SortType)
		}

		if filters.SortDirection != "" {
			req = req.SortDirection(filters.SortDirection)
		}
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// GetSecurityTest returns one catalog entry with its organization
// configuration and the issue counts still open under it.
func GetSecurityTest(ctx context.Context, securityTestUID string) (*v3.GetSecurityTest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.SecurityTestsAPI.GetSecurityTest(ctx, securityTestUID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
