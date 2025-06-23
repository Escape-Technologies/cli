package monitor

import (
	"context"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func sendLogs(ctx context.Context, ch ssh.Channel) {
	go func() {
		<-ctx.Done()
		log.RemoveHook("monitor")
	}()
	log.AddHook("monitor", func(log log.Entry) {
		// Log levels: trace: 6, debug: 5, info: 4, warn: 3, error: 2, fatal: 1, panic: 0
		if log.Level <= 4 {
			ch.SendRequest("log", false, []byte(log.Message)) //nolint:errcheck
		}
	})
}
