package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

const shutdownTimeout = 10 * time.Second

// ServerOptions configures the embedded HTTP-based MCP server.
type ServerOptions struct {
	Version      string
	Port         int
	PublicAPIURL string
	Tools        []ToolSpec
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

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpHandler)
	mux.HandleFunc("/health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.options.Port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		// Detach from the just-cancelled parent context; graceful shutdown
		// needs its own bounded window to drain in-flight tool executions.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		_ = mcpHandler.Shutdown(shutdownCtx)
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return fmt.Errorf("mcp http server: %w", err)
}
