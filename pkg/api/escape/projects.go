package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListProjectsFilters holds optional filters for listing projects
type ListProjectsFilters struct {
	Search string
}

// ListProjects lists all projects with optional search filter
func ListProjects(ctx context.Context, next string, filters *ListProjectsFilters) ([]v3.ListProjects200ResponseDataInner, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ProjectsAPI.ListProjects(ctx)
	if next != "" {
		req = req.Cursor(next)
	}
	if filters != nil && filters.Search != "" {
		req = req.Search(filters.Search)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetProject gets a project by ID
func GetProject(ctx context.Context, projectID string) (*v3.CreateProject200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ProjectsAPI.GetProject(ctx, projectID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// CreateProject creates a new project from raw JSON bytes
func CreateProject(ctx context.Context, body []byte) (*v3.CreateProject200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	var req v3.CreateProjectRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	data, _, err := client.ProjectsAPI.CreateProject(ctx).CreateProjectRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// UpdateProject updates a project by ID from raw JSON bytes
func UpdateProject(ctx context.Context, projectID string, body []byte) (*v3.CreateProject200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	var req v3.UpdateProjectRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	data, _, err := client.ProjectsAPI.UpdateProject(ctx, projectID).UpdateProjectRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// DeleteProject deletes a project by ID.
func DeleteProject(ctx context.Context, projectID string) error {
	if err := rawRequest(ctx, http.MethodDelete, "/projects/"+projectID, nil, nil); err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}
