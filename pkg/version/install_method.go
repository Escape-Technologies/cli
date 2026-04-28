package version //nolint:revive

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// InstallMethod describes how the CLI binary was installed.
type InstallMethod string

const (
	// InstallMethodCurl uses the Unix install script.
	InstallMethodCurl InstallMethod = "curl"
	// InstallMethodDocker uses the published Docker image.
	InstallMethodDocker InstallMethod = "docker"
	// InstallMethodGitHubActions runs through the GitHub Action.
	InstallMethodGitHubActions InstallMethod = "github-actions"
	// InstallMethodHomebrew uses the published Homebrew cask.
	InstallMethodHomebrew InstallMethod = "homebrew"
	// InstallMethodWindowsScript uses the PowerShell install script.
	InstallMethodWindowsScript InstallMethod = "windows-script"
	// InstallMethodManual covers binaries installed without a known wrapper.
	InstallMethodManual InstallMethod = "manual"
)

// InstallInfo stores the detected installation method and binary path.
type InstallInfo struct {
	Method InstallMethod
	Path   string
}

var (
	installInfoOnce sync.Once
	installInfo     InstallInfo
	// InstallMethodOverride is injected at build time for wrapped distributions
	// like the published Docker image.
	InstallMethodOverride = ""
)

// GetInstallInfo returns the cached install details for the current process.
func GetInstallInfo() InstallInfo {
	installInfoOnce.Do(func() {
		path := getExecutablePath()
		installInfo = InstallInfo{
			Method: detectInstallMethod(path),
			Path:   path,
		}
	})

	return installInfo
}

func getExecutablePath() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}

	resolvedPath, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolvedPath
	}

	return path
}

func detectInstallMethod(path string) InstallMethod {
	return detectInstallMethodForGOOS(path, runtime.GOOS)
}

func detectInstallMethodForGOOS(path string, goos string) InstallMethod {
	if method, ok := parseInstallMethod(InstallMethodOverride); ok {
		return method
	}

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return InstallMethodGitHubActions
	}

	if goos == "windows" {
		return InstallMethodWindowsScript
	}

	if isHomebrewCaskPath(path, goos) {
		return InstallMethodHomebrew
	}

	if path == "/usr/local/bin/escape-cli" {
		return InstallMethodCurl
	}

	return InstallMethodManual
}

func isHomebrewCaskPath(path string, goos string) bool {
	if goos != "darwin" {
		return false
	}

	normalized := filepath.ToSlash(filepath.Clean(path))
	for _, prefix := range []string{
		"/opt/homebrew/Caskroom/escape-cli/",
		"/usr/local/Caskroom/escape-cli/",
	} {
		rest, ok := strings.CutPrefix(normalized, prefix)
		if !ok {
			continue
		}

		version, binary, ok := strings.Cut(rest, "/")
		return ok && version != "" && binary == "escape-cli"
	}

	return false
}

func parseInstallMethod(value string) (InstallMethod, bool) {
	switch InstallMethod(strings.ToLower(strings.TrimSpace(value))) {
	case InstallMethodCurl:
		return InstallMethodCurl, true
	case InstallMethodDocker:
		return InstallMethodDocker, true
	case InstallMethodGitHubActions:
		return InstallMethodGitHubActions, true
	case InstallMethodHomebrew:
		return InstallMethodHomebrew, true
	case InstallMethodWindowsScript:
		return InstallMethodWindowsScript, true
	case InstallMethodManual:
		return InstallMethodManual, true
	default:
		return "", false
	}
}

// DisplayName returns a user-facing description of the installation method.
func (i InstallInfo) DisplayName() string {
	switch i.Method {
	case InstallMethodCurl:
		if i.Path == "" {
			return "curl script"
		}
		return fmt.Sprintf("curl script (%s)", i.Path)
	case InstallMethodDocker:
		return "docker image"
	case InstallMethodGitHubActions:
		return "GitHub Actions"
	case InstallMethodHomebrew:
		if i.Path == "" {
			return "Homebrew cask"
		}
		return fmt.Sprintf("Homebrew cask (%s)", i.Path)
	case InstallMethodWindowsScript:
		if i.Path == "" {
			return "PowerShell script"
		}
		return fmt.Sprintf("PowerShell script (%s)", i.Path)
	case InstallMethodManual:
		if i.Path == "" {
			return "manual install"
		}
		return fmt.Sprintf("manual install (%s)", i.Path)
	default:
		return "unknown install"
	}
}

// UpgradeCommand returns an executable upgrade instruction for the install method.
func UpgradeCommand(method InstallMethod, latestVersion string) string {
	switch method {
	case InstallMethodCurl:
		return "curl -sf https://raw.githubusercontent.com/Escape-Technologies/cli/refs/heads/main/scripts/install.sh | sudo bash"
	case InstallMethodDocker:
		return "docker pull escapetech/cli:latest"
	case InstallMethodGitHubActions:
		if latestVersion == "" {
			return "uses: Escape-Technologies/cli@latest"
		}
		return "uses: Escape-Technologies/cli@v" + normalizeVersion(latestVersion)
	case InstallMethodHomebrew:
		return "brew upgrade --cask escape-technologies/tap/escape-cli"
	case InstallMethodWindowsScript:
		return `powershell -c "irm https://raw.githubusercontent.com/Escape-Technologies/cli/refs/heads/main/scripts/install.ps1 | iex"`
	case InstallMethodManual:
		return "curl -sf https://raw.githubusercontent.com/Escape-Technologies/cli/refs/heads/main/scripts/install.sh | sudo bash"
	default:
		return ""
	}
}
