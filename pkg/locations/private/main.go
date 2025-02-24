package private

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func StartLocation(ctx context.Context, locationId string, sshPrivateKey ed25519.PrivateKey) error {
	log.Info("Starting location")

	client, err := DialSSH(ctx, locationId, sshPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()


	listener, err := StartListener(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer (*listener).Close()

	err = StartSocks5Server(ctx, listener)
	if err != nil {
		return fmt.Errorf("failed to start socks5 server: %w", err)
	}

	// Wait for context cancellation
	// <-ctx.Done()
	return nil
}

