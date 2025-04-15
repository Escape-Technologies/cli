package kube

import (
	"context"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/oapi-codegen/runtime/types"
)

func UpsertIntegration(ctx context.Context, locationId *types.UUID, locationName string) error {
	log.Trace("Upserting location %s", locationName)
	// TODO(quentin@escape.tech): Implement this
	// res, err := client.UpsertKubernetesIntegrationWithResponse(ctx, v1.UpsertKubernetesIntegrationJSONRequestBody{
	// 	Name:       locationName,
	// 	LocationId: locationId,
	// })
	// if err != nil {
	// 	return err
	// }
	// if res.JSON200 != nil {
	// 	log.Info("Kubernetes integration %s created successfully", locationName)
	// 	return nil
	// } else if res.JSON400 != nil {
	// 	log.Error("Kubernetes integration %s creation failed", locationName)
	// 	for _, evt := range res.JSON400.Events {
	// 		log.Debug("Event: %s", evt.Logline)
	// 	}
	// 	return fmt.Errorf("unable to create integration: %s", res.JSON400.Message)
	// } else if res.JSON500 != nil {
	// 	return fmt.Errorf("unable to create integration: %s", res.JSON500.Message)
	// } else {
	// 	return fmt.Errorf("unable to create integration: Unknown error")
	// }
	return nil
}
