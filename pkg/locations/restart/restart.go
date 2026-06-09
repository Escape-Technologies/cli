// Package restart auto restarts the location agent to pull the latest version of the agent
package restart

import (
	"math/rand/v2"
	"os"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

// jitter takes a duration and returns the same duration with a random jitter
// between duration-10% and duration+10%
func jitter(duration time.Duration) time.Duration {
	random := rand.Float64()*0.2 - 0.1 // nolint:mnd
	ratio := 1 + random
	return time.Duration(float64(duration) * ratio)
}

// Start starts the restart timeout
func Start() {
	restartTimeout := os.Getenv("ESCAPE_CLI_RESTART_INTERVAL")
	if restartTimeout == "" {
		log.Debug("ESCAPE_CLI_RESTART_INTERVAL not set, not restarting the private location")
		return
	}

	restartTimeoutDuration, err := time.ParseDuration(restartTimeout)
	if err != nil {
		log.Error("Failed to parse ESCAPE_CLI_RESTART_INTERVAL(%s): %v", restartTimeout, err)
		return
	}
	if restartTimeoutDuration <= 0 {
		log.Error("ESCAPE_CLI_RESTART_INTERVAL must be > 0, got %s", restartTimeoutDuration)
		return
	}

	finalDelay := jitter(restartTimeoutDuration)
	log.Debug("Restarting the private location at %s (in %s)", time.Now().Add(finalDelay), finalDelay.String())
	time.Sleep(finalDelay)

	log.Debug("Restart timeout reached, restarting the private location")
	time.Sleep(30 * time.Second) // nolint:mnd
	os.Exit(0)
}
