package escape

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// ListProfilesFilters holds optional filters for listing profiles
type ListProfilesFilters struct {
	AssetIDs          []string
	AssetSchemaIDs    string
	AssetTypes        []string
	AssetStatuses     []string
	Domains           []string
	IDs               string
	IssueIDs          []string
	TagsIDs           []string
	ProjectIDs        []string
	ScanIDs           string
	LastScanStatuses  []string
	Search            string
	Initiators        []string
	Kinds             []string
	Risks             []string
	ProblemCodes      []string
	ProblemSeverities []string
	NoProjects        string
	NoTags            string
	Dnf               string
	SortType          string
	SortDirection     string
}

// ListProfiles lists one page of profiles.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListProfiles(ctx context.Context, next string, filters *ListProfilesFilters, size int) ([]v3.ProfileSummarized, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ProfilesAPI.ListProfiles(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
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

		if filters.AssetSchemaIDs != "" {
			req = req.AssetSchemaIds(v3.ListProfilesAssetSchemaIdsParameter{String: &filters.AssetSchemaIDs})
		}

		if len(filters.AssetTypes) > 0 {
			req = req.AssetTypes(filters.AssetTypes)
		}

		if len(filters.AssetStatuses) > 0 {
			req = req.AssetStatuses(filters.AssetStatuses)
		}

		if len(filters.Domains) > 0 {
			req = req.Domains(strings.Join(filters.Domains, ","))
		}

		if filters.IDs != "" {
			req = req.Ids(filters.IDs)
		}

		if len(filters.IssueIDs) > 0 {
			req = req.IssueIds(strings.Join(filters.IssueIDs, ","))
		}

		if len(filters.TagsIDs) > 0 {
			req = req.TagIds(strings.Join(filters.TagsIDs, ","))
		}

		if len(filters.ProjectIDs) > 0 {
			ids := filters.ProjectIDs
			req = req.ProjectIds(v3.ListProfilesProjectIdsParameter{ArrayOfString: &ids})
		}

		if filters.ScanIDs != "" {
			req = req.ScanIds(filters.ScanIDs)
		}

		if len(filters.LastScanStatuses) > 0 {
			req = req.LastScanStatuses(filters.LastScanStatuses)
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

		if len(filters.ProblemCodes) > 0 {
			req = req.ProblemCodes(filters.ProblemCodes)
		}

		if len(filters.ProblemSeverities) > 0 {
			req = req.ProblemSeverities(filters.ProblemSeverities)
		}

		if filters.NoProjects != "" {
			req = req.NoProjects(filters.NoProjects)
		}

		if filters.NoTags != "" {
			req = req.NoTags(filters.NoTags)
		}

		if filters.Dnf != "" {
			req = req.Dnf(filters.Dnf)
		}
	}

	data, _, err := req.Execute()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
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

// CreateProfileAiPentest creates an AI pentest profile (POST /profiles/ai-pentest).
func CreateProfileAiPentest(ctx context.Context, data []byte) (interface{}, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.CreateAiPentestProfileRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	profile, _, err := client.ProfilesAPI.CreateAiPentestProfile(ctx).CreateAiPentestProfileRequest(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return profile, nil
}

// CreateContinuousPentest sets up continuous pentesting on an AI pentest profile.
func CreateContinuousPentest(ctx context.Context, profileID string, data []byte) (*v3.UpdateContinuousPentest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.ContinuousPentestCreateInput
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for ContinuousPentestCreateInput: %w", err)
	}

	result, _, err := client.ProfilesAPI.CreateContinuousPentest(ctx, profileID).ContinuousPentestCreateInput(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// UpdateContinuousPentest edits the repositories and cadence of an enabled continuous pentest.
func UpdateContinuousPentest(ctx context.Context, profileID string, data []byte) (*v3.UpdateContinuousPentest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.ContinuousPentestUpdateInput
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for ContinuousPentestUpdateInput: %w", err)
	}

	result, _, err := client.ProfilesAPI.UpdateContinuousPentest(ctx, profileID).ContinuousPentestUpdateInput(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// EnableContinuousPentest turns the timer back on for a continuous pentest.
func EnableContinuousPentest(ctx context.Context, profileID string, data []byte) (*v3.UpdateContinuousPentest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	var payload v3.ContinuousPentestEnableInput
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON for ContinuousPentestEnableInput: %w", err)
	}

	result, _, err := client.ProfilesAPI.EnableContinuousPentest(ctx, profileID).ContinuousPentestEnableInput(payload).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
}

// DisableContinuousPentest stops the timer of a continuous pentest and keeps the repositories.
func DisableContinuousPentest(ctx context.Context, profileID string) (*v3.UpdateContinuousPentest200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	result, _, err := client.ProfilesAPI.DisableContinuousPentest(ctx, profileID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return result, nil
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
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return asset, nil
}

// DownloadSignedURL streams the body of a pre-signed URL to the writer.
// Used to download schema bytes after `GetProfile` returns a signedUrl.
// The caller is the sole authority on deadlines: pass a ctx with the desired
// deadline via context.WithTimeout / context.WithDeadline. We deliberately
// don't set http.Client.Timeout here — stacking two independent timeout
// mechanisms (ctx deadline + client timeout) produces confusing failure
// modes where the shorter one wins silently.
func DownloadSignedURL(ctx context.Context, signedURL string, dst io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

// DeleteProfile deletes a profile by ID
func DeleteProfile(ctx context.Context, profileID string) (*v3.DeleteProfile200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	data, _, err := client.ProfilesAPI.DeleteProfile(ctx, profileID).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
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

// ListProblems lists one page of profile scan problems.
// size 0 keeps the API default page size. The returned total is totalCount.
func ListProblems(ctx context.Context, next string, filters *ListProblemsFilters, size int) ([]v3.ProfileScanProblemsRow, *string, int, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.ProfilesAPI.Problems(ctx)
	if next != "" {
		req = req.Cursor(next)
	}

	if size > 0 {
		req = req.Size(size)
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
		return nil, nil, 0, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data.Data, data.NextCursor, data.GetTotalCount(), nil
}
