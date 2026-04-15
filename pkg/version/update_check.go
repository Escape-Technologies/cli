package version //nolint:revive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Escape-Technologies/cli/pkg/env"
)

const updateCheckTimeout = 1000 * time.Millisecond

type githubRelease struct {
	TagName string `json:"tag_name"`
}

type UpdateInfo struct {
	Current   string `json:"current,omitempty"`
	Latest    string `json:"latest,omitempty"`
	Available bool   `json:"available,omitempty"`
}

var (
	updateInfoOnce sync.Once
	updateInfo     *UpdateInfo
)

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}

func getLatestReleaseTag(ctx context.Context) (string, error) {
	client := env.GetHTTPClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/Escape-Technologies/cli/releases/latest", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("perform request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return strings.TrimSpace(release.TagName), nil
}

// CheckForUpdate checks the latest GitHub release and caches the result for the current process.
func CheckForUpdate(parentCtx context.Context) *UpdateInfo {
	v := GetVersion()
	current := normalizeVersion(v.Version)
	if current == "" {
		return &UpdateInfo{}
	}

	if current == "local" {
		return &UpdateInfo{Current: current}
	}

	updateInfoOnce.Do(func() {
		info := &UpdateInfo{Current: current}

		ctx, cancel := context.WithTimeout(parentCtx, updateCheckTimeout)
		defer cancel()

		latest, err := getLatestReleaseTag(ctx)
		if err != nil || latest == "" {
			updateInfo = info
			return
		}

		info.Latest = normalizeVersion(latest)
		info.Available = info.Current != "" && info.Current != info.Latest
		updateInfo = info
	})

	if updateInfo == nil {
		return &UpdateInfo{Current: current}
	}

	return updateInfo
}
