package escape

import (
	"context"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListVaultVariablesFilters holds the optional filters of the variable listing.
type ListVaultVariablesFilters struct {
	Search string
}

// ListVaultVariables lists one page of the organization's vault variables.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListVaultVariables(ctx context.Context, next string, size int, filters *ListVaultVariablesFilters) ([]v3.VaultVariableSummary, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.VaultVariablesAPI.ListVaultVariables(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	if filters != nil && filters.Search != "" {
		req = req.Search(filters.Search)
	}

	data, _, listErr := req.Execute()
	if listErr != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(listErr))
	}

	nextCursor := data.GetNextCursor()

	return data.GetData(), &nextCursor, data.GetTotalCount(), nil
}

// CreateVaultVariable stores a secret reference. vaultID is required for an
// EXTERNAL variable and empty otherwise.
func CreateVaultVariable(ctx context.Context, key, variableType, content, vaultID string) (*v3.CreateVaultVariable200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := v3.NewCreateVaultVariableRequest(
		key,
		v3.ENUMPROPERTIESDATAITEMSPROPERTIESTYPE(variableType),
		content,
	)
	if vaultID != "" {
		req.SetVaultId(vaultID)
	}

	data, _, err := client.VaultVariablesAPI.CreateVaultVariable(ctx).CreateVaultVariableRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// DeleteVaultVariable deletes a vault variable by ID.
func DeleteVaultVariable(ctx context.Context, variableID string) (*v3.DeleteProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.VaultVariablesAPI.DeleteVaultVariable(ctx, variableID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// RotateVaultVariable replaces the value of a vault variable.
func RotateVaultVariable(ctx context.Context, variableID, content string) (*v3.CreateVaultVariable200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := v3.NewRotateVaultVariableRequest(content)
	data, _, err := client.VaultVariablesAPI.RotateVaultVariable(ctx, variableID).RotateVaultVariableRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ListVaults lists one page of the vaults reachable by the organization.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListVaults(ctx context.Context, next string, size int) ([]v3.VaultSummary1, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.VaultVariablesAPI.ListVaults(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
	}

	data, _, listErr := req.Execute()
	if listErr != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(listErr))
	}

	nextCursor := data.GetNextCursor()

	return data.GetData(), &nextCursor, data.GetTotalCount(), nil
}

// LinkVault links a provider vault to the organization.
func LinkVault(ctx context.Context, integrationID, externalID string) (*v3.LinkVault200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := v3.NewLinkVaultRequest(integrationID, externalID)
	data, _, err := client.VaultVariablesAPI.LinkVault(ctx).LinkVaultRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ListVaultEntries lists the entries of a vault, as the provider names them.
func ListVaultEntries(ctx context.Context, vaultID string) ([]v3.VaultEntry, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.VaultVariablesAPI.ListVaultEntries(ctx, vaultID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}
