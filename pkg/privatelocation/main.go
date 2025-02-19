package privatelocation

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func StartLocation(ctx context.Context, key string) error {
	log.Info("Location started with %s", key)
	time.Sleep(2 * time.Second)
	return nil
}

func GenSSHKeys() (string, ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", nil
	}

	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", nil
	}

	publicKeyString := "ssh-ed25519 " + base64.StdEncoding.EncodeToString(sshPublicKey.Marshal())

	return publicKeyString, privateKey
}
