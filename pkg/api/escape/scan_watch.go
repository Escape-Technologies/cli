package escape

import (
	"context"
	"fmt"
	"time"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/log"
)

const (
	watchScanInterval = 5 * time.Second
	watchScanMaxTries = 5
)

// WatchScan watches scans status and logs
func WatchScan(ctx context.Context, scanID string) (chan *v3.ScanDetailed1, error) {
	client, err := NewAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	ch := make(chan *v3.ScanDetailed1)
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
			if scan.Status != "STARTING" &&
				scan.Status != "RUNNING" {
				log.Info("Scan ended with status %s", scan.Status)
				ch <- scan
				return
			}
			if lastProgressRatio == scan.ProgressRatio {
				continue
			}
			ch <- scan
			lastProgressRatio = scan.ProgressRatio
		}
	}()
	return ch, nil
}
