package env

import (
	"os"
)

// GetColorPreference returns the color preference
// Default to true if not set
func GetColorPreference() (bool) {
	isColorDisabled := os.Getenv("ESCAPE_COLOR_DISABLED")
	return isColorDisabled == "true"
}
