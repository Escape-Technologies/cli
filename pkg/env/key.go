package env

import (
	"fmt"
	"os"
)

func GetAPIKey() (string, error) {
	key := os.Getenv("ESCAPE_API_KEY")
	if key == "" {
		return "", fmt.Errorf("ESCAPE_API_KEY environment variable is not set")
	}
	return key, nil
}
