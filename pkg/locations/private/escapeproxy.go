package private

import (
	"os"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func SetupProxyURL() string {
	proxyURL := os.Getenv("ESCAPE_REPEATER_PROXY_URL")
	if proxyURL != "" {
		log.Info("Using custom proxy url: %s", proxyURL)
	}
	return proxyURL
}