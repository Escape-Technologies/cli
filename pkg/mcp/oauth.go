package mcp

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

const (
	oauthKeyBits           = 3072
	oauthCodeExpiryWindow  = 75 * time.Second // jti kept > code exp to catch late replays
	oauthAccessTokenTTL    = 31536000         // 1 year (seconds) — real expiry is api-key revocation
	oauthClientID          = "escape-mcp-public"
	oauthCodeChallengeAlg  = "S256"
	// oauthKeyDirPerm is the mode for the directory that holds the OAuth
	// RSA private key PEM (u=rwx, g=-, o=-).
	oauthKeyDirPerm        os.FileMode = 0o700
	// oauthKeyFilePerm matches the sensitive-file convention for secrets
	// at rest (u=rw, g=-, o=-).
	oauthKeyFilePerm os.FileMode = 0o600
	// jwksCacheFilePerm kept for backwards-compat with external writers.
	jwksCacheFilePerm      = oauthKeyFilePerm
	validationCacheTTL     = 60 * time.Second
	upstreamValidationPath = "/v3/me"
)

// oauthHandlers owns the keypair, per-request state (jti seen-set,
// upstream validation cache), and the compiled redirect allowlist.
// It is created once at server start and shared across requests.
type oauthHandlers struct {
	privateKey         *rsa.PrivateKey
	publicKey          *rsa.PublicKey
	kid                string
	issuerURL          string
	resourceURL        string
	prmURL             string
	jwksURL            string
	tokenURL           string
	authorizeURL       string
	registrationURL    string
	publicAPIURL       string
	allowlist          *redirectAllowlist
	seenJTI            map[string]int64
	seenJTIMu          sync.Mutex
	upstreamClient     *http.Client
	validationCache    map[string]int64
	validationCacheMu  sync.Mutex
	// TODO(mcp-oauth): seenJTI and validationCache are process-local.
	// mcp.escape.tech must run as a single replica (or with sticky routing
	// keyed on jti) until we move jti to a shared short-TTL store such as
	// Redis (SET jti 1 EX 90 NX). Same applies to validationCache: every
	// replica caches independently.
}

// oauthCodePayload is the JSON structure encrypted inside the authorization
// code JWE. Fields mirror the OAuth 2.1 authorization-code flow bindings.
type oauthCodePayload struct {
	APIKey              string `json:"api_key"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Exp                 int64  `json:"exp"`
	Iat                 int64  `json:"iat"`
	JTI                 string `json:"jti"`
}

// oauthConfig is what server.go passes when building the handlers.
type oauthConfig struct {
	IssuerURL          string
	ResourceURL        string
	PublicAPIURL       string
	OAuthPrivateKeyPath string
	ExtraRedirectHosts []string
}

// newOAuthHandlers bootstraps the RSA keypair, compiles the allowlist, and
// prepares the shared state. It never returns a partially-initialized
// handler — on any failure it errors out so the serve command aborts.
func newOAuthHandlers(cfg oauthConfig) (*oauthHandlers, error) {
	privateKey, err := loadOrGenerateKey(cfg.OAuthPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load oauth private key: %w", err)
	}
	if privateKey.N.BitLen() < 2048 { //nolint:mnd // RFC 7518 §6.3 minimum.
		return nil, fmt.Errorf("oauth private key is too small: got %d bits, need at least 2048", privateKey.N.BitLen())
	}

	publicKey := &privateKey.PublicKey
	kid, err := keyID(publicKey)
	if err != nil {
		return nil, fmt.Errorf("compute key id: %w", err)
	}

	resourceURL := strings.TrimSpace(cfg.ResourceURL)
	if resourceURL == "" {
		return nil, errors.New("ResourceURL is required")
	}
	prmURL, err := derivePRMURL(resourceURL)
	if err != nil {
		return nil, fmt.Errorf("derive prm url: %w", err)
	}

	resourceOrigin, err := originOf(resourceURL)
	if err != nil {
		return nil, fmt.Errorf("parse resource url: %w", err)
	}

	issuerURL := strings.TrimSpace(cfg.IssuerURL)
	if issuerURL == "" {
		return nil, errors.New("IssuerURL is required")
	}

	handlers := &oauthHandlers{
		privateKey:      privateKey,
		publicKey:       publicKey,
		kid:             kid,
		issuerURL:       strings.TrimRight(issuerURL, "/"),
		resourceURL:     resourceURL,
		prmURL:          prmURL,
		jwksURL:         resourceOrigin + "/oauth/mcp/jwks",
		tokenURL:        resourceOrigin + "/oauth/mcp/token",
		authorizeURL:    strings.TrimRight(issuerURL, "/") + "/oauth/mcp/authorize",
		registrationURL: "", // Hosted on the Node API; advertised via AS metadata, not by this server.
		publicAPIURL:    strings.TrimRight(cfg.PublicAPIURL, "/"),
		allowlist:       buildRedirectAllowlist(cfg.ExtraRedirectHosts),
		seenJTI:         make(map[string]int64),
		validationCache: make(map[string]int64),
		upstreamClient:  &http.Client{Timeout: 5 * time.Second}, //nolint:mnd
	}

	return handlers, nil
}

// ServePRM serves the RFC 9728 Protected Resource Metadata document.
// The JSON is deterministic, so callers can cache it at the edge —
// however, per the OAuth spec we still send no-store to avoid any
// accidental staleness if the document ever changes.
func (h *oauthHandlers) ServePRM(w http.ResponseWriter, _ *http.Request) {
	doc := map[string]any{
		"resource":                 h.resourceURL,
		"authorization_servers":    []string{h.issuerURL},
		"bearer_methods_supported": []string{"header"},
		"scopes_supported":         []string{"mcp"},
	}
	writeJSON(w, http.StatusOK, doc, true)
}

// ServeJWKS serves the public RSA key as a JWKS document. The Node API
// reads this to encrypt authorization codes.
func (h *oauthHandlers) ServeJWKS(w http.ResponseWriter, _ *http.Request) {
	jwk := jose.JSONWebKey{
		Key:       h.publicKey,
		KeyID:     h.kid,
		Algorithm: string(jose.RSA_OAEP_256),
		Use:       "enc",
	}
	doc := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	writeJSON(w, http.StatusOK, doc, false)
}

// ServeToken is the OAuth 2.1 token endpoint. It decrypts the JWE code,
// validates PKCE + jti + exp + redirect_uri + client_id, and returns the
// embedded api key as the access token.
func (h *oauthHandlers) ServeToken(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		h.writeOAuthError(w, http.StatusMethodNotAllowed, "invalid_request", "method not allowed")
		return
	}
	if err := req.ParseForm(); err != nil {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_request", "malformed form body")
		return
	}

	if grantType := req.Form.Get("grant_type"); grantType != "authorization_code" {
		h.writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type",
			fmt.Sprintf("grant_type %q is not supported", grantType))
		return
	}

	code := req.Form.Get("code")
	codeVerifier := req.Form.Get("code_verifier")
	redirectURI := req.Form.Get("redirect_uri")
	clientID := req.Form.Get("client_id")

	if code == "" || codeVerifier == "" || redirectURI == "" || clientID == "" {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_request",
			"missing one of code, code_verifier, redirect_uri, client_id")
		return
	}

	payload, err := h.decryptCode(code)
	if err != nil {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "unable to decrypt authorization code")
		return
	}

	now := time.Now().Unix()
	if payload.Exp < now {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "authorization code expired")
		return
	}
	if payload.CodeChallengeMethod != oauthCodeChallengeAlg {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "unsupported code_challenge_method")
		return
	}
	if !verifyPKCE(codeVerifier, payload.CodeChallenge) {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "code_verifier does not match code_challenge")
		return
	}

	// Exact-match client_id and redirect_uri with the encrypted payload.
	if subtle.ConstantTimeCompare([]byte(clientID), []byte(payload.ClientID)) != 1 {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
		return
	}
	if subtle.ConstantTimeCompare([]byte(redirectURI), []byte(payload.RedirectURI)) != 1 {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
		return
	}
	// Defense-in-depth: validate the decrypted redirect_uri against the
	// server's allowlist even though the authorize endpoint should have
	// rejected it already. Protects against a compromised authorize path.
	if !h.allowlist.allow(payload.RedirectURI) {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "redirect_uri is not in the server allowlist")
		return
	}
	if payload.ClientID != oauthClientID {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "unknown client_id")
		return
	}
	if payload.APIKey == "" {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "authorization code has no api_key")
		return
	}

	// Single-use jti check. Keep entries until a safe margin past exp so
	// any late replay (within clock skew) still rejects.
	if !h.markJTISeen(payload.JTI, payload.Exp) {
		h.writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "authorization code already used")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": payload.APIKey,
		"token_type":   "Bearer",
		"expires_in":   oauthAccessTokenTTL,
		"scope":        "mcp",
	}, true)
}

// WriteUnauthorized emits the canonical 401 + WWW-Authenticate response
// used whenever the MCP /mcp endpoint receives a missing or invalid
// credential. Also sets the no-store cache headers.
func (h *oauthHandlers) WriteUnauthorized(w http.ResponseWriter, errorCode, description string) {
	h.setNoCacheHeaders(w)
	w.Header().Set(
		"WWW-Authenticate",
		fmt.Sprintf(
			`Bearer realm="mcp", error=%q, error_description=%q, resource_metadata=%q`,
			errorCode, description, h.prmURL,
		),
	)
	writeJSON(w, http.StatusUnauthorized, map[string]any{
		"error":             errorCode,
		"error_description": description,
	}, false) // Headers already set above; writeJSON won't override.
}

// ValidateAPIKey hits the upstream API's /v3/me to check the key is
// active (not revoked). Cached per-hash for validationCacheTTL.
// Returns true if the key is valid. Takes a context so upstream
// validation inherits the caller's timeout / cancellation.
func (h *oauthHandlers) ValidateAPIKey(ctx context.Context, apiKey string) bool {
	if apiKey == "" {
		return false
	}

	hash := sha256.Sum256([]byte(apiKey))
	cacheKey := hex.EncodeToString(hash[:])

	now := time.Now().Unix()
	h.validationCacheMu.Lock()
	// Lazy sweep.
	for k, v := range h.validationCache {
		if v < now {
			delete(h.validationCache, k)
		}
	}
	if expAt, ok := h.validationCache[cacheKey]; ok && expAt >= now {
		h.validationCacheMu.Unlock()
		return true
	}
	h.validationCacheMu.Unlock()

	if h.publicAPIURL == "" {
		// In local dev without a PublicAPIURL we cannot validate; accept
		// presence as sufficient. Not a production path.
		return true
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		h.publicAPIURL+upstreamValidationPath,
		http.NoBody,
	)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Key "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := h.upstreamClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	const httpOKFloor = 200
	const httpOKCeil = 300
	if resp.StatusCode >= httpOKFloor && resp.StatusCode < httpOKCeil {
		h.validationCacheMu.Lock()
		h.validationCache[cacheKey] = now + int64(validationCacheTTL.Seconds())
		h.validationCacheMu.Unlock()
		return true
	}
	return false
}

// decryptCode decrypts the JWE-encoded authorization code.
func (h *oauthHandlers) decryptCode(code string) (*oauthCodePayload, error) {
	jwe, err := jose.ParseEncrypted(
		code,
		[]jose.KeyAlgorithm{jose.RSA_OAEP_256},
		[]jose.ContentEncryption{jose.A256GCM},
	)
	if err != nil {
		return nil, fmt.Errorf("parse jwe: %w", err)
	}
	plaintext, err := jwe.Decrypt(h.privateKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt jwe: %w", err)
	}

	var payload oauthCodePayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}
	return &payload, nil
}

// markJTISeen records jti as used. Returns false if it was already seen.
// Sweeps expired entries on each call to keep the map bounded.
func (h *oauthHandlers) markJTISeen(jti string, exp int64) bool {
	if jti == "" {
		// Without a jti the token is already not single-use; reject.
		return false
	}
	h.seenJTIMu.Lock()
	defer h.seenJTIMu.Unlock()

	now := time.Now().Unix()
	for k, v := range h.seenJTI {
		if v < now {
			delete(h.seenJTI, k)
		}
	}
	if _, exists := h.seenJTI[jti]; exists {
		return false
	}
	keepUntil := exp + int64(oauthCodeExpiryWindow.Seconds())
	h.seenJTI[jti] = keepUntil
	return true
}

func (h *oauthHandlers) writeOAuthError(w http.ResponseWriter, status int, errorCode, description string) {
	h.setNoCacheHeaders(w)
	writeJSON(w, status, map[string]any{
		"error":             errorCode,
		"error_description": description,
	}, false)
}

func (h *oauthHandlers) setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}

// EncryptCodeForTest builds a JWE using the handler's public key. Only
// used from tests to avoid a circular test helper.
func (h *oauthHandlers) EncryptCodeForTest(payload oauthCodePayload) (string, error) {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}
	recipient := jose.Recipient{
		Algorithm: jose.RSA_OAEP_256,
		Key:       h.publicKey,
		KeyID:     h.kid,
	}
	encrypter, err := jose.NewEncrypter(jose.A256GCM, recipient, nil)
	if err != nil {
		return "", fmt.Errorf("new encrypter: %w", err)
	}
	object, err := encrypter.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("encrypt payload: %w", err)
	}
	serialized, err := object.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("serialize jwe: %w", err)
	}
	return serialized, nil
}

// verifyPKCE validates the code_verifier against an S256 code_challenge.
// Uses a constant-time compare.
func verifyPKCE(verifier, challenge string) bool {
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

// keyID derives a stable kid from the public key bytes.
func keyID(pub *rsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", fmt.Errorf("marshal pub: %w", err)
	}
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:8]), nil //nolint:mnd
}

func loadOrGenerateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		// Development / ad-hoc mode: generate an ephemeral key. Tokens
		// minted before a restart become unredeemable, which is fine for
		// dev — the authorize flow is interactive and short-lived.
		key, err := rsa.GenerateKey(rand.Reader, oauthKeyBits)
		if err != nil {
			return nil, fmt.Errorf("generate ephemeral rsa key: %w", err)
		}
		return key, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			key, genErr := rsa.GenerateKey(rand.Reader, oauthKeyBits)
			if genErr != nil {
				return nil, fmt.Errorf("generate key: %w", genErr)
			}
			if writeErr := writePrivateKeyPEM(path, key); writeErr != nil {
				return nil, fmt.Errorf("persist generated key: %w", writeErr)
			}
			return key, nil
		}
		return nil, fmt.Errorf("read key file: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("oauth private key PEM could not be decoded")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, parseErr := x509.ParsePKCS1PrivateKey(block.Bytes)
		if parseErr != nil {
			return nil, fmt.Errorf("parse pkcs1 key: %w", parseErr)
		}
		return key, nil
	case "PRIVATE KEY":
		parsed, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
		if parseErr != nil {
			return nil, fmt.Errorf("parse pkcs8 key: %w", parseErr)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("oauth private key is not RSA")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}

func writePrivateKeyPEM(path string, key *rsa.PrivateKey) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, oauthKeyDirPerm); err != nil {
			return fmt.Errorf("mkdir key dir: %w", err)
		}
	}
	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, oauthKeyFilePerm)
	if err != nil {
		return fmt.Errorf("open key file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if err := pem.Encode(file, block); err != nil {
		return fmt.Errorf("encode pem: %w", err)
	}
	return nil
}

// derivePRMURL returns `<origin>/.well-known/oauth-protected-resource` for
// the resource URL. Using only the origin avoids the bug where appending
// to the full resource path (e.g. https://mcp.escape.tech/mcp) produces
// an incorrect discovery URL under /mcp/.well-known/....
func derivePRMURL(resource string) (string, error) {
	origin, err := originOf(resource)
	if err != nil {
		return "", err
	}
	return origin + "/.well-known/oauth-protected-resource", nil
}

func originOf(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("missing scheme or host in %q", raw)
	}
	return u.Scheme + "://" + u.Host, nil
}

// writeJSON marshals `body` as JSON and writes it with the given status.
// When setNoCache is true also sets Cache-Control/Pragma headers before
// writing (required for OAuth token endpoint responses).
func writeJSON(w http.ResponseWriter, status int, body any, setNoCache bool) {
	if setNoCache {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
	}
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(body)
	if err != nil {
		// JSON marshal errors for these structs are programmer errors;
		// fall back to a hardcoded body so we still terminate the HTTP
		// response cleanly.
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"server_error"}`)
		return
	}
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

// redirectAllowlist is a compiled host matcher shared between the
// authorize page (frontend, defense layer 1) and the token endpoint
// (this file, defense layer 2 — the credential-exposing one).
type redirectAllowlist struct {
	allowedHosts []allowedHost
}

type allowedHost struct {
	scheme    string
	host      string
	wildcard  bool
	loopback  bool
	allowPort bool
}

// buildRedirectAllowlist constructs the compiled allowlist. `extras` are
// HTTPS-only hostnames (or `*.host`) passed via ESCAPE_MCP_EXTRA_REDIRECT_HOSTS
// for staging/QA. Keep in sync with services/api/src/lib/oauth/redirect-allowlist.ts
// and services/frontend/src/routes/oauth/mcp/_lib/allowlist.ts (Anthropic +
// Cowork + Cursor + OpenAI/ChatGPT + Continue.dev + Zed + Windsurf/Codeium,
// each as apex + wildcard subdomains, HTTPS only).
func buildRedirectAllowlist(extras []string) *redirectAllowlist {
	hosts := []allowedHost{
		{scheme: "http", host: "127.0.0.1", loopback: true, allowPort: true},
		{scheme: "http", host: "localhost", loopback: true, allowPort: true},
		{scheme: "https", host: "claude.ai"},
		{scheme: "https", host: "anthropic.com", wildcard: true},
		{scheme: "https", host: "cowork.ai"},
		{scheme: "https", host: "cowork.ai", wildcard: true},
		{scheme: "https", host: "cursor.com"},
		{scheme: "https", host: "cursor.com", wildcard: true},
		{scheme: "https", host: "cursor.sh"},
		{scheme: "https", host: "cursor.sh", wildcard: true},
		{scheme: "https", host: "openai.com"},
		{scheme: "https", host: "openai.com", wildcard: true},
		{scheme: "https", host: "chatgpt.com"},
		{scheme: "https", host: "chatgpt.com", wildcard: true},
		{scheme: "https", host: "continue.dev"},
		{scheme: "https", host: "continue.dev", wildcard: true},
		{scheme: "https", host: "zed.dev"},
		{scheme: "https", host: "zed.dev", wildcard: true},
		{scheme: "https", host: "windsurf.com"},
		{scheme: "https", host: "windsurf.com", wildcard: true},
		{scheme: "https", host: "codeium.com"},
		{scheme: "https", host: "codeium.com", wildcard: true},
	}
	for _, raw := range extras {
		host := strings.TrimSpace(raw)
		if host == "" {
			continue
		}
		// Tolerate scheme prefixes and paths so operators can paste a
		// URL. We always compile as HTTPS (enforced at match time).
		host = strings.TrimPrefix(host, "https://")
		host = strings.TrimPrefix(host, "http://")
		if idx := strings.IndexAny(host, "/?#"); idx >= 0 {
			host = host[:idx]
		}
		wildcard := false
		if strings.HasPrefix(host, "*.") {
			wildcard = true
			host = strings.TrimPrefix(host, "*.")
		}
		if host == "" {
			continue
		}
		hosts = append(hosts, allowedHost{scheme: "https", host: host, wildcard: wildcard})
	}
	return &redirectAllowlist{allowedHosts: hosts}
}

// allow returns true when the URI is permitted by the allowlist.
// Rejects non-http(s) schemes, userinfo, fragments, and unrecognized hosts.
func (a *redirectAllowlist) allow(raw string) bool {
	if a == nil || raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.User != nil || u.Fragment != "" {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}
	// net/url.Hostname() already strips brackets from IPv6 literals in
	// Go ≥1.5, but normalize defensively so the matcher stays identical
	// in shape to the TS/Node peer (where URL.hostname *does* preserve
	// brackets).
	host := stripBrackets(strings.ToLower(u.Hostname()))
	if host == "" {
		return false
	}
	port := u.Port()

	for _, allowed := range a.allowedHosts {
		if allowed.scheme != scheme {
			continue
		}
		if !allowed.allowPort && port != "" {
			continue
		}
		if allowed.loopback {
			if host == allowed.host {
				return true
			}
			// Also accept IPv6 loopback.
			if allowed.host == "127.0.0.1" && host == "::1" {
				return true
			}
			continue
		}
		if allowed.wildcard {
			if strings.HasSuffix(host, "."+allowed.host) {
				return true
			}
			continue
		}
		if host == allowed.host {
			return true
		}
	}
	return false
}

func stripBrackets(host string) string {
	if len(host) >= 2 && host[0] == '[' && host[len(host)-1] == ']' { //nolint:mnd
		return host[1 : len(host)-1]
	}
	return host
}

// IsLoopbackHost is a small helper exposed for tests.
func IsLoopbackHost(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return host == "localhost"
	}
	return ip.IsLoopback()
}

// asJWTClaims provides a path to decode as standard JWT claims if we ever
// need to cross-check. Currently unused — reserved so the go-jose/v4 import
// doesn't get pruned when oauth_test.go is excluded from non-test builds.
var _ = jwt.Expected{}

// Ensure crypto.Hash is referenced so go build doesn't drop the import on
// toolchains where the linker is over-eager. (defensive; inexpensive)
var _ = crypto.SHA256
