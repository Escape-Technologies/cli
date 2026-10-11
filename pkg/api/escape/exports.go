package escape

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/env"
)

const (
	csvAccept   = "text/csv"
	sarifAccept = "application/sarif+json"
)

// Export is one synchronous CSV or SARIF download.
// ContentType is the response media type, without parameters.
type Export struct {
	ContentType string
	Body        []byte
}

// ExportIssuesCSV downloads issues as CSV.
// where is a JSON IssueWhereInput object, or empty for the whole organization.
func ExportIssuesCSV(ctx context.Context, where string) (Export, error) {
	document, err := getExport(ctx, rawPath("exports", "issues", "csv"), csvAccept, where)
	if err != nil {
		return Export{}, fmt.Errorf("unable to export issues csv: %w", err)
	}

	return document, nil
}

// ExportIssuesSARIF downloads issues as a SARIF 2.1.0 document.
// where is a JSON IssueWhereInput object, or empty for the whole organization.
func ExportIssuesSARIF(ctx context.Context, where string) (Export, error) {
	document, err := getExport(ctx, rawPath("exports", "issues", "sarif"), sarifAccept, where)
	if err != nil {
		return Export{}, fmt.Errorf("unable to export issues sarif: %w", err)
	}

	return document, nil
}

// ExportAssetsCSV downloads assets as CSV.
// where is a JSON AssetWhereInput object, or empty for the whole organization.
func ExportAssetsCSV(ctx context.Context, where string) (Export, error) {
	document, err := getExport(ctx, rawPath("exports", "assets", "csv"), csvAccept, where)
	if err != nil {
		return Export{}, fmt.Errorf("unable to export assets csv: %w", err)
	}

	return document, nil
}

// getExport downloads one synchronous export.
// rawRequest decodes JSON objects. These responses are CSV text, or a SARIF
// document whose media type is application/sarif+json. Accept names that
// media type, and where is the JSON predicate the route already expects.
func getExport(ctx context.Context, path, accept, where string) (Export, error) {
	baseURL, err := env.GetAPIURL()
	if err != nil {
		return Export{}, fmt.Errorf("failed to get API URL: %w", err)
	}

	authorization, err := env.GetAuthorizationHeader()
	if err != nil {
		return Export{}, fmt.Errorf("failed to get authorization header: %w", err)
	}

	endpoint := baseURL.String() + "/v3" + path
	if where != "" {
		query := url.Values{}
		query.Set("where", where)
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Export{}, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Authorization", authorization)
	req.Header.Set("Accept", accept)

	client := env.GetHTTPClient()
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && client.Timeout == 0 {
		client.Timeout = defaultRawRequestTimeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return Export{}, fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Export{}, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = resp.Status
		}

		return Export{}, fmt.Errorf("api error (%s): %s", resp.Status, msg)
	}

	return Export{ContentType: mediaType(resp.Header.Get("Content-Type")), Body: data}, nil
}

func mediaType(header string) string {
	if cut, _, found := strings.Cut(header, ";"); found {
		return strings.TrimSpace(cut)
	}

	return strings.TrimSpace(header)
}
