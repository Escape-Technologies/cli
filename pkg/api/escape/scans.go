package escape

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListScansFilters holds optional filters for listing scans
type ListScansFilters struct {
	After      string
	Before     string
	ProfileIDs *[]string
	Ignored    string
	Initiator  *[]string
	Kinds      *[]string
	Status     *[]string
}

// ListScans lists all scans for an application
func ListScans(ctx context.Context, next string, filters *ListScansFilters) ([]v3.ScanSummarized1, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ScansAPI.ListScans(ctx)

	batchSize := 100
	req = req.SortType("createdAt").SortDirection("desc").Size(batchSize)
	if next != "" {
		req = req.Cursor(next)
	}
	if filters != nil {
		if filters.After != "" {
			req = req.After(filters.After)
		}
		if filters.Before != "" {
			req = req.Before(filters.Before)
		}
		if filters.ProfileIDs != nil && len(*filters.ProfileIDs) > 0 {
			req = req.ProfileIds(strings.Join(*filters.ProfileIDs, ","))
		}
		if filters.Ignored != "" {
			req = req.Ignored(filters.Ignored)
		}
		if filters.Initiator != nil && len(*filters.Initiator) > 0 {
			req = req.Initiator(strings.Join(*filters.Initiator, ","))
		}
		if filters.Kinds != nil && len(*filters.Kinds) > 0 {
			req = req.Kinds(strings.Join(*filters.Kinds, ","))
		}
		if filters.Status != nil && len(*filters.Status) > 0 {
			req = req.Status(strings.Join(*filters.Status, ","))
		}
	}
	data, _, err := req.Execute()

	if err != nil {
		return nil, nil, fmt.Errorf("unable to list scans: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetScan returns a scan by its ID
func GetScan(ctx context.Context, scanID string) (*v3.ScanDetailed1, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ScansAPI.GetScan(ctx, scanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get scan: %w", err)
	}
	return data, nil
}

// GetScanIssues returns issues found in a scan
func GetScanIssues(ctx context.Context, scanID string) ([]v3.IssueSummarized, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	// Use IssuesAPI.ListIssues with scanIds filter
	req := client.IssuesAPI.ListIssues(ctx)
	if scanID != "" {
		req = req.ScanIds(scanID)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get scan issues: %w", err)
	}
	return data.Data, nil
}

// StartScan starts a scan for an application
func StartScan(
	ctx context.Context,
	profileID string,
	commitHash string,
	commitLink string,
	commitBranch string,
	commitAuthor string,
	commitAuthorProfilePictureLink string,
	configurationOverride map[string]interface{},
	additionalProperties map[string]interface{},
	initiator v3.ENUMPROPERTIESDATAITEMSPROPERTIESINITIATORSITEMS,
) (*v3.ScanDetailed1, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := v3.StartScanRequest{
		ProfileId:                      profileID,
		CommitHash:                     &commitHash,
		CommitLink:                     &commitLink,
		CommitBranch:                   &commitBranch,
		CommitAuthor:                   &commitAuthor,
		CommitAuthorProfilePictureLink: &commitAuthorProfilePictureLink,
		Initiator:                      (*v3.ENUMPROPERTIESINITIATOR)(&initiator),
		ConfigurationOverride:          configurationOverride,
		AdditionalProperties:           additionalProperties,
	}

	data, _, err := client.ScansAPI.StartScan(ctx).StartScanRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to start scan: %w", err)
	}
	return data, nil
}

// CancelScan cancels a scan
func CancelScan(ctx context.Context, scanID string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}

	_, httpResp, err := client.ScansAPI.CancelScan(ctx, scanID).Execute()
	if err != nil {
		if httpResp.StatusCode == http.StatusBadRequest {
			return errors.New("scan is not running or already canceled")
		}
		return fmt.Errorf("unable to cancel scan: %w", err)
	}
	return nil
}

// IgnoreScan ignore a scan
func IgnoreScan(ctx context.Context, scanID string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}

	ignoreScanRequest := v3.IgnoreScanRequest{
		Ignored: true,
	}

	_, _, err = client.ScansAPI.IgnoreScan(ctx, scanID).IgnoreScanRequest(ignoreScanRequest).Execute()
	if err != nil {
		return fmt.Errorf("unable to ignore scan: %w", err)
	}
	return nil
}
