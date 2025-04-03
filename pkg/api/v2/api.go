package v2

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Escape-Technologies/cli/pkg/env"
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

	proxyURL := env.GetFrontendProxyURL()

	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{Transport: transport}
	opts = append(opts, WithHTTPClient(client))
	opts = append(opts, WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		log.Trace("Sending request %s %s", req.Method, req.URL)
		req.Header.Set("Authorization", fmt.Sprintf("Key %s", key))
		return nil
	}))
	return NewClientWithResponses(server, opts...)
}
