package private

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/locations/private/monitor"
	"github.com/Escape-Technologies/cli/pkg/log"

	"golang.org/x/crypto/ssh"
)

const sshDialTimeout = 30 * time.Second

func getClient(target string, conn net.Conn, config *ssh.ClientConfig) (*ssh.Client, error) {
	c, chans, reqs, err := ssh.NewClientConn(conn, target, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn: %w", err)
	}

	return ssh.NewClient(c, chans, reqs), nil
}

func dialSSH(ctx context.Context, ep endpoint, locationID string, sshPrivateKey ed25519.PrivateKey, healthy *atomic.Bool) error {
	log.Debug("Creating signer from private key")
	signer, err := ssh.NewSignerFromKey(sshPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer: %w", err)
	}

	config := &ssh.ClientConfig{
		User: locationID,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	proxyURL := env.GetFrontendProxyURL()

	log.Trace("Getting %s conn for target: %s", ep.transport, ep.addr)
	dialCtx, cancel := context.WithTimeout(ctx, sshDialTimeout)
	defer cancel()
	conn, err := getConn(dialCtx, ep, proxyURL)
	if err != nil {
		return fmt.Errorf("failed to get conn: %w", err)
	}

	if ep.transport != transportSSH {
		// ssh.NewClientConn ignores config.Timeout (only ssh.Dial reads it), so the conn deadline
		// is what bounds the handshake. Over HTTPS a stalled proxy or poll session does not
		// reset the connection the way a dead TCP peer does, and would hang here forever.
		config.Timeout = sshDialTimeout
		_ = conn.SetDeadline(time.Now().Add(sshDialTimeout))
	}

	client, err := getClient(ep.addr, conn, config)
	if err != nil {
		return fmt.Errorf("failed to create Secure Tunnel client: %w", err)
	}
	// Ends the WebSocket or poll session goroutines once the tunnel is gone.
	defer client.Close() //nolint:errcheck

	_ = conn.SetDeadline(time.Time{})

	log.Debug("Secure Tunnel connection established to Escape Platform")
	listenerCtx, listenerCancel := context.WithCancel(ctx)
	go monitor.Start(listenerCtx, client)

	log.Trace("Starting listener")
	err = startListener(listenerCtx, client, healthy)
	listenerCancel()
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	return nil
}
