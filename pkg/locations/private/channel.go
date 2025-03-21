package private

import (
	"context"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func openEscapeChannel(ctx context.Context, client *ssh.Client) (ssh.Channel, error) {
	ch, _, err := client.OpenChannel("escape", nil)
	if err != nil {
		log.Error("failed to open escape channel: %v", err)
		return nil, err
	}
	log.Info("escape channel opened")

	go func() {
		<-ctx.Done()
		ch.Close()
	}()

	return ch, nil
}