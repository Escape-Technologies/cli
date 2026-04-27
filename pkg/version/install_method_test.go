package version

import (
	"runtime"
	"testing"
)

func TestDetectInstallMethodHonorsBuildOverride(t *testing.T) {
	previous := InstallMethodOverride
	InstallMethodOverride = "docker"
	t.Cleanup(func() {
		InstallMethodOverride = previous
	})

	t.Setenv("GITHUB_ACTIONS", "")

	if got := detectInstallMethod("/tmp/escape-cli"); got != InstallMethodDocker {
		t.Fatalf("detectInstallMethod() = %q, want %q", got, InstallMethodDocker)
	}
}

func TestDetectInstallMethodFallsBackToManual(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defaults to the PowerShell installer")
	}

	previous := InstallMethodOverride
	InstallMethodOverride = ""
	t.Cleanup(func() {
		InstallMethodOverride = previous
	})

	t.Setenv("GITHUB_ACTIONS", "")

	if got := detectInstallMethod("/tmp/escape-cli"); got != InstallMethodManual {
		t.Fatalf("detectInstallMethod() = %q, want %q", got, InstallMethodManual)
	}
}

func TestDetectInstallMethodKeepsCurlPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows defaults to the PowerShell installer")
	}

	previous := InstallMethodOverride
	InstallMethodOverride = ""
	t.Cleanup(func() {
		InstallMethodOverride = previous
	})

	t.Setenv("GITHUB_ACTIONS", "")

	if got := detectInstallMethod("/usr/local/bin/escape-cli"); got != InstallMethodCurl {
		t.Fatalf("detectInstallMethod() = %q, want %q", got, InstallMethodCurl)
	}
}

func TestDetectInstallMethodDetectsHomebrewCaskOnDarwin(t *testing.T) {
	previous := InstallMethodOverride
	InstallMethodOverride = ""
	t.Cleanup(func() {
		InstallMethodOverride = previous
	})

	t.Setenv("GITHUB_ACTIONS", "")

	tests := []string{
		"/opt/homebrew/Caskroom/escape-cli/1.2.3/escape-cli",
		"/usr/local/Caskroom/escape-cli/1.2.3/escape-cli",
	}

	for _, path := range tests {
		path := path
		t.Run(path, func(t *testing.T) {
			if got := detectInstallMethodForGOOS(path, "darwin"); got != InstallMethodHomebrew {
				t.Fatalf("detectInstallMethodForGOOS(%q, darwin) = %q, want %q", path, got, InstallMethodHomebrew)
			}
		})
	}
}

func TestDetectInstallMethodDoesNotDetectHomebrewCaskOnNonDarwin(t *testing.T) {
	previous := InstallMethodOverride
	InstallMethodOverride = ""
	t.Cleanup(func() {
		InstallMethodOverride = previous
	})

	t.Setenv("GITHUB_ACTIONS", "")

	if got := detectInstallMethodForGOOS("/opt/homebrew/Caskroom/escape-cli/1.2.3/escape-cli", "linux"); got != InstallMethodManual {
		t.Fatalf("detectInstallMethodForGOOS() = %q, want %q", got, InstallMethodManual)
	}
}

func TestParseInstallMethodAcceptsHomebrewOverride(t *testing.T) {
	if got, ok := parseInstallMethod("homebrew"); !ok || got != InstallMethodHomebrew {
		t.Fatalf("parseInstallMethod() = %q, %t, want %q, true", got, ok, InstallMethodHomebrew)
	}
}

func TestParseInstallMethodRejectsUnknownOverride(t *testing.T) {
	if got, ok := parseInstallMethod("containerd"); ok {
		t.Fatalf("parseInstallMethod() = %q, want unknown value to be rejected", got)
	}
}
