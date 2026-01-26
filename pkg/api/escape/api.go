// Package escape provides the API client for the Escape Platform
package escape

import (
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/version"
)

// Debug is a flag to enable debug mode for the API client
var Debug = false

// newAPIV3Client creates a new API v3 client
func newAPIV3Client() (*v3.APIClient, error) {
	log.Trace("Initializing v3 client")
	url, err := env.GetAPIURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get API URL: %w", err)
	}
	key, err := env.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	cfg := v3.Configuration{
		Host:   url.Host,
		Scheme: url.Scheme,
		DefaultHeader: map[string]string{
			"Authorization": "Key " + key,
		},
		UserAgent:  version.GetVersion().UserAgent(),
		Debug:      Debug,
		HTTPClient: env.GetHTTPClient(),
		Servers: []v3.ServerConfiguration{
			{
				URL: url.String() + "/v3",
			},
		},
	}

	return v3.NewAPIClient(&cfg), nil
}
