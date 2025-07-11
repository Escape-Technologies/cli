// Package health implements the health check server
package health

import (
	"context"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func buildHandler(healthy *atomic.Bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		var msg string
		if healthy.Load() {
			w.WriteHeader(http.StatusOK)
			msg = "OK"
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			msg = "Not connected"
		}
		_, err := w.Write([]byte(msg))
		if err != nil {
			log.Debug("Error during health check: %v", err)
		}
	})

	if os.Getenv("ESCAPE_ENABLE_LOGS_ENDPOINT") == "true" {
		mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			log.Debug("Forwarded log: %s", string(bodyBytes))
		})
	}

	return mux
}
// Start the health check server
func Start(ctx context.Context, healthy *atomic.Bool) {
	if os.Getenv("HEALTH_CHECK_PORT") == "" {
		log.Trace("HEALTH_CHECK_PORT is not set, skipping health check")
		return
	}

	srv := &http.Server{
		Addr:    ":" + os.Getenv("HEALTH_CHECK_PORT"),
		Handler: buildHandler(healthy),
	}
	go func() {
		<-ctx.Done()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Error("Failed to shutdown health check server: %v", err)
		}
	}()
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Error("Failed to start health check server: %v", err)
		}
	}()
	log.Debug("Health check server started on http://0.0.0.0:%s/health", os.Getenv("HEALTH_CHECK_PORT"))
	if os.Getenv("ESCAPE_ENABLE_LOGS_ENDPOINT") == "true" {
		log.Debug("Log endpoint available at http://0.0.0.0:%s/log", os.Getenv("HEALTH_CHECK_PORT"))
	}
}
