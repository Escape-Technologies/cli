package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func NewAPIClient() (*Client, error) {
	log.Trace("Initializing client")
	server := os.Getenv("ESCAPE_API_URL")
	if server == "" {
		server = "https://public.escape.tech/"
	}
	key := os.Getenv("ESCAPE_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ESCAPE_API_KEY is not set")
	}
	return NewClient(
		server,
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			log.Debug("Sending request %s %s", req.Method, req.URL)
			if req.Body != nil {
				clone := req.Clone(context.Background())
				body, err := io.ReadAll(clone.Body)
				if err != nil {
					return err
				}
				log.Trace("Body %s", string(body))
			}
			req.Header.Set("Authorization", fmt.Sprintf("Key %s", key))
			return nil
		}),
	)
}
