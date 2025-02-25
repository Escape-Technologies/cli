package private

import (
	"context"
	"net"
	"net/url"
)

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
