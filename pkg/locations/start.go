package locations

import (
	"context"
	"fmt"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/locations/private"
	"github.com/Escape-Technologies/cli/pkg/log"
)


func Start(ctx context.Context, client *api.ClientWithResponses, name string) error {
	sshPublicKey, sshPrivateKey := private.GenSSHKeys(name)
	log.Info("Generated public SSH Key: %s", sshPublicKey)

	log.Info("Upserting location %s with public key %s", name, sshPublicKey)
	y := true
	location, err := client.UpsertLocationWithResponse(ctx, api.UpsertLocationJSONRequestBody{
		Name:              name,
		PrivateLocationV2: &y,
		Key:               &sshPublicKey,
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if location.JSON200 != nil {
		for {
			err := private.StartLocation(ctx,location.JSON200.Id.String(), sshPrivateKey)
			if err != nil {
				log.Error("Error starting location: %s", err)
			} else {
				log.Error("Unknown error starting location")
			}
			time.Sleep(1 * time.Second)
			if ctx.Err() != nil {
				return nil
			}
		}
	} else if location.JSON400 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON400.Message)
	} else if location.JSON500 != nil {
		return fmt.Errorf("unable to start location: %s", location.JSON500.Message)
	} else {
		return fmt.Errorf("unable to start location: Unknown error")
	}
}
