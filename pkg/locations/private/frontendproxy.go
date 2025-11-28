package private

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

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

func getConn(ctx context.Context, target string, frontendProxyURL *url.URL) (net.Conn, error) {
	if frontendProxyURL == nil {
		conn, err := netDialerWithTCPKeepalive().DialContext(ctx, "tcp", target)
		if err != nil {
			return conn, fmt.Errorf("failed to dial target: %w", err)
		}
		return conn, nil
	}

	dialer := proxyDialer(frontendProxyURL)
	return dialer(ctx, target)
}
