package private

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

func StartListener(ctx context.Context, client *ssh.Client) (*net.Listener, error) {
	listener, err := client.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create reverse tunnel: %w", err)
	}
	defer listener.Close()

	return &listener, nil
}
