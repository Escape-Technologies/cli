package private

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
)

type bufConn struct {
	net.Conn
	r io.Reader
}

func (c *bufConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func netDialerWithTCPKeepalive() *net.Dialer {
	return &net.Dialer{
		KeepAlive: time.Duration(-1),
		Control: func(_, _ string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1)
				if err != nil {
					log.Error("failed to set SO_KEEPALIVE: %v", err)
				}
			})
		},
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func sendHTTPRequest(ctx context.Context, req *http.Request, conn net.Conn) error {
	req = req.WithContext(ctx)
	if err := req.Write(conn); err != nil {
		return fmt.Errorf("failed to write the HTTP request: %v", err)
	}
	return nil
}

func doHTTPConnectHandshake(ctx context.Context, conn net.Conn, backendAddr string, proxyURL url.URL) (_ net.Conn, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: backendAddr},
		Header: make(http.Header),
	}
	if t := proxyURL.User; t != nil {
		u := t.Username()
		p, _ := t.Password()
		req.Header.Add("Proxy-Authorization", "Basic "+basicAuth(u, p))
	}

	if err := sendHTTPRequest(ctx, req, conn); err != nil {
		return nil, fmt.Errorf("failed to write the HTTP request: %v", err)
	}

	r := bufio.NewReader(conn)
	resp, err := http.ReadResponse(r, req)
	if err != nil {
		return nil, fmt.Errorf("reading server HTTP response: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to do connect handshake, status code: %s", resp.Status)
	}

	return &bufConn{Conn: conn, r: r}, nil
}

func proxyDialer(proxy string) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		proxyAddr := proxyURL.Host

		conn, err := netDialerWithTCPKeepalive().DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, err
		}
		return doHTTPConnectHandshake(ctx, conn, addr, *proxyURL)
	}
}

func getConn(ctx context.Context, target, proxyURL string) (net.Conn, error) {
	if proxyURL == "" {
		return netDialerWithTCPKeepalive().DialContext(ctx, "tcp", target)
	}

	dialer := proxyDialer(proxyURL)
	return dialer(ctx, target)
}

func getClient(ctx context.Context, target string, conn net.Conn, config *ssh.ClientConfig) (*ssh.Client, error) {
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

	target := "private-location.escape.tech:2222"
	proxyURL := os.Getenv("ESCAPE_REPEATER_PROXY_URL")
	
	conn, err := getConn(ctx, target, proxyURL)
	if err != nil {
		return fmt.Errorf("failed to get conn: %w", err)
	}
	
	client, err := getClient(ctx, target, conn, config)
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
