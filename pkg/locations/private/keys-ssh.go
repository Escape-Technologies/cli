package private

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/ssh"
)

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
