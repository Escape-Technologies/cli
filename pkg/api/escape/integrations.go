package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"gopkg.in/yaml.v2"
)

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

func CreateIntegration(ctx context.Context, integration *v2.UpdateIntegrationRequest) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, _, err = client.IntegrationsAPI.CreateIntegration(ctx).UpdateIntegrationRequest(*integration).Execute()
	if err != nil {
		return fmt.Errorf("unable to create integration: %w", err)
	}
	return nil
}

func UpdateIntegration(ctx context.Context, id string, integration *v2.UpdateIntegrationRequest) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, _, err = client.IntegrationsAPI.UpdateIntegration(ctx, id).UpdateIntegrationRequest(*integration).Execute()
	if err != nil {
		return fmt.Errorf("unable to update integration: %w", err)
	}
	return nil
}

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

func UpsertIntegration(ctx context.Context, integration *v2.UpdateIntegrationRequest) error {
	err := CreateIntegration(ctx, integration)
	if err == nil {
		return nil
	}
	id, err := extractConflict(err)
	if err != nil {
		return err
	}
	return UpdateIntegration(ctx, id, integration)
}

func UpsertIntegrationFromFile(ctx context.Context, filePath string) error {
	integration := &v2.UpdateIntegrationRequest{}
	format := "json"
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		format = "yaml"
	} else if strings.HasSuffix(filePath, ".json") {
		format = "json"
	} else {
		return fmt.Errorf("unsupported file format: %s", filePath)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()
	if format == "json" {
		err = json.NewDecoder(file).Decode(integration)
		if err != nil {
			return fmt.Errorf("failed to decode file %s as JSON: %w", filePath, err)
		}
	} else {
		err = yaml.NewDecoder(file).Decode(integration)
		if err != nil {
			return fmt.Errorf("failed to decode file %s as YAML: %w", filePath, err)
		}
	}
	return UpsertIntegration(ctx, integration)
}
