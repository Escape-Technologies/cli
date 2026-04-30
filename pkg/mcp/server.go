package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	// OAuth options. When IssuerURL and ResourceURL are both empty the
	// server skips OAuth wiring and keeps the legacy X-ESCAPE-API-KEY
	// behavior (useful for locally-run CLI MCP). In prod, both are set.
	IssuerURL           string
	ResourceURL         string
	OAuthPrivateKeyPath string
	ExtraRedirectHosts  []string
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
	if err := RegisterPublicAPITools(rootServer, PublicAPIOptions{
		PublicAPIURL: s.options.PublicAPIURL,
	}); err != nil {
		return fmt.Errorf("register public api tools: %w", err)
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

	var oauth *oauthHandlers
	if s.options.IssuerURL != "" || s.options.ResourceURL != "" {
		handlers, err := newOAuthHandlers(oauthConfig{
			IssuerURL:           s.options.IssuerURL,
			ResourceURL:         s.options.ResourceURL,
			PublicAPIURL:        s.options.PublicAPIURL,
			OAuthPrivateKeyPath: s.options.OAuthPrivateKeyPath,
			ExtraRedirectHosts:  s.options.ExtraRedirectHosts,
		})
		if err != nil {
			return fmt.Errorf("initialize oauth handlers: %w", err)
		}
		oauth = handlers
		if s.options.OAuthPrivateKeyPath == "" {
			slog.WarnContext(
				ctx,
				"oauth: using ephemeral RSA keypair (no --oauth-private-key); tokens minted before restart become unredeemable",
			)
		}
	}

	mainMux := http.NewServeMux()
	// wrapWithCORS wraps the /mcp handler. Claude Desktop and Cowork (and
	// any future browser-based MCP client) send an OPTIONS preflight before
	// the actual POST. Without CORS headers the browser silently drops the
	// request and reports "Couldn't reach the MCP server". We allow the
	// same set of origins that the OAuth allowlist permits; the pre-flight
	// does not require credentials, so `Access-Control-Allow-Credentials`
	// is intentionally omitted here.
	// CORS must wrap auth: OPTIONS preflights carry no credentials and
	// must receive CORS headers (status 204) without triggering a 401.
	mainMux.Handle("/mcp", wrapWithCORS(wrapWithAuthMiddleware(interceptedMCP, oauth)))
	mainMux.HandleFunc("/health", healthHandler)
	if oauth != nil {
		mainMux.Handle("/.well-known/oauth-protected-resource", wrapWithCORS(http.HandlerFunc(oauth.ServePRM)))
		mainMux.Handle("/oauth/mcp/jwks", wrapWithCORS(http.HandlerFunc(oauth.ServeJWKS)))
		mainMux.HandleFunc("/oauth/mcp/token", oauth.ServeToken)
	}

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.options.Port),
		Handler:           mainMux,
		ReadHeaderTimeout: 10 * time.Second, //nolint:mnd
	}

	var healthServer *http.Server
	healthPort := parseHealthCheckPort(s.options.Port)
	if healthPort != 0 {
		healthMux := http.NewServeMux()
		healthMux.HandleFunc("/health", healthHandler)
		healthServer = &http.Server{
			Addr:              fmt.Sprintf(":%d", healthPort),
			Handler:           healthMux,
			ReadHeaderTimeout: 10 * time.Second, //nolint:mnd
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

// wrapWithAuthMiddleware wraps the MCP handler with:
//  1. auth_method telemetry (see auth.go AuthMethod).
//  2. a requireValidAuth gate when OAuth is enabled — absent creds AND
//     revoked/invalid creds both produce 401 + WWW-Authenticate so MCP
//     clients re-run discovery. When OAuth is disabled (locally-run CLI)
//     only telemetry is added; the downstream handler surfaces auth
//     errors via its existing JSON-RPC path.
func wrapWithAuthMiddleware(next http.Handler, oauth *oauthHandlers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// InjectAuthContext has already run via the mcp-go HTTPContextFunc,
		// so the method+key are already available in context. Re-derive
		// here as well because the streamable handler constructs its own
		// context internally and this HTTP middleware runs earlier.
		ctx := InjectAuthContext(req.Context(), req)
		req = req.WithContext(ctx)

		method := authMethodFromContext(ctx)
		slog.InfoContext(ctx, "mcp.request", "auth_method", string(method))

		if oauth != nil {
			apiKey := apiKeyFromRequest(req)
			if apiKey == "" {
				oauth.WriteUnauthorized(w, "invalid_token", "missing credentials")
				return
			}
			if !oauth.ValidateAPIKey(ctx, apiKey) {
				oauth.WriteUnauthorized(w, "invalid_token", "revoked or invalid api key")
				return
			}
		}

		next.ServeHTTP(w, req)
	})
}

// authMethodFromContext returns the method classification recorded by
// InjectAuthContext, or AuthMethodNone if the context is empty.
func authMethodFromContext(ctx context.Context) AuthMethod {
	raw := ctx.Value(authContextKey{})
	auth, ok := raw.(Auth)
	if !ok {
		return AuthMethodNone
	}
	return auth.Method
}

// wrapWithCORS adds the CORS headers required by browser-based MCP clients
// (Claude Desktop, Cowork) so OPTIONS preflights don't silently drop the
// request and produce "Couldn't reach the MCP server" errors. We reflect
// the origin from the request rather than a hardcoded list to avoid
// maintenance — the actual API-key auth gates the sensitive operations;
// CORS here only controls which origins the browser allows to read
// responses, not who can mint tokens.
func wrapWithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
			w.Header().Set("Access-Control-Allow-Headers",
				"Authorization, Content-Type, Mcp-Session-Id, Accept")
			w.Header().Set("Vary", "Origin")
		}
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// apiKeyFromRequest extracts the api key from headers in the same order
// as InjectAuthContext, without requiring the context value (so the
// middleware can gate before handing off to the MCP streamable handler).
func apiKeyFromRequest(req *http.Request) string {
	if key := strings.TrimSpace(req.Header.Get("X-ESCAPE-API-KEY")); key != "" {
		return key
	}
	return apiKeyFromAuthorization(strings.TrimSpace(req.Header.Get("Authorization")))
}

func parseHealthCheckPort(mainPort int) int {
	raw := os.Getenv(healthCheckPortEnv)
	if raw == "" {
		return 0
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port <= 0 || port > 65535 || port == mainPort { //nolint:mnd
		return 0
	}
	return port
}
