package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
)

func TestProfileUpdateForwardsNullsAndMergesFlags(t *testing.T) {
	const profileID = "00000000-0000-0000-0000-000000000001"
	var raw []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v3/profiles/"+profileID {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}

		raw, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
		  "id": "`+profileID+`",
		  "name": "Renamed",
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
		}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
	t.Setenv("ESCAPE_AUTHORIZATION", "")
	if err := out.SetOutput("json"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })
	out.SetInputSchema(false)

	for _, name := range []string{"name", "description", "cron", "off-limits-areas", "additional-context", "app-description"} {
		resetFlag(t, profileUpdateCmd, name)
	}

	if err := profileUpdateCmd.Flags().Set("name", "Renamed"); err != nil {
		t.Fatalf("set name: %v", err)
	}

	profileUpdateCmd.SetIn(strings.NewReader(`{"proxyId":null,"offLimitsAreas":null,"context":{"additional":null,"appDescription":"keep"}}`))
	profileUpdateCmd.SetOut(io.Discard)
	profileUpdateCmd.SetErr(io.Discard)
	profileUpdateCmd.SetContext(context.Background())

	if err := profileUpdateCmd.RunE(profileUpdateCmd, []string{profileID}); err != nil {
		t.Fatalf("update: %v", err)
	}

	wire := string(raw)
	for _, want := range []string{`"proxyId":null`, `"offLimitsAreas":null`, `"additional":null`, `"appDescription":"keep"`, `"name":"Renamed"`} {
		if !strings.Contains(wire, want) {
			t.Fatalf("body missing %s: %s", want, wire)
		}
	}

	if strings.Contains(wire, `"cron"`) || strings.Contains(wire, `"description"`) {
		t.Fatalf("unset field was sent: %s", wire)
	}
}

func TestProfileUpdateOmitsFieldsWhoseFlagsAreUnset(t *testing.T) {
	const profileID = "00000000-0000-0000-0000-000000000001"
	var raw []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
		  "id": "`+profileID+`",
		  "name": "Only",
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
		}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
	t.Setenv("ESCAPE_AUTHORIZATION", "")
	if err := out.SetOutput("json"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })
	out.SetInputSchema(false)

	for _, name := range []string{"name", "description", "cron", "off-limits-areas", "additional-context"} {
		resetFlag(t, profileUpdateCmd, name)
	}

	if err := profileUpdateCmd.Flags().Set("name", "Only"); err != nil {
		t.Fatalf("set name: %v", err)
	}

	profileUpdateCmd.SetIn(strings.NewReader(""))
	profileUpdateCmd.SetOut(io.Discard)
	profileUpdateCmd.SetErr(io.Discard)
	profileUpdateCmd.SetContext(context.Background())

	if err := profileUpdateCmd.RunE(profileUpdateCmd, []string{profileID}); err != nil {
		t.Fatalf("update: %v", err)
	}

	wire := string(raw)
	if !strings.Contains(wire, `"name":"Only"`) {
		t.Fatalf("body missing name: %s", wire)
	}

	for _, absent := range []string{`"proxyId"`, `"offLimitsAreas"`, `"context"`, `"cron"`, `"description"`} {
		if strings.Contains(wire, absent) {
			t.Fatalf("unset %s was sent: %s", absent, wire)
		}
	}
}
