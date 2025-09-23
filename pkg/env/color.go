package env

import (
	"os"
)

// GetColorPreference returns the color preference
// Default to true if not set
func GetColorPreference() (bool, error) {
	noColor := os.Getenv("ESCAPE_NO_COLOR")
	if noColor == "true" {
		return true, nil
	}
	return false, nil
}
