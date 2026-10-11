package escape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/log"
)

// ListKubernetesIntegrationsFilters holds optional filters for listing Kubernetes integrations
type ListKubernetesIntegrationsFilters struct {
	ProjectIDs  []string
	LocationIDs string
	Search      string
}

// ListIntegrationsFilters holds optional filters for listing integrations.
type ListIntegrationsFilters struct {
	ProjectIDs  []string
	IDs         []string
	LocationIDs []string
	Search      string
}

type listIntegrationsResponse struct {
	Data       []map[string]interface{} `json:"data"`
	NextCursor *string                  `json:"nextCursor"`
	TotalCount int                      `json:"totalCount"`
}

// UpsertKubernetesIntegration creates a Kubernetes integration if it doesn't exist
func UpsertKubernetesIntegration(ctx context.Context, req v3.CreatekubernetesIntegrationRequest) (*v3.CreateakamaiIntegration200Response, error) {
	list, _, err := listKubernetesIntegrations(ctx, "", &ListKubernetesIntegrationsFilters{
		LocationIDs: strings.Join([]string{*req.ProxyId}, ","),
	})
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	if len(list) > 0 {
		for _, integration := range list {
			if integration.Location.Id == *req.ProxyId {
				log.Info("Kubernetes integration already exists")
				return nil, nil
			}
		}
	}

	log.Info("Creating Kubernetes integration..")
	resp, err := createKubernetesIntegration(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unable to create Kubernetes integration: %w", err)
	}

	log.Info("Kubernetes integration created")

	return resp, nil
}

func createKubernetesIntegration(ctx context.Context, req v3.CreatekubernetesIntegrationRequest) (*v3.CreateakamaiIntegration200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	resp, _, err := client.IntegrationsAPI.CreatekubernetesIntegration(ctx).
		CreatekubernetesIntegrationRequest(req).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return resp, nil
}

// ListKubernetesIntegrations lists Kubernetes integrations
func listKubernetesIntegrations(ctx context.Context, next string, filters *ListKubernetesIntegrationsFilters) ([]v3.ListIntegrations200ResponseDataInner, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}

	rSize := 50
	req := client.IntegrationsAPI.ListkubernetesIntegrations(ctx).Size(rSize)
	if next != "" {
		req = req.Cursor(next)
	}

	if filters != nil {
		if len(filters.ProjectIDs) > 0 {
			req = req.ProjectIds(filters.ProjectIDs)
		}

		if len(filters.LocationIDs) > 0 {
			req = req.LocationIds([]string{filters.LocationIDs})
		}

		if filters.Search != "" {
			req = req.Search(filters.Search)
		}
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, nil
}

// ListIntegrations lists one page of integrations of a given kind.
// size 0 omits the query parameter so the API keeps its default page size.
// The returned total is totalCount.
func ListIntegrations(ctx context.Context, kind, next string, filters *ListIntegrationsFilters, size int) ([]map[string]interface{}, *string, int, error) {
	values := url.Values{}
	if next != "" {
		values.Set("cursor", next)
	}

	if size > 0 {
		values.Set("size", strconv.Itoa(size))
	}

	if filters != nil {
		if len(filters.ProjectIDs) > 0 {
			values.Set("projectIds", strings.Join(filters.ProjectIDs, ","))
		}

		if len(filters.IDs) > 0 {
			values.Set("ids", strings.Join(filters.IDs, ","))
		}

		if len(filters.LocationIDs) > 0 {
			values.Set("locationIds", strings.Join(filters.LocationIDs, ","))
		}

		if filters.Search != "" {
			values.Set("search", filters.Search)
		}
	}

	path := rawPath("integrations", kind)
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp listIntegrationsResponse
	if err := rawRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return resp.Data, resp.NextCursor, resp.TotalCount, nil
}

// GetIntegration gets an integration by kind and ID.
func GetIntegration(ctx context.Context, kind, integrationID string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := rawRequest(ctx, http.MethodGet, rawPath("integrations", kind, integrationID), nil, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return resp, nil
}

// CreateIntegration creates an integration of a given kind from raw JSON bytes.
func CreateIntegration(ctx context.Context, kind string, body []byte) (map[string]interface{}, error) {
	if !json.Valid(body) {
		return nil, errors.New("invalid JSON")
	}

	var resp map[string]interface{}
	if err := rawRequest(ctx, http.MethodPost, rawPath("integrations", kind), body, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return resp, nil
}

// UpdateIntegration updates an integration by kind and ID from raw JSON bytes.
func UpdateIntegration(ctx context.Context, kind, integrationID string, body []byte) (map[string]interface{}, error) {
	if !json.Valid(body) {
		return nil, errors.New("invalid JSON")
	}

	var resp map[string]interface{}
	if err := rawRequest(ctx, http.MethodPut, rawPath("integrations", kind, integrationID), body, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return resp, nil
}

// DeleteIntegration deletes an integration by kind and ID.
// Every generated delete operation returns CreateakamaiIntegration200Response,
// so the kind-generic client decodes that one shape.
func DeleteIntegration(ctx context.Context, kind, integrationID string) (*v3.CreateakamaiIntegration200Response, error) {
	var resp v3.CreateakamaiIntegration200Response
	if err := rawRequest(ctx, http.MethodDelete, rawPath("integrations", kind, integrationID), nil, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &resp, nil
}

// integrationRunResponse is the per-kind run-listing envelope: every kind
// returns the same paginated document.
type integrationRunResponse struct {
	Data       []map[string]interface{} `json:"data"`
	NextCursor *string                  `json:"nextCursor"`
	TotalCount int                      `json:"totalCount"`
}

// TestIntegration asks the integration to check its credentials.
func TestIntegration(ctx context.Context, kind, integrationID string) (*v3.SyncAkamaiIntegration200Response, error) {
	var resp v3.SyncAkamaiIntegration200Response
	if err := rawRequest(ctx, http.MethodPost, rawPath("integrations", kind, integrationID, "test"), nil, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &resp, nil
}

// SyncIntegration schedules a synchronization run of the integration.
func SyncIntegration(ctx context.Context, kind, integrationID string) (*v3.SyncAkamaiIntegration200Response, error) {
	var resp v3.SyncAkamaiIntegration200Response
	if err := rawRequest(ctx, http.MethodPost, rawPath("integrations", kind, integrationID, "sync"), nil, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &resp, nil
}

// ListIntegrationRuns lists one page of the runs of an integration.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListIntegrationRuns(ctx context.Context, kind, integrationID, next string, size int) ([]map[string]interface{}, *string, int, error) {
	path := rawPath("integrations", kind, integrationID, "runs")
	query := url.Values{}
	if next != "" {
		query.Set("cursor", next)
	}

	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}

	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp integrationRunResponse
	if err := rawRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return resp.Data, resp.NextCursor, resp.TotalCount, nil
}

// ListGithubRepositoriesFilters narrows the GitHub repository listing.
// FullNames filters by owner/name; Search matches a substring of it.
type ListGithubRepositoriesFilters struct {
	Search    string
	FullNames []string
}

// GithubRepository is one repository a GitHub integration can reach.
// FullName, CloneURL and IntegrationIDs are the coordinates an asset create
// of type GITHUB_REPOSITORY needs. Description is null when the provider
// has none.
type GithubRepository struct {
	ID             string   `json:"id"`
	FullName       string   `json:"fullName"`
	Name           string   `json:"name"`
	OwnerLogin     string   `json:"ownerLogin"`
	CloneURL       string   `json:"cloneUrl"`
	HTMLURL        string   `json:"htmlUrl"`
	DefaultBranch  string   `json:"defaultBranch"`
	Private        bool     `json:"private"`
	Archived       bool     `json:"archived"`
	Description    *string  `json:"description"`
	IntegrationIDs []string `json:"integrationIds"`
}

// RemoteVault is a vault an integration can see at its provider, before it
// is linked. ExternalID is the coordinate `vault link` sends.
type RemoteVault struct {
	Name       string `json:"name"`
	ExternalID string `json:"externalId"`
}

// repositoryResponse is the GitHub repository listing envelope. The route
// pages like the integration listings do.
type repositoryResponse struct {
	Data       []GithubRepository `json:"data"`
	NextCursor *string            `json:"nextCursor"`
	TotalCount int                `json:"totalCount"`
}

// ListGithubRepositories lists one page of the repositories the organization's
// GitHub credentials can see. An empty integrationID lists every repository
// across the GitHub integrations; otherwise it lists one integration's.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListGithubRepositories(ctx context.Context, integrationID, next string, size int, filters *ListGithubRepositoriesFilters) ([]GithubRepository, *string, int, error) {
	query := url.Values{}
	if next != "" {
		query.Set("cursor", next)
	}

	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}

	if filters != nil {
		if filters.Search != "" {
			query.Set("search", filters.Search)
		}

		// Repeated keys, not a comma-joined value: the route coerces one value or
		// repeated values into an array and does not split on commas, so
		// "owner/a,owner/b" would be looked up as a single repository name.
		for _, fullName := range filters.FullNames {
			query.Add("fullNames", fullName)
		}
	}

	path := rawPath("integrations", "github", "repositories")
	if integrationID != "" {
		path = rawPath("integrations", "github", integrationID, "repositories")
	}

	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp repositoryResponse
	if err := rawRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return resp.Data, resp.NextCursor, resp.TotalCount, nil
}

// ListIntegrationVaults lists the vaults the integration can see at its
// provider, before any of them is linked to the organization. Only names and
// provider coordinates are returned; values never leave the provider.
// search narrows the listing by vault name.
func ListIntegrationVaults(ctx context.Context, kind, integrationID, search string) ([]RemoteVault, error) {
	if strings.TrimSpace(kind) == "" {
		return nil, errors.New("kind is required")
	}

	if strings.TrimSpace(integrationID) == "" {
		return nil, errors.New("integrationID is required")
	}

	path := rawPath("integrations", kind, integrationID, "vaults")
	if search != "" {
		path += "?search=" + url.QueryEscape(search)
	}

	var vaults []RemoteVault
	if err := rawRequest(ctx, http.MethodGet, path, nil, &vaults); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return vaults, nil
}

// JiraProject is one Jira project the integration can file tickets against.
type JiraProject struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	ProjectTypeKey string  `json:"projectTypeKey"`
	EntityID       *string `json:"entityId"`
}

// JiraIssueType is one issue type inside a Jira project.
type JiraIssueType struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Subtask     *bool   `json:"subtask"`
}

// ListJiraProjects lists every Jira project an integration can file tickets
// against, following X-Next-Cursor until the API stops sending one. search
// narrows the listing by project name and is sent on every page. A deployment
// that omits the header returns its array in one response.
//
// The generated client decodes a page envelope, so the bare array is read
// directly and the response headers are kept for the cursor.
func ListJiraProjects(ctx context.Context, integrationID, search string) ([]JiraProject, error) {
	if strings.TrimSpace(integrationID) == "" {
		return nil, errors.New("integrationID is required")
	}

	projects, err := drainCursorPages(func(cursor string) ([]JiraProject, *http.Response, error) {
		query := url.Values{}
		if search != "" {
			query.Set("search", search)
		}

		if cursor != "" {
			query.Set("cursor", cursor)
		}

		path := rawPath("integrations", "jira", integrationID, "projects")
		if encoded := query.Encode(); encoded != "" {
			path += "?" + encoded
		}

		var page []JiraProject
		resp, err := rawResponse(ctx, http.MethodGet, path, nil, &page)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to list Jira projects: %w", humanizeAPIError(err))
		}

		return page, resp, nil
	})
	if err != nil {
		return nil, err
	}

	if projects == nil {
		projects = []JiraProject{}
	}

	return projects, nil
}

// ListJiraIssueTypes lists the issue types of one Jira project.
// projectID is the Jira project id (the projectId query param), not the Escape integration id.
// The route returns a bare array, so this listing uses rawRequest.
func ListJiraIssueTypes(ctx context.Context, integrationID, projectID string) ([]JiraIssueType, error) {
	if strings.TrimSpace(integrationID) == "" {
		return nil, errors.New("integrationID is required")
	}

	if strings.TrimSpace(projectID) == "" {
		return nil, errors.New("projectID is required")
	}

	query := url.Values{}
	query.Set("projectId", projectID)
	path := rawPath("integrations", "jira", integrationID, "issue-types") + "?" + query.Encode()

	var issueTypes []JiraIssueType
	if err := rawRequest(ctx, http.MethodGet, path, nil, &issueTypes); err != nil {
		return nil, fmt.Errorf("unable to list Jira issue types: %w", humanizeAPIError(err))
	}

	if issueTypes == nil {
		issueTypes = []JiraIssueType{}
	}

	return issueTypes, nil
}

// ListAzureBoardsAreas lists the area paths of one Azure Boards project.
// project and workItemType are both required by the route. The route returns
// a bare array of strings, so this listing uses rawRequest.
func ListAzureBoardsAreas(ctx context.Context, integrationID, project, workItemType string) ([]string, error) {
	if strings.TrimSpace(integrationID) == "" {
		return nil, errors.New("integrationID is required")
	}

	if strings.TrimSpace(project) == "" {
		return nil, errors.New("project is required")
	}

	if strings.TrimSpace(workItemType) == "" {
		return nil, errors.New("workItemType is required")
	}

	query := url.Values{}
	query.Set("project", project)
	query.Set("workItemType", workItemType)
	path := rawPath("integrations", "azure-boards", integrationID, "areas") + "?" + query.Encode()

	var areas []string
	if err := rawRequest(ctx, http.MethodGet, path, nil, &areas); err != nil {
		return nil, fmt.Errorf("unable to list Azure Boards areas: %w", humanizeAPIError(err))
	}

	if areas == nil {
		areas = []string{}
	}

	return areas, nil
}
