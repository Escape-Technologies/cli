package private

import (
	"context"
	"crypto/ed25519"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func StartLocation(ctx context.Context, locationId string, sshPrivateKey ed25519.PrivateKey) error {
	log.Info("Starting location")

	for {
		_, err := dialSSH(ctx, locationId, sshPrivateKey)
		if err != nil {
			log.Info("failed to dial ssh, retrying...")
			continue
		}

	}

	// <-ctx.Done()
	return nil
}

