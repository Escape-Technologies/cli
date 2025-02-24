package private

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func startListener(ctx context.Context, client *ssh.Client, healthy *atomic.Bool) error {
	listener, err := client.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("failed to create reverse tunnel: %w", err)
	}
	defer listener.Close()

	log.Info("Established reverse tunnel on remote port %d", listener.Addr().(*net.TCPAddr).Port)

	err = startSocks5Server(ctx, listener, healthy)
	healthy.Store(false)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("failed to start socks5 server: %w", err)
	}
	return nil
}
