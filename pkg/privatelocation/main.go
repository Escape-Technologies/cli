package privatelocation

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/socks5"
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

	client, err := ssh.Dial("tcp", "a814bdc744e1147dd86d66114ed8edcc-2eb18fcf1bd8afa3.elb.eu-west-3.amazonaws.com:2222", config)
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
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			log.Info("Waiting for connection")
			conn, err := listener.Accept()
			if err != nil {
				log.Error("Failed to accept connection: %v", err)
				continue
			}
			log.Info("Accepted connection, creating socks5 server")
			go handleSocks5Connection(conn)
		}
	}
}
func handleSocks5Connection(conn net.Conn) {
	socks5Server, err := socks5.New(&socks5.Config{})

	if err != nil {
		log.Error("Failed to create socks5 server: %v", err)
		return
	}
	log.Info("Socks5 server created")

	socks5Server.ServeConn(conn)
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
