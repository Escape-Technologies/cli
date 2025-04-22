// Package ssh provides SSH key generation
package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// GenSSHKeys generates a new SSH key pair
func GenSSHKeys(name string) (string, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	publicKeyString := "ssh-ed25519 " + base64.StdEncoding.EncodeToString(sshPublicKey.Marshal()) + " " + name + "@escape.tech"

	return publicKeyString, privateKey, nil
}
