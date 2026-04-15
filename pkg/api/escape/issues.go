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
	Names        []string
}

// GetIssue gets an issue by ID
func GetIssue(ctx context.Context, issueID string) (*v3.GetIssue200Response, error) {
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
func ListIssues(ctx context.Context, next string, filters *ListIssuesFilters, sortType string, sortDirection string) ([]v3.IssueSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.IssuesAPI.ListIssues(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if sortType == "" {
		sortType = "LAST_SEEN"
	}
	req = req.SortType(sortType)
	if sortDirection != "" {
		req = req.SortDirection(sortDirection)
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
			req = req.TagIds(strings.Join(filters.TagsIDs, ","))
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
		if len(filters.Names) > 0 {
			req = req.Names(v3.ListIssuesNamesParameter{ArrayOfString: &filters.Names})
		}
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// UpdateIssue updates an issue status with an optional comment
func UpdateIssue(ctx context.Context, issueID string, status v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS, comment string) (bool, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return false, fmt.Errorf("unable to init client: %w", err)
	}

	if comment == "" {
		comment = "Updated via CLI"
	}
	req := client.IssuesAPI.UpdateIssue(ctx, issueID).UpdateIssueRequest(v3.UpdateIssueRequest{
		Status: &status,
		AdditionalProperties: map[string]interface{}{
			"comment": comment,
		},
	})

	_, httpRes, err := req.Execute()
	if err != nil && httpRes.StatusCode != http.StatusOK {
		return false, fmt.Errorf("api error: %w", err)
	}

	return true, nil
}

// CommentIssue adds a comment to an issue
func CommentIssue(ctx context.Context, issueID string, comment string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, _, err = client.IssuesAPI.CreateIssueComment(ctx, issueID).CreateAssetCommentRequest(v3.CreateAssetCommentRequest{
		Comment: comment,
	}).Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
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

// GetIssueFunnel returns the issue funnel breakdown
func GetIssueFunnel(ctx context.Context, projectIDs []string) ([]v3.IssueFunnelInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.IssuesAPI.GetIssueFunnel(ctx)
	if len(projectIDs) > 0 {
		req = req.ProjectIds(v3.GetIssueFunnelProjectIdsParameter{ArrayOfString: &projectIDs})
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// GetIssueTrends returns severity trends over time
func GetIssueTrends(ctx context.Context, after, before, interval string, applicationIDs, projectIDs []string) ([]v3.IssueTrendsInner, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.IssuesAPI.GetIssueTrends(ctx).After(after).Before(before)
	if interval != "" {
		req = req.Interval(interval)
	}
	if len(applicationIDs) > 0 {
		req = req.ApplicationIds(v3.GetIssueTrendsApplicationIdsParameter{ArrayOfString: &applicationIDs})
	}
	if len(projectIDs) > 0 {
		req = req.ProjectIds(v3.GetIssueTrendsProjectIdsParameter{ArrayOfString: &projectIDs})
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// BulkUpdateIssues updates multiple issues matching a filter
func BulkUpdateIssues(ctx context.Context, status v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS, where *v3.BulkUpdateIssuesRequestWhere) (*v3.BulkUpdateIssues200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	body := v3.BulkUpdateIssuesRequest{Status: status}
	if where != nil {
		body.Where = where
	}
	data, _, err := client.IssuesAPI.BulkUpdateIssues(ctx).BulkUpdateIssuesRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// NotifyIssueOwners sends a notification to asset owners about an issue
func NotifyIssueOwners(ctx context.Context, issueID, scanID string) (bool, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return false, fmt.Errorf("unable to init client: %w", err)
	}
	body := v3.NotifyIssueOwnersRequest{ScanId: scanID}
	data, _, err := client.IssuesAPI.NotifyIssueOwners(ctx, issueID).NotifyIssueOwnersRequest(body).Execute()
	if err != nil {
		return false, fmt.Errorf("api error: %w", err)
	}
	return data.GetNotified(), nil
}
