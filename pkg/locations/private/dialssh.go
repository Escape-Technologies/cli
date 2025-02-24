package private

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"sync/atomic"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

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

	log.Info("Dialing locationID: %s", locationId)
	client, err := ssh.Dial("tcp", "private-location.escape.tech:2222", config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
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
