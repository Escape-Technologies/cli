package mcp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestOAuthEndToEnd exercises the full handshake against the MCP server
// mux. Proves: 401 discovery, token exchange with cache headers, Bearer
// auth accepted at /mcp AFTER normalization (no leak into child env),
// revoked keys also trigger 401, and replay is rejected.
func TestOAuthEndToEnd(t *testing.T) {
	t.Parallel()

	const validAPIKey = "valid-api-key-12345"

	// Upstream API stub. Validates keys via /v3/me and captures the
	// Authorization header so we can assert the Bearer-normalization fix.
	var lastAuthHeaderSeen string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastAuthHeaderSeen = r.Header.Get("Authorization")
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
	defer upstream.Close()

	// Build the OAuth handlers pointed at the upstream stub.
	oauth, err := newOAuthHandlers(oauthConfig{
		IssuerURL:    "https://app.test",
		ResourceURL:  "https://mcp.test/mcp",
		PublicAPIURL: upstream.URL,
	})
	if err != nil {
		t.Fatalf("build oauth: %v", err)
	}

	// Stub /mcp handler that records the Auth struct it saw.
	var seenAuth Auth
	mcpEcho := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, _ := AuthFromContext(InjectAuthContext(r.Context(), r))
		seenAuth = auth
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	mux := http.NewServeMux()
	mux.Handle("/mcp", wrapWithAuthMiddleware(mcpEcho, oauth))
	mux.HandleFunc("/.well-known/oauth-protected-resource", oauth.ServePRM)
	mux.HandleFunc("/oauth/mcp/jwks", oauth.ServeJWKS)
	mux.HandleFunc("/oauth/mcp/token", oauth.ServeToken)

	server := httptest.NewServer(mux)
	defer server.Close()

	// 1. POST /mcp without creds → 401 + WWW-Authenticate + PRM URL at origin root.
	t.Run("unauthenticated_returns_401_with_discovery", func(t *testing.T) {
		resp, err := http.Post(server.URL+"/mcp", "application/json", strings.NewReader(`{}`))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		wwwAuth := resp.Header.Get("WWW-Authenticate")
		// The PRM URL advertised must be at the origin root, not under /mcp.
		if !strings.Contains(wwwAuth, "/.well-known/oauth-protected-resource") {
			t.Fatalf("PRM URL missing from WWW-Authenticate: %q", wwwAuth)
		}
		if strings.Contains(wwwAuth, "/mcp/.well-known/") {
			t.Fatalf("PRM URL must NOT be under /mcp: %q", wwwAuth)
		}
	})

	// 2. POST /mcp with an invalid/revoked key → 401 + WWW-Authenticate.
	// Regression for reviewer's point: revoked keys must also trigger discovery.
	t.Run("invalid_key_also_returns_401_discovery", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/mcp", strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer revoked-key")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		if resp.Header.Get("WWW-Authenticate") == "" {
			t.Fatalf("missing WWW-Authenticate")
		}
	})

	// 3. GET /.well-known/oauth-protected-resource returns PRM JSON.
	t.Run("prm_endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/.well-known/oauth-protected-resource")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
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

	// 4. GET /oauth/mcp/jwks parses.
	t.Run("jwks_endpoint", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/oauth/mcp/jwks")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	// 5. Mint a JWE in-test with a valid api key, exchange it at /oauth/mcp/token,
	//    assert cache headers + expires_in=1y + access_token matches the embedded key.
	verifier, challenge := pkcePair(t)
	jwe, err := oauth.EncryptCodeForTest(oauthCodePayload{
		APIKey:              validAPIKey,
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

	t.Run("token_exchange_returns_cache_headers", func(t *testing.T) {
		resp, err := http.PostForm(server.URL+"/oauth/mcp/token", tokenForm)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer drain(resp.Body)
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
		if body["access_token"] != validAPIKey {
			t.Fatalf("unexpected access_token: %v", body["access_token"])
		}
		if v, ok := body["expires_in"].(float64); !ok || int64(v) != 31536000 {
			t.Fatalf("unexpected expires_in: %v", body["expires_in"])
		}
	})

	// 6. POST /mcp with the Bearer token → 200, handler-observed Auth has
	//    APIKey set and Authorization cleared (Bearer normalization). The
	//    upstream API also saw `Key <key>` (not `Bearer ...`), preventing
	//    JWT-verify 401 at the backend.
	t.Run("bearer_accepted_and_normalized", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/mcp", strings.NewReader(`{}`))
		req.Header.Set("Authorization", "Bearer "+validAPIKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer drain(resp.Body)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, body)
		}
		if seenAuth.APIKey != validAPIKey {
			t.Fatalf("expected handler to see APIKey %q, got %q", validAPIKey, seenAuth.APIKey)
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
		// Upstream validation hit should have used "Key <apikey>".
		if !strings.HasPrefix(lastAuthHeaderSeen, "Key ") {
			t.Fatalf("upstream saw unexpected authorization: %q", lastAuthHeaderSeen)
		}
	})

	// 7. Replay: re-POST the same code → 400 invalid_grant.
	t.Run("replay_rejected", func(t *testing.T) {
		resp, err := http.PostForm(server.URL+"/oauth/mcp/token", tokenForm)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
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

func drain(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
