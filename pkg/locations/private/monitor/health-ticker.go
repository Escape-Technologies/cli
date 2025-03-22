package monitor

import (
	"context"
	"time"

	"golang.org/x/crypto/ssh"
)

func healthTicker(ctx context.Context, ch ssh.Channel) {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()
	for range ticker.C {
		ch.SendRequest("health", false, nil)
	}
}
