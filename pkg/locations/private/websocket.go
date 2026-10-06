package private

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/gorilla/websocket"
)

// errUpgradeRefused marks a WebSocket handshake answered without a 101: the distributor or a
// proxy on the path stripped or refused the Upgrade. Long-poll on the same URL can still work.
var errUpgradeRefused = errors.New("WebSocket upgrade refused")

const wsCloseWait = time.Second

// dialWebSocket opens the distributor's WebSocket on u and returns it as a byte stream.
func dialWebSocket(ctx context.Context, u *url.URL, dial dialFunc, tlsConfig *tls.Config) (net.Conn, error) {
	dialer := websocket.Dialer{
		NetDialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
			return dial(ctx, addr)
		},
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: sshDialTimeout,
	}

	wsURL := *u
	wsURL.Scheme = "wss"
	if u.Scheme == "http" {
		wsURL.Scheme = "ws"
	}

	ws, resp, err := dialer.DialContext(ctx, wsURL.String(), nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close() //nolint:errcheck
	}

	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("%w: %s", errUpgradeRefused, resp.Status)
		}

		if errors.Is(err, websocket.ErrBadHandshake) {
			return nil, fmt.Errorf("%w: %w", errUpgradeRefused, err)
		}

		return nil, fmt.Errorf("failed to open WebSocket: %w", err)
	}

	log.Debug("WebSocket opened to %s", u)

	return &wsConn{ws: ws, reader: nil, readMu: sync.Mutex{}, writeMu: sync.Mutex{}, closeOnce: sync.Once{}}, nil
}

// wsConn is a WebSocket as a byte stream: each Write is one binary message, and Read
// concatenates incoming binary messages, ignoring their boundaries. Read and Write are safe
// from different goroutines. Pings from the distributor are answered while Read runs.
type wsConn struct {
	ws     *websocket.Conn
	reader io.Reader

	readMu    sync.Mutex
	writeMu   sync.Mutex
	closeOnce sync.Once
}

func (c *wsConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	for {
		if c.reader != nil {
			n, err := c.reader.Read(p)
			if errors.Is(err, io.EOF) {
				c.reader = nil
				if n == 0 {
					continue
				}

				err = nil
			}

			return n, err //nolint:wrapcheck
		}

		msgType, reader, err := c.ws.NextReader()
		if err != nil {
			return 0, err //nolint:wrapcheck
		}

		if msgType != websocket.BinaryMessage {
			return 0, errors.New("unexpected non-binary WebSocket message")
		}

		c.reader = reader
	}
}

func (c *wsConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	err := c.ws.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err //nolint:wrapcheck
	}

	return len(p), nil
}

func (c *wsConn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		_ = c.ws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(wsCloseWait))
		err = c.ws.Close()
	})

	return err //nolint:wrapcheck
}

func (c *wsConn) LocalAddr() net.Addr  { return c.ws.LocalAddr() }
func (c *wsConn) RemoteAddr() net.Addr { return c.ws.RemoteAddr() }

func (c *wsConn) SetDeadline(t time.Time) error {
	err := c.ws.SetReadDeadline(t)
	if err != nil {
		return err //nolint:wrapcheck
	}

	return c.ws.SetWriteDeadline(t) //nolint:wrapcheck
}

func (c *wsConn) SetReadDeadline(t time.Time) error  { return c.ws.SetReadDeadline(t) }  //nolint:wrapcheck
func (c *wsConn) SetWriteDeadline(t time.Time) error { return c.ws.SetWriteDeadline(t) } //nolint:wrapcheck

var _ net.Conn = (*wsConn)(nil)
