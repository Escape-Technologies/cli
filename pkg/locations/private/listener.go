package private

import (
	"context"
	"fmt"
	"net"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func startListener(ctx context.Context, client *ssh.Client) error {
	listener, err := client.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("failed to create reverse tunnel: %w", err)
	}

	for {
		log.Info("Established reverse tunnel on remote port %d", listener.Addr().(*net.TCPAddr).Port)
		err = startSocks5Server(ctx, listener)
		if err == nil {
			log.Info("failed to serve socks5 server, retrying...")
			continue
		}
	}
}
