package locations

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	v1 "github.com/Escape-Technologies/cli/pkg/api/v1"
	"github.com/Escape-Technologies/cli/pkg/locations/health"
	"github.com/Escape-Technologies/cli/pkg/locations/private"
	"github.com/Escape-Technologies/cli/pkg/locations/private/kube"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func Start(ctx context.Context, client *v1.ClientWithResponses, name string) error {
	healthy := &atomic.Bool{}
	healthy.Store(false)
	go health.Start(ctx, healthy)

	log.Trace("Generating SSH Keys")
	sshPublicKey, sshPrivateKey := private.GenSSHKeys(name)
	log.Debug("Generated SSH Key: %s", sshPublicKey)

	log.Trace("Upserting location %s with public key %s", name, sshPublicKey)
	y := true
	location, err := client.UpsertLocationWithResponse(ctx, v1.UpsertLocationJSONRequestBody{
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
			log.Info("Private location %s in sync with Escape Platform, starting location...", name)
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
