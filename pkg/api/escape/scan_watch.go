package escape

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/log"
)

// ScanWatchResult is a struct that contains an event alongside with the scan status
type ScanWatchResult struct {
	Status        v2.EnumE48dd51fe8a350a4154904abf16320d7 `json:"status"`
	ProgressRatio float32                                 `json:"progressRatio"`
	CreatedAt     time.Time                               `json:"createdAt"`
	Level         v2.EnumAc8825c946764c840068c1a5eddeee84 `json:"level"`
	Title         string                                  `json:"title"`
	Description   string                                  `json:"description"`
}

const (
	watchScanInterval = 5 * time.Second
	watchScanMaxTries = 5
)

// WatchScan watches scans status and logs
func WatchScan(ctx context.Context, scanID string) (chan *ScanWatchResult, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	scan, _, err := client.ScansAPI.GetScan(ctx, scanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get scan: %w", err)
	}
	var scanLock sync.Mutex
	var lastEventID string
	var after *string
	ch := make(chan *ScanWatchResult)
	shouldStop := atomic.Bool{}
	shouldStop.Store(false)
	go func() {
		tries := 0
		for {
			if shouldStop.Load() {
				return
			}
			time.Sleep(watchScanInterval)
			data, _, err := client.ScansAPI.GetScan(ctx, scanID).Execute()
			if err != nil {
				log.Info("unable to get scan: %s", err.Error())
				tries++
				if tries > watchScanMaxTries {
					log.Error("unable to get scan more than %d times, stopping watch: %s", watchScanMaxTries, err.Error())
					close(ch)
					shouldStop.Store(true)
					return
				}
				continue
			}
			tries = 0
			scanLock.Lock()
			scan = data
			scanLock.Unlock()
		}
	}()
	go func() {
		tries := 0
		hasMore := true
		for {
			if !hasMore {
				time.Sleep(watchScanInterval)
			}
			if shouldStop.Load() {
				return
			}
			req := client.ScansAPI.ListEvents(ctx, scanID)
			if after != nil {
				req = req.After(*after)
			}
			data, _, err := req.Execute()
			if err != nil {
				log.Info("unable to get scan events: %s", err.Error())
				tries++
				if tries > watchScanMaxTries {
					log.Error("unable to get scan events more than %d times, stopping watch: %s", watchScanMaxTries, err.Error())
					close(ch)
					shouldStop.Store(true)
					return
				}
				continue
			}
			tries = 0
			after = &data.NextCursor
			hasMore = false
			if len(data.Data) == 0 {
				continue
			}
			for _, event := range filterEvents(data.Data, lastEventID) {
				scanLock.Lock()
				ch <- &ScanWatchResult{
					Status:        scan.Status,
					ProgressRatio: scan.ProgressRatio,
					CreatedAt:     event.CreatedAt,
					Description:   event.Description,
					Level:         event.Level,
					Title:         event.Title,
				}
				scanLock.Unlock()
				hasMore = true
			}
			lastEventID = data.Data[len(data.Data)-1].Id
		}
	}()
	return ch, nil
}

func filterEvents(events []v2.ListEvents200ResponseDataInner, lastEventID string) []v2.ListEvents200ResponseDataInner {
	for i, event := range events {
		if event.Id == lastEventID {
			return events[i+1:]
		}
	}
	return events
}
