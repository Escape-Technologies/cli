package private

import (
	"context"
	"net"
	"net/url"
)

func proxyDialer(proxyURL *url.URL) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, addr string) (net.Conn, error) {
		proxyAddr := proxyURL.Host

		conn, err := netDialerWithTCPKeepalive().DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, err
		}
		return doHTTPConnectHandshake(ctx, conn, addr, *proxyURL)
	}
}

func getConn(ctx context.Context, target string, frontendProxyURL *url.URL) (net.Conn, error) {
	if frontendProxyURL == nil {
		return netDialerWithTCPKeepalive().DialContext(ctx, "tcp", target)
	}

	dialer := proxyDialer(frontendProxyURL)
	return dialer(ctx, target)
}
