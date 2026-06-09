// Package dns provides the DNS integration for private locations
package dns

import (
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

const (
	listenAddr    = "127.0.0.1:1053"
	minBackoff    = 100 * time.Millisecond
	maxBackoff    = 5 * time.Second
	backoffFactor = 2
)

// Start runs the private-location DNS forwarder under a supervised loop: it
// discovers upstream resolvers, serves until the listener exits, then backs off
// and retries. Upstreams are recomputed on each iteration so network changes
// are picked up. This function blocks and is meant to run in its own goroutine.
func Start() {
	backoff := minBackoff
	for {
		upstreams := resolveUpstreams()
		log.Debug("[DNS] starting TCP DNS server on %s, forwarding to %s", listenAddr, strings.Join(upstreams, ", "))

		err := serve(listenAddr, newHandler(upstreams))
		log.Error("[DNS] server stopped: %v, restarting in %s", err, backoff)

		time.Sleep(backoff)
		backoff = nextBackoff(backoff)
	}
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * backoffFactor
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}
