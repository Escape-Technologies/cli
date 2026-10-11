package escape

import (
	"context"
	"fmt"
	"net/url"

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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// StatisticsSource selects which scanner family a sub-resource counts.
type StatisticsSource string

const (
	// StatisticsSourceDAST counts DAST findings.
	StatisticsSourceDAST StatisticsSource = "DAST"
	// StatisticsSourcePentest counts AI pentest findings.
	StatisticsSourcePentest StatisticsSource = "PENTEST"
)

// GetStatisticsEndpoints returns how many API routes, GraphQL resolvers and
// web pages the scans of one source reached. source empty keeps the API
// default; projectID empty keeps every project.
func GetStatisticsEndpoints(ctx context.Context, source StatisticsSource, projectID string) (*v3.GetStatisticsEndpoints200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.StatisticsAPI.GetStatisticsEndpoints(ctx)
	if source != "" {
		req = req.Source(string(source))
	}

	if projectID != "" {
		req = req.ProjectId(projectID)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// GetStatisticsIssues returns the open and recently resolved issue counts of
// one source, grouped by status and by severity.
func GetStatisticsIssues(ctx context.Context, source StatisticsSource, projectID string) (*v3.GetStatisticsIssues200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.StatisticsAPI.GetStatisticsIssues(ctx)
	if source != "" {
		req = req.Source(string(source))
	}

	if projectID != "" {
		req = req.ProjectId(projectID)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// GetStatisticsTrends returns the time-bucketed trends of the organization:
// open versus resolved issue counts, and the average API health score. start
// and end are ISO 8601 bounding the half-open window; interval empty keeps
// the API default bucket width; an empty projectIDs keeps every project.
func GetStatisticsTrends(ctx context.Context, start, end, interval string, projectIDs []string) (*v3.GetStatisticsTrends200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	values := url.Values{}
	addArrayQuery(values, "projectIds", projectIDs)
	ctx = withArrayQuery(ctx, values)

	req := client.StatisticsAPI.GetStatisticsTrends(ctx).Start(start).End(end)
	if interval != "" {
		req = req.Interval(interval)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// TriggerAsmScans triggers ASM scans on assets matching an optional filter
func TriggerAsmScans(ctx context.Context, where *v3.TriggerAsmScansRequestWhere) (*v3.TriggerAsmScans200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.AsmAPI.TriggerAsmScans(ctx)
	body := v3.TriggerAsmScansRequest{}
	if where != nil {
		body.Where = where
	}

	req = req.TriggerAsmScansRequest(body)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
