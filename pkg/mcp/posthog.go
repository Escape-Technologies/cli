package mcp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	posthogAPIKeyEnv     = "POSTHOG_API_KEY"
	posthogHostEnv       = "POSTHOG_HOST"
	posthogDeploymentEnv = "ESCAPE_CLI_POSTHOG_DEPLOYMENT"
	// phc_ project keys are public write-only keys. Same project as public-api.
	// Committed so every MCP process reports, including local `escape mcp serve`.
	posthogDefaultAPIKey     = "phc_I2jCMr0uUAnrp5qQXDWnwU8jcQa4yAnGjW1VLGT2VoA"
	posthogDefaultHost       = "https://b.escape.tech"
	posthogDeploymentLocal   = "local"
	posthogDeploymentStaging = "staging"
	posthogDeploymentProd    = "prod"
	posthogCapturePath       = "/capture/"
	posthogEvent             = "mcp_request"
	posthogLib               = "escape-mcp"
	posthogIdentityPath      = "/v3/me"
	identityCacheTTL         = 5 * time.Minute
	maxIdentityCacheSize     = 10_000
	posthogCaptureTimeout    = 5 * time.Second
	posthogIdentityTimeout   = 5 * time.Second
	// Hosted MCP is a single 256Mi replica. Each event can block on /v3/me
	// and then on PostHog. A fixed worker pool plus a short queue drops the
	// overflow instead of spawning a goroutine per request.
	posthogWorkers = 4
	posthogQueue   = 64
)

// posthogIdentity is the user/org resolved from GET /v3/me.
type posthogIdentity struct {
	UserID           string
	UserEmail        string
	OrganizationID   string
	OrganizationName string
}

// posthogPeek holds the JSON-RPC fields we are allowed to send to PostHog.
// Tool arguments and response bodies are never captured.
type posthogPeek struct {
	Method        string
	Tool          string
	ClientName    string
	ClientVersion string
}

type identityCacheEntry struct {
	identity  posthogIdentity
	expiresAt time.Time
}

// posthogCapture sends mcp_request events to PostHog for every MCP process.
// The project key is committed; POSTHOG_API_KEY overrides it.
type posthogCapture struct {
	apiKey       string
	host         string
	deployment   string
	publicAPIURL string
	client       *http.Client

	cacheMu sync.Mutex
	cache   map[string]identityCacheEntry

	// queue is never closed; workers live for the process.
	queue chan posthogEventPayload
}

// posthogAPIKey returns POSTHOG_API_KEY when set, otherwise the committed project key.
func posthogAPIKey() string {
	key := strings.TrimSpace(os.Getenv(posthogAPIKeyEnv))
	if key == "" {
		return posthogDefaultAPIKey
	}

	return key
}

// posthogDeployment is the PostHog label for where this process runs.
// Hosted pods set ESCAPE_CLI_POSTHOG_DEPLOYMENT to staging or prod.
// Unset or any other value is local.
func posthogDeployment() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(posthogDeploymentEnv))) {
	case posthogDeploymentStaging:
		return posthogDeploymentStaging
	case posthogDeploymentProd:
		return posthogDeploymentProd
	default:
		return posthogDeploymentLocal
	}
}

func newPosthogCapture(ctx context.Context, publicAPIURL string) *posthogCapture {
	host := strings.TrimSpace(os.Getenv(posthogHostEnv))
	if host == "" {
		host = posthogDefaultHost
	}

	capture := &posthogCapture{
		apiKey:       posthogAPIKey(),
		host:         strings.TrimRight(host, "/"),
		deployment:   posthogDeployment(),
		publicAPIURL: strings.TrimRight(publicAPIURL, "/"),
		client:       &http.Client{Timeout: posthogCaptureTimeout},
		cache:        make(map[string]identityCacheEntry),
		queue:        make(chan posthogEventPayload, posthogQueue),
	}
	for range posthogWorkers {
		go capture.loop(ctx)
	}

	return capture
}

// wrapWithPosthog peeks the JSON-RPC envelope on POST /mcp and fires a
// fire-and-forget mcp_request event after the handler returns.
// ctx is the server lifetime; workers stop when it ends.
func wrapWithPosthog(ctx context.Context, next http.Handler, publicAPIURL string) http.Handler {
	capture := newPosthogCapture(ctx, publicAPIURL)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			next.ServeHTTP(w, req)
			return
		}

		peek, body, err := peekPosthogBody(req)
		if err != nil {
			// IntentModeOff skips the intent middleware, and mcp-go reads the
			// body with no size limit. Put back whatever we already consumed
			// so a failed peek does not turn a valid request into an empty body.
			req.Body = &bodyRestorer{
				Reader: io.MultiReader(bytes.NewReader(body), req.Body),
				closer: req.Body,
			}
			next.ServeHTTP(w, req)

			return
		}

		req.Body = io.NopCloser(bytes.NewReader(body))

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, req)

		fingerprint := credentialFingerprint(req)
		authType := authMethodFromRequest(req)
		capture.enqueue(posthogEventPayload{
			peek:        peek,
			status:      recorder.status,
			authType:    authType,
			fingerprint: fingerprint,
			userAgent:   req.Header.Get("User-Agent"),
			apiKey:      apiKeyFromRequest(req),
		})
	})
}

// bodyRestorer replays bytes already read and still closes the original body.
type bodyRestorer struct {
	io.Reader
	closer io.Closer
}

func (b *bodyRestorer) Close() error {
	if b.closer == nil {
		return nil
	}

	if err := b.closer.Close(); err != nil {
		return fmt.Errorf("close request body: %w", err)
	}

	return nil
}

func (c *posthogCapture) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-c.queue:
			c.send(ctx, payload)
		}
	}
}

// enqueue drops the event when every worker is busy and the queue is full.
func (c *posthogCapture) enqueue(payload posthogEventPayload) bool {
	select {
	case c.queue <- payload:
		return true
	default:
		slog.Warn("posthog: dropping mcp_request, telemetry queue full")

		return false
	}
}

type posthogEventPayload struct {
	peek        posthogPeek
	status      int
	authType    AuthMethod
	fingerprint string
	userAgent   string
	apiKey      string
}

func (c *posthogCapture) send(ctx context.Context, payload posthogEventPayload) {
	id := c.resolveIdentity(ctx, payload.apiKey, payload.fingerprint)

	distinctID := "anonymous"
	switch {
	case id != nil && id.UserID != "":
		distinctID = id.UserID
	case payload.fingerprint != "":
		distinctID = payload.fingerprint
	}

	properties := map[string]any{
		"$lib":        posthogLib,
		"deployment":  c.deployment,
		"method":      payload.peek.Method,
		"status":      payload.status,
		"success":     payload.status < http.StatusBadRequest,
		"authType":    string(payload.authType),
		"fingerprint": payload.fingerprint,
		"userAgent":   payload.userAgent,
	}
	if payload.peek.Tool != "" {
		properties["tool"] = payload.peek.Tool
	}

	if payload.peek.ClientName != "" {
		properties["clientName"] = payload.peek.ClientName
	}

	if payload.peek.ClientVersion != "" {
		properties["clientVersion"] = payload.peek.ClientVersion
	}

	if id != nil {
		properties["userId"] = id.UserID
		properties["userEmail"] = id.UserEmail
		properties["organizationId"] = id.OrganizationID
		properties["organizationName"] = id.OrganizationName
	}

	body, err := json.Marshal(map[string]any{
		"api_key":     c.apiKey,
		"event":       posthogEvent,
		"distinct_id": distinctID,
		"properties":  properties,
	})
	if err != nil {
		slog.Warn("posthog: marshal capture payload", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.host+posthogCapturePath,
		bytes.NewReader(body),
	)
	if err != nil {
		slog.Warn("posthog: build capture request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Warn("posthog: capture request failed", "error", err)
		return
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= http.StatusBadRequest {
		slog.Warn("posthog: capture rejected", "status", resp.StatusCode)
	}
}

func (c *posthogCapture) resolveIdentity(
	ctx context.Context,
	apiKey, fingerprint string,
) *posthogIdentity {
	if apiKey == "" || c.publicAPIURL == "" {
		return nil
	}

	cacheKey := fingerprint
	if cacheKey == "" {
		sum := sha256.Sum256([]byte(apiKey))
		cacheKey = hex.EncodeToString(sum[:])
	}

	now := time.Now()
	c.cacheMu.Lock()
	if entry, ok := c.cache[cacheKey]; ok && entry.expiresAt.After(now) {
		identity := entry.identity
		c.cacheMu.Unlock()

		return &identity
	}

	c.pruneLocked(now)
	c.cacheMu.Unlock()

	identity, err := c.fetchIdentity(ctx, apiKey)
	if err != nil || identity == nil {
		return nil
	}

	c.cacheMu.Lock()
	c.cache[cacheKey] = identityCacheEntry{
		identity:  *identity,
		expiresAt: now.Add(identityCacheTTL),
	}
	c.cacheMu.Unlock()

	return identity
}

func (c *posthogCapture) fetchIdentity(ctx context.Context, apiKey string) (*posthogIdentity, error) {
	reqCtx, cancel := context.WithTimeout(ctx, posthogIdentityTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		reqCtx,
		http.MethodGet,
		c.publicAPIURL+posthogIdentityPath,
		http.NoBody,
	)
	if err != nil {
		return nil, fmt.Errorf("build identity request: %w", err)
	}

	req.Header.Set("Authorization", "Key "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("identity request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil
	}

	var parsed struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
		Organization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organization"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode identity response: %w", err)
	}

	if parsed.User.ID == "" {
		return nil, nil
	}

	return &posthogIdentity{
		UserID:           parsed.User.ID,
		UserEmail:        parsed.User.Email,
		OrganizationID:   parsed.Organization.ID,
		OrganizationName: parsed.Organization.Name,
	}, nil
}

func (c *posthogCapture) pruneLocked(now time.Time) {
	for key, entry := range c.cache {
		if entry.expiresAt.Before(now) || entry.expiresAt.Equal(now) {
			delete(c.cache, key)
		}
	}

	if len(c.cache) <= maxIdentityCacheSize {
		return
	}

	overflow := len(c.cache) - maxIdentityCacheSize
	i := 0
	for key := range c.cache {
		delete(c.cache, key)
		i++
		if i >= overflow {
			break
		}
	}
}

func peekPosthogBody(req *http.Request) (posthogPeek, []byte, error) {
	// LimitReader avoids MaxBytesReader's nil-ResponseWriter panic on overflow.
	limited := io.LimitReader(req.Body, maxRequestBodyBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		// Leave req.Body open so the caller can replay body plus the remainder.
		return posthogPeek{}, body, fmt.Errorf("read mcp request body: %w", err)
	}

	if len(body) > maxRequestBodyBytes {
		return posthogPeek{}, body, errors.New("request body too large")
	}

	_ = req.Body.Close()

	var peek struct {
		Method string `json:"method"`
		Params struct {
			Name       string `json:"name"`
			ClientInfo struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"clientInfo"`
		} `json:"params"`
	}
	// Malformed JSON still proceeds; we just leave method empty.
	_ = json.Unmarshal(body, &peek)

	out := posthogPeek{Method: peek.Method}
	if peek.Method == "tools/call" {
		out.Tool = peek.Params.Name
	}

	if peek.Method == "initialize" {
		out.ClientName = peek.Params.ClientInfo.Name
		out.ClientVersion = peek.Params.ClientInfo.Version
	}

	return out, body, nil
}

// credentialFingerprint hashes the API key sent to /v3/me.
// Precedence matches apiKeyFromRequest: X-ESCAPE-API-KEY wins, so a stale
// Authorization header cannot make two keys share one identity cache entry.
func credentialFingerprint(req *http.Request) string {
	credential := apiKeyFromRequest(req)
	if credential == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(credential))

	return hex.EncodeToString(sum[:])
}

func authMethodFromRequest(req *http.Request) AuthMethod {
	return authMethodFromContext(InjectAuthContext(req.Context(), req))
}

// statusRecorder records the HTTP status without buffering the response body.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}

	r.wroteHeader = true
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	n, err := r.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("write response: %w", err)
	}

	return n, nil
}
