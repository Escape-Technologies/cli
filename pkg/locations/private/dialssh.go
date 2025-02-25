package private

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net"
	"os"
	"sync/atomic"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func getClient(target string, conn net.Conn, config *ssh.ClientConfig) (*ssh.Client, error) {
	c, chans, reqs, err := ssh.NewClientConn(conn, target, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func dialSSH(ctx context.Context, locationId string, sshPrivateKey ed25519.PrivateKey, healthy *atomic.Bool) error {
	log.Debug("Creating signer from private key")
	signer, err := ssh.NewSignerFromKey(sshPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer: %w", err)
	}

	config := &ssh.ClientConfig{
		User: locationId,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	targetURL := os.Getenv("ESCAPE_PRIVATE_LOCATION_URL")
	if targetURL == "" {
		targetURL = "private-location.escape.tech:2222"
	}
	proxyURL := os.Getenv("ESCAPE_REPEATER_PROXY_URL")

	log.Info("Getting conn for target: %s", targetURL)
	conn, err := getConn(ctx, targetURL, proxyURL)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("failed to get conn: %w", err)
	}

	client, err := getClient(targetURL, conn, config)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	log.Info("Starting listener")
	err = startListener(ctx, client, healthy)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	return nil
}
