package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// GetStatistics returns organization-level statistics
func GetStatistics(ctx context.Context) (*v3.GetStatistics200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.StatisticsAPI.GetStatistics(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// TriggerAsmScans triggers ASM scans on assets matching an optional filter
func TriggerAsmScans(ctx context.Context, where *v3.TriggerAsmScansRequestWhere) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	req := client.AsmAPI.TriggerAsmScans(ctx)
	body := v3.TriggerAsmScansRequest{}
	if where != nil {
		body.Where = where
	}
	req = req.TriggerAsmScansRequest(body)
	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}
