package kube

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/oapi-codegen/runtime/types"
)


func CreateIntegration(ctx context.Context, client *api.ClientWithResponses, locationId *types.UUID, locationName string) error {

	var rawJson = []byte(`
	{
		"name": "` + locationName + `",
		"locationId": "` + locationId.String() + `"
	}
	`)
	log.Trace("Creating integration %s", locationName)
	res, err := client.CreateIntegrationWithBodyWithResponse(ctx, "application/json", bytes.NewBuffer(rawJson))
	if err != nil {
		return err
	}
	if res.JSON200 != nil {
		log.Info("Kubernetes integration %s created successfully", locationName)
		return nil
	} else if res.JSON400 != nil {
		log.Error("Kubernetes integration %s creation failed", locationName)
		return fmt.Errorf("unable to create integration: %s", res.JSON400.Message)
	} else if res.JSON500 != nil {
		return fmt.Errorf("unable to create integration: %s", res.JSON500.Message)
	} else {
		return fmt.Errorf("unable to create integration: Unknown error")
	}
}
