// Package private provides the private location tunnel implementation
package private

import (
	"context"
	"crypto/ed25519"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
)

// StartLocation starts a private location tunnel
func StartLocation(ctx context.Context, locationID string, sshPrivateKey ed25519.PrivateKey, healthy *atomic.Bool) error {
	log.Trace("Starting private location %s", locationID)
	ep, err := resolveEndpoint()
	if err != nil {
		return err
	}

	log.Debug("Connecting to the Escape Platform via Secure Tunnel at %s", ep.addr)

	var hasEverConnected atomic.Bool
	var hasLoggedTimeoutHint atomic.Bool
	var failureStartTime time.Time
	const failureThreshold = 1 * time.Minute

	for {
		err := dialSSH(ctx, ep, locationID, sshPrivateKey, healthy)
		if err != nil {
			if failureStartTime.IsZero() {
				failureStartTime = time.Now()
			}

			isTimeout := isDialTimeout(err)
			shouldLog := isTimeout && !hasEverConnected.Load()
			if !shouldLog && (!hasEverConnected.Load() || time.Since(failureStartTime) >= failureThreshold) {
				shouldLog = true
			}

			if shouldLog {
				log.Error("Failed to dial Secure Tunnel: %s, retrying...", err)
			}

			errMsg := err.Error()

			if isTimeout {
				logDialTimeoutHints(ep, &hasLoggedTimeoutHint)
			} else if strings.Contains(errMsg, "connection refused") {
				if !hasEverConnected.Load() {
					_, port, _ := net.SplitHostPort(ep.addr)
					log.Error("Firewall is likely blocking outbound connections to port %s", port)
					log.Error("Ensure %s outbound access is allowed", ep.addr)
				}
			} else if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "could not resolve") {
				host, _, _ := net.SplitHostPort(ep.addr)
				log.Error("DNS resolution failed for %s", host)
			} else if errors.Is(err, errUpgradeRefused) && ep.transport == transportWSS && !hasEverConnected.Load() {
				log.Error("A proxy or firewall on the path to %s refused or stripped the WebSocket Upgrade", ep.addr)
				log.Error("Set ESCAPE_TRANSPORT=auto to fall back to HTTP long-poll on the same port")
			}
		} else {
			hasEverConnected.Store(true)
			hasLoggedTimeoutHint.Store(false)
			failureStartTime = time.Time{}
			// First failure may just be a network issue, we dont want to notify the customer yet
			log.Info("Secure Tunnel connection lost, retrying...")
		}

		time.Sleep(1 * time.Second)
	}
}

func isDialTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	errMsg := err.Error()

	return strings.Contains(errMsg, "context deadline exceeded") || strings.Contains(errMsg, "i/o timeout")
}

func logDialTimeoutHints(ep endpoint, hasLoggedTimeoutHint *atomic.Bool) {
	if hasLoggedTimeoutHint.Load() {
		return
	}

	hasLoggedTimeoutHint.Store(true)
	dest := ep.addr
	if ep.connectURL != nil {
		dest = ep.connectURL.String()
	}

	log.Error("Timed out connecting to Escape Secure Tunnel endpoint (%s)", dest)

	if env.GetFrontendProxyURL() == nil {
		log.Error("Outbound traffic may require a proxy: set ESCAPE_FRONTEND_PROXY_URL on the deployment")
	}

	log.Error("Ensure %s is reachable from this network", ep.addr)
	if ep.transport == transportHTTPS || ep.transport == transportAuto {
		log.Error("HTTP long-poll also needs a proxy that does not buffer responses from %s", ep.addr)
	}
}
