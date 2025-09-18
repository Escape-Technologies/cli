package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/env"
)

const updateCheckTimeout = 1000 * time.Millisecond

type githubRelease struct {
    TagName string `json:"tag_name"`
}

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

// WarnIfNotLatestVersion checks the latest GitHub release and prints if the current version is not the latest.
func WarnIfNotLatestVersion(parentCtx context.Context) {
    v := GetVersion()
    if strings.TrimSpace(v.Version) == "" || v.Version == "local" {
        return
    }

    ctx, cancel := context.WithTimeout(parentCtx, updateCheckTimeout)
    defer cancel()

    latest, err := getLatestReleaseTag(ctx)
    if err != nil || latest == "" {
        return
    }

    current := normalizeVersion(v.Version)
    latestNorm := normalizeVersion(latest)
    if current == latestNorm {
        return
    }

    yellow := "\x1b[33m"
    reset := "\x1b[0m"
    msg := fmt.Sprintf("A new version of escape-cli is available: %s (you have %s). Update: https://docs.escape.tech/documentation/tooling/cli/?h=cli", latest, v.Version)
    fmt.Fprintf(os.Stderr, "%s%s%s\n", yellow, msg, reset)
}


