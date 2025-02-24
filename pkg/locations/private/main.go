package private

import (
	"context"
	"crypto/ed25519"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func StartLocation(ctx context.Context, locationId string, sshPrivateKey ed25519.PrivateKey) error {
	log.Info("Starting location")

	client, err := dialSSH(ctx, locationId, sshPrivateKey)
	if err != nil {
		dialSSH(ctx, locationId, sshPrivateKey)
	}
	defer client.Close()

	// <-ctx.Done()
	return nil
}

