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

// ListAssetsFilters holds optional filters for listing assets. The slice
// fields are sent as comma-separated query values, matching the public API.
type ListAssetsFilters struct {
	AssetTypes      []string
	AssetStatuses   []string
	ProjectIDs      []string
	TagIDs          []string
	Classes         []string
	Environments    []string
	Domains         []string
	IntegrationIDs  []string
	OwnerEmails     []string
	TechnologyKeys  []string
	Ports           []string
	Severities      []string
	Risks           []string
	Frameworks      []string
	DNF             string
	Search          string
	ManuallyCreated bool
	SortType        string
	SortDirection   string
}

// ListAssets lists one page of assets.
// size 0 keeps the historical page size of 100. The returned total is totalCount.
func ListAssets(ctx context.Context, next string, filters *ListAssetsFilters, size int) ([]v3.AssetSummarized, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	rSize := 100
	if size > 0 {
		rSize = size
	}

	req := client.AssetsAPI.ListAssets(ctx).Size(rSize)
	if next != "" {
		req = req.Cursor(next)
	}

	if filters != nil {
		if filters.SortType != "" {
			req = req.SortType(filters.SortType)
		}

		if filters.SortDirection != "" {
			req = req.SortDirection(filters.SortDirection)
		}

		if len(filters.AssetTypes) > 0 {
			req = req.Types(filters.AssetTypes)
		}

		if len(filters.AssetStatuses) > 0 {
			req = req.Statuses(filters.AssetStatuses)
		}

		if len(filters.ProjectIDs) > 0 {
			req = req.ProjectIds(strings.Join(filters.ProjectIDs, ","))
		}

		if len(filters.TagIDs) > 0 {
			req = req.TagIds(strings.Join(filters.TagIDs, ","))
		}

		if len(filters.Classes) > 0 {
			req = req.Classes(filters.Classes)
		}

		if len(filters.Environments) > 0 {
			req = req.Environments(filters.Environments)
		}

		if len(filters.Domains) > 0 {
			req = req.Domains(strings.Join(filters.Domains, ","))
		}

		if len(filters.IntegrationIDs) > 0 {
			req = req.IntegrationIds(strings.Join(filters.IntegrationIDs, ","))
		}

		if len(filters.OwnerEmails) > 0 {
			req = req.OwnerEmails(strings.Join(filters.OwnerEmails, ","))
		}

		if len(filters.TechnologyKeys) > 0 {
			req = req.TechnologyKeys(strings.Join(filters.TechnologyKeys, ","))
		}

		if len(filters.Ports) > 0 {
			req = req.Ports(strings.Join(filters.Ports, ","))
		}

		if len(filters.Severities) > 0 {
			req = req.Severities(filters.Severities)
		}

		if len(filters.Risks) > 0 {
			req = req.Risks(filters.Risks)
		}

		if len(filters.Frameworks) > 0 {
			req = req.Frameworks(filters.Frameworks)
		}

		if filters.DNF != "" {
			req = req.Dnf(filters.DNF)
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
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}

// GetAsset gets an asset by ID
func GetAsset(ctx context.Context, id string) (*v3.AssetDetailed1, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.AssetsAPI.GetAsset(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// GetAssetContent gets the download metadata (signed URL, content type, filename) of an asset content
func GetAssetContent(ctx context.Context, id string) (*v3.GetAssetContent200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.AssetsAPI.GetAssetContent(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// DeleteAsset deletes an asset by ID and returns the API's acknowledgement.
func DeleteAsset(ctx context.Context, id string) (*v3.DeleteProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, httpRes, err := client.AssetsAPI.DeleteAsset(ctx, id).Execute()
	if err != nil {
		if ack, ok, ackErr := deleteAssetAcknowledgement(httpRes, err); ok {
			return ack, ackErr
		}

		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// deleteAssetAcknowledgement handles the body DELETE /v3/assets/{id} really
// sends. The route declares {"message": string} but returns the GraphQL
// mutation result, a bare JSON boolean, so the generated client fails to
// decode a successful response. ok is false when err is not that decode
// failure on a 2xx response.
func deleteAssetAcknowledgement(httpRes *http.Response, err error) (*v3.DeleteProfile200Response, bool, error) {
	if httpRes == nil || httpRes.StatusCode < http.StatusOK || httpRes.StatusCode >= http.StatusMultipleChoices {
		return nil, false, nil
	}

	var apiErr *v3.GenericOpenAPIError
	if !errors.As(err, &apiErr) || apiErr == nil {
		return nil, false, nil
	}

	var deleted bool
	if jsonErr := json.Unmarshal(apiErr.Body(), &deleted); jsonErr != nil {
		return nil, false, nil
	}

	if !deleted {
		return nil, true, errors.New("api error: the API reported that the asset was not deleted")
	}

	return v3.NewDeleteProfile200Response("Asset deleted successfully"), true, nil
}

// UpdateAssetInput is the optional body of PUT /v3/assets/{id}. Nil fields and
// empty collections are omitted, so the API only touches what the caller set.
type UpdateAssetInput struct {
	Description       *string
	Framework         *v3.ENUMPROPERTIESFRAMEWORK
	Owners            *[]string
	Status            *v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS
	TagIDs            *[]string
	ProjectIDs        *[]string
	Name              *string
	Environment       *v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSERVICEPROPERTIESENVIRONMENT
	ManuallySetStatus *bool
	BaseURL           *string
	Address           *string
	BestSchemaID      *string
	SchemaIDs         []string
	ParentLinks       *v3.UpdateAssetParentLinks
	ChildLinks        *v3.UpdateAssetChildLinks
}

// UpdateAsset updates an asset by ID
func UpdateAsset(ctx context.Context, id string, input UpdateAssetInput) (*v3.UpdateAsset200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	updateAssetRequest := v3.UpdateAsset{
		Description:       input.Description,
		Framework:         input.Framework,
		Status:            input.Status,
		Name:              input.Name,
		Environment:       input.Environment,
		ManuallySetStatus: input.ManuallySetStatus,
		BaseUrl:           input.BaseURL,
		Address:           input.Address,
		BestSchemaId:      input.BestSchemaID,
		ParentLinks:       input.ParentLinks,
		ChildLinks:        input.ChildLinks,
	}

	if input.Owners != nil && len(*input.Owners) > 0 {
		updateAssetRequest.Owners = &v3.UpdateAssetOwners{
			ArrayOfString: input.Owners,
		}
	}

	if input.TagIDs != nil && len(*input.TagIDs) > 0 {
		updateAssetRequest.TagIds = &v3.UpdateAssetTagIds{
			ArrayOfString: input.TagIDs,
		}
	}

	if input.ProjectIDs != nil && len(*input.ProjectIDs) > 0 {
		updateAssetRequest.ProjectIds = *input.ProjectIDs
	}

	if len(input.SchemaIDs) > 0 {
		updateAssetRequest.SchemaIds = input.SchemaIDs
	}

	data, apiRes, err := client.AssetsAPI.UpdateAsset(ctx, id).UpdateAsset(updateAssetRequest).Execute()
	if err != nil {
		if apiRes.StatusCode == http.StatusBadRequest {
			body, _ := io.ReadAll(apiRes.Body)
			return nil, fmt.Errorf("api error: %s", string(body))
		}

		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// BulkOperationResult is the {"success": bool} body returned by the bulk asset
// endpoints. The generated client names the bulk response model after whichever
// operation the spec lists first (BulkUpdateAssets200Response,
// BulkUpdateProfiles200Response, ...), so the CLI decodes into this stable local
// type instead of the generated one and keeps compiling across regenerations.
type BulkOperationResult struct {
	Success bool `json:"success"`
}

// BulkImportAssets starts an asynchronous job that creates multiple assets. The
// raw JSON body is the BulkImportAssets request, so callers can pass the stdin
// payload straight through.
func BulkImportAssets(ctx context.Context, data []byte) (*v3.BulkImportAssets200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var body v3.BulkImportAssets
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	result, _, err := client.AssetsAPI.BulkImportAssets(ctx).BulkImportAssets(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// BulkUpdateAssets updates multiple assets matching a filter
func BulkUpdateAssets(ctx context.Context, where v3.BulkUpdateAssetsRequestWhere, tagIDs, projectIDs []string, status *v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS) (*BulkOperationResult, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	body := v3.BulkUpdateAssetsRequest{Where: where}
	if len(tagIDs) > 0 {
		body.TagIds = tagIDs
	}

	if len(projectIDs) > 0 {
		body.ProjectIds = projectIDs
	}

	if status != nil {
		body.Status = status
	}

	data, _, err := client.AssetsAPI.BulkUpdateAssets(ctx).BulkUpdateAssetsRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &BulkOperationResult{Success: data.GetSuccess()}, nil
}

// BulkDeleteAssets deletes multiple assets matching a filter
func BulkDeleteAssets(ctx context.Context, where v3.BulkUpdateAssetsRequestWhere) (*BulkOperationResult, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	body := v3.BulkDeleteAssetsRequest{Where: where}
	data, _, err := client.AssetsAPI.BulkDeleteAssets(ctx).BulkDeleteAssetsRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return &BulkOperationResult{Success: data.GetSuccess()}, nil
}

// CommentAsset adds a comment to an asset
func CommentAsset(ctx context.Context, assetID, comment string) (*v3.CreateAssetComment200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	body := v3.CreateAssetCommentRequest{Comment: comment}
	data, _, err := client.AssetsAPI.CreateAssetComment(ctx, assetID).CreateAssetCommentRequest(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// ListAssetActivities lists activities for an asset
func ListAssetActivities(ctx context.Context, assetID string) ([]v3.ActivitySummarized, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.AssetsAPI.ListAssetActivities(ctx, assetID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// normalizeAssetType normalizes asset type tokens to match generated method names
// e.g. GITHUB_REPOSITORY -> GITHUBREPOSITORY, http-endpoint -> httpendpoint
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
				client, err := newAPIV3Client()
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

				return unpackExecute(executeMethod.Call(nil))
			}
		}
	}

	return nil, fmt.Errorf("asset type %s not found", assetType)
}

// unpackExecute reads the (body, ..., error) values returned by generated Execute methods.
func unpackExecute(results []reflect.Value) (interface{}, error) {
	if len(results) == 0 {
		return nil, errors.New("unexpected Execute signature")
	}

	last := results[len(results)-1]
	if last.IsValid() && last.Kind() == reflect.Interface && !last.IsNil() {
		if err, ok := last.Interface().(error); ok {
			return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
		}
	}

	value := results[0].Interface()
	if isNilResponse(value) {
		return nil, errors.New("api error: empty response")
	}

	return value, nil
}

func isNilResponse(value interface{}) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	nillable := rv.Kind() == reflect.Pointer ||
		rv.Kind() == reflect.Interface ||
		rv.Kind() == reflect.Map ||
		rv.Kind() == reflect.Slice

	return nillable && rv.IsNil()
}
