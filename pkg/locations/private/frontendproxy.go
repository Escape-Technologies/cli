package private

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/net/proxy"
)

func proxyDialer(proxyURL *url.URL) func(context.Context, string) (net.Conn, error) {
	scheme := strings.ToLower(proxyURL.Scheme)

	// Handle SOCKS5 proxies
	if scheme == "socks5" || scheme == "socks5h" {
		proxyDialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return func(_ context.Context, _ string) (net.Conn, error) {
				return nil, fmt.Errorf("failed to create SOCKS5 proxy dialer: %w", err)
			}
		}

		return func(ctx context.Context, addr string) (net.Conn, error) {
			// Create a channel to handle context cancellation
			type result struct {
				conn net.Conn
				err  error
			}
			resultChan := make(chan result, 1)

			go func() {
				conn, err := proxyDialer.Dial("tcp", addr)
				resultChan <- result{conn: conn, err: err}
			}()

			select {
			case <-ctx.Done():
				// If context is cancelled, try to receive the connection and close it
				// to avoid leaking resources
				select {
				case res := <-resultChan:
					if res.conn != nil {
						res.conn.Close() //nolint:errcheck
					}
				default:
				}

				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			case res := <-resultChan:
				if res.err != nil {
					return nil, fmt.Errorf("failed to dial through SOCKS5 proxy: %w", res.err)
				}

				return res.conn, nil
			}
		}
	}

	// Handle HTTP/HTTPS proxies (default behavior)
	return func(ctx context.Context, addr string) (net.Conn, error) {
		proxyAddr := proxyURL.Host

		conn, err := netDialerWithTCPKeepalive().DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to dial proxy: %w", err)
		}

		return doHTTPConnectHandshake(ctx, conn, addr, *proxyURL)
	}
}

// dialFunc opens a TCP stream to addr, through the frontend proxy when one is set.
type dialFunc func(ctx context.Context, addr string) (net.Conn, error)

func tcpDialer(frontendProxyURL *url.URL) dialFunc {
	if frontendProxyURL == nil {
		return func(ctx context.Context, addr string) (net.Conn, error) {
			conn, err := netDialerWithTCPKeepalive().DialContext(ctx, "tcp", addr)
			if err != nil {
				return conn, fmt.Errorf("failed to dial target: %w", err)
			}

			return conn, nil
		}
	}

	return proxyDialer(frontendProxyURL)
}

// getConn opens the outer socket of the tunnel. Whatever the transport, the returned
// net.Conn is a plain byte stream that carries the SSH session.
//
// The HTTPS transports reach the frontend proxy with a CONNECT (or SOCKS5) to host:443, then
// do TLS and the WebSocket Upgrade or long-poll themselves: the proxy never has to speak WebSocket.
func getConn(ctx context.Context, ep endpoint, frontendProxyURL *url.URL) (net.Conn, error) {
	dial := tcpDialer(frontendProxyURL)
	switch ep.transport {
	case transportSSH:
		return dial(ctx, ep.addr)
	case transportWSS, transportHTTPS, transportAuto:
		return getHTTPSConn(ctx, ep, dial)
	}

	return nil, fmt.Errorf("unknown transport %q", ep.transport)
}

func getHTTPSConn(ctx context.Context, ep endpoint, dial dialFunc) (net.Conn, error) {
	tlsConfig, err := env.GetCertificates()
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS configuration: %w", err)
	}

	if ep.transport == transportHTTPS {
		return dialPoll(ctx, ep.connectURL, dial, tlsConfig)
	}

	conn, err := dialWebSocket(ctx, ep.connectURL, dial, tlsConfig)
	if err == nil || ep.transport == transportWSS || ctx.Err() != nil {
		return conn, err
	}

	log.Warn("WebSocket to %s failed, falling back to HTTP long-poll on the same URL: %s", ep.connectURL, err)

	return dialPoll(ctx, ep.connectURL, dial, tlsConfig)
}
