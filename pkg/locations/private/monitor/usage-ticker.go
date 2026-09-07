package monitor

import (
	"context"
	"time"

	"github.com/Escape-Technologies/cli/pkg/locations/stats"
	"github.com/Escape-Technologies/cli/pkg/log"
)

const usageTickerInterval = 5 * time.Minute

func usageTicker(ctx context.Context) {
	ticker := time.NewTicker(usageTickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := stats.SnapshotAndReset()
			if snapshot.Requests == 0 && snapshot.DNS == 0 {
				continue
			}
			log.Info(
				"Sent %d requests and %d DNS requests in the last 5 minutes",
				snapshot.Requests,
				snapshot.DNS,
			)
		}
	}
}
