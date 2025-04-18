package escape

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/Escape-Technologies/cli/pkg/env"
)

func createUploadSignedURL(ctx context.Context) (string, string, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return "", "", fmt.Errorf("unable to init client: %w", err)
	}
	data, _, err := client.UploadAPI.CreateUploadSignedUrl(ctx).Execute()
	if err != nil {
		return "", "", fmt.Errorf("unable to get upload url: %w", err)
	}
	return data.Url, data.Id, nil
}

func uploadToS3(ctx context.Context, data string) (string, error) {
	url, id, err := createUploadSignedURL(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get upload url: %w", err)
	}

	body := bytes.NewBuffer([]byte(data))
	req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}
	resp, err := env.GetHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to upload data: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	return id, nil
}
