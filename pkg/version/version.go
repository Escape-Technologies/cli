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

func (v V) DisplayVersion() string {
	if strings.TrimSpace(v.Version) == "" {
		return "unknown"
	}

	if v.Version == "local" {
		return "dev"
	}

	return "v" + normalizeVersion(v.Version)
}

// String returns the version information as a string
func (v V) String() string {
	installInfo := GetInstallInfo()
	lines := []string{
		"Escape CLI",
		fmt.Sprintf("  Version:    %s", v.DisplayVersion()),
		fmt.Sprintf("  Commit:     %s", v.Commit),
		fmt.Sprintf("  Build date: %s", v.BuildDate),
		fmt.Sprintf("  Install:    %s", installInfo.DisplayName()),
	}

	status := "Unknown"
	switch {
	case v.Version == "local":
		status = "Development build"
	case v.LatestVersion == "":
		status = "Unable to check"
	case normalizeVersion(v.LatestVersion) == normalizeVersion(v.Version):
		status = "Up to date"
	default:
		status = fmt.Sprintf("Update available (%s)", "v"+normalizeVersion(v.LatestVersion))
	}
	lines = append(lines, fmt.Sprintf("  Status:     %s", status))

	if strings.TrimSpace(v.UpgradeCommand) != "" {
		lines = append(lines, fmt.Sprintf("  Upgrade:    %s", v.UpgradeCommand))
	}

	return strings.Join(lines, "\n")
}

func (v V) LogString() string {
	return fmt.Sprintf("version=%s commit=%s buildDate=%s", normalizeVersion(v.Version), v.Commit, v.BuildDate)
}

// UserAgent returns the user agent string
func (v V) UserAgent() string {
	return "escape-cli/" + v.Version
}
