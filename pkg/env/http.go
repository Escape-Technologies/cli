package env

import (
	"net/http"
)

func GetHTTPClient() *http.Client {
	transport := &http.Transport{}
	proxyURL := GetFrontendProxyURL()
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Transport: transport}
}
