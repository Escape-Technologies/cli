package private

import (
	"context"
	"fmt"
	"net"

	"github.com/Escape-Technologies/cli/pkg/log"
	socks5 "github.com/Escape-Technologies/go-socks5"
)

func startSocks5Server(ctx context.Context, listener net.Listener) error {
	log.Info("Starting socks5 server")
	socks5Server, err := socks5.New(&socks5.Config{})
	if err != nil {
		return fmt.Errorf("failed to create socks5 server config: %w", err)
	}
	log.Info("Socks5 server started")

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		listener.Close()
		errChan <- nil
	}()
	go func() {
		err = socks5Server.Serve(listener)
		if err != nil {
			errChan <- fmt.Errorf("failed to serve socks5 server: %w", err)
		} else {
			errChan <- nil
		}
	}()

	return <-errChan
}
