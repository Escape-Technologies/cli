package private

import (
	"context"
	"crypto/ed25519"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func StartLocation(ctx context.Context, locationId string, sshPrivateKey ed25519.PrivateKey) error {
	log.Info("Starting location")

	for {
		err := dialSSH(ctx, locationId, sshPrivateKey)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			log.Info("failed to dial ssh: %v, retrying...", err)

		} else {
			log.Info("Disconnected from SSH, retrying...")
		}
		time.Sleep(1 * time.Second)
	}
}
