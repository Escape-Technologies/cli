package env

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// GetAPIURL returns the escape api url
// Default to https://public.escape.tech if not set
func GetAPIURL() (*url.URL, error) {
	rawURL := os.Getenv("ESCAPE_API_URL")
	if rawURL == "" {
		rawURL = "https://public.escape.tech"
	}
	rawURL = strings.TrimSuffix(rawURL, "/")
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ESCAPE_API_URL: %w", err)
	}
	return parsedURL, nil
}
