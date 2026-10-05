package escape

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// retestListPageSize is the public API maximum page size for GET /retests.
const retestListPageSize = 100

// ListRetestsFilters holds optional filters for listing retests.
type ListRetestsFilters struct {
	ProfileIDs    []string
	SortType      string
	SortDirection string
}

// ListRetests lists one page of issue retests. The id of each item is the scan id.
// Empty sort fields default to createdAt desc so the newest retest is first.
// size <= 0 keeps the historical request page size (the API maximum). The
// returned int is totalCount.
func ListRetests(ctx context.Context, next string, filters *ListRetestsFilters, size int) ([]v3.RetestDetailed, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	sortType := "createdAt"
	sortDirection := "desc"
	var profileIDs []string
	if filters != nil {
		if filters.SortType != "" {
			sortType = filters.SortType
		}

		if filters.SortDirection != "" {
			sortDirection = filters.SortDirection
		}

		profileIDs = filters.ProfileIDs
	}

	pageSize := size
	if pageSize <= 0 {
		pageSize = retestListPageSize
	}

	req := client.RetestsAPI.ListRetests(ctx).
		Size(pageSize).
		SortType(sortType).
		SortDirection(sortDirection)
	if next != "" {
		req = req.Cursor(next)
	}

	if len(profileIDs) > 0 {
		req = req.ProfileIds(strings.Join(profileIDs, ","))
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to list retests: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// GetRetest gets one page of a retest by scan id. Reports are cursor-paginated.
// size <= 0 leaves the API default page size in place.
func GetRetest(ctx context.Context, retestID string, cursor string, size int) (*v3.GetRetest200Response, error) {
	if strings.TrimSpace(retestID) == "" {
		return nil, errors.New("retest ID is required")
	}

	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.RetestsAPI.GetRetest(ctx, retestID)
	if cursor != "" {
		req = req.Cursor(cursor)
	}

	if size > 0 {
		req = req.Size(size)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get retest: %w", humanizeAPIError(err))
	}

	return data, nil
}

// StartRetest starts an issue retest. The request must contain a profile id and
// exactly one of issueIds or filter. The returned id is the scan id.
func StartRetest(ctx context.Context, request v3.StartRetestRequest) (*v3.StartRetest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.RetestsAPI.StartRetest(ctx).StartRetestRequest(request).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to start retest: %w", humanizeAPIError(err))
	}

	return data, nil
}
