package monitor

import (
	"context"

	"golang.org/x/crypto/ssh"
)

func Start(ctx context.Context, client *ssh.Client) {
	ch, err := openEscapeChannel(ctx, client)
	if err != nil {
		return
	}

	sendLogs(ctx, ch)
	go healthTicker(ctx, ch)
}
