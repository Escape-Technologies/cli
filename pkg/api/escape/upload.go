package escape

import (
	"context"
	"fmt"
	"net/http"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

type UploadUrl struct {
	Url string `json:"url"`
}

func (u *UploadUrl) String() string {
	return u.Url
}

func GetUrl(ctx context.Context) (*UploadUrl, error) {
	client, err := v2.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("unable to init client: %w", err)
	}
	resp, err := client.CreateUploadSignedUrlWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get upload url: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unable to get upload url: status code %d", resp.StatusCode())
	}
	return &UploadUrl{
		Url: resp.JSON200.Url,
	}, nil
}
