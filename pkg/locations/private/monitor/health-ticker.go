package monitor

import (
	"context"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

const (
	healthTickerInterval = 5 * time.Second
)

func healthTicker(ctx context.Context, ch ssh.Channel) {
	ticker := time.NewTicker(healthTickerInterval)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()
	var err error
	for range ticker.C {
		_, err = ch.SendRequest("health", false, nil)
		if err != nil {
			log.Debug("Failed to send health request: %s", err)
		}
	}
}
