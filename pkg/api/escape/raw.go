package escape

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/env"
)

func rawRequest(ctx context.Context, method, path string, body []byte, out any) error {
	baseURL, err := env.GetAPIURL()
	if err != nil {
		return fmt.Errorf("failed to get API URL: %w", err)
	}
	key, err := env.GetAPIKey()
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	url := baseURL.String() + "/v3" + path
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Key "+key)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	client := env.GetHTTPClient()
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && client.Timeout == 0 {
		client.Timeout = 60 * time.Second
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("api error (%s): %s", resp.Status, msg)
	}

	if out == nil || resp.ContentLength == 0 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

func rawPath(segments ...string) string {
	escaped := make([]string, 0, len(segments))
	for _, segment := range segments {
		escaped = append(escaped, url.PathEscape(segment))
	}
	return "/" + strings.Join(escaped, "/")
}
