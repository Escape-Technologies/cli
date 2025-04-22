package private

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
	socks5 "github.com/Escape-Technologies/go-socks5"
)

func startSocks5Server(ctx context.Context, listener net.Listener, healthy *atomic.Bool) error {
	log.Trace("Starting socks5 server")
	socks5Server, err := socks5.New(&socks5.Config{
		Dial: env.BuildProxyDialer(env.GetBackendProxyURL()),
	})
	if err != nil {
		return fmt.Errorf("failed to create socks5 server config: %w", err)
	}
	log.Info("Private location ready to accept connections")
	healthy.Store(true)

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		listener.Close() //nolint:errcheck
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
