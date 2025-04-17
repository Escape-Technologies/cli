// Package version contains the version information for the CLI
package version

import "fmt"

var (
	// Version injected at build time
	Version = "local"
	// Commit injected at build time
	Commit = "unknown"
	// BuildDate injected at build time
	BuildDate = "unknown"
)

// V Version JSON struct
type V struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
}

// GetVersion returns the version information
func GetVersion() V {
	return V{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

// String returns the version information as a string
func (v V) String() string {
	return fmt.Sprintf("Version: %s, Commit: %s, BuildDate: %s", v.Version, v.Commit, v.BuildDate)
}

// UserAgent returns the user agent string
func (v V) UserAgent() string {
	return "escape-cli/" + v.Version
}
