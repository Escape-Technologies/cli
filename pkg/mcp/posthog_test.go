package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPosthogCaptureUsesCommittedKeyAndLocalDeployment(t *testing.T) {
	t.Setenv(posthogAPIKeyEnv, "")
	t.Setenv(posthogHostEnv, "")
	t.Setenv(posthogDeploymentEnv, "")

	got := newPosthogCapture(t.Context(), "https://public.example")
	if got == nil {
		t.Fatal("expected capture; the project key is committed")
	}

	if got.apiKey != posthogDefaultAPIKey {
		t.Fatalf("expected committed key, got %q", got.apiKey)
	}

	if got.host != posthogDefaultHost {
		t.Fatalf("expected default host %q, got %q", posthogDefaultHost, got.host)
	}

	if got.deployment != posthogDeploymentLocal {
		t.Fatalf("expected deployment %q, got %q", posthogDeploymentLocal, got.deployment)
	}
}

func TestPosthogDeploymentLabel(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want string
	}{
		{name: "unset", env: "", want: posthogDeploymentLocal},
		{name: "blank", env: "   ", want: posthogDeploymentLocal},
		{name: "staging", env: "staging", want: posthogDeploymentStaging},
		{name: "prod", env: "prod", want: posthogDeploymentProd},
		{name: "prod case", env: "PROD", want: posthogDeploymentProd},
		{name: "unknown", env: "dev", want: posthogDeploymentLocal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(posthogDeploymentEnv, tc.env)
			if got := posthogDeployment(); got != tc.want {
				t.Fatalf("posthogDeployment() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestWrapWithPosthogPassthroughWithoutKey(t *testing.T) {
	t.Setenv(posthogAPIKeyEnv, "")
	t.Setenv(posthogDeploymentEnv, "")

	var (
		mu       sync.Mutex
		captured map[string]any
		done     = make(chan struct{})
	)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != posthogCapturePath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode capture: %v", err)
			return
		}

		mu.Lock()
		captured = payload
		mu.Unlock()
		close(done)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	t.Setenv(posthogHostEnv, upstream.URL)

	var hits atomic.Int32
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})
	handler := wrapWithPosthog(t.Context(), next, "https://public.example")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if hits.Load() != 1 {
		t.Fatalf("expected next to run once, got %d", hits.Load())
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for posthog capture")
	}

	mu.Lock()
	payload := captured
	mu.Unlock()
	if payload["api_key"] != posthogDefaultAPIKey {
		t.Fatalf("expected committed key, got %v", payload["api_key"])
	}

	props, ok := payload["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties object, got %T", payload["properties"])
	}

	if props["deployment"] != posthogDeploymentLocal {
		t.Fatalf("expected deployment %q, got %v", posthogDeploymentLocal, props["deployment"])
	}
}

func TestPeekPosthogBodyToolsCall(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"jsonrpc":"2.0",
		"id":1,
		"method":"tools/call",
		"params":{
			"name":"scans_list",
			"arguments":{"secret":"must-not-leak","profileId":"abc"}
		}
	}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", bytes.NewReader(body))

	peek, restored, err := peekPosthogBody(req)
	if err != nil {
		t.Fatalf("peek: %v", err)
	}

	if peek.Method != "tools/call" {
		t.Fatalf("expected method tools/call, got %q", peek.Method)
	}

	if peek.Tool != "scans_list" {
		t.Fatalf("expected tool scans_list, got %q", peek.Tool)
	}

	if peek.ClientName != "" || peek.ClientVersion != "" {
		t.Fatalf("expected empty client info on tools/call, got %+v", peek)
	}

	if !bytes.Equal(restored, body) {
		t.Fatal("peek must restore the original body unchanged")
	}

	// Redaction guarantee: the peek struct has no Arguments field. Confirm
	// the restored body still contains them (downstream still gets them)
	// while the peek does not expose them as a first-class field.
	if !bytes.Contains(restored, []byte(`"secret":"must-not-leak"`)) {
		t.Fatal("restored body should still carry arguments for the handler")
	}
}

func TestPeekPosthogBodyInitialize(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"jsonrpc":"2.0",
		"id":1,
		"method":"initialize",
		"params":{
			"protocolVersion":"2025-06-18",
			"clientInfo":{"name":"cursor","version":"1.2.3"},
			"capabilities":{}
		}
	}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", bytes.NewReader(body))

	peek, _, err := peekPosthogBody(req)
	if err != nil {
		t.Fatalf("peek: %v", err)
	}

	if peek.Method != "initialize" {
		t.Fatalf("expected method initialize, got %q", peek.Method)
	}

	if peek.Tool != "" {
		t.Fatalf("expected empty tool on initialize, got %q", peek.Tool)
	}

	if peek.ClientName != "cursor" || peek.ClientVersion != "1.2.3" {
		t.Fatalf("expected cursor/1.2.3, got %s/%s", peek.ClientName, peek.ClientVersion)
	}
}

func TestPosthogCaptureSendIncludesIdentityOmitsArguments(t *testing.T) {
	var (
		mu       sync.Mutex
		captured map[string]any
		meHits   atomic.Int32
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == posthogIdentityPath:
			meHits.Add(1)
			if got := r.Header.Get("Authorization"); got != "Key test-api-key" {
				t.Errorf("expected Authorization Key test-api-key, got %q", got)
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"user":         map[string]any{"id": "user-1", "email": "a@escape.tech"},
				"organization": map[string]any{"id": "org-1", "name": "Escape"},
			})
		case r.Method == http.MethodPost && r.URL.Path == posthogCapturePath:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode capture: %v", err)
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			mu.Lock()
			captured = payload
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(upstream.Close)

	t.Setenv(posthogAPIKeyEnv, "phc_test")
	t.Setenv(posthogHostEnv, upstream.URL)
	t.Setenv(posthogDeploymentEnv, "staging")

	capture := newPosthogCapture(t.Context(), upstream.URL)
	if capture == nil {
		t.Fatal("expected capture when key is set")
	}

	capture.send(context.Background(), posthogEventPayload{
		peek: posthogPeek{
			Method: "tools/call",
			Tool:   "profiles_list",
		},
		status:      http.StatusOK,
		authType:    AuthMethodAuthorizationBearer,
		fingerprint: "abc123",
		userAgent:   "Claude/1.0",
		apiKey:      "test-api-key",
	})

	mu.Lock()
	payload := captured
	mu.Unlock()
	if payload == nil {
		t.Fatal("expected a capture payload")
	}

	if payload["event"] != posthogEvent {
		t.Fatalf("expected event %q, got %v", posthogEvent, payload["event"])
	}

	if payload["distinct_id"] != "user-1" {
		t.Fatalf("expected distinct_id user-1, got %v", payload["distinct_id"])
	}

	props, ok := payload["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties object, got %T", payload["properties"])
	}

	if props["deployment"] != posthogDeploymentStaging {
		t.Fatalf("expected deployment %q, got %v", posthogDeploymentStaging, props["deployment"])
	}

	if props["method"] != "tools/call" {
		t.Fatalf("expected method tools/call, got %v", props["method"])
	}

	if props["tool"] != "profiles_list" {
		t.Fatalf("expected tool profiles_list, got %v", props["tool"])
	}

	if props["userId"] != "user-1" || props["userEmail"] != "a@escape.tech" {
		t.Fatalf("unexpected user props: %+v", props)
	}

	if props["organizationId"] != "org-1" || props["organizationName"] != "Escape" {
		t.Fatalf("unexpected org props: %+v", props)
	}

	if props["authType"] != string(AuthMethodAuthorizationBearer) {
		t.Fatalf("expected authType bearer, got %v", props["authType"])
	}

	if _, exists := props["arguments"]; exists {
		t.Fatal("capture must not include tool arguments")
	}

	if _, exists := props["requestBody"]; exists {
		t.Fatal("capture must not include requestBody")
	}

	if _, exists := props["responseBody"]; exists {
		t.Fatal("capture must not include responseBody")
	}

	if meHits.Load() != 1 {
		t.Fatalf("expected one /v3/me hit, got %d", meHits.Load())
	}

	// Second send with same fingerprint should reuse the identity cache.
	capture.send(context.Background(), posthogEventPayload{
		peek:        posthogPeek{Method: "tools/list"},
		status:      http.StatusOK,
		authType:    AuthMethodAuthorizationBearer,
		fingerprint: "abc123",
		apiKey:      "test-api-key",
	})
	if meHits.Load() != 1 {
		t.Fatalf("expected cached identity (still 1 /v3/me hit), got %d", meHits.Load())
	}
}

func TestWrapWithPosthogCapturesJSONRPC(t *testing.T) {
	var (
		mu       sync.Mutex
		captured map[string]any
		done     = make(chan struct{})
	)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == posthogIdentityPath:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user":         map[string]any{"id": "u-2", "email": "b@escape.tech"},
				"organization": map[string]any{"id": "o-2", "name": "Acme"},
			})
		case r.Method == http.MethodPost && r.URL.Path == posthogCapturePath:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode: %v", err)
				return
			}

			mu.Lock()
			captured = payload
			mu.Unlock()
			close(done)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(upstream.Close)

	t.Setenv(posthogAPIKeyEnv, "phc_test")
	t.Setenv(posthogHostEnv, upstream.URL)
	t.Setenv(posthogDeploymentEnv, "prod")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("downstream read: %v", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if !bytes.Contains(body, []byte(`"name":"scans_start"`)) {
			t.Errorf("downstream must still see the original body")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{}}`))
	})

	handler := wrapWithPosthog(t.Context(), next, upstream.URL)

	reqBody := `{
		"jsonrpc":"2.0",
		"id":7,
		"method":"tools/call",
		"params":{"name":"scans_start","arguments":{"secret":"must-not-leak"}}
	}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer test-api-key")
	req.Header.Set("User-Agent", "Cursor/0.0")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for posthog capture")
	}

	mu.Lock()
	payload := captured
	mu.Unlock()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if strings.Contains(string(raw), "must-not-leak") {
		t.Fatalf("posthog payload leaked tool arguments: %s", raw)
	}

	props, ok := payload["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties, got %T", payload["properties"])
	}

	if props["deployment"] != posthogDeploymentProd {
		t.Fatalf("expected deployment %q, got %v", posthogDeploymentProd, props["deployment"])
	}

	if props["method"] != "tools/call" || props["tool"] != "scans_start" {
		t.Fatalf("unexpected method/tool: %+v", props)
	}

	if props["status"] != float64(http.StatusOK) {
		t.Fatalf("expected status 200, got %v", props["status"])
	}

	if props["userId"] != "u-2" {
		t.Fatalf("expected identity u-2, got %v", props["userId"])
	}

	if props["authType"] != string(AuthMethodAuthorizationBearer) {
		t.Fatalf("expected bearer authType, got %v", props["authType"])
	}
}

func TestCredentialFingerprint(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	fp := credentialFingerprint(req)
	if fp == "" {
		t.Fatal("expected non-empty fingerprint")
	}

	if strings.Contains(fp, "secret-token") {
		t.Fatal("fingerprint must not contain the raw credential")
	}

	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", nil)
	req2.Header.Set("Authorization", "Bearer secret-token")
	if credentialFingerprint(req2) != fp {
		t.Fatal("same credential must produce the same fingerprint")
	}
}

func TestCredentialFingerprintPrefersAPIKeyHeader(t *testing.T) {
	t.Parallel()

	both := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", nil)
	both.Header.Set("Authorization", "Bearer stale-token")
	both.Header.Set("X-ESCAPE-API-KEY", "real-key")

	onlyKey := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", nil)
	onlyKey.Header.Set("X-ESCAPE-API-KEY", "real-key")
	if credentialFingerprint(both) != credentialFingerprint(onlyKey) {
		t.Fatal("X-ESCAPE-API-KEY must win over Authorization")
	}

	other := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", nil)
	other.Header.Set("Authorization", "Bearer stale-token")
	other.Header.Set("X-ESCAPE-API-KEY", "other-key")
	if credentialFingerprint(other) == credentialFingerprint(both) {
		t.Fatal("different API keys must not share a fingerprint")
	}
}

func TestWrapWithPosthogRestoresOversizedBody(t *testing.T) {
	t.Setenv(posthogAPIKeyEnv, "phc_test")

	body := bytes.Repeat([]byte("x"), maxRequestBodyBytes+32)
	var got []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("downstream read: %v", err)
		}

		got = b
		w.WriteHeader(http.StatusOK)
	})
	handler := wrapWithPosthog(t.Context(), next, "http://127.0.0.1:1")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !bytes.Equal(got, body) {
		t.Fatalf("downstream body length %d, want %d", len(got), len(body))
	}
}

type errAfterReader struct {
	payload []byte
	off     int
}

func (r *errAfterReader) Read(p []byte) (int, error) {
	if r.off >= len(r.payload) {
		return 0, io.ErrUnexpectedEOF
	}

	n := copy(p, r.payload[r.off:])
	r.off += n

	return n, nil
}

func (r *errAfterReader) Close() error { return nil }

func TestWrapWithPosthogRestoresBodyOnReadError(t *testing.T) {
	t.Setenv(posthogAPIKeyEnv, "phc_test")

	payload := []byte(`{"jsonrpc":"2.0","method":"tools/list"}`)
	var got []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err == nil {
			t.Error("expected the original read error to surface downstream")
		}

		got = b
		w.WriteHeader(http.StatusOK)
	})
	handler := wrapWithPosthog(t.Context(), next, "http://127.0.0.1:1")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", &errAfterReader{payload: payload})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !bytes.Equal(got, payload) {
		t.Fatalf("downstream got %q, want %q", got, payload)
	}
}

func TestEnqueueDropsWhenQueueFull(t *testing.T) {
	block := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(block) }) }

	var inCapture atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != posthogCapturePath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		inCapture.Add(1)
		<-block
		w.WriteHeader(http.StatusOK)
	}))
	// Release handlers before Close. Otherwise Close waits on connections
	// stuck in the handler, and the test never finishes.
	t.Cleanup(func() {
		release()
		upstream.Close()
	})

	t.Setenv(posthogAPIKeyEnv, "phc_test")
	t.Setenv(posthogHostEnv, upstream.URL)

	capture := newPosthogCapture(t.Context(), "http://127.0.0.1:1")
	if capture == nil {
		t.Fatal("expected capture")
	}

	payload := posthogEventPayload{peek: posthogPeek{Method: "tools/list"}}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if inCapture.Load() < int32(posthogWorkers) {
			_ = capture.enqueue(payload)
			time.Sleep(5 * time.Millisecond)
			continue
		}

		for capture.enqueue(payload) {
			if time.Now().After(deadline) {
				t.Fatal("queue never filled")
			}
		}

		if capture.enqueue(payload) {
			t.Fatal("queue accepted an event while workers were blocked and the queue was full")
		}

		return
	}

	t.Fatalf("workers never blocked, in capture=%d", inCapture.Load())
}
