package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

type ListIntegrationsFilters struct {
	ProjectIDs  []string
	IDs         []string
	LocationIDs []string
	Search      string
}

type listIntegrationsResponse struct {
	Data       []map[string]interface{} `json:"data"`
	NextCursor *string                  `json:"nextCursor"`
}

// UpsertKubernetesIntegration creates a Kubernetes integration if it doesn't exist
func UpsertKubernetesIntegration(ctx context.Context, req v3.CreatekubernetesIntegrationRequest) (*v3.CreatekubernetesIntegration200Response, error) {
	list, _, err := listKubernetesIntegrations(ctx, "", &ListKubernetesIntegrationsFilters{
		LocationIDs: strings.Join([]string{*req.ProxyId}, ","),
	})
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
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

func createKubernetesIntegration(ctx context.Context, req v3.CreatekubernetesIntegrationRequest) (*v3.CreatekubernetesIntegration200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, _, err := client.IntegrationsAPI.CreatekubernetesIntegration(ctx).
		CreatekubernetesIntegrationRequest(req).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
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
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

func ListIntegrations(ctx context.Context, kind, next string, filters *ListIntegrationsFilters) ([]map[string]interface{}, *string, error) {
	values := url.Values{}
	if next != "" {
		values.Set("cursor", next)
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

	path := "/integrations/" + kind
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp listIntegrationsResponse
	if err := rawRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return resp.Data, resp.NextCursor, nil
}

func GetIntegration(ctx context.Context, kind, integrationID string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := rawRequest(ctx, http.MethodGet, "/integrations/"+kind+"/"+integrationID, nil, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return resp, nil
}

func CreateIntegration(ctx context.Context, kind string, body []byte) (map[string]interface{}, error) {
	if !json.Valid(body) {
		return nil, fmt.Errorf("invalid JSON")
	}

	var resp map[string]interface{}
	if err := rawRequest(ctx, http.MethodPost, "/integrations/"+kind, body, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return resp, nil
}

func UpdateIntegration(ctx context.Context, kind, integrationID string, body []byte) (map[string]interface{}, error) {
	if !json.Valid(body) {
		return nil, fmt.Errorf("invalid JSON")
	}

	var resp map[string]interface{}
	if err := rawRequest(ctx, http.MethodPut, "/integrations/"+kind+"/"+integrationID, body, &resp); err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return resp, nil
}

func DeleteIntegration(ctx context.Context, kind, integrationID string) error {
	if err := rawRequest(ctx, http.MethodDelete, "/integrations/"+kind+"/"+integrationID, nil, nil); err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}
