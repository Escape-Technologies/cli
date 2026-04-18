package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	shutdownTimeout    = 10 * time.Second
	healthCheckPortEnv = "HEALTH_CHECK_PORT"
)

// ServerOptions configures the embedded HTTP-based MCP server.
type ServerOptions struct {
	Version      string
	Port         int
	PublicAPIURL string
	Tools        []ToolSpec
	// IntentMode controls the tools/list interceptor. When empty the server
	// defaults to IntentModeCompactOnly, which serves compact stubs but never
	// calls a classifier. Set to IntentModeOn to enable classifier-driven
	// expansion (requires MCP_CLASSIFIER_* env vars). IntentModeOff disables
	// the middleware entirely.
	IntentMode IntentMode
	// Classifier is an optional ranker used when IntentMode == IntentModeOn.
	// If nil while IntentModeOn, the server behaves as IntentModeCompactOnly.
	Classifier Classifier
}

// Server is the embedded MCP server that exposes CLI commands over HTTP.
type Server struct {
	options ServerOptions
}

// NewServer builds a non-started embedded MCP server from the supplied options.
func NewServer(options ServerOptions) *Server {
	return &Server{options: options}
}

// Serve starts the embedded MCP server and blocks until the supplied context
// is cancelled. The shutdown path uses a detached context to give in-flight
// handlers a bounded drain window.
func (s *Server) Serve(ctx context.Context) error {
	rootServer := mcpserver.NewMCPServer(
		"Escape.tech-API-MCP",
		s.options.Version,
		mcpserver.WithToolCapabilities(false),
	)
	RegisterBuiltinTools(rootServer, s.options.Tools)
	if err := RegisterKnowledgeTools(rootServer, KnowledgeOptions{}); err != nil {
		return fmt.Errorf("register knowledge tools: %w", err)
	}
	RegisterCommandTools(rootServer, s.options.Tools, CommandExecutionOptions{
		PublicAPIURL: s.options.PublicAPIURL,
	})

	mcpHandler := mcpserver.NewStreamableHTTPServer(
		rootServer,
		mcpserver.WithStateLess(true),
		mcpserver.WithDisableStreaming(true),
		mcpserver.WithHTTPContextFunc(func(ctx context.Context, req *http.Request) context.Context {
			return InjectAuthContext(ctx, req)
		}),
	)

	healthHandler := func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	}

	mcpMode := s.options.IntentMode
	if mcpMode == "" {
		mcpMode = IntentModeCompactOnly
	}
	interceptedMCP := NewIntentMiddleware(mcpHandler, IntentOptions{
		Mode:       mcpMode,
		Classifier: s.options.Classifier,
		Specs:      s.options.Tools,
	})

	mainMux := http.NewServeMux()
	mainMux.Handle("/mcp", interceptedMCP)
	mainMux.HandleFunc("/health", healthHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.options.Port),
		Handler: mainMux,
	}

	var healthServer *http.Server
	healthPort := parseHealthCheckPort(s.options.Port)
	if healthPort != 0 {
		healthMux := http.NewServeMux()
		healthMux.HandleFunc("/health", healthHandler)
		healthServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", healthPort),
			Handler: healthMux,
		}
	}

	go func() {
		<-ctx.Done()
		// Detach from the just-cancelled parent context; graceful shutdown
		// needs its own bounded window to drain in-flight tool executions.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		_ = mcpHandler.Shutdown(shutdownCtx)
		_ = httpServer.Shutdown(shutdownCtx)
		if healthServer != nil {
			_ = healthServer.Shutdown(shutdownCtx)
		}
	}()
	if healthServer != nil {
		errCh := make(chan error, 1)
		go func() {
			err := healthServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				errCh <- nil
				return
			}
			errCh <- fmt.Errorf("health http server: %w", err)
		}()
		defer func() {
			if err := <-errCh; err != nil {
				// Health server failure shouldn't mask main server error; best-effort log via stderr.
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}()
	}
	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return fmt.Errorf("mcp http server: %w", err)
}

func parseHealthCheckPort(mainPort int) int {
	raw := os.Getenv(healthCheckPortEnv)
	if raw == "" {
		return 0
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port <= 0 || port > 65535 || port == mainPort {
		return 0
	}
	return port
}
