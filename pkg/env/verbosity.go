package env

import (
	"os"
	"strconv"
)

// GetVerbosity returns the verbosity level
// Default to 0 if not set
func GetVerbosity() int {
	verbosity := os.Getenv("ESCAPE_VERBOSITY")
	if verbosity == "" {
		return 0
	}

	verbosityInt, err := strconv.Atoi(verbosity)
	if err != nil {
		return 0
	}
	return verbosityInt
}
