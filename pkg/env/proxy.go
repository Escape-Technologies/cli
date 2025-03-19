package env

import (
	"context"
	"net"
	"net/url"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/net/proxy"
)

func GetFrontendProxyURL() *url.URL {
	proxyURL := os.Getenv("ESCAPE_REPEATER_PROXY_URL")
	if proxyURL == "" {
		proxyURL = os.Getenv("ESCAPE_FRONTEND_PROXY_URL")
		if proxyURL == "" {
			return nil
		}
	}
	url, err := url.Parse(proxyURL)
	if err != nil {
		log.Warn("Failed to parse proxy url: %s", err)
		return nil
	}
	log.Debug("Using custom proxy url: %s", url.Host)
	return url
}


func GetBackendProxyURL() *url.URL {
	proxyURL := os.Getenv("ESCAPE_BACKEND_PROXY_URL")
	if proxyURL == "" {
		return nil
	}
	url, err := url.Parse(proxyURL)
	if err != nil {
		log.Warn("Failed to parse proxy url: %s", err)
		return nil
	}
	log.Debug("Using custom backend proxy url: %s", url.Host)
	return url
}

func BuildProxyDialer(ctx context.Context, proxyURL *url.URL) func(ctx context.Context, network string, addr string) (net.Conn, error) {
	defaultDialer := proxy.Direct

	if proxyURL == nil {
		return defaultDialer.DialContext
	}

	log.Debug("Building proxy dialer for %s", proxyURL.Host)
	proxyDialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		log.Error("Failed to create proxy dialer: %s", err)
		return defaultDialer.DialContext
	}

	log.Debug("Testing proxy connection through %s", proxyURL.String())		
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == "127.0.0.1:8001" {
			log.Debug("SOCKS5 dialing: network=%s addr=%s without proxy", network, addr)
			return defaultDialer.DialContext(ctx, network, addr)
		}
		log.Debug("SOCKS5 dialing: network=%s addr=%s through proxy=%s", network, addr, proxyURL.String())
		return proxyDialer.Dial(network, addr)
	}
}