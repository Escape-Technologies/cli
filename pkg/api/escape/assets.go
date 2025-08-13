package escape

import (
	"context"
	"fmt"
	"io"
	"net/http"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListAssets lists all assets
func ListAssets(ctx context.Context, next string, assetTypes []string, assetStatuses []string) ([]v3.AssetSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	rSize := 100
	req := client.AssetsAPI.ListAssets(ctx).Size(rSize)
	if next != "" {
		req.Cursor(next)
	}
	if len(assetTypes) > 0 {
		req = req.Types(assetTypes)
	}
	if len(assetStatuses) > 0 {
		req = req.Statuses(assetStatuses)
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetAsset gets an asset by ID
func GetAsset(ctx context.Context, id string) (*v3.AssetDetailed, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.AssetsAPI.GetAsset(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// DeleteAsset deletes an asset by ID
func DeleteAsset(ctx context.Context, id string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	_, httpRes, err := client.AssetsAPI.DeleteAsset(ctx, id).Execute()
	if err != nil && httpRes.StatusCode != http.StatusOK {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}

// UpdateAsset updates an asset by ID
func UpdateAsset(
	ctx context.Context,
	id string,
	assetDescription *string,
	assetFramework *v3.ENUMPROPERTIESFRAMEWORK,
	assetOwners *[]string,
	assetStatus *v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUS,
	assetTagIDs *[]string,
) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}

	updateAssetRequest := v3.UpdateAssetRequest{}

	if assetDescription != nil {
		updateAssetRequest.Description = assetDescription
	}

	if assetFramework != nil {
		updateAssetRequest.Framework = assetFramework
	}

	if assetOwners != nil && len(*assetOwners) > 0 {
		updateAssetRequest.Owners = &v3.UpdateAssetRequestOwners{
			ArrayOfString: assetOwners,
		}
	}

	if assetStatus != nil {
		updateAssetRequest.Status = assetStatus
	}

	if assetTagIDs != nil && len(*assetTagIDs) > 0 {
		updateAssetRequest.TagIds = &v3.UpdateAssetRequestTagIds{
			ArrayOfString: assetTagIDs,
		}
	}

	_, apiRes, err := client.AssetsAPI.UpdateAsset(ctx, id).UpdateAssetRequest(updateAssetRequest).Execute()
	if err != nil {
		if apiRes.StatusCode == http.StatusBadRequest {
			body, _ := io.ReadAll(apiRes.Body)
			return fmt.Errorf("api error: %s", string(body))
		}
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}
