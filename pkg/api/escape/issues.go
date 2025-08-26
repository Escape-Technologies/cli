package escape

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListIssuesFilters holds optional filters for listing issues
type ListIssuesFilters struct {
	Status       []string
	Severities   []string
	ProfileIDs   []string
	AssetIDs     []string
	Domains      []string
	IssueIDs     []string
	ScanIDs      []string
	TagsIDs      []string
	Search       string
	JiraTicket   string
	Risks        []string
	AssetClasses []string
	ScannerKinds []string
}

// GetIssue gets an issue by ID
func GetIssue(ctx context.Context, issueID string) (*v3.IssueDetailed, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.IssuesAPI.GetIssue(ctx, issueID)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// ListIssues lists all issues.
func ListIssues(ctx context.Context, next string, filters *ListIssuesFilters) ([]v3.IssueSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.IssuesAPI.ListIssues(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if filters != nil {
		if len(filters.Status) > 0 {
			req = req.Status(strings.Join(filters.Status, ","))
		}
		if len(filters.Severities) > 0 {
			req = req.Severities(strings.Join(filters.Severities, ","))
		}
		if len(filters.ProfileIDs) > 0 {
			req = req.ProfileIds(strings.Join(filters.ProfileIDs, ","))
		}
		if len(filters.AssetIDs) > 0 {
			req = req.AssetIds(strings.Join(filters.AssetIDs, ","))
		}
		if len(filters.Domains) > 0 {
			req = req.Domains(strings.Join(filters.Domains, ","))
		}
		if len(filters.IssueIDs) > 0 {
			req = req.Ids(strings.Join(filters.IssueIDs, ","))
		}
		if len(filters.ScanIDs) > 0 {
			req = req.ScanIds(strings.Join(filters.ScanIDs, ","))
		}
		if len(filters.TagsIDs) > 0 {
			req = req.TagsIds(strings.Join(filters.TagsIDs, ","))
		}
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}
		if filters.JiraTicket != "" {
			req = req.JiraTicket(filters.JiraTicket)
		}
		if len(filters.Risks) > 0 {
			req = req.Risks(filters.Risks)
		}
		if len(filters.AssetClasses) > 0 {
			req = req.AssetClasses(strings.Join(filters.AssetClasses, ","))
		}
		if len(filters.ScannerKinds) > 0 {
			req = req.ScannerKinds(strings.Join(filters.ScannerKinds, ","))
		}
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// UpdateIssue updates an issue
func UpdateIssue(ctx context.Context, issueID string, status v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS) (bool, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return false, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.IssuesAPI.UpdateIssue(ctx, issueID).UpdateIssueRequest(v3.UpdateIssueRequest{
		Status: &status,
		AdditionalProperties: map[string]interface{}{
			"comment": "Updated via CLI",
		},
	})

	_, httpRes, err := req.Execute()
	if err != nil && httpRes.StatusCode != http.StatusOK {
		return false, fmt.Errorf("api error: %w", err)
	}

	return true, nil
}

// ListIssueActivities lists the activities of an issue
func ListIssueActivities(ctx context.Context, issueID string) ([]v3.ActivitySummarized, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.IssuesAPI.ListIssueActivities(ctx, issueID)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}
