package monitor

import (
	"context"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

const maxLogLevel = 4

func sendLogs(ctx context.Context, ch ssh.Channel) {
	go func() {
		<-ctx.Done()
		log.RemoveHook("monitor")
	}()
	log.AddHook("monitor", func(log log.Entry) {
		// Log levels: trace: 6, debug: 5, info: 4, warn: 3, error: 2, fatal: 1, panic: 0
		if log.Level <= maxLogLevel {
			ch.SendRequest("log", false, []byte(log.Message)) //nolint:errcheck
		}
	})
}
package monitor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

const (
	maxLogLevel     = 4
	logChannelSize  = 100
)

type logPayload struct {
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp_ns,omitempty"`
}

func sendLogs(ctx context.Context, ch ssh.Channel) {
	logChan := make(chan log.Entry, logChannelSize)
	ready := make(chan struct{})
	
	go func() {
		close(ready)
		for {
			select {
			case <-ctx.Done():
				log.RemoveHook("monitor")
				close(logChan)
				return
			case entry, ok := <-logChan:
				if !ok {
					return
				}
				payload := logPayload{
					Message:   entry.Message,
					Timestamp: time.Now().UnixMilli(),
				}
				data, err := json.Marshal(payload)
				if err != nil {
					ch.SendRequest("log", false, []byte(entry.Message)) //nolint:errcheck
				} else {
					ch.SendRequest("log", false, data) //nolint:errcheck
				}
			}
		}
	}()
	
	<-ready
	
	log.AddHook("monitor", func(entry log.Entry) {
		// Log levels: trace: 6, debug: 5, info: 4, warn: 3, error: 2, fatal: 1, panic: 0
		if entry.Level <= maxLogLevel {
			select {
			case logChan <- entry:
			case <-ctx.Done():
				return
			}
		}
	})
}
