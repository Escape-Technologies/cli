// Package private provides the private location tunnel implementation
package private

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

// StartLocation starts a private location tunnel
func StartLocation(ctx context.Context, locationID string, sshPrivateKey ed25519.PrivateKey, healthy *atomic.Bool) error {
	log.Trace("Starting private location %s", locationID)
	for {
		err := dialSSH(ctx, locationID, sshPrivateKey, healthy)
		if ctx.Err() != nil {
			return fmt.Errorf("dialSSH: %w", ctx.Err())
		}
		if err != nil {
			log.Info("failed to dial ssh: %v, retrying...", err)
		} else {
			log.Info("Disconnected from SSH, retrying...")
		}
		time.Sleep(1 * time.Second)
	}
}
