package private

import (
	"context"
	"fmt"
	"net"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/socks5"
)

func StartSocks5Server(ctx context.Context, listener *net.Listener) error {
	socks5Server, err := socks5.New(&socks5.Config{})
	if err != nil {
		return fmt.Errorf("failed to create socks5 server: %w", err)
	}

	log.Info("Established reverse tunnel on remote port %d", (*listener).Addr().(*net.TCPAddr).Port)
	log.Info("SOCKS5 server created, waiting for connections")

	return socks5Server.Serve(*listener)
}

