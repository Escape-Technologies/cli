package env

import (
	"fmt"
	"net/url"
	"os"
)

func GetAPIURL() (*url.URL, error) {
	rawURL := os.Getenv("ESCAPE_API_URL")
	if rawURL == "" {
		return nil, fmt.Errorf("ESCAPE_API_URL environment variable is not set")
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ESCAPE_API_URL: %w", err)
	}
	return parsedURL, nil
}
