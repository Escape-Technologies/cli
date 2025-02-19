package privatelocation

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func StartLocation(ctx context.Context, locationId string, sshPrivateKey ed25519.PrivateKey) error {
	log.Info("Starting location")

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

	client, err := ssh.Dial("tcp", "a6afba49141654b028ed54a37057b33b-ce4ca0de8f229555.elb.us-east-1.amazonaws.com:2222", config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	listener, err := client.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("failed to create reverse tunnel: %w", err)
	}
	defer listener.Close()

	log.Info("Established reverse tunnel on remote port %d", listener.Addr().(*net.TCPAddr).Port)

	<-ctx.Done()
	return nil
}

func GenSSHKeys(name string) (string, ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", nil
	}

	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", nil
	}

	publicKeyString := "ssh-ed25519 " + base64.StdEncoding.EncodeToString(sshPublicKey.Marshal()) + " " + name + "@escape.tech"

	return publicKeyString, privateKey
}
