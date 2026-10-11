package escape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// JobsFilters narrows a job listing. Empty fields are left out of the query
// so the server keeps its own default.
type JobsFilters struct {
	Search        string
	Kinds         []string
	Statuses      []string
	SortType      string
	SortDirection string
}

// ListJobs returns one page of the async jobs of the organization, most
// recent first. size 0 keeps the API default page size. The returned total
// is totalCount.
func ListJobs(ctx context.Context, next string, size int, filters *JobsFilters) ([]v3.JobSummary, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.JobsAPI.ListJobs(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if filters != nil {
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}

		if len(filters.Kinds) > 0 {
			req = req.Kinds(filters.Kinds)
		}

		if len(filters.Statuses) > 0 {
			req = req.Statuses(filters.Statuses)
		}

		if filters.SortType != "" {
			req = req.SortType(filters.SortType)
		}

		if filters.SortDirection != "" {
			req = req.SortDirection(filters.SortDirection)
		}
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// GetJob gets the status and result of an async job
func GetJob(ctx context.Context, jobID string) (*v3.GetJob200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.JobsAPI.GetJob(ctx, jobID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ExportOptions carries the optional fields of the export request. The where
// fields hold JSON GraphQL where-inputs, validated by the caller.
type ExportOptions struct {
	ScanID       string
	Name         string
	AssetWhere   json.RawMessage
	ProfileWhere json.RawMessage
	IssueWhere   json.RawMessage
	// Dry validates the selection and schedules no job.
	Dry bool
	// Notify is nil when the caller did not ask for a change: the API
	// defaults to emailing the report when it is ready.
	Notify *bool
}

// TriggerExport triggers an async PDF/report export job
func TriggerExport(ctx context.Context, blocks []string, opts ExportOptions) (*v3.TriggerExport200Response, error) {
	if len(blocks) == 0 {
		return nil, errors.New("at least one export block is required")
	}

	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	blockItems := make([]v3.TriggerExportRequestBlocksInner, 0, len(blocks))
	for _, b := range blocks {
		kind := v3.ENUMPROPERTIESBLOCKSITEMSPROPERTIESKIND(b)
		blockItems = append(blockItems, v3.TriggerExportRequestBlocksInner{Kind: kind})
	}

	req := v3.TriggerExportRequest{Blocks: blockItems}
	if opts.ScanID != "" {
		req.ScanId = &opts.ScanID
	}

	if opts.Name != "" {
		req.Name = &opts.Name
	}

	if len(opts.AssetWhere) > 0 {
		req.AssetWhere = opts.AssetWhere
	}

	if len(opts.ProfileWhere) > 0 {
		req.ProfileWhere = opts.ProfileWhere
	}

	if len(opts.IssueWhere) > 0 {
		req.IssueWhere = opts.IssueWhere
	}

	if opts.Dry {
		req.Dry = &opts.Dry
	}

	if opts.Notify != nil {
		req.Notify = opts.Notify
	}

	data, _, err := client.JobsAPI.TriggerExport(ctx).TriggerExportRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
