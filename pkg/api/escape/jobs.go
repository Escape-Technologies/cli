package escape

import (
	"context"
	"errors"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

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

// TriggerExport triggers an async PDF/report export job
func TriggerExport(ctx context.Context, blocks []string, scanID string) (*v3.TriggerExport200Response, error) {
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
	if scanID != "" {
		req.ScanId = &scanID
	}
	data, _, err := client.JobsAPI.TriggerExport(ctx).TriggerExportRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return data, nil
}
