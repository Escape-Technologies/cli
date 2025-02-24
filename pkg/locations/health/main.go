package health

import (
	"context"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Escape-Technologies/cli/pkg/log"
)

func buildHandler(healthy *atomic.Bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func Start(ctx context.Context, healthy *atomic.Bool) {
	if os.Getenv("HEALTH_CHECK_PORT") == "" {
		log.Trace("HEALTH_CHECK_PORT is not set, skipping health check")
		return
	}

	srv := &http.Server{
		Addr:    ":" + os.Getenv("HEALTH_CHECK_PORT"),
		Handler: http.HandlerFunc(buildHandler(healthy)),
	}
	go func() {
		<-ctx.Done()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Error("Error shutting down health check server: %v", err)
		}
	}()
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Error("Error starting the health check server: %v", err)
		}
	}()
	log.Info("Health check server started on http://0.0.0.0:%s/health", os.Getenv("HEALTH_CHECK_PORT"))
}
