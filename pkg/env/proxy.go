package env

import (
	"context"
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


func BuildProxyDialer(ctx context.Context, proxyURL *url.URL) proxy.Dialer {
	defaultDialer := proxy.Direct

	if proxyURL.Scheme == "socks5" || proxyURL.Scheme == "socks5h" {
		auth := &proxy.Auth{}
		if proxyURL.User != nil {
			auth.User = proxyURL.User.Username()
			if password, ok := proxyURL.User.Password(); ok {
				auth.Password = password
			}
		}

		socksDialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			log.Error("Failed to create SOCKS5 dialer: %w", err)
			return defaultDialer
		}
		return socksDialer
	}

	proxyDialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		log.Error("Failed to create HTTP proxy dialer: %w", err)
		return defaultDialer
	}

	return proxyDialer
}
