package env

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	uuidRegex = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	// APIKeyHint tells users where to retrieve a valid API key.
	APIKeyHint = "Get your key here: https://app.escape.tech/user/profile/"
)

var uuidRegexp = regexp.MustCompile(uuidRegex)

func validateUUIDKey(key string) error {
	if !uuidRegexp.MatchString(key) {
		return fmt.Errorf("%s does not match regex %s", key, uuidRegex)
	}
	return nil
}

// GetAPIKey returns the escape api key or an error if it is not set
func GetAPIKey() (string, error) {
	key := os.Getenv("ESCAPE_API_KEY")
	if key == "" {
		return "", errors.New("ESCAPE_API_KEY environment variable is not set. " + APIKeyHint)
	}
	err := validateUUIDKey(key)
	if err != nil {
		return "", fmt.Errorf("ESCAPE_API_KEY invalid UUID format: %w", err)
	}
	return key, nil
}

// GetAuthorizationHeader returns the raw Authorization header to use for API calls.
// ESCAPE_AUTHORIZATION takes precedence to preserve non-API-key auth flows.
func GetAuthorizationHeader() (string, error) {
	authorization := strings.TrimSpace(os.Getenv("ESCAPE_AUTHORIZATION"))
	if authorization != "" {
		return authorization, nil
	}

	key, err := GetAPIKey()
	if err != nil {
		return "", err
	}

	return "Key " + key, nil
}
