// Package private provides the private location tunnel implementation
package private

import (
	"context"
	"crypto/ed25519"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

// StartLocation starts a private location tunnel
func StartLocation(ctx context.Context, locationID string, sshPrivateKey ed25519.PrivateKey, healthy *atomic.Bool) error {
	log.Trace("Starting private location %s", locationID)
	for {
		err := dialSSH(ctx, locationID, sshPrivateKey, healthy)
		if err != nil {
			log.Error("Failed to dial SSH: %s, retrying...", err)
		} else {
			log.Error("SSH connection lost, retrying...")
		}
		time.Sleep(1 * time.Second)
	}
}
