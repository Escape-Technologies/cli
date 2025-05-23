package escape

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/env"
)

// ListScans lists all scans for an application
func ListScans(ctx context.Context, applicationID, next string) ([]v2.ListScans200ResponseDataInner, string, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, "", fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ApplicationsAPI.ListScans(ctx, applicationID)
	if next != "" {
		req.After(next)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, "", fmt.Errorf("unable to list scans: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetScanIssues returns issues found in a scan
func GetScanIssues(ctx context.Context, scanID string) ([]v2.ListIssues200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ScansAPI.ListIssues(ctx, scanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get scan: %w", err)
	}
	return data, nil
}

// StartScan starts a scan for an application
func StartScan(
	ctx context.Context,
	applicationID string,
	commitHash string,
	commitLink string,
	commitBranch string,
	commitAuthor string,
	commitAuthorProfilePictureLink string,
	configurationOverride *v2.CreateApplicationRequestAnyOfConfiguration,
) (*v2.ListScans200ResponseDataInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := v2.StartScanRequest{
		ApplicationId:         applicationID,
		ConfigurationOverride: configurationOverride,
	}
	if commitHash != "" {
		req.CommitHash = &commitHash
	}
	if commitLink != "" {
		req.CommitLink = &commitLink
	}
	if commitBranch != "" {
		req.CommitBranch = &commitBranch
	}
	if commitAuthor != "" {
		req.CommitAuthor = &commitAuthor
	}
	if commitAuthorProfilePictureLink != "" {
		req.CommitAuthorProfilePictureLink = &commitAuthorProfilePictureLink
	}
	data, _, err := client.ScansAPI.StartScan(ctx).StartScanRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to start scan: %w", err)
	}
	return data, nil
}

// DownloadScanExchangesZip downloads the zip file of the scan exchanges and puts them in outPath
func DownloadScanExchangesZip(ctx context.Context, scanID string, outPath string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ScansAPI.GetScanExchangesArchive(ctx, scanID).Execute()
	if err != nil {
		return fmt.Errorf("unable to get scan exchanges zip url: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, data.Archive, nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	resp, err := env.GetHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %w", fmt.Errorf("status code: %d", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: %w", err)
	}
	err = os.WriteFile(outPath, body, 0644) //nolint:mnd
	if err != nil {
		return fmt.Errorf("unable to write file: %w", err)
	}
	return nil
}

// GetScan returns a scan by its ID
func GetScan(ctx context.Context, scanID string) (*v2.ListScans200ResponseDataInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ScansAPI.GetScan(ctx, scanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get scan: %w", err)
	}
	return data, nil
}
