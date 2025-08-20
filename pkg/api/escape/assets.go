package escape

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

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
		req = req.Cursor(next)
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

// CreateAsset creates an asset
func CreateAsset(ctx context.Context, data []byte, assetType string, body interface{}) (interface{}, error) {
	typ := reflect.TypeOf((*v3.AssetsAPIService)(nil))
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if strings.HasPrefix(method.Name, "Create") && !strings.HasSuffix(method.Name, "Execute") {
			if strings.Contains(method.Name, strings.ToUpper(assetType)) {
				fmt.Println(method.Name)
				client, err := newAPIV3Client()
				if err != nil {
					return nil, fmt.Errorf("unable to init client: %w", err)
				}
				// Create request object
				req := method.Func.Call([]reflect.Value{reflect.ValueOf(client.AssetsAPI), reflect.ValueOf(ctx)})[0]

				executeMethod := req.MethodByName("Execute")
				if !executeMethod.IsValid() {
					return nil, errors.New("failed to find Execute method")
				}

				fmt.Println("Executed method: ", method.Name)

				results := executeMethod.Call(nil)
				fmt.Println(results[1].Interface())

				// Return the response (first return value)
				return results[0].Interface(), nil
			}
		}
	}
	return nil, fmt.Errorf("asset type %s not found", assetType)
}
