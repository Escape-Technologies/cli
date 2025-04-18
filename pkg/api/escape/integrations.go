package escape

import (
	"context"
	"fmt"
	"os"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

// ListIntegrations Lists all integrations
func ListIntegrations(ctx context.Context) ([]v2.ListIntegrations200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.IntegrationsAPI.ListIntegrations(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get integrations: %w", err)
	}
	return data, nil
}

// GetIntegration Get an integration by ID
func GetIntegration(ctx context.Context, id string) (*v2.GetIntegration200Response, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.IntegrationsAPI.GetIntegration(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get integration: %w", err)
	}
	return data, nil
}

// DeleteIntegration Deletes an integration
func DeleteIntegration(ctx context.Context, id string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, _, err = client.IntegrationsAPI.DeleteIntegration(ctx, id).Execute()
	if err != nil {
		return fmt.Errorf("unable to delete integration: %w", err)
	}
	return nil
}

// UpsertIntegration Upserts an integration
func UpsertIntegration(ctx context.Context, name string, locationID *string, integration *v2.GetIntegration200ResponseData) error {
	err := createIntegration(ctx, name, locationID, integration)
	if err == nil {
		return nil
	}
	id, err := extractConflict(err)
	if err != nil {
		return err
	}
	return updateIntegration(ctx, id, name, locationID, integration)
}

// UpsertIntegrationFromFile Upserts an integration from a file
func UpsertIntegrationFromFile(ctx context.Context, filePath string) error {
	body, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}
	integration, err := parseJSONOrYAML(body, &v2.UpdateIntegrationRequest{})
	if err != nil {
		return fmt.Errorf("unable to parse file: %w", err)
	}
	return UpsertIntegration(ctx, integration.Name, integration.LocationId, &integration.Data)
}

func createIntegration(ctx context.Context, name string, locationID *string, integration *v2.GetIntegration200ResponseData) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := v2.UpdateIntegrationRequest{
		Name:       name,
		LocationId: locationID,
	}
	if integration != nil {
		req.Data = *integration
	}
	_, _, err = client.IntegrationsAPI.CreateIntegration(ctx).UpdateIntegrationRequest(req).Execute()
	if err != nil {
		return fmt.Errorf("unable to create integration: %w", err)
	}
	return nil
}

func updateIntegration(ctx context.Context, id string, name string, locationID *string, integration *v2.GetIntegration200ResponseData) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := v2.UpdateIntegrationRequest{
		Name:       name,
		LocationId: locationID,
	}
	if integration != nil {
		req.Data = *integration
	}
	_, _, err = client.IntegrationsAPI.UpdateIntegration(ctx, id).UpdateIntegrationRequest(req).Execute()
	if err != nil {
		return fmt.Errorf("unable to update integration: %w", err)
	}
	return nil
}
