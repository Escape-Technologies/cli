package escape

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/tabwriter"
	"time"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Integration struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type IntegrationWithParameters struct {
	Id string `json:"id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Parameters map[string]interface{} `json:"data"`
	LastValidationAt time.Time `json:"lastValidationAt"`
}

func FormatIntegrationTable(integrations []IntegrationWithParameters) string {
	var sb strings.Builder

	w := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 1)
	fmt.Fprintf(w, "ID\tName\tCreatedAt\tUpdatedAt\tLastValidationAt\n")
	for _, integration := range integrations {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
			integration.Id, 
			integration.Name,
			integration.CreatedAt.Format(time.RFC3339), 
			integration.UpdatedAt.Format(time.RFC3339), 
			integration.LastValidationAt.Format(time.RFC3339))
	}
	w.Flush()
	
	return sb.String()
}

func FormatIntegrationsTable(integrations []Integration) string {
	sb := strings.Builder{}

	w := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 1)
	fmt.Fprintf(w, "ID\tName\tKind\tCreatedAt\tUpdatedAt\n")
	for _, integration := range integrations {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", integration.Id, integration.Name, integration.Kind, integration.CreatedAt.Format(time.RFC3339), integration.UpdatedAt.Format(time.RFC3339))
	}
	w.Flush()
	return sb.String()
}

func GetIntegrations(ctx context.Context) ([]Integration, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetIntegrationsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get integrations: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get integrations: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get integrations: no data received")
	}
	integrations := []Integration{}
	for _, integration := range *resp.JSON200 {
		integrations = append(integrations, Integration{
			Id: integration.Id.String(),
			Name: integration.Name,
			Kind: string(integration.Kind),
			CreatedAt: integration.CreatedAt,
			UpdatedAt: integration.UpdatedAt,
		})
	}
	return integrations, nil
}

func GetIntegration(ctx context.Context, id openapi_types.UUID) (*IntegrationWithParameters, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.GetIntegrationWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to get integration: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get integration: status code %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, errors.New("unable to get integration: no data received")
	}
	
	// Convert the parameters directly to a map
	parametersData, err := json.Marshal(resp.JSON200.Parameters.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal integration parameters: %w", err)
	}
	
	var parameters map[string]interface{}
	if err := json.Unmarshal(parametersData, &parameters); err != nil {
		return nil, fmt.Errorf("unable to parse integration parameters: %w", err)
	}
	
	return &IntegrationWithParameters{
		Id: resp.JSON200.Id.String(),
		Name: resp.JSON200.Name,
		CreatedAt: resp.JSON200.CreatedAt,
		UpdatedAt: resp.JSON200.UpdatedAt,
		Parameters: parameters,
		LastValidationAt: *resp.JSON200.LastValidationAt,
	}, nil
}

type IntegrationInput struct {
	Name       string                 `json:"name"`
	Parameters struct {
		Data struct {
			Kind       string                 `json:"kind"`
			Parameters map[string]interface{} `json:"parameters"`
		} `json:"data"`
	} `json:"parameters"`
}

func CreateIntegration(ctx context.Context, integration IntegrationInput) (*IntegrationWithParameters, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	jsonData, err := json.Marshal(integration)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal integration: %w", err)
	}
	
	resp, err := client.CreateIntegrationWithBodyWithResponse(
		ctx, 
		"application/json",
		bytes.NewReader(jsonData),
	)
	log.Printf("DEBUG - Integration response: %s\n", string(resp.Body))

	if err != nil {		
		return nil, fmt.Errorf("unable to create integration: %w", err)
	}
		
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to create integration: status code %d", resp.StatusCode())
	}
	
	var integrationResp IntegrationWithParameters
	
	if err := json.Unmarshal(resp.Body, &integrationResp); err != nil {
		return nil, fmt.Errorf("unable to unmarshal integration response: %w", err)
	}

	paramBytes, err := json.Marshal(integration.Parameters.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal parameters: %w", err)
	}
	
	var paramMap map[string]interface{}
	if err := json.Unmarshal(paramBytes, &paramMap); err != nil {
		return nil, fmt.Errorf("unable to parse parameters: %w", err)
	}

	fmt.Printf("DEBUG - Integration response: %+v\n", integrationResp)
	
	integrationResp.Parameters = paramMap
	
	return &integrationResp, nil
}
