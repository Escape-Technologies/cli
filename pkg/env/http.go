// Package env provides environment variables parsing and utils builders
package env

import (
	"net/http"
)

// GetHTTPClient returns a new http client with the proxy set if needed
// This client should be used for calls to the Escape Platform
func GetHTTPClient() *http.Client {
	transport := &http.Transport{}
	proxyURL := GetFrontendProxyURL()
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Transport: transport}
}
