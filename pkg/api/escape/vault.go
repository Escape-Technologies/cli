package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// VariableContent is a vault variable as the public API returns it.
// The value is write-only and is never part of this struct.
type VariableContent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// variablePage is one page of GET /vault-variables.
type variablePage struct {
	Data       []VariableContent `json:"data"`
	NextCursor *string           `json:"nextCursor"`
	TotalCount int               `json:"totalCount"`
}

// ListVaultVariablesFilters holds the optional filters of the variable listing.
type ListVaultVariablesFilters struct {
	Search string
}

// ListVaultVariables lists one page of the organization's vault variables.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListVaultVariables(ctx context.Context, next string, size int, filters *ListVaultVariablesFilters) ([]VariableContent, *string, int, error) {
	query := url.Values{}
	if next != "" {
		query.Set("cursor", next)
	}

	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}

	if filters != nil && filters.Search != "" {
		query.Set("search", filters.Search)
	}

	path := rawPath("vault-variables")
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var page variablePage
	if err := rawRequest(ctx, http.MethodGet, path, nil, &page); err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return page.Data, page.NextCursor, page.TotalCount, nil
}

// CreatePlaintextVaultVariable stores a plaintext value under name.
func CreatePlaintextVaultVariable(ctx context.Context, name, value string) (*VariableContent, error) {
	return createStoredVariable(ctx, "plaintext", "PLAINTEXT", name, value)
}

// CreateSecretVaultVariable stores a masked secret under name.
func CreateSecretVaultVariable(ctx context.Context, name, value string) (*VariableContent, error) {
	return createStoredVariable(ctx, "secret", "SECRET", name, value)
}

func createStoredVariable(ctx context.Context, segment, variableType, name, value string) (*VariableContent, error) {
	body, err := json.Marshal(map[string]string{
		"type":  variableType,
		"name":  name,
		"value": value,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	var variable VariableContent
	if err := rawRequest(ctx, http.MethodPost, rawPath("vault-variables", segment), body, &variable); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &variable, nil
}

// CreateExternalVaultVariable stores a pointer at a provider entry.
// The API confirms the vault and the entry exist before it stores the variable.
func CreateExternalVaultVariable(ctx context.Context, name, externalID, externalEntry string) (*VariableContent, error) {
	body, err := json.Marshal(map[string]string{
		"type":          "EXTERNAL",
		"name":          name,
		"externalId":    externalID,
		"externalEntry": externalEntry,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	var variable VariableContent
	if err := rawRequest(ctx, http.MethodPost, rawPath("vault-variables", "external"), body, &variable); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &variable, nil
}

// UpdateVaultVariable replaces the value of a plaintext or secret variable,
// or the entry name of an external one. Exactly one of value and externalEntry
// is sent.
func UpdateVaultVariable(ctx context.Context, variableID, value, externalEntry string, external bool) (*VariableContent, error) {
	payload := map[string]string{}
	if external {
		payload["externalEntry"] = externalEntry
	} else {
		payload["value"] = value
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	var variable VariableContent
	if err := rawRequest(ctx, http.MethodPut, rawPath("vault-variables", variableID), body, &variable); err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &variable, nil
}

// DeleteVaultVariable deletes a vault variable by ID.
func DeleteVaultVariable(ctx context.Context, variableID string) error {
	if err := rawRequest(ctx, http.MethodDelete, rawPath("vault-variables", variableID), nil, &struct{}{}); err != nil {
		return fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return nil
}
