package escape

import (
	"context"
	"fmt"
	"net/http"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

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

// ListIssues lists all issues
func ListIssues(ctx context.Context, next string, issueStatus []string, issueSeverity []string) ([]v3.IssueSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.IssuesAPI.ListIssues(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if issueStatus != nil {
		req = req.Status(issueStatus)
	}

	if issueSeverity != nil {
		req = req.Severities(issueSeverity)
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
func ListIssueActivities(ctx context.Context, issueID string) (*v3.ActivitySummarized, error) {
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
