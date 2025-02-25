package kube

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/oapi-codegen/runtime/types"
)

func UpsertIntegration(ctx context.Context, locationId *types.UUID, locationName string) error {
	client, err := api.NewAPIClient()
	if err != nil {
		return err
	}
	integrations, err := client.GetV1IntegrationsKubernetesWithResponse(ctx)
	if err != nil {
		return err
	}

	for _, integration := range *integrations.JSON200 {
		if integration.LocationId == locationId {
			log.Info("Integration already exists, skipping")
			return nil
		}
		log.Info("Integration not found, creating")
		res, err := client.PostV1IntegrationsKubernetesWithResponse(ctx, api.PostV1IntegrationsKubernetesJSONRequestBody{
			LocationId: locationId,
			Name:       locationName,
		})
		if err != nil {
			return err
		}
		if res.JSON200 != nil {
			log.Info("Integration created successfully")
			return nil
		} else if res.JSON400 != nil {
			return fmt.Errorf("unable to create integration: %s", res.JSON400.Message)
		} else if res.JSON500 != nil {
			return fmt.Errorf("unable to create integration: %s", res.JSON500.Message)
		} else {
			return fmt.Errorf("unable to create integration: Unknown error")
		}
	}
	return nil
}
