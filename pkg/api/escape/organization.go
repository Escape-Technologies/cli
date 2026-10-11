package escape

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// GetOrganizationSettings reads the organization settings.
func GetOrganizationSettings(ctx context.Context) (*v3.GetOrganizationSettings200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.OrganizationSettingsAPI.GetOrganizationSettings(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// UpdateOrganizationSettings updates the organization settings from raw JSON bytes.
func UpdateOrganizationSettings(ctx context.Context, body []byte) (*v3.GetOrganizationSettings200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var req v3.UpdateOrganizationSettingsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	data, _, err := client.OrganizationSettingsAPI.UpdateOrganizationSettings(ctx).UpdateOrganizationSettingsRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// CreateOrganizationScannerConfiguration publishes a new organization-wide
// scanner configuration from raw JSON bytes. The body carries the `content`
// object; profiles without their own override follow it.
func CreateOrganizationScannerConfiguration(ctx context.Context, body []byte) (*v3.CreateOrganizationScannerConfiguration200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var req v3.CreateOrganizationScannerConfigurationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	data, _, err := client.OrganizationSettingsAPI.CreateOrganizationScannerConfiguration(ctx).CreateOrganizationScannerConfigurationRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
