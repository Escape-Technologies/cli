package locations

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/privatelocation"
)


func Start(ctx context.Context, client *api.ClientWithResponses, name string) error {
	publicSSHKey, privateSSHKey := privatelocation.GenSSHKeys()
	log.Info("Private SSH Key: %s", privateSSHKey)
	log.Info("Public SSH Key: %s", publicSSHKey)

	log.Info("Upserting location %s", name)
	y := true
	location, err := client.UpsertLocationWithResponse(ctx, api.UpsertLocationJSONRequestBody{
		Name:              name,
		PrivateLocationV2: &y,
		Key:               &publicSSHKey,
	})
	if err != nil {
		return err
	}

	if location.JSON200 != nil {
		return privatelocation.StartLocation(ctx, string(privateSSHKey))
	} else if location.JSON400 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON400.Message)
	} else if location.JSON500 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON500.Message)
	} else {
		return fmt.Errorf("unable to start location: Unknown error")
	}
}
