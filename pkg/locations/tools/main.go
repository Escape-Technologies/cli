// Package tools exposes location-local helper tools over a loopback-only HTTP API.
package tools

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

const (
	// Port is the loopback-only private-location tools API port.
	Port              = "53216"
	address           = "127.0.0.1:" + Port
	maxJobs           = 7
	maxPendingJobs    = 100
	jobTTL            = 10 * time.Second
	retryDelay        = 5 * time.Second
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	shutdownWait      = 5 * time.Second
)

// Start runs the private-location tools API on a loopback-only listener.
func Start(ctx context.Context) {
	for ctx.Err() == nil {
		serverCtx, cancel := context.WithCancel(ctx)
		srv := &http.Server{
			Addr:              address,
			Handler:           buildHandler(serverCtx, maxJobs, maxPendingJobs, jobTTL),
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			<-serverCtx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownWait)
			defer shutdownCancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error("Failed to shutdown tools server: %v", err)
			}
		}()

		log.Debug("Tools server starting on http://%s/api/tools", address)
		err := srv.ListenAndServe()
		cancel()
		if errors.Is(err, http.ErrServerClosed) || ctx.Err() != nil {
			<-done
			return
		}

		log.Error("Failed to start tools server: %v", err)
		select {
		case <-time.After(retryDelay):
		case <-ctx.Done():
			<-done
			return
		}
	}
}
