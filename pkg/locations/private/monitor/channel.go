package monitor

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func openEscapeChannel(ctx context.Context, client *ssh.Client) (ssh.Channel, error) {
	ch, _, err := client.OpenChannel("escape", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open log channel: %w", err)
	}
	log.Trace("escape channel opened")

	go func() {
		<-ctx.Done()
		ch.Close() //nolint:errcheck
	}()

	return ch, nil
}
