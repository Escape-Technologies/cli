package mcp

import (
	"context"
	"net/http"
	"testing"
)

func TestInjectAuthContextWithAPIKey(t *testing.T) {
	t.Parallel()

	req := buildTestRequest(t, "X-ESCAPE-API-KEY", "1234")

	auth, err := AuthFromContext(InjectAuthContext(context.Background(), req))
	if err != nil {
		t.Fatalf("expected auth context, got error: %v", err)
	}

	if auth.APIKey != "1234" {
		t.Fatalf("expected api key 1234, got %q", auth.APIKey)
	}
	if auth.Authorization != "" {
		t.Fatalf("expected empty authorization, got %q", auth.Authorization)
	}
	if auth.Method != AuthMethodAPIKeyHeader {
		t.Fatalf("expected method %q, got %q", AuthMethodAPIKeyHeader, auth.Method)
	}
}

func TestInjectAuthContextWithKeyAuthorizationHeader(t *testing.T) {
	t.Parallel()

	req := buildTestRequest(t, "Authorization", "Key 1234")

	auth, err := AuthFromContext(InjectAuthContext(context.Background(), req))
	if err != nil {
		t.Fatalf("expected auth context, got error: %v", err)
	}

	if auth.APIKey != "1234" {
		t.Fatalf("expected api key 1234, got %q", auth.APIKey)
	}
	if auth.Authorization != "Key 1234" {
		t.Fatalf("expected authorization header to be preserved, got %q", auth.Authorization)
	}
	if auth.Method != AuthMethodAuthorizationKey {
		t.Fatalf("expected method %q, got %q", AuthMethodAuthorizationKey, auth.Method)
	}
}

// TestInjectAuthContextWithBearerAuthorizationHeader is the regression that
// Codex flagged as a critical blocker: a raw Bearer header passed to the
// child CLI would be forwarded via ESCAPE_AUTHORIZATION and beat the
// normalized API key, which in turn would 401 against the backend API
// (which treats Bearer as a JWT). We must extract the api key from the
// Bearer header AND clear the Authorization field so only ESCAPE_API_KEY
// propagates downstream.
func TestInjectAuthContextWithBearerAuthorizationHeader(t *testing.T) {
	t.Parallel()

	req := buildTestRequest(t, "Authorization", "Bearer my-api-key")

	auth, err := AuthFromContext(InjectAuthContext(context.Background(), req))
	if err != nil {
		t.Fatalf("expected auth context, got error: %v", err)
	}

	if auth.APIKey != "my-api-key" {
		t.Fatalf("expected api key my-api-key, got %q", auth.APIKey)
	}
	if auth.Authorization != "" {
		t.Fatalf(
			"expected authorization to be cleared to prevent leak into child CLI, got %q",
			auth.Authorization,
		)
	}
	if auth.Method != AuthMethodAuthorizationBearer {
		t.Fatalf("expected method %q, got %q", AuthMethodAuthorizationBearer, auth.Method)
	}
}

func TestInjectAuthContextBearerCaseInsensitive(t *testing.T) {
	t.Parallel()

	for _, scheme := range []string{"Bearer", "bearer", "BEARER", "BeArEr"} {
		scheme := scheme
		t.Run(scheme, func(t *testing.T) {
			t.Parallel()
			req := buildTestRequest(t, "Authorization", scheme+" my-api-key")
			auth, err := AuthFromContext(InjectAuthContext(context.Background(), req))
			if err != nil {
				t.Fatalf("expected auth context, got error: %v", err)
			}
			if auth.APIKey != "my-api-key" {
				t.Fatalf("expected api key my-api-key, got %q", auth.APIKey)
			}
			if auth.Method != AuthMethodAuthorizationBearer {
				t.Fatalf("expected bearer method, got %q", auth.Method)
			}
		})
	}
}

// TestInjectAuthContextXEscapeAPIKeyTakesPrecedenceOverBearer documents the
// explicit precedence rule: when both headers are present, X-ESCAPE-API-KEY
// wins AND the losing Authorization header is dropped. If we kept the raw
// Bearer header around, executor.go would forward it as
// ESCAPE_AUTHORIZATION and env/key.go would prefer that over the winning
// API key, so a mixed-header request would still 401 the backend.
func TestInjectAuthContextXEscapeAPIKeyTakesPrecedenceOverBearer(t *testing.T) {
	t.Parallel()

	req := buildTestRequest(t, "X-ESCAPE-API-KEY", "from-header")
	req.Header.Set("Authorization", "Bearer from-bearer")

	auth, err := AuthFromContext(InjectAuthContext(context.Background(), req))
	if err != nil {
		t.Fatalf("expected auth context, got error: %v", err)
	}

	if auth.APIKey != "from-header" {
		t.Fatalf("expected X-ESCAPE-API-KEY to win, got %q", auth.APIKey)
	}
	if auth.Authorization != "" {
		t.Fatalf(
			"expected losing Authorization to be cleared to prevent leak into child CLI, got %q",
			auth.Authorization,
		)
	}
	if auth.Method != AuthMethodAPIKeyHeader {
		t.Fatalf("expected method %q, got %q", AuthMethodAPIKeyHeader, auth.Method)
	}
}

func TestInjectAuthContextMissingCredentialsMethodNone(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	ctx := InjectAuthContext(context.Background(), req)

	// AuthFromContext must surface the missing-credentials error for the
	// existing defense-in-depth path.
	if _, err := AuthFromContext(ctx); err == nil {
		t.Fatalf("expected missing-credentials error")
	}

	// But the Method must still be captured so the middleware can log
	// the "none" bucket for telemetry.
	raw, ok := ctx.Value(authContextKey{}).(Auth)
	if !ok {
		t.Fatalf("expected Auth struct in context")
	}
	if raw.Method != AuthMethodNone {
		t.Fatalf("expected method %q, got %q", AuthMethodNone, raw.Method)
	}
}

func buildTestRequest(t *testing.T, header, value string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set(header, value)
	return req
}
