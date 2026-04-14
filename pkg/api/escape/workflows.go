package escape

import (
	"context"
	"encoding/json"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListWorkflowsFilters holds optional filters for listing workflows
type ListWorkflowsFilters struct {
	Triggers []string
	Search   string
}

// ListWorkflows lists all workflows with optional filters
func ListWorkflows(ctx context.Context, next string, filters *ListWorkflowsFilters) ([]v3.WorkflowSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.WorkflowsAPI.ListWorkflows(ctx)
	if next != "" {
		req = req.Cursor(next)
	}
	if filters != nil {
		if len(filters.Triggers) > 0 {
			req = req.Triggers(filters.Triggers)
		}
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetWorkflow gets a workflow by ID
func GetWorkflow(ctx context.Context, workflowID string) (*v3.CreateWorkflow200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.WorkflowsAPI.GetWorkflow(ctx, workflowID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// CreateWorkflow creates a new workflow from raw JSON bytes
func CreateWorkflow(ctx context.Context, body []byte) (*v3.CreateWorkflow200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	var req v3.CreateWorkflowRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	data, _, err := client.WorkflowsAPI.CreateWorkflow(ctx).CreateWorkflowRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// UpdateWorkflow updates a workflow by ID from raw JSON bytes
func UpdateWorkflow(ctx context.Context, workflowID string, body []byte) (*v3.CreateWorkflow200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	var req v3.UpdateWorkflowRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	data, _, err := client.WorkflowsAPI.UpdateWorkflow(ctx, workflowID).UpdateWorkflowRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// DeleteWorkflow deletes a workflow by ID
func DeleteWorkflow(ctx context.Context, workflowID string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, _, err = client.WorkflowsAPI.DeleteWorkflow(ctx, workflowID).Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}
