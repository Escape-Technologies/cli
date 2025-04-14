package escape

import (
	"context"
	"fmt"
	"net/http"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

func CreateUploadSignedUrl(ctx context.Context) (*v2.CreateUploadSignedUrl200Response, error) {
	client, err := newAPIV2Client()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	data, resp, err := client.UploadAPI.CreateUploadSignedUrl(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("unable to get upload url: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get upload url: status code %d", resp.StatusCode)
	}
	return data, nil
}
