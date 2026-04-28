package version

import "testing"

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		current  string
		latest   string
		expected int
	}{
		{
			name:     "same version",
			current:  "1.2.3",
			latest:   "v1.2.3",
			expected: 0,
		},
		{
			name:     "latest is newer",
			current:  "1.2.3",
			latest:   "1.2.4",
			expected: -1,
		},
		{
			name:     "current is newer",
			current:  "1.2.4",
			latest:   "1.2.3",
			expected: 1,
		},
		{
			name:     "build metadata does not change precedence",
			current:  "1.2.3+build.1",
			latest:   "1.2.3+build.2",
			expected: 0,
		},
		{
			name:     "invalid semver is treated as unknown ordering",
			current:  "release-candidate",
			latest:   "1.2.3",
			expected: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := compareVersions(tc.current, tc.latest); got != tc.expected {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", tc.current, tc.latest, got, tc.expected)
			}
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  V
		expected string
	}{
		{
			name: "development build",
			version: V{
				Version: "local",
			},
			expected: "Development build",
		},
		{
			name: "unable to check",
			version: V{
				Version: "1.2.3",
			},
			expected: "Unable to check",
		},
		{
			name: "update available",
			version: V{
				Version:       "1.2.3",
				LatestVersion: "1.2.4",
			},
			expected: "Update available (v1.2.4)",
		},
		{
			name: "metadata only difference stays up to date",
			version: V{
				Version:       "1.2.3+build.1",
				LatestVersion: "1.2.3+build.2",
			},
			expected: "Up to date",
		},
		{
			name: "current version can be ahead",
			version: V{
				Version:       "1.2.4",
				LatestVersion: "1.2.3",
			},
			expected: "Ahead of latest release",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := updateStatus(tc.version); got != tc.expected {
				t.Fatalf("updateStatus(%+v) = %q, want %q", tc.version, got, tc.expected)
			}
		})
	}
}

func TestUpgradeCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   InstallMethod
		latest   string
		expected string
	}{
		{
			name:     "manual install stays executable",
			method:   InstallMethodManual,
			expected: "curl -sf https://raw.githubusercontent.com/Escape-Technologies/cli/refs/heads/main/scripts/install.sh | sudo bash",
		},
		{
			name:     "github actions uses tagged release",
			method:   InstallMethodGitHubActions,
			latest:   "1.2.3",
			expected: "uses: Escape-Technologies/cli@v1.2.3",
		},
		{
			name:     "homebrew uses cask upgrade",
			method:   InstallMethodHomebrew,
			latest:   "1.2.3",
			expected: "brew upgrade --cask escape-technologies/tap/escape-cli",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := UpgradeCommand(tc.method, tc.latest); got != tc.expected {
				t.Fatalf("UpgradeCommand(%q, %q) = %q, want %q", tc.method, tc.latest, got, tc.expected)
			}
		})
	}
}
