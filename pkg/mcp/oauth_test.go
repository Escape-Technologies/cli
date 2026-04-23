package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// newTestOAuth builds a handlers instance with an ephemeral key and an
// upstream API that accepts a single fixed api key. Used across tests.
func newTestOAuth(t *testing.T, apiKey string) (*oauthHandlers, *httptest.Server) {
	t.Helper()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != upstreamValidationPath {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") == "Key "+apiKey {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"user"}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(upstream.Close)

	h, err := newOAuthHandlers(oauthConfig{
		IssuerURL:    "https://app.test",
		ResourceURL:  "https://mcp.test/mcp",
		PublicAPIURL: upstream.URL,
	})
	if err != nil {
		t.Fatalf("build handlers: %v", err)
	}
	return h, upstream
}

func pkcePair(t *testing.T) (verifier, challenge string) {
	t.Helper()
	verifier = "test-verifier-with-enough-entropy-to-be-long-enough-for-s256"
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return
}

func mintJWE(t *testing.T, h *oauthHandlers, mutate func(*oauthCodePayload)) string {
	t.Helper()
	_, challenge := pkcePair(t)
	payload := oauthCodePayload{
		APIKey:              "test-api-key",
		CodeChallenge:       challenge,
		CodeChallengeMethod: oauthCodeChallengeAlg,
		ClientID:            oauthClientID,
		RedirectURI:         "https://claude.ai/cb",
		Exp:                 time.Now().Add(1 * time.Minute).Unix(),
		Iat:                 time.Now().Unix(),
		JTI:                 fmt.Sprintf("jti-%d", time.Now().UnixNano()),
	}
	if mutate != nil {
		mutate(&payload)
	}
	jwe, err := h.EncryptCodeForTest(payload)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return jwe
}

func TestPRMHasOriginRootPath(t *testing.T) {
	t.Parallel()
	h, _ := newTestOAuth(t, "test-api-key")

	// The PRM discovery URL advertised in WWW-Authenticate must live
	// at the origin root, not under the /mcp resource path. See review
	// comment: deriving from ResourceURL produced /mcp/.well-known/...
	// which is the wrong host-level URL.
	if h.prmURL != "https://mcp.test/.well-known/oauth-protected-resource" {
		t.Fatalf("unexpected prmURL: %q", h.prmURL)
	}
}

func TestPRMResponse(t *testing.T) {
	t.Parallel()
	h, _ := newTestOAuth(t, "test-api-key")
	rec := httptest.NewRecorder()
	h.ServePRM(
		rec,
		httptest.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"/.well-known/oauth-protected-resource",
			nil,
		),
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %q", ct)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc["resource"] != "https://mcp.test/mcp" {
		t.Fatalf("unexpected resource: %v", doc["resource"])
	}
	if servers, ok := doc["authorization_servers"].([]any); !ok || len(servers) != 1 || servers[0] != "https://app.test" {
		t.Fatalf("unexpected authorization_servers: %v", doc["authorization_servers"])
	}
}

func TestJWKSShape(t *testing.T) {
	t.Parallel()
	h, _ := newTestOAuth(t, "test-api-key")
	rec := httptest.NewRecorder()
	h.ServeJWKS(
		rec,
		httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/oauth/mcp/jwks", nil),
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var doc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Use string `json:"use"`
			Alg string `json:"alg"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(doc.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(doc.Keys))
	}
	k := doc.Keys[0]
	if k.Kty != "RSA" || k.Use != "enc" || k.Alg != "RSA-OAEP-256" || k.Kid == "" || k.N == "" || k.E == "" {
		t.Fatalf("unexpected jwk: %+v", k)
	}
}

func TestUnauthorizedResponse(t *testing.T) {
	t.Parallel()
	h, _ := newTestOAuth(t, "test-api-key")
	rec := httptest.NewRecorder()
	h.WriteUnauthorized(rec, "invalid_token", "missing credentials")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("missing Cache-Control: %q", rec.Header().Get("Cache-Control"))
	}
	wwwAuth := rec.Header().Get("WWW-Authenticate")
	if !strings.Contains(wwwAuth, `realm="mcp"`) {
		t.Fatalf("WWW-Authenticate missing realm: %q", wwwAuth)
	}
	if !strings.Contains(wwwAuth, `resource_metadata="https://mcp.test/.well-known/oauth-protected-resource"`) {
		t.Fatalf("WWW-Authenticate missing PRM URL: %q", wwwAuth)
	}
	if !strings.Contains(wwwAuth, `error="invalid_token"`) {
		t.Fatalf("WWW-Authenticate missing error param: %q", wwwAuth)
	}
}

type tokenTestCase struct {
	name         string
	form         url.Values
	expectStatus int
	expectError  string
}

func TestServeToken(t *testing.T) {
	t.Parallel()

	h, _ := newTestOAuth(t, "test-api-key")
	verifier, challenge := pkcePair(t)

	validForm := func() url.Values {
		code := mintJWE(t, h, func(p *oauthCodePayload) { p.CodeChallenge = challenge })
		return url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"code_verifier": {verifier},
			"redirect_uri":  {"https://claude.ai/cb"},
			"client_id":     {oauthClientID},
		}
	}

	cases := []tokenTestCase{
		{
			name:         "unsupported grant_type",
			form:         url.Values{"grant_type": {"password"}, "code": {"x"}, "code_verifier": {"y"}, "redirect_uri": {"https://claude.ai/cb"}, "client_id": {oauthClientID}},
			expectStatus: http.StatusBadRequest,
			expectError:  "unsupported_grant_type",
		},
		{
			name: "missing code_verifier",
			form: url.Values{"grant_type": {"authorization_code"}, "code": {"x"}, "redirect_uri": {"https://claude.ai/cb"}, "client_id": {oauthClientID}},
			expectStatus: http.StatusBadRequest,
			expectError:  "invalid_request",
		},
		{
			name:         "malformed code",
			form:         url.Values{"grant_type": {"authorization_code"}, "code": {"not-a-jwe"}, "code_verifier": {verifier}, "redirect_uri": {"https://claude.ai/cb"}, "client_id": {oauthClientID}},
			expectStatus: http.StatusBadRequest,
			expectError:  "invalid_grant",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := newFormRequest(tc.form)
			rec := httptest.NewRecorder()
			h.ServeToken(rec, req)
			assertOAuthError(t, rec, tc.expectStatus, tc.expectError)
		})
	}

	// Happy path.
	t.Run("happy_path", func(t *testing.T) {
		t.Parallel()
		req := newFormRequest(validForm())
		rec := httptest.NewRecorder()
		h.ServeToken(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("missing Cache-Control")
		}
		if rec.Header().Get("Pragma") != "no-cache" {
			t.Fatalf("missing Pragma")
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if body["access_token"] != "test-api-key" {
			t.Fatalf("unexpected access_token: %v", body["access_token"])
		}
		if body["token_type"] != "Bearer" {
			t.Fatalf("unexpected token_type: %v", body["token_type"])
		}
		// Regression: expires_in must be 1y (31,536,000) not 10y.
		if v, ok := body["expires_in"].(float64); !ok || int64(v) != 31536000 {
			t.Fatalf("unexpected expires_in: %v", body["expires_in"])
		}
	})

	t.Run("wrong_verifier", func(t *testing.T) {
		t.Parallel()
		form := validForm()
		form.Set("code_verifier", "wrong-verifier-wrong-verifier-wrong")
		rec := httptest.NewRecorder()
		h.ServeToken(rec, newFormRequest(form))
		assertOAuthError(t, rec, http.StatusBadRequest, "invalid_grant")
	})

	t.Run("expired_code", func(t *testing.T) {
		t.Parallel()
		verifierLocal, challengeLocal := pkcePair(t)
		code := mintJWE(t, h, func(p *oauthCodePayload) {
			p.CodeChallenge = challengeLocal
			p.Exp = time.Now().Add(-1 * time.Minute).Unix()
		})
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"code_verifier": {verifierLocal},
			"redirect_uri":  {"https://claude.ai/cb"},
			"client_id":     {oauthClientID},
		}
		rec := httptest.NewRecorder()
		h.ServeToken(rec, newFormRequest(form))
		assertOAuthError(t, rec, http.StatusBadRequest, "invalid_grant")
	})

	t.Run("replay_rejected", func(t *testing.T) {
		t.Parallel()
		form := validForm()
		// First redemption succeeds.
		rec1 := httptest.NewRecorder()
		h.ServeToken(rec1, newFormRequest(form))
		if rec1.Code != http.StatusOK {
			t.Fatalf("first redemption failed: %d %s", rec1.Code, rec1.Body.String())
		}
		// Second must fail with invalid_grant.
		rec2 := httptest.NewRecorder()
		h.ServeToken(rec2, newFormRequest(form))
		assertOAuthError(t, rec2, http.StatusBadRequest, "invalid_grant")
	})

	t.Run("wrong_redirect_uri_form", func(t *testing.T) {
		t.Parallel()
		form := validForm()
		form.Set("redirect_uri", "https://claude.ai/other")
		rec := httptest.NewRecorder()
		h.ServeToken(rec, newFormRequest(form))
		assertOAuthError(t, rec, http.StatusBadRequest, "invalid_grant")
	})

	t.Run("redirect_uri_outside_allowlist_defense_in_depth", func(t *testing.T) {
		t.Parallel()
		verifierLocal, challengeLocal := pkcePair(t)
		// Mint a code where the payload's redirect_uri is NOT on the
		// allowlist. This simulates a compromised authorize step.
		code := mintJWE(t, h, func(p *oauthCodePayload) {
			p.CodeChallenge = challengeLocal
			p.RedirectURI = "https://evil.com/cb"
		})
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"code_verifier": {verifierLocal},
			"redirect_uri":  {"https://evil.com/cb"},
			"client_id":     {oauthClientID},
		}
		rec := httptest.NewRecorder()
		h.ServeToken(rec, newFormRequest(form))
		assertOAuthError(t, rec, http.StatusBadRequest, "invalid_grant")
	})

	t.Run("wrong_client_id", func(t *testing.T) {
		t.Parallel()
		form := validForm()
		form.Set("client_id", "someone-else")
		rec := httptest.NewRecorder()
		h.ServeToken(rec, newFormRequest(form))
		assertOAuthError(t, rec, http.StatusBadRequest, "invalid_grant")
	})

	t.Run("plain_challenge_rejected", func(t *testing.T) {
		t.Parallel()
		verifierLocal := "some-verifier"
		code := mintJWE(t, h, func(p *oauthCodePayload) {
			p.CodeChallenge = verifierLocal // plain = verifier
			p.CodeChallengeMethod = "plain"
		})
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"code_verifier": {verifierLocal},
			"redirect_uri":  {"https://claude.ai/cb"},
			"client_id":     {oauthClientID},
		}
		rec := httptest.NewRecorder()
		h.ServeToken(rec, newFormRequest(form))
		assertOAuthError(t, rec, http.StatusBadRequest, "invalid_grant")
	})
}

func TestRedirectAllowlist(t *testing.T) {
	t.Parallel()
	a := buildRedirectAllowlist([]string{"qa.staging.example"})

	cases := []struct {
		raw  string
		want bool
	}{
		// Allowed.
		{"https://claude.ai/cb", true},
		{"https://foo.anthropic.com/cb", true},
		{"https://cowork.ai/cb", true},
		{"https://sub.cowork.ai/cb", true},
		{"http://localhost:12345/cb", true},
		{"http://127.0.0.1:8080/cb", true},
		// IPv6 loopback — URL parsers may or may not keep brackets on
		// Hostname(); both forms must match so MCP clients binding to
		// [::1] don't get rejected at the loopback gate.
		{"http://[::1]:8080/cb", true},
		{"https://qa.staging.example/cb", true},

		// Rejected — phishing vectors.
		{"https://evil.claude.ai.attacker.com/cb", false},
		{"https://claude.ai.attacker.com/cb", false},
		{"https://anthropic.com.evil/cb", false},
		{"https://fakeclaude.ai/cb", false},

		// Rejected — wrong scheme.
		{"javascript:alert(1)", false},
		{"data:text/html,x", false},
		{"ftp://cowork.ai/cb", false},
		{"http://claude.ai/cb", false}, // must be https for non-loopback

		// Rejected — userinfo + fragments.
		{"https://claude.ai@evil/cb", false},
		{"https://claude.ai/cb#fragment", false},

		// Empty.
		{"", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()
			if got := a.allow(tc.raw); got != tc.want {
				t.Fatalf("allow(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func newFormRequest(form url.Values) *http.Request {
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/oauth/mcp/token",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func assertOAuthError(t *testing.T, rec *httptest.ResponseRecorder, status int, expectError string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("expected status %d, got %d body=%s", status, rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("missing Cache-Control on error response")
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["error"] != expectError {
		t.Fatalf("expected error %q, got %q", expectError, body["error"])
	}
}

// Drain response body, used in some tests to silence linter.
func drainBody(body io.ReadCloser) { _, _ = io.Copy(io.Discard, body); _ = body.Close() }

var _ = drainBody
