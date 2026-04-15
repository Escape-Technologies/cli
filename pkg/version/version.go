// Package version contains the version information for the CLI
package version //nolint:revive

import (
	"context"
	"fmt"
	"strings"
)

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
	Version        string `json:"version"`
	Commit         string `json:"commit"`
	BuildDate      string `json:"buildDate"`
	InstallMethod  string `json:"installMethod,omitempty"`
	LatestVersion  string `json:"latestVersion,omitempty"`
	UpgradeCommand string `json:"upgradeCommand,omitempty"`
}

// GetVersion returns the version information
func GetVersion() V {
	return V{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

// GetDetailedVersion returns version metadata plus update information.
func GetDetailedVersion(ctx context.Context) V {
	v := GetVersion()
	installInfo := GetInstallInfo()
	updateInfo := CheckForUpdate(ctx)

	v.InstallMethod = string(installInfo.Method)
	v.LatestVersion = updateInfo.Latest
	if updateInfo.Available {
		v.UpgradeCommand = UpgradeCommand(installInfo.Method, updateInfo.Latest)
	}

	return v
}

// DisplayVersion returns a human-readable version string.
func (v V) DisplayVersion() string {
	if strings.TrimSpace(v.Version) == "" {
		return "unknown"
	}

	if v.Version == "local" {
		return "dev"
	}

	return "v" + normalizeVersion(v.Version)
}

func updateStatus(v V) string {
	switch {
	case v.Version == "local":
		return "Development build"
	case strings.TrimSpace(v.LatestVersion) == "":
		return "Unable to check"
	}

	switch compareVersions(v.Version, v.LatestVersion) {
	case -1:
		return fmt.Sprintf("Update available (%s)", "v"+normalizeVersion(v.LatestVersion))
	case 1:
		return "Ahead of latest release"
	default:
		return "Up to date"
	}
}

// String returns the version information as a string
func (v V) String() string {
	installInfo := GetInstallInfo()
	lines := []string{
		"Escape CLI",
		"  Version:    " + v.DisplayVersion(),
		"  Commit:     " + v.Commit,
		"  Build date: " + v.BuildDate,
		"  Install:    " + installInfo.DisplayName(),
	}

	lines = append(lines, "  Status:     "+updateStatus(v))

	if strings.TrimSpace(v.UpgradeCommand) != "" {
		lines = append(lines, "  Upgrade:    "+v.UpgradeCommand)
	}

	return strings.Join(lines, "\n")
}

// LogString returns the version metadata in logfmt-friendly form.
func (v V) LogString() string {
	return fmt.Sprintf("version=%s commit=%s buildDate=%s", normalizeVersion(v.Version), v.Commit, v.BuildDate)
}

// UserAgent returns the user agent string
func (v V) UserAgent() string {
	return "escape-cli/" + v.Version
}
