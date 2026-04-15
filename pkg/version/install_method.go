package version //nolint:revive

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type InstallMethod string

const (
	InstallMethodCurl          InstallMethod = "curl"
	InstallMethodDocker        InstallMethod = "docker"
	InstallMethodGitHubActions InstallMethod = "github-actions"
	InstallMethodWindowsScript InstallMethod = "windows-script"
	InstallMethodManual        InstallMethod = "manual"
)

type InstallInfo struct {
	Method InstallMethod
	Path   string
}

var (
	installInfoOnce sync.Once
	installInfo     InstallInfo
)

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
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return InstallMethodGitHubActions
	}

	if isRunningInDocker() {
		return InstallMethodDocker
	}

	if runtime.GOOS == "windows" {
		return InstallMethodWindowsScript
	}

	if path == "/usr/local/bin/escape-cli" {
		return InstallMethodCurl
	}

	return InstallMethodManual
}

func isRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	cgroup, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}

	content := strings.ToLower(string(cgroup))
	return strings.Contains(content, "docker") ||
		strings.Contains(content, "containerd") ||
		strings.Contains(content, "kubepods")
}

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
	case InstallMethodWindowsScript:
		if i.Path == "" {
			return "PowerShell script"
		}
		return fmt.Sprintf("PowerShell script (%s)", i.Path)
	default:
		if i.Path == "" {
			return "manual install"
		}
		return fmt.Sprintf("manual install (%s)", i.Path)
	}
}

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
		return fmt.Sprintf("uses: Escape-Technologies/cli@v%s", normalizeVersion(latestVersion))
	case InstallMethodWindowsScript:
		return `powershell -c "irm https://raw.githubusercontent.com/Escape-Technologies/cli/refs/heads/main/scripts/install.ps1 | iex"`
	default:
		return "https://github.com/Escape-Technologies/cli/releases/latest"
	}
}
