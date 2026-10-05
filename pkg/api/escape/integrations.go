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
