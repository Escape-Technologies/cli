package escape

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/env"
)

// ListApplications lists all applications
func ListApplications(ctx context.Context) ([]v2.ListApplications200ResponseInner, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ApplicationsAPI.ListApplications(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	if data == nil {
		return nil, errors.New("no data received")
	}
	return data, nil
}

// GetApplication gets an application by ID
func GetApplication(ctx context.Context, id string) (*v2.GetApplication200Response, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.ApplicationsAPI.GetApplication(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	if data == nil {
		return nil, errors.New("no data received")
	}
	return data, nil
}

// UpdateApplicationSchema updates an application schema
func UpdateApplicationSchema(ctx context.Context, id string, schemaPathOrURL string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	blobID, url, err := schemaToS3(ctx, schemaPathOrURL)
	if err != nil {
		return fmt.Errorf("unable to upload schema: %w", err)
	}
	_, _, err = client.ApplicationsAPI.UpdateSchema(ctx, id).CreateApplicationRequestSchema(v2.CreateApplicationRequestSchema{
		Url:    url,
		BlobId: blobID,
	}).Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}

func schemaToS3(ctx context.Context, schemaPathOrURL string) (string, *string, error) {
	body, url, err := pullSchema(ctx, schemaPathOrURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid schema: %w", err)
	}
	blobID, err := uploadToS3(ctx, body)
	if err != nil {
		return "", nil, fmt.Errorf("unable to upload schema: %w", err)
	}
	return blobID, url, nil
}

func pullSchema(ctx context.Context, schemaPathOrURL string) (string, *string, error) {
	if strings.HasPrefix(schemaPathOrURL, "http") {
		body, err := pullSchemaFromURL(ctx, schemaPathOrURL)
		if err != nil {
			return "", nil, fmt.Errorf("unable to pull schema from URL %s: %w", schemaPathOrURL, err)
		}
		return body, &schemaPathOrURL, nil
	}
	body, err := pullSchemaFromPath(schemaPathOrURL)
	if err != nil {
		return "", nil, fmt.Errorf("unable to read file %s: %w", schemaPathOrURL, err)
	}
	return body, nil, nil
}

func pullSchemaFromPath(path string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	if !isJSONOrYAML(body) {
		return "", errors.New("file is neither json nor yaml")
	}
	return string(body), nil
}

func pullSchemaFromURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}
	resp, err := env.GetHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %w", fmt.Errorf("status code: %d", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read body: %w", err)
	}
	if !isJSONOrYAML(body) {
		return "", errors.New("file is neither json nor yaml")
	}
	return string(body), nil
}

// UpdateApplicationConfig updates an application configuration
func UpdateApplicationConfig(ctx context.Context, id string, configPath string) error {
	client, err := newAPIV2Client()
	if err != nil {
		return fmt.Errorf("unable to init client: %w", err)
	}
	cfg, err := readConfig(configPath)
	if err != nil {
		return fmt.Errorf("unable to read config at %s: %w", configPath, err)
	}
	_, _, err = client.ApplicationsAPI.UpdateConfiguration(ctx, id).CreateApplicationRequestConfiguration(*cfg).Execute()
	if err != nil {
		return fmt.Errorf("api error: %w", err)
	}
	return nil
}

func readConfig(path string) (*v2.CreateApplicationRequestConfiguration, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}
	cfg, err := parseJSONOrYAML(body, &v2.NullableCreateApplicationRequestConfiguration{})
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}
	return cfg.Get(), nil
}
