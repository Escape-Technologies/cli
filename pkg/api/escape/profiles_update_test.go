package escape

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const profileUpdateFixture = `{
  "id": "00000000-0000-0000-0000-000000000001",
  "name": "profile",
  "description": "",
  "scannerKind": "BLST_REST",
  "createdAt": "2026-01-01T00:00:00Z",
  "updatedAt": "2026-01-01T00:00:00Z",
  "configuration": {},
  "extraAssets": [],
  "asset": {
    "id": "00000000-0000-0000-0000-000000000002",
    "class": "API_SERVICE",
    "type": "REST",
    "name": "api",
    "createdAt": "2026-01-01T00:00:00Z",
    "lastSeenAt": "2026-01-01T00:00:00Z",
    "status": "MONITORED",
    "tags": [],
    "risks": [],
    "projectIds": [],
    "links": {"assetOverview": "https://example.com/asset"}
  },
  "risks": [],
  "statistics": {"issues": {"all": 0, "severities": [], "categories": []}},
  "links": {"profileSummary": "https://example.com/profile"}
}`

func TestUpdateProfilePreservesExplicitNulls(t *testing.T) {
	const profileID = "00000000-0000-0000-0000-000000000001"
	body := []byte(`{"name":"Renamed","proxyId":null,"offLimitsAreas":null,"context":{"additional":null,"appDescription":"keep"}}`)

	var raw []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v3/profiles/"+profileID {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}

		raw, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, profileUpdateFixture)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	profile, err := UpdateProfile(context.Background(), profileID, body)
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}

	if profile.GetId() != profileID {
		t.Fatalf("id = %q", profile.GetId())
	}

	wire := string(raw)
	for _, want := range []string{`"proxyId":null`, `"offLimitsAreas":null`, `"additional":null`, `"appDescription":"keep"`, `"name":"Renamed"`} {
		if !strings.Contains(wire, want) {
			t.Fatalf("body missing %s: %s", want, wire)
		}
	}

	if strings.Contains(wire, `"cron"`) {
		t.Fatalf("unset cron was sent: %s", wire)
	}
}

func TestUpdateProfileUnauthorizedIsInvalidAPIKey(t *testing.T) {
	const profileID = "00000000-0000-0000-0000-000000000001"
	body := []byte(`{"name":"renamed","proxyId":null}`)

	var raw []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"message":"Not authorized."}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	_, err := UpdateProfile(context.Background(), profileID, body)
	if !IsInvalidAPIKey(err) {
		t.Fatalf("IsInvalidAPIKey(%v) = false", err)
	}

	if !strings.Contains(err.Error(), InvalidAPIKeyMessage) {
		t.Fatalf("message = %q, want %q", err.Error(), InvalidAPIKeyMessage)
	}

	if !strings.Contains(string(raw), `"name":"renamed"`) || !strings.Contains(string(raw), `"proxyId":null`) {
		t.Fatalf("body = %s", raw)
	}
}

func TestUpdateProfileBadRequestSurfacesAPIMessage(t *testing.T) {
	const profileID = "00000000-0000-0000-0000-000000000001"
	const apiMessage = "Bad Request: name: must not be empty"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"message":"Bad Request","details":"name: must not be empty"}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	_, err := UpdateProfile(context.Background(), profileID, []byte(`{"name":"renamed"}`))
	if err == nil {
		t.Fatal("expected an error")
	}

	if IsInvalidAPIKey(err) {
		t.Fatalf("400 classified as an invalid API key: %v", err)
	}

	if !strings.Contains(err.Error(), apiMessage) {
		t.Fatalf("message = %q, want %q", err.Error(), apiMessage)
	}
}

func TestUpdateProfileRejectsABodyTheModelRejects(t *testing.T) {
	t.Setenv("ESCAPE_API_URL", "http://127.0.0.1:1")
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	_, err := UpdateProfile(context.Background(), "profile-1", []byte(`{"proxyId":1}`))
	if err == nil || !strings.Contains(err.Error(), "invalid JSON for UpdateProfileRequest") {
		t.Fatalf("err = %v, want a model validation error", err)
	}
}
