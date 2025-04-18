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
		ch.SendRequest("log", false, []byte(log.Message)) //nolint:errcheck
	})
}
