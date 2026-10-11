package escape

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func schemaUploadContentType(data []byte) string {
	if json.Valid(data) {
		return "application/json"
	}

	return "application/yaml"
}

// GetUploadSignedURL gets a signed url
func GetUploadSignedURL(ctx context.Context) (*v3.CreateUploadSignedUrl200Response, error) {
	client, err := newAPIV3Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}

	req := client.UploadAPI.CreateUploadSignedUrl(ctx)
	data, _, err := req.Execute()
	if err != nil {
		return nil, fmt.Errorf("api error: %w", humanizeAPIError(err))
	}

	return data, nil
}

// UploadFileContent PUTs file bytes at a signed upload URL. The content type
// follows the file name: the knowledge library keeps documents, while the
// schema upload only ever sends JSON or YAML.
func UploadFileContent(ctx context.Context, url, name string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", fileContentType(name))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to upload file: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to upload file: %s", resp.Status)
	}

	return nil
}

// fileContentType guesses the MIME type of a file name. An unknown extension
// falls back to the octet stream a signed URL accepts for any payload.
func fileContentType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".md", ".markdown":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"
	case ".pdf":
		return "application/pdf"
	case ".csv":
		return "text/csv"
	case ".txt", ".text":
		return "text/plain"
	}

	if guessed := mime.TypeByExtension(ext); guessed != "" {
		return guessed
	}

	return "application/octet-stream"
}

// UploadSchema uploads a schema file to the signed url
func UploadSchema(ctx context.Context, url string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", schemaUploadContentType(data))
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
