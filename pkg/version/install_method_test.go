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

func TestParseInstallMethodRejectsUnknownOverride(t *testing.T) {
	if got, ok := parseInstallMethod("containerd"); ok {
		t.Fatalf("parseInstallMethod() = %q, want unknown value to be rejected", got)
	}
}
