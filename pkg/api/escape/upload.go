package escape

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// GetUploadSignedURL gets a signed url
func GetUploadSignedURL(ctx context.Context) (*v3.UploadSummary, error) {
	client, err := NewAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	req := client.UploadAPI.CreateUploadSignedUrl(ctx)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", err)
	}
	return data, nil
}

// UploadSchema uploads a file to the signed url
func UploadSchema(ctx context.Context, url string, data []byte) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to upload schema: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to upload schema: %s", resp.Status)
	}

	return nil
}


