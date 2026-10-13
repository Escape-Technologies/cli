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

const defaultRawRequestTimeout = 60 * time.Second

func rawRequest(ctx context.Context, method, path string, body []byte, out any) error {
	_, err := rawResponse(ctx, method, path, body, out)

	return err
}

// rawResponse is rawRequest that also returns the response. The body is closed
// before return; the headers stay readable, which is how a bare-array listing
// follows X-Next-Cursor. Callers that only need the decoded body use rawRequest.
func rawResponse(ctx context.Context, method, path string, body []byte, out any) (*http.Response, error) {
	baseURL, err := env.GetAPIURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get API URL: %w", err)
	}

	authorization, err := env.GetAuthorizationHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization header: %w", err)
	}

	url := baseURL.String() + "/v3" + path
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Authorization", authorization)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	client := env.GetHTTPClient()
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && client.Timeout == 0 {
		client.Timeout = defaultRawRequestTimeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)

		return nil, &rawAPIError{status: resp.Status, body: data}
	}

	if out == nil || resp.ContentLength == 0 {
		_, _ = io.Copy(io.Discard, resp.Body)

		return resp, nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		if err == io.EOF {
			return resp, nil
		}

		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resp, nil
}

// rawAPIError is an HTTP failure from rawRequest. The status and body stay
// available so humanizeAPIError can turn a 401 into InvalidAPIKeyMessage and
// a {message, details} body into the same text the generated client produces.
// Error() keeps the previous "api error (status): body" string.
type rawAPIError struct {
	status string
	body   []byte
}

func (e *rawAPIError) Error() string {
	msg := strings.TrimSpace(string(e.body))
	if msg == "" {
		msg = e.status
	}

	return fmt.Sprintf("api error (%s): %s", e.status, msg)
}

func rawPath(segments ...string) string {
	escaped := make([]string, 0, len(segments))
	for _, segment := range segments {
		escaped = append(escaped, url.PathEscape(segment))
	}

	return "/" + strings.Join(escaped, "/")
}
