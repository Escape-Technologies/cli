package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListViewIssuesFilters holds the optional sort of a view's issue listing.
type ListViewIssuesFilters struct {
	SortType      string
	SortDirection string
}

// ListViews lists the saved views of one target. next selects the
// next-generation views (the API default) or the legacy table views.
func ListViews(ctx context.Context, target, projectID string, next bool) ([]v3.ViewDetail, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ViewsAPI.ListViews(ctx).Target(target)
	if projectID != "" {
		req = req.ProjectId(projectID)
	}

	req = req.Next(strconv.FormatBool(next))

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.GetData(), nil
}

// CreateView creates a view from raw JSON bytes.
func CreateView(ctx context.Context, body []byte) (*v3.CreateView200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var req v3.CreateViewRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	data, _, err := client.ViewsAPI.CreateView(ctx).CreateViewRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// UpdateView updates a view by ID from raw JSON bytes.
func UpdateView(ctx context.Context, viewID string, body []byte) (*v3.CreateView200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var req v3.UpdateViewRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	data, _, err := client.ViewsAPI.UpdateView(ctx, viewID).UpdateViewRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// DeleteView deletes a view by ID.
func DeleteView(ctx context.Context, viewID string) (*v3.DeleteProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.ViewsAPI.DeleteView(ctx, viewID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ReorderViews stores the display order of one target's views from raw JSON bytes.
func ReorderViews(ctx context.Context, body []byte) (*v3.ReorderViews200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var req v3.ReorderViewsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	data, _, err := client.ViewsAPI.ReorderViews(ctx).ReorderViewsRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ListViewIssues lists one page of the issues a saved view selects.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListViewIssues(ctx context.Context, viewID, next string, size int, filters *ListViewIssuesFilters) ([]v3.IssueSummarized, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ViewsAPI.ListViewIssues(ctx, viewID)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if filters != nil {
		if filters.SortType != "" {
			req = req.SortType(filters.SortType)
		}

		if filters.SortDirection != "" {
			req = req.SortDirection(filters.SortDirection)
		}
	}

	data, _, listErr := req.Execute()
	if listErr != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(listErr))
	}

	nextCursor := data.GetNextCursor()

	return data.GetData(), &nextCursor, data.GetTotalCount(), nil
}
