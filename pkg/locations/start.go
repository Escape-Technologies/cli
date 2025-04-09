package locations

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/locations/health"
	"github.com/Escape-Technologies/cli/pkg/locations/private"
	"github.com/Escape-Technologies/cli/pkg/locations/private/kube"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func Start(ctx context.Context, name string) error {
	healthy := &atomic.Bool{}
	healthy.Store(false)
	go health.Start(ctx, healthy)

	log.Trace("Generating SSH Keys")
	sshPublicKey, sshPrivateKey := private.GenSSHKeys(name)
	log.Debug("Generated SSH Key: %s", sshPublicKey)

	log.Trace("Creating location %s with public key %s", name, sshPublicKey)
	location, err := escape.CreateLocation(ctx, escape.Location{
		Name:              name,
		SSHPublicKey:      sshPublicKey,
	})
	if err != nil {
		return err
	}

	go kube.Start(ctx, &location.Id, location.Name, healthy)
	for {
		log.Info("Private location %s in sync with Escape Platform, starting location...", name)
		err := private.StartLocation(ctx, location.Id.String(), sshPrivateKey, healthy)
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
}
