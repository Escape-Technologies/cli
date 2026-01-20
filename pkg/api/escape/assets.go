package escape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListAssetsFilters holds optional filters for listing assets
type ListAssetsFilters struct {
	AssetTypes      []string
	AssetStatuses   []string
	Search          string
	ManuallyCreated bool
}

// ListAssets lists all assets
func ListAssets(ctx context.Context, next string, filters *ListAssetsFilters) ([]v3.AssetSummarized, *string, error) {
	client, err := NewAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	rSize := 100
	req := client.AssetsAPI.ListAssets(ctx).Size(rSize)
	if next != "" {
		req = req.Cursor(next)
	}
	if filters != nil {
		if len(filters.AssetTypes) > 0 {
			req = req.Types(strings.Join(filters.AssetTypes, ","))
		}
		if len(filters.AssetStatuses) > 0 {
			req = req.Statuses(strings.Join(filters.AssetStatuses, ","))
		}
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}
		if filters.ManuallyCreated {
			req = req.ManuallyCreated(strconv.FormatBool(filters.ManuallyCreated))
		}
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", err)
	}
	return data.Data, data.NextCursor, nil
}

// GetAsset gets an asset by ID
func GetAsset(ctx context.Context, id string) (*v3.AssetDetailed, error) {
	client, err := NewAPIV3Client()
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
	client, err := NewAPIV3Client()
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
	client, err := NewAPIV3Client()
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

// normalizeAssetType normalizes asset type tokens to match generated method names
// e.g. KUBERNETES_CLUSTER -> KUBERNETESCLUSTER, http-endpoint -> httpendpoint
func normalizeAssetType(s string) string {
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

// CreateAsset creates an asset
func CreateAsset(ctx context.Context, data []byte, assetType string) (interface{}, error) {
	typ := reflect.TypeOf((*v3.AssetsAPIService)(nil))
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if strings.HasPrefix(method.Name, "Create") && !strings.HasSuffix(method.Name, "Execute") {
			if strings.Contains(strings.ToUpper(method.Name), strings.ToUpper(normalizeAssetType(assetType))) {
				client, err := NewAPIV3Client()
				if err != nil {
					return nil, fmt.Errorf("unable to init client: %w", err)
				}
				// Create request object like ApiCreateAssetDNSRequest
				req := method.Func.Call([]reflect.Value{reflect.ValueOf(client.AssetsAPI), reflect.ValueOf(ctx)})[0]

				// Find and call the typed body setter, e.g. CreateAssetDNSRequest(payload)
				typedBodyName := method.Name + "Request"
				setter := req.MethodByName(typedBodyName)
				if !setter.IsValid() {
					return nil, fmt.Errorf("failed to find body setter %s", typedBodyName)
				}
				// Build the typed payload from raw JSON
				if setter.Type().NumIn() != 1 {
					return nil, errors.New("unexpected setter signature")
				}

				// unmarshal raw JSON to typed payload
				payloadType := setter.Type().In(0)
				payloadPtr := reflect.New(payloadType)

				err = json.Unmarshal(data, payloadPtr.Interface())
				if err != nil {
					return nil, fmt.Errorf("invalid JSON for %s: %w", payloadType.Name(), err)
				}
				// attach body to request
				reqWithBody := setter.Call([]reflect.Value{payloadPtr.Elem()})[0]

				// execute the request
				executeMethod := reqWithBody.MethodByName("Execute")
				if !executeMethod.IsValid() {
					return nil, errors.New("failed to find Execute method")
				}
				results := executeMethod.Call(nil)

				return results[0].Interface(), nil
			}
		}
	}
	return nil, fmt.Errorf("asset type %s not found", assetType)
}
