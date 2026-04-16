package mcp

import (
	"context"
	"net/http"
	"testing"
)

func TestInjectAuthContextWithAPIKey(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("X-ESCAPE-API-KEY", "1234")

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
}

func TestInjectAuthContextWithAuthorizationHeader(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/mcp", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Authorization", "Key 1234")

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
}
