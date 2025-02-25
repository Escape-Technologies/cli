package locations

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/locations/health"
	"github.com/Escape-Technologies/cli/pkg/locations/private"
	"github.com/Escape-Technologies/cli/pkg/locations/private/kube"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func Start(ctx context.Context, client *api.ClientWithResponses, name string) error {
	healthy := &atomic.Bool{}
	healthy.Store(false)
	go health.Start(ctx, healthy)

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
	if location.JSON200 != nil {
		go kube.Start(ctx, client, location.JSON200.Id, *location.JSON200.Name, healthy)
		for {
			err := private.StartLocation(ctx, location.JSON200.Id.String(), sshPrivateKey, healthy)
			if err != nil {
				log.Error("Error starting location: %s", err)
			} else {
				log.Error("Unknown error starting location")
			}
			time.Sleep(100 * time.Millisecond)
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
