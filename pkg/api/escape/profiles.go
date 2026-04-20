package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListProfilesFilters holds optional filters for listing profiles
type ListProfilesFilters struct {
	AssetIDs      []string
	Domains       []string
	IssueIDs      []string
	TagsIDs       []string
	Search        string
	Initiators    []string
	Kinds         []string
	Risks         []string
	SortType      string
	SortDirection string
}

// ListProfiles lists all profiles
func ListProfiles(ctx context.Context, next string, filters *ListProfilesFilters) ([]v3.ProfileSummarized, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ProfilesAPI.ListProfiles(ctx)
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
		if len(filters.AssetIDs) > 0 {
			req = req.AssetIds(strings.Join(filters.AssetIDs, ","))
		}
		if len(filters.Domains) > 0 {
			req = req.Domains(strings.Join(filters.Domains, ","))
		}
		if len(filters.IssueIDs) > 0 {
			req = req.IssueIds(strings.Join(filters.IssueIDs, ","))
		}
		if len(filters.TagsIDs) > 0 {
			req = req.TagIds(strings.Join(filters.TagsIDs, ","))
		}
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}
		if len(filters.Initiators) > 0 {
			req = req.Initiators((filters.Initiators))
		}
		if len(filters.Kinds) > 0 {
			req = req.Kinds((filters.Kinds))
		}
		if len(filters.Risks) > 0 {
			req = req.Risks((filters.Risks))
		}
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return data.Data, data.NextCursor, nil
}

// GetProfile gets a profile by ID
func GetProfile(ctx context.Context, profileID string) (*v3.GetProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ProfilesAPI.GetProfile(ctx, profileID)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return data, nil
}

// CreateProfileRest creates a profile for a REST application
func CreateProfileRest(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastRestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.ProfilesAPI.CreateDastRestProfile(ctx)
	profile, _, err := req.CreateDastRestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// CreateProfileWebapp creates a profile for a web application
func CreateProfileWebapp(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastRestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.ProfilesAPI.CreateDastWebAppProfile(ctx)
	profile, _, err := req.CreateDastRestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// CreateProfileGraphql creates a profile for a GraphQL application
func CreateProfileGraphql(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastRestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	req := client.ProfilesAPI.CreateDastGraphqlProfile(ctx)
	profile, _, err := req.CreateDastRestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// UpdateProfile updates profile metadata (name, description, cron, extra assets)
func UpdateProfile(ctx context.Context, profileID string, data []byte) (*v3.GetProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.UpdateProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for UpdateProfileRequest: %w", err)
	}

	profile, _, err := client.ProfilesAPI.UpdateProfile(ctx, profileID).UpdateProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// UpdateProfileConfiguration updates profile configuration (auth, scope, frontend_dast, etc.)
func UpdateProfileConfiguration(ctx context.Context, profileID string, data []byte) (map[string]interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.UpdateProfileConfigurationRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for UpdateProfileConfigurationRequest: %w", err)
	}

	result, _, err := client.ProfilesAPI.UpdateProfileConfiguration(ctx, profileID).UpdateProfileConfigurationRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return result, nil
}

// CreateSchemaAsset turns a temporary upload object key into a SCHEMA asset.
// POST /v3/assets/schema polls the schema-build workflow internally; the
// server returns the finalized asset or a 400 on workflow timeout.
func CreateSchemaAsset(ctx context.Context, temporaryObjectKey string, name string) (*v3.UpdateAsset200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	upload := v3.CreateSchemaViaUpload{
		AssetType: v3.ENUMSCHEMA_SCHEMA,
		Upload: v3.CreateSchemaViaUploadUpload{
			TemporaryObjectKey: temporaryObjectKey,
		},
	}
	if name != "" {
		upload.Name = &name
	}

	asset, _, err := client.AssetsAPI.CreateAssetSchema(ctx).
		CreateAssetSchemaRequest(v3.CreateAssetSchemaRequest{CreateSchemaViaUpload: &upload}).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return asset, nil
}

// DownloadSignedURL streams the body of a pre-signed URL to the writer.
// Used to download schema bytes after `GetProfile` returns a signedUrl.
// The caller controls the timeout via ctx (typical schemas are small but the
// signed-URL fetch is the one leg the generated v3 client does not cover).
func DownloadSignedURL(ctx context.Context, signedURL string, dst io.Writer, timeout time.Duration) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to fetch signed url: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to fetch signed url: %s", resp.Status)
	}

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return fmt.Errorf("unable to write schema body: %w", err)
	}
	return nil
}

// UpdateProfileSchema updates the schema attached to a profile
func UpdateProfileSchema(ctx context.Context, profileID string, schemaID string) (*v3.GetProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	payload := v3.UpdateProfileSchemaRequest{
		SchemaId: schemaID,
	}

	profile, _, err := client.ProfilesAPI.UpdateProfileSchema(ctx, profileID).UpdateProfileSchemaRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// CreateProfilePentestRest creates an AI Pentest profile for a REST application
func CreateProfilePentestRest(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastRestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	profile, _, err := client.ProfilesAPI.CreatePentestRestProfile(ctx).CreateDastRestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// CreateProfilePentestGraphql creates an AI Pentest profile for a GraphQL application
func CreateProfilePentestGraphql(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastRestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	profile, _, err := client.ProfilesAPI.CreatePentestGraphqlProfile(ctx).CreateDastRestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// CreateProfilePentestWebapp creates an AI Pentest profile for a web application
func CreateProfilePentestWebapp(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateDastRestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	profile, _, err := client.ProfilesAPI.CreatePentestWebappProfile(ctx).CreateDastRestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return profile, nil
}

// DeleteProfile deletes a profile by ID
func DeleteProfile(ctx context.Context, profileID string) error {
	client, err := newAPIV3Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ProfilesAPI.DeleteProfile(ctx, profileID)
	_, _, err = req.Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return nil
}

// ListProblemsFilters holds optional filters for listing problems
type ListProblemsFilters struct {
	AssetIDs   []string
	Domains    []string
	IssueIDs   []string
	TagsIDs    []string
	Search     string
	Initiators []string
	Kinds      []string
	Risks      []string
}

// ListProblems lists all scan problems
func ListProblems(ctx context.Context, next string, filters *ListProblemsFilters) ([]v3.ProfileScanProblemsRow, *string, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.ProfilesAPI.Problems(ctx)
	if next != "" {
		req = req.Cursor(next)
	}
	if filters != nil {
		if len(filters.AssetIDs) > 0 {
			req = req.AssetIds(strings.Join(filters.AssetIDs, ","))
		}
		if len(filters.Domains) > 0 {
			req = req.Domains(strings.Join(filters.Domains, ","))
		}
		if len(filters.IssueIDs) > 0 {
			req = req.IssueIds(strings.Join(filters.IssueIDs, ","))
		}
		if len(filters.TagsIDs) > 0 {
			req = req.TagIds(strings.Join(filters.TagsIDs, ","))
		}
		if filters.Search != "" {
			req = req.Search(filters.Search)
		}
		if len(filters.Initiators) > 0 {
			req = req.Initiators((filters.Initiators))
		}
		if len(filters.Kinds) > 0 {
			req = req.Kinds((filters.Kinds))
		}
		if len(filters.Risks) > 0 {
			req = req.Risks((filters.Risks))
		}
	}
	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}
	return data.Data, data.NextCursor, nil
}
