// Package escape provides the API client for the Escape Platform
package escape

import (
	"fmt"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/version"
)

// Debug is a flag to enable debug mode for the API client
var Debug = false

func newAPIV2Client() (*v2.APIClient, error) {
	log.Debug("Initializing v2 client")
	url, err := env.GetAPIURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get API URL: %w", err)
	}
	key, err := env.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	cfg := v2.Configuration{
		Host:   url.Host,
		Scheme: url.Scheme,
		DefaultHeader: map[string]string{
			"Authorization": "Key " + key,
		},
		UserAgent:  version.GetVersion().UserAgent(),
		Debug:      Debug,
		HTTPClient: env.GetHTTPClient(),
		Servers: []v2.ServerConfiguration{
			{
				URL: url.String() + "/v2",
			},
		},
	}

	return v2.NewAPIClient(&cfg), nil
}
