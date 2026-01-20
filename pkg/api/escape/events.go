package escape

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListEventsFilters holds optional filters for listing events
type ListEventsFilters struct {
	Search string
	ScanIDs []string
	AssetIDs []string
	IssueIDs []string
	Levels []string
	Stages []string
	HasAttachments bool
	Attachments []string
}

// ListEvents lists events
func ListEvents(ctx context.Context, next string, filters *ListEventsFilters) ([]v3.EventSummarized, *string, error) {
	client, err := NewAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	maxSize := 100
	req := client.EventsAPI.ListEvents(ctx).SortDirection("desc").Size(maxSize)
	if next != "" {
		req = req.Cursor(next)
	}
	if filters != nil {
		if filters.Search != "" {
			req = req.Search(filters.Search)
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
			req = req.Stages(strings.Join(filters.Stages, ","))
		}
		if filters.HasAttachments {
			req = req.HasAttachments(strconv.FormatBool(filters.HasAttachments))
		}
		if len(filters.Attachments) > 0 {
			req = req.Attachments(strings.Join(filters.Attachments, ","))
		}
		if len(filters.Levels) == 0 {
			req = req.Levels(string(v3.ENUMPROPERTIESDATAITEMSPROPERTIESLEVEL_ERROR))
		} else {
			req = req.Levels(strings.Join(filters.Levels, ","))
		}
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetEvent gets an event
func GetEvent(ctx context.Context, eventID string) (*v3.EventDetailed, error) {
	client, err := NewAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.EventsAPI.GetEvent(ctx, eventID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}