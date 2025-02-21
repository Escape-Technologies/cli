package privatelocation

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strconv"

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
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Error("Failed to accept connection: %v", err)
				continue
			}
			go handleSocks5Connection(conn)
		}
	}
}
//this shit works, TODO: make this readable
func handleSocks5Connection(conn net.Conn) {
	defer conn.Close()

	// Read SOCKS5 version and number of authentication methods
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Error("Failed to read SOCKS5 header: %v", err)
		return
	}

	if buf[0] != 5 { // SOCKS5 version
		log.Error("Unsupported SOCKS version: %d", buf[0])
		return
	}

	// Read authentication methods
	methods := make([]byte, buf[1])
	if _, err := io.ReadFull(conn, methods); err != nil {
		log.Error("Failed to read auth methods: %v", err)
		return
	}

	// Respond with no authentication required
	if _, err := conn.Write([]byte{5, 0}); err != nil {
		log.Error("Failed to write auth response: %v", err)
		return
	}

	// Read connection request
	buf = make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Error("Failed to read connection request: %v", err)
		return
	}

	if buf[0] != 5 || buf[1] != 1 { // Only support CONNECT command
		log.Error("Unsupported SOCKS5 command: %d", buf[1])
		return
	}

	// Read address type
	var host string
	switch buf[3] {
	case 1: // IPv4
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			log.Error("Failed to read IPv4 address: %v", err)
			return
		}
		host = net.IPv4(addr[0], addr[1], addr[2], addr[3]).String()
	case 3: // Domain name
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			log.Error("Failed to read domain length: %v", err)
			return
		}
		domainBuf := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, domainBuf); err != nil {
			log.Error("Failed to read domain: %v", err)
			return
		}
		host = string(domainBuf)
	case 4: // IPv6
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			log.Error("Failed to read IPv6 address: %v", err)
			return
		}
		host = net.IP(addr).String()
	default:
		log.Error("Unsupported address type: %d", buf[3])
		return
	}

	// Read port
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		log.Error("Failed to read port: %v", err)
		return
	}
	port := int(portBuf[0])<<8 | int(portBuf[1])

	// Connect to target
	target, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		log.Error("Failed to connect to target: %v", err)
		conn.Write([]byte{5, 1, 0, 1, 0, 0, 0, 0, 0, 0}) // Connection refused
		return
	}
	defer target.Close()

	// Send success response
	localAddr := target.LocalAddr().(*net.TCPAddr)
	response := make([]byte, 10)
	response[0] = 5    // SOCKS5
	response[1] = 0    // Success
	response[2] = 0    // Reserved
	response[3] = 1    // IPv4
	copy(response[4:8], localAddr.IP.To4())
	response[8] = byte(localAddr.Port >> 8)
	response[9] = byte(localAddr.Port)
	if _, err := conn.Write(response); err != nil {
		log.Error("Failed to write success response: %v", err)
		return
	}

	// Start proxying data
	go func() {
		io.Copy(target, conn)
	}()
	io.Copy(conn, target)
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
