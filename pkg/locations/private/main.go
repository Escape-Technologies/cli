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
	for {
		err := dialSSH(ctx, locationID, sshPrivateKey, healthy)
		if err != nil {
			log.Error("Failed to dial SSH: %s, retrying...", err)
			errMsg := err.Error()
			if strings.Contains(errMsg, "connection refused") {
				log.Error("Firewall is likely blocking outbound connections to port 2222")
				log.Error("Ensure private-location.escape.tech:2222 outbound access is allowed")
			} else if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "could not resolve") {
				log.Error("DNS resolution failed for private-location.escape.tech")
			} 
		} else {
			log.Error("SSH connection lost, retrying...")
		}
		time.Sleep(1 * time.Second)
	}
}
