package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// GetCompliancePosture returns the organization compliance posture: every
// framework, whether it is enabled, and the security tests mapped to each of
// its compliance items.
func GetCompliancePosture(ctx context.Context) (*v3.GetCompliancePosture200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.ComplianceAPI.GetCompliancePosture(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// GetEnabledCompliances returns the compliance frameworks the organization
// turned on.
func GetEnabledCompliances(ctx context.Context) (*v3.GetEnabledCompliances200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.ComplianceAPI.GetEnabledCompliances(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// GetScanCompliance returns the compliance posture of one scan: per-framework
// enabled status and the compliance items the scan failed.
func GetScanCompliance(ctx context.Context, scanID string) (*v3.GetScanCompliance200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.ComplianceAPI.GetScanCompliance(ctx, scanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ListCompliancePerApplication returns one page of per-application compliance
// statistics. profileIDs is the comma-separated filter the route takes; an
// empty string keeps every application. size 0 keeps the API default page
// size. The returned total is totalCount.
func ListCompliancePerApplication(ctx context.Context, next string, size int, profileIDs string) ([]v3.GetCompliancePerApplication200ResponseDataInner, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ComplianceAPI.GetCompliancePerApplication(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if profileIDs != "" {
		req = req.ProfileIds(profileIDs)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}
