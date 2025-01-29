package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func initClient() (*api.Client, error) {
	log.Trace("Initializing client")
	server := os.Getenv("ESCAPE_API_URL")
	if server == "" {
		server = "https://public.escape.tech/"
	}
	key := os.Getenv("ESCAPE_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ESCAPE_API_KEY is not set")
	}
	return api.NewClient(
		server,
		api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			log.Trace("Sending request %s %s", req.Method, req.URL)
			req.Header.Set("Authorization", fmt.Sprintf("Key %s", key))
			return nil
		}),
	)
}
