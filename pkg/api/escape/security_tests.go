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

// UpdateSecurityTest changes the organization configuration of one security
// test: whether it is enabled, and the severity override its issues report.
// Omitting customSeverity leaves the override untouched. A JSON null, carried
// on AdditionalProperties, resets it to the scanner default. An empty string
// is not a severity and is not sent.
func UpdateSecurityTest(ctx context.Context, securityTestUID string, body v3.UpdateSecurityTestRequest) (*v3.GetSecurityTest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.SecurityTestsAPI.UpdateSecurityTest(ctx, securityTestUID).UpdateSecurityTestRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// BulkUpdateSecurityTests changes the organization configuration of every
// security test matching a filter, or of all of them when body.All is set.
// A body with All set cannot carry a filter. DryRun previews the matching
// security test IDs without changing anything.
func BulkUpdateSecurityTests(ctx context.Context, body v3.BulkUpdateSecurityTestsRequest) (*v3.BulkUpdateSecurityTests200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.SecurityTestsAPI.BulkUpdateSecurityTests(ctx).BulkUpdateSecurityTestsRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
