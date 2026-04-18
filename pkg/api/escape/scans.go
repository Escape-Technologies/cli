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
	After         string
	Before        string
	AssetIDs      *[]string
	ProfileIDs    *[]string
	ProjectIDs    *[]string
	Ignored       string
	Initiator     *[]string
	Kinds         *[]string
	Status        *[]string
	SortType      string
	SortDirection string
}

// ListScans lists all scans for an application
func ListScans(ctx context.Context, next string, filters *ListScansFilters) ([]v3.ScanSummarized2, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ScansAPI.ListScans(ctx)

	batchSize := 100
	if filters != nil && filters.SortType != "" {
		req = req.SortType(filters.SortType)
	} else {
		req = req.SortType("createdAt")
	}
	if filters != nil && filters.SortDirection != "" {
		req = req.SortDirection(filters.SortDirection)
	} else {
		req = req.SortDirection("desc")
	}
	req = req.Size(batchSize)
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
		if filters.AssetIDs != nil && len(*filters.AssetIDs) > 0 {
			req = req.AssetIds(strings.Join(*filters.AssetIDs, ","))
		}
		if filters.ProfileIDs != nil && len(*filters.ProfileIDs) > 0 {
			req = req.ProfileIds(strings.Join(*filters.ProfileIDs, ","))
		}
		if filters.ProjectIDs != nil && len(*filters.ProjectIDs) > 0 {
			req = req.ProjectIds(v3.ListScansProjectIdsParameter{ArrayOfString: filters.ProjectIDs})
		}
		if filters.Ignored != "" {
			req = req.Ignored(filters.Ignored)
		}
		if filters.Initiator != nil && len(*filters.Initiator) > 0 {
			req = req.Initiator(*filters.Initiator)
		}
		if filters.Kinds != nil && len(*filters.Kinds) > 0 {
			req = req.Kinds(*filters.Kinds)
		}
		if filters.Status != nil && len(*filters.Status) > 0 {
			req = req.Status(*filters.Status)
		}
	}
	data, _, err := req.Execute()

	if err != nil {
		return nil, nil, fmt.Errorf("unable to list scans: %w", humanizeAPIError(err))
	}
	return data.Data, data.NextCursor, nil
}

// GetScan returns a scan by its ID
func GetScan(ctx context.Context, scanID string) (*v3.StartScan200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ScansAPI.GetScan(ctx, scanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get scan: %w", humanizeAPIError(err))
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
		return nil, fmt.Errorf("unable to get scan issues: %w", humanizeAPIError(err))
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
) (*v3.StartScan200Response, error) {
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
		return nil, fmt.Errorf("unable to start scan: %w", humanizeAPIError(err))
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
		return fmt.Errorf("unable to cancel scan: %w", humanizeAPIError(err))
	}
	return nil
}

// ListScanTargets lists all targets discovered during a scan
func ListScanTargets(ctx context.Context, scanID string, next string, targetTypes string, size int) ([]v3.TargetDetailed, *string, error) {
	if strings.TrimSpace(scanID) == "" {
		return nil, nil, errors.New("scanID is required")
	}

	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ScansAPI.ListScanTargets(ctx, scanID)
	if next != "" {
		req = req.Cursor(next)
	}
	if targetTypes != "" {
		req = req.Types(strings.Split(targetTypes, ","))
	}
	if size > 0 {
		req = req.Size(size)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list scan targets: %w", humanizeAPIError(err))
	}
	return data.Data, data.NextCursor, nil
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
		return fmt.Errorf("unable to ignore scan: %w", humanizeAPIError(err))
	}
	return nil
}
