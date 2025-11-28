package private

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/Escape-Technologies/cli/pkg/log"
)

type bufConn struct {
	net.Conn
	r io.Reader
}

func (c *bufConn) Read(b []byte) (int, error) {
	n, err := c.r.Read(b)
	if err != nil {
		return n, fmt.Errorf("failed to read from buffer: %w", err)
	}
	return n, nil
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
		if err != nil && conn != nil {
			closeErr := conn.Close()
			if closeErr != nil {
				log.Debug("Failed to close connection: %s", closeErr)
			}
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
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to do connect handshake, status code: %s", resp.Status)
	}

	return &bufConn{Conn: conn, r: r}, nil
}
