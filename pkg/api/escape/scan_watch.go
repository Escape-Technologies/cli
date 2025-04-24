package escape

import (
	"context"
	"fmt"
	"time"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/log"
)

const (
	watchScanInterval = 5 * time.Second
	watchScanMaxTries = 5
)

// WatchScan watches scans status and logs
func WatchScan(ctx context.Context, scanID string) (chan *v2.ListScans200ResponseDataInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	ch := make(chan *v2.ListScans200ResponseDataInner)
	go func() {
		defer close(ch)
		tries := 0
		lastProgressRatio := float32(0.0)
		for {
			time.Sleep(watchScanInterval)
			scan, _, err := client.ScansAPI.GetScan(ctx, scanID).Execute()
			if err != nil {
				log.Info("unable to get scan: %s", err.Error())
				tries++
				if tries > watchScanMaxTries {
					log.Error("unable to get scan more than %d times, stopping watch: %s", watchScanMaxTries, err.Error())
					return
				}
				continue
			}
			tries = 0
			if lastProgressRatio == scan.ProgressRatio {
				continue
			}
			ch <- scan
			lastProgressRatio = scan.ProgressRatio
			if scan.Status != v2.ENUME48DD51FE8A350A4154904ABF16320D7_STARTING &&
				scan.Status != v2.ENUME48DD51FE8A350A4154904ABF16320D7_RUNNING {
				log.Info("Scan ended with status %s", scan.Status)
				return
			}
		}
	}()
	return ch, nil
}
