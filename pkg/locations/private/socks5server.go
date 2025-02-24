package private

import (
	"context"
	"fmt"
	"net"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/socks5"
)

func startSocks5Server(ctx context.Context, listener net.Listener) error {

	log.Info("Starting socks5 server")
	socks5Server, err := socks5.New(&socks5.Config{})
	if err != nil {
		return fmt.Errorf("failed to create socks5 server: %w", err)
	}

	log.Info("Socks5 server started")
	err = socks5Server.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to serve socks5 server: %w", err)
	}

	return nil
}

