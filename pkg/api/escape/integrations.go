package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/log"
)

// ListKubernetesIntegrationsFilters holds optional filters for listing Kubernetes integrations
type ListKubernetesIntegrationsFilters struct {
	ProjectIDs []string
	LocationID string
	Search     string
}

// UpsertKubernetesIntegration creates a Kubernetes integration if it doesn't exist
func UpsertKubernetesIntegration(ctx context.Context, req v3.CreatekubernetesIntegrationRequest) (*v3.CreatekubernetesIntegration200Response, error) {
	list, _, err := listKubernetesIntegrations(ctx, "", &ListKubernetesIntegrationsFilters{
		LocationID: *req.ProxyId,
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
		if len(filters.LocationID) > 0 {
			req = req.LocationId(filters.LocationID)
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
