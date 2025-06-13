// Package env provides environment variables parsing and utils builders
package env

import (
	"net/http"

	"github.com/Escape-Technologies/cli/pkg/log"
)

// GetHTTPClient returns a new http client with the proxy set if needed
// This client should be used for calls to the Escape Platform
func GetHTTPClient() *http.Client {
	transport := &http.Transport{}
	certificates, err := GetCertificates()
	if err != nil {
		log.Warn("Failed to get certificates: %s", err)
	}
	if certificates != nil {
		transport.TLSClientConfig = certificates
	}
	proxyURL := GetFrontendProxyURL()
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Transport: transport}
}
