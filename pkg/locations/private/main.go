// Package private provides the private location tunnel implementation
package private

import (
	"context"
	"crypto/ed25519"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

// StartLocation starts a private location tunnel
func StartLocation(ctx context.Context, locationID string, sshPrivateKey ed25519.PrivateKey, healthy *atomic.Bool) error {
	log.Trace("Starting private location %s", locationID)
	var hasEverConnected atomic.Bool
	var failureStartTime time.Time
	const failureThreshold = 1 * time.Minute

	for {
		err := dialSSH(ctx, locationID, sshPrivateKey, healthy)
		if err != nil {
			if failureStartTime.IsZero() {
				failureStartTime = time.Now()
			}

			if !hasEverConnected.Load() || time.Since(failureStartTime) >= failureThreshold {
				log.Error("Failed to dial SSH: %s, retrying...", err)
			}
			errMsg := err.Error()

			if strings.Contains(errMsg, "connection refused") {
				if !hasEverConnected.Load() {
					log.Error("Firewall is likely blocking outbound connections to port 2222")
					log.Error("Ensure private-location.escape.tech:2222 outbound access is allowed")
				}
			} else if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "could not resolve") {
				log.Error("DNS resolution failed for private-location.escape.tech")
			}
		} else {
			hasEverConnected.Store(true)
			failureStartTime = time.Time{}
			// First failure may just be a network issue, we dont want to notify the customer yet
			log.Info("SSH connection lost, retrying...")
		}
		time.Sleep(1 * time.Second)
	}
}
