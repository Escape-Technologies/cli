// Package locations provides the location start implementation
package locations

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/locations/health"
	"github.com/Escape-Technologies/cli/pkg/locations/private"
	"github.com/Escape-Technologies/cli/pkg/locations/private/kube"
	"github.com/Escape-Technologies/cli/pkg/locations/ssh"
	"github.com/Escape-Technologies/cli/pkg/log"
)

const (
	retryInterval = 100 * time.Millisecond
)

// Start the private location
func Start(ctx context.Context, name string) error {
	healthy := &atomic.Bool{}
	healthy.Store(false)
	go health.Start(ctx, healthy)

	log.Trace("Generating SSH Keys")
	sshPublicKey, sshPrivateKey, err := ssh.GenSSHKeys(name)
	if err != nil {
		return fmt.Errorf("unable to generate SSH keys: %w", err)
	}
	log.Debug("Generated SSH Key: %s", sshPublicKey)

	log.Trace("Creating location %s with public key %s", name, sshPublicKey)
	id, err := escape.UpsertLocation(ctx, name, sshPublicKey)
	if err != nil {
		return fmt.Errorf("unable to update private location on Escape Platform: %w", err)
	}

	go kube.Start(ctx, id, name, healthy)
	for {
		log.Trace("Creating location %s with public key %s", name, sshPublicKey)
		id, err := escape.UpsertLocation(ctx, name, sshPublicKey)
		if err != nil {
			log.Error("unable to update private location on Escape Platform: %s", err)
			time.Sleep(retryInterval)
			continue
		}
		log.Info("Private location %s in sync with Escape Platform, starting location...", name)
		err = private.StartLocation(ctx, id, sshPrivateKey, healthy)
		if err != nil {
			log.Error("Error starting location: %s", err)
		} else {
			log.Error("Unknown error starting location")
		}
		time.Sleep(retryInterval)
		if ctx.Err() != nil {
			return nil
		}
	}
}
