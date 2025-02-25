package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func NewAPIClient(opts ...ClientOption) (*ClientWithResponses, error) {
	log.Trace("Initializing client")
	server := os.Getenv("ESCAPE_API_URL")
	if server == "" {
		server = "https://public.escape.tech/"
	}
	key := os.Getenv("ESCAPE_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ESCAPE_API_KEY is not set")
	}

	transport := &http.Transport{}

	proxyURL := os.Getenv("ESCAPE_REPEATER_PROXY_URL")
	if proxyURL != "" {
		log.Info("Using custom proxy url: %s", proxyURL)
		url, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(url)
	}

	client := &http.Client{Transport: transport}
	opts = append(opts, WithHTTPClient(client))
	opts = append(opts, WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		log.Debug("Sending request %s %s", req.Method, req.URL)
		req.Header.Set("Authorization", fmt.Sprintf("Key %s", key))
		return nil
	}))
	return NewClientWithResponses(server, opts...)
}
