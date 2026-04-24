package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// observedUpstream captures the Authorization header seen by the upstream
// stub. httptest runs handlers on separate goroutines, so writes MUST be
// synchronized with the test goroutine's reads — otherwise `go test -race`
// flags it and the assertion can be flaky.
type observedUpstream struct {
	mu               sync.Mutex
	lastAuthHeader   string
	seenMCPHandlerAt Auth
}

func (o *observedUpstream) recordUpstream(header string) {
	o.mu.Lock()
	o.lastAuthHeader = header
	o.mu.Unlock()
}

func (o *observedUpstream) recordMCPHandler(auth Auth) {
	o.mu.Lock()
	o.seenMCPHandlerAt = auth
	o.mu.Unlock()
}

func (o *observedUpstream) snapshot() (string, Auth) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.lastAuthHeader, o.seenMCPHandlerAt
}

// e2eFixture holds the live httptest.Server + helpers used by each
// subtest. Extracted so TestOAuthEndToEnd itself stays under the
// cyclomatic complexity limit.
type e2eFixture struct {
	t          *testing.T
	server     *httptest.Server
	oauth      *oauthHandlers
	validKey   string
	observed   *observedUpstream
}

func newE2EFixture(t *testing.T) *e2eFixture {
	t.Helper()

	const validAPIKey = "valid-api-key-12345"
	observed := &observedUpstream{}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed.recordUpstream(r.Header.Get("Authorization"))
		if r.URL.Path != upstreamValidationPath {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") == "Key "+validAPIKey {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"user-1"}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(upstream.Close)

	oauth, err := newOAuthHandlers(oauthConfig{
		IssuerURL:    "https://app.test",
		ResourceURL:  "https://mcp.test/mcp",
		PublicAPIURL: upstream.URL,
	})
	if err != nil {
		t.Fatalf("build oauth: %v", err)
	}

	mcpEcho := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, _ := AuthFromContext(InjectAuthContext(r.Context(), r))
		observed.recordMCPHandler(auth)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	mux := http.NewServeMux()
	mux.Handle("/mcp", wrapWithCORS(wrapWithAuthMiddleware(mcpEcho, oauth)))
	mux.Handle("/.well-known/oauth-protected-resource", wrapWithCORS(http.HandlerFunc(oauth.ServePRM)))
	mux.Handle("/oauth/mcp/jwks", wrapWithCORS(http.HandlerFunc(oauth.ServeJWKS)))
	mux.HandleFunc("/oauth/mcp/token", oauth.ServeToken)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return &e2eFixture{
		t:        t,
		server:   server,
		oauth:    oauth,
		validKey: validAPIKey,
		observed: observed,
	}
}

func (f *e2eFixture) post(path, contentType, body string, headers map[string]string) *http.Response {
	f.t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		f.server.URL+path,
		strings.NewReader(body),
	)
	if err != nil {
		f.t.Fatalf("build request: %v", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		f.t.Fatalf("post %s: %v", path, err)
	}
	return resp
}

func (f *e2eFixture) get(path string) *http.Response {
	f.t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, f.server.URL+path, http.NoBody)
	if err != nil {
		f.t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		f.t.Fatalf("get %s: %v", path, err)
	}
	return resp
}

// TestOAuthEndToEnd exercises the full handshake against the MCP server
// mux. Proves: 401 discovery, token exchange with cache headers, Bearer
// auth accepted at /mcp AFTER normalization (no leak into child env),
// revoked keys also trigger 401, and replay is rejected.
func TestOAuthEndToEnd(t *testing.T) {
	t.Parallel()

	f := newE2EFixture(t)
	verifier, challenge := pkcePair(t)

	// Pre-mint a valid JWE now so per-subtest complexity stays low.
	jwe, err := f.oauth.EncryptCodeForTest(oauthCodePayload{
		APIKey:              f.validKey,
		CodeChallenge:       challenge,
		CodeChallengeMethod: oauthCodeChallengeAlg,
		ClientID:            oauthClientID,
		RedirectURI:         "https://claude.ai/cb",
		Exp:                 time.Now().Add(time.Minute).Unix(),
		Iat:                 time.Now().Unix(),
		JTI:                 "e2e-jti-1",
	})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	tokenForm := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {jwe},
		"code_verifier": {verifier},
		"redirect_uri":  {"https://claude.ai/cb"},
		"client_id":     {oauthClientID},
	}

	t.Run("unauthenticated_returns_401_with_discovery", func(t *testing.T) {
		resp := f.post("/mcp", "application/json", `{}`, nil) //nolint:bodyclose // drained below
		defer drain(resp.Body)
		assertStatusAndDiscoveryHeader(t, resp, http.StatusUnauthorized)
	})

	// Regression for "Couldn't reach the MCP server" from Claude.
	// Browser-based MCP clients (Cowork, Claude web) send an OPTIONS
	// preflight before the POST /mcp. Without CORS headers the browser
	// treats the silence as a rejection and drops the request entirely —
	// no 401, no JSON-RPC error, just a network-level block that shows
	// up as "Couldn't reach the MCP server" in the client UI.
	t.Run("options_preflight_returns_204_with_cors_headers", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodOptions, f.server.URL+"/mcp", http.NoBody)
		req.Header.Set("Origin", "https://claude.ai")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Authorization,Content-Type")
		resp, err := http.DefaultClient.Do(req) //nolint:bodyclose // drained below
		if err != nil {
			t.Fatalf("options: %v", err)
		}
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204 for OPTIONS preflight, got %d", resp.StatusCode)
		}
		if resp.Header.Get("Access-Control-Allow-Origin") != "https://claude.ai" {
			t.Fatalf("ACAO header missing or wrong: %q", resp.Header.Get("Access-Control-Allow-Origin"))
		}
		if resp.Header.Get("Access-Control-Allow-Methods") == "" {
			t.Fatalf("ACAM header missing")
		}
		// Preflights must NOT carry WWW-Authenticate (that would break
		// the browser's CORS check before any auth happens).
		if resp.Header.Get("Www-Authenticate") != "" {
			t.Fatalf("WWW-Authenticate must be absent on OPTIONS preflight, got: %q", resp.Header.Get("Www-Authenticate"))
		}
	})

	t.Run("invalid_key_also_returns_401_discovery", func(t *testing.T) {
		resp := f.post("/mcp", "application/json", `{}`, map[string]string{ //nolint:bodyclose // drained below
			"Authorization": "Bearer revoked-key",
		})
		defer drain(resp.Body)
		assertStatusAndDiscoveryHeader(t, resp, http.StatusUnauthorized)
	})

	t.Run("prm_endpoint", func(t *testing.T) {
		resp := f.get("/.well-known/oauth-protected-resource") //nolint:bodyclose // drained below
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var doc map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&doc)
		if doc["resource"] != "https://mcp.test/mcp" {
			t.Fatalf("unexpected resource: %v", doc["resource"])
		}
	})

	t.Run("jwks_endpoint", func(t *testing.T) {
		resp := f.get("/oauth/mcp/jwks") //nolint:bodyclose // drained below
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("token_exchange_returns_cache_headers", func(t *testing.T) {
		resp := f.post( //nolint:bodyclose // drained below
			"/oauth/mcp/token",
			"application/x-www-form-urlencoded",
			tokenForm.Encode(),
			nil,
		)
		defer drain(resp.Body)
		assertTokenResponse(t, resp, f.validKey)
	})

	t.Run("bearer_accepted_and_normalized", func(t *testing.T) {
		resp := f.post("/mcp", "application/json", `{}`, map[string]string{ //nolint:bodyclose // drained below
			"Authorization": "Bearer " + f.validKey,
		})
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, body)
		}
		lastAuth, seenAuth := f.observed.snapshot()
		if seenAuth.APIKey != f.validKey {
			t.Fatalf("expected handler to see APIKey %q, got %q", f.validKey, seenAuth.APIKey)
		}
		if seenAuth.Authorization != "" {
			t.Fatalf(
				"expected Authorization to be cleared to prevent leak into child CLI, got %q",
				seenAuth.Authorization,
			)
		}
		if seenAuth.Method != AuthMethodAuthorizationBearer {
			t.Fatalf("expected bearer method, got %q", seenAuth.Method)
		}
		if !strings.HasPrefix(lastAuth, "Key ") {
			t.Fatalf("upstream saw unexpected authorization: %q", lastAuth)
		}
	})

	t.Run("replay_rejected", func(t *testing.T) {
		resp := f.post( //nolint:bodyclose // drained below
			"/oauth/mcp/token",
			"application/x-www-form-urlencoded",
			tokenForm.Encode(),
			nil,
		)
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if body["error"] != "invalid_grant" {
			t.Fatalf("expected invalid_grant, got %v", body["error"])
		}
	})
}

func assertStatusAndDiscoveryHeader(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("expected status %d, got %d", want, resp.StatusCode)
	}
	wwwAuth := resp.Header.Get("WWW-Authenticate")
	if !strings.Contains(wwwAuth, "/.well-known/oauth-protected-resource") {
		t.Fatalf("PRM URL missing from WWW-Authenticate: %q", wwwAuth)
	}
	if strings.Contains(wwwAuth, "/mcp/.well-known/") {
		t.Fatalf("PRM URL must NOT be under /mcp: %q", wwwAuth)
	}
}

func assertTokenResponse(t *testing.T, resp *http.Response, expectedAPIKey string) {
	t.Helper()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, body)
	}
	if resp.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("missing Cache-Control: %q", resp.Header.Get("Cache-Control"))
	}
	if resp.Header.Get("Pragma") != "no-cache" {
		t.Fatalf("missing Pragma: %q", resp.Header.Get("Pragma"))
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["access_token"] != expectedAPIKey {
		t.Fatalf("unexpected access_token: %v", body["access_token"])
	}
	const oneYearSeconds = 31536000
	if v, ok := body["expires_in"].(float64); !ok || int64(v) != oneYearSeconds {
		t.Fatalf("unexpected expires_in: %v", body["expires_in"])
	}
}

func drain(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
