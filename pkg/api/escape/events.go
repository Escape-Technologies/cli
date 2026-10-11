package escape

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListEventsFilters holds optional filters for listing events
type ListEventsFilters struct {
	Search              string
	After               string
	Before              string
	ProfileIDs          []string
	ScanIDs             []string
	AssetIDs            []string
	IssueIDs            []string
	EventIDs            string
	TargetIDs           string
	WorkflowIDs         string
	Groups              string
	Levels              []string
	Stages              []string
	Severities          []string
	Risks               []string
	ScanProblemCodes    []string
	ResponseStatusCodes string
	HasAttachments      bool
	Attachments         []string
	Public              string
	Dnf                 string
	SortType            string
	SortDirection       string
}

// eventsArrayQuery collects the event list filters the generated client wraps
// in a union-typed query parameter and serializes as a Go type name.
func eventsArrayQuery(filters *ListEventsFilters) url.Values {
	values := url.Values{}
	if filters == nil {
		return values
	}

	addScalarQuery(values, "eventIds", filters.EventIDs)
	addScalarQuery(values, "targetIds", filters.TargetIDs)
	addScalarQuery(values, "workflowIds", filters.WorkflowIDs)
	addScalarQuery(values, "groups", filters.Groups)
	addScalarQuery(values, "responseStatusCodes", filters.ResponseStatusCodes)

	return values
}

// ListEvents lists one page of events.
// size 0 keeps the historical page size of 100. The returned total is totalCount.
func ListEvents(ctx context.Context, next string, filters *ListEventsFilters, size int) ([]v3.EventSummarized, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	maxSize := 100
	if size > 0 {
		maxSize = size
	}

	ctx = withArrayQuery(ctx, eventsArrayQuery(filters))

	req := client.EventsAPI.ListEvents(ctx).Size(maxSize)
	if filters != nil && filters.SortType != "" {
		req = req.SortType(filters.SortType)
	}

	if filters != nil && filters.SortDirection != "" {
		req = req.SortDirection(filters.SortDirection)
	} else {
		req = req.SortDirection("desc")
	}

	if next != "" {
		req = req.Cursor(next)
	}

	if filters != nil {
		req = applyListEventsFilters(req, filters)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// applyListEventsFilters adds every set filter to the request.
func applyListEventsFilters(req v3.ApiListEventsRequest, filters *ListEventsFilters) v3.ApiListEventsRequest {
	if filters.Search != "" {
		req = req.Search(filters.Search)
	}

	if filters.After != "" {
		req = req.After(filters.After)
	}

	if filters.Before != "" {
		req = req.Before(filters.Before)
	}

	if len(filters.ProfileIDs) > 0 {
		req = req.ProfileIds(strings.Join(filters.ProfileIDs, ","))
	}

	if len(filters.ScanIDs) > 0 {
		req = req.ScanIds(strings.Join(filters.ScanIDs, ","))
	}

	if len(filters.AssetIDs) > 0 {
		req = req.AssetIds(strings.Join(filters.AssetIDs, ","))
	}

	if len(filters.IssueIDs) > 0 {
		req = req.IssueIds(strings.Join(filters.IssueIDs, ","))
	}

	if len(filters.Stages) > 0 {
		req = req.Stages(filters.Stages)
	}

	if filters.HasAttachments {
		req = req.HasAttachments(strconv.FormatBool(filters.HasAttachments))
	}

	if len(filters.Attachments) > 0 {
		req = req.Attachments(filters.Attachments)
	}

	if len(filters.Levels) > 0 {
		req = req.Levels(filters.Levels)
	}

	if len(filters.Severities) > 0 {
		req = req.Severities(filters.Severities)
	}

	if len(filters.Risks) > 0 {
		req = req.Risks(filters.Risks)
	}

	if len(filters.ScanProblemCodes) > 0 {
		req = req.ScanProblemCodes(filters.ScanProblemCodes)
	}

	if filters.Public != "" {
		req = req.Public(filters.Public)
	}

	if filters.Dnf != "" {
		req = req.Dnf(filters.Dnf)
	}

	return req
}

// GetEvent gets an event
func GetEvent(ctx context.Context, eventID string) (*v3.GetEvent200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.EventsAPI.GetEvent(ctx, eventID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
