// Package monitor provides a way to monitor the private location on the Escape Platform
package monitor

import (
	"context"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

// Start the reporting data to Escape Platform
func Start(ctx context.Context, client *ssh.Client) {
	ch, err := openEscapeChannel(ctx, client)
	if err != nil {
		log.Error("Failed to establish log forwarding to Escape Platform: %s", err)
		return
	}

	sendLogs(ctx, ch)
	go healthTicker(ctx, ch)
}
