package escape

import (
	"fmt"
	"net/http"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/version"
)

func newAPIV2Client() (*v2.APIClient, error) {
	log.Trace("Initializing v2 client")
	url, err := env.GetAPIURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get API URL: %w", err)
	}
	key, err := env.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	transport := &http.Transport{}
	proxyURL := env.GetFrontendProxyURL()
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	cfg := v2.Configuration{
		Host:   url.Host,
		Scheme: url.Scheme,
		DefaultHeader: map[string]string{
			"Authorization": fmt.Sprintf("Key %s", key),
		},
		UserAgent:  version.GetVersion().UserAgent(),
		Debug:      false,
		HTTPClient: &http.Client{Transport: transport},
	}

	return v2.NewAPIClient(&cfg), nil
}
