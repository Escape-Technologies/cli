package escape

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListRegressionTests returns one page of regression tests.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListRegressionTests(ctx context.Context, next string, search string, size int) ([]v3.RegressionTestSummary, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.RegressionTestsAPI.ListRegressionTests(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if search != "" {
		req = req.Search(search)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// GetRegressionTest returns a regression test by ID, including its recent runs.
func GetRegressionTest(ctx context.Context, regressionTestID string) (*v3.CreateRegressionTest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.RegressionTestsAPI.GetRegressionTest(ctx, regressionTestID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// CreateRegressionTest creates a regression test from a source report uploaded
// beforehand via POST /upload/signed-url.
func CreateRegressionTest(ctx context.Context, data []byte) (*v3.CreateRegressionTest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateRegressionTestInput
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for CreateRegressionTestInput: %w", err)
	}

	result, _, err := client.RegressionTestsAPI.CreateRegressionTest(ctx).CreateRegressionTestInput(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// UpdateRegressionTest updates the name and/or the additional context of a test.
func UpdateRegressionTest(ctx context.Context, regressionTestID string, data []byte) (*v3.CreateRegressionTest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.UpdateRegressionTestRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for UpdateRegressionTestRequest: %w", err)
	}

	result, _, err := client.RegressionTestsAPI.UpdateRegressionTest(ctx, regressionTestID).UpdateRegressionTestRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// DeleteRegressionTest deletes a regression test by ID.
func DeleteRegressionTest(ctx context.Context, regressionTestID string) (*v3.DeleteProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	result, _, err := client.RegressionTestsAPI.DeleteRegressionTest(ctx, regressionTestID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// RunRegressionTest starts a run of the regression test. The run is asynchronous.
func RunRegressionTest(ctx context.Context, regressionTestID string) (*v3.CreateRegressionTest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	result, _, err := client.RegressionTestsAPI.RunRegressionTest(ctx, regressionTestID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// StopRegressionTest stops the run in progress for a regression test.
func StopRegressionTest(ctx context.Context, regressionTestID string) (*v3.CreateRegressionTest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	result, _, err := client.RegressionTestsAPI.StopRegressionTest(ctx, regressionTestID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// ListRegressionTestHistory returns one page of runs of a regression test.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListRegressionTestHistory(ctx context.Context, regressionTestID string, next string, size int) ([]v3.RegressionTestRun, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.RegressionTestsAPI.ListRegressionTestHistory(ctx, regressionTestID)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// ReplyRegressionTestClarification answers the clarification question of a run.
func ReplyRegressionTestClarification(ctx context.Context, regressionTestID string, runID string, data []byte) (*v3.ReplyRegressionTestClarification200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.ReplyRegressionTestClarificationInput
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for ReplyRegressionTestClarificationInput: %w", err)
	}

	result, _, err := client.RegressionTestsAPI.ReplyRegressionTestClarification(ctx, regressionTestID, runID).ReplyRegressionTestClarificationInput(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}
