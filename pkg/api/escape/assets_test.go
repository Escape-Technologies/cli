package escape

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestListAssetsSendsAllFilters(t *testing.T) {
	var query url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"nextCursor":null,"totalCount":0}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	_, _, _, err := ListAssets(context.Background(), "", &ListAssetsFilters{
		ProjectIDs:     []string{"project-1", "project-2"},
		TagIDs:         []string{"tag-1"},
		Classes:        []string{"FRONTEND", "HOST"},
		Environments:   []string{"PRODUCTION"},
		Domains:        []string{"example.com"},
		IntegrationIDs: []string{"integration-1"},
		OwnerEmails:    []string{"owner@example.com"},
		TechnologyKeys: []string{"npm/lodash"},
		Ports:          []string{"80", "443"},
		Severities:     []string{"HIGH"},
		Risks:          []string{"EXPOSED"},
		Frameworks:     []string{"REST_GIN"},
		DNF:            `{"operator":"AND"}`,
	}, 5)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	want := map[string]string{
		"projectIds":     "project-1,project-2",
		"tagIds":         "tag-1",
		"environments":   "PRODUCTION",
		"domains":        "example.com",
		"integrationIds": "integration-1",
		"ownerEmails":    "owner@example.com",
		"technologyKeys": "npm/lodash",
		"ports":          "80,443",
		"severities":     "HIGH",
		"risks":          "EXPOSED",
		"frameworks":     "REST_GIN",
		"dnf":            `{"operator":"AND"}`,
	}
	for key, value := range want {
		if got := query.Get(key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}

	if got := query["classes"]; !reflect.DeepEqual(got, []string{"FRONTEND", "HOST"}) {
		t.Errorf("classes = %v, want [FRONTEND HOST]", got)
	}
}

func TestGetAssetContentReturnsDownloadMetadata(t *testing.T) {
	const assetID = "00000000-0000-0000-0000-000000000001"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/assets/"+assetID+"/content" {
			t.Errorf("path = %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"signedUrl":"https://download.example/schema","contentType":"application/json","filename":"schema.json"}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	content, err := GetAssetContent(context.Background(), assetID)
	if err != nil {
		t.Fatalf("get asset content: %v", err)
	}

	if content.GetSignedUrl() != "https://download.example/schema" ||
		content.GetContentType() != "application/json" ||
		content.GetFilename() != "schema.json" {
		t.Fatalf("content = %+v", content)
	}
}

func errorResult(err error) reflect.Value {
	wrapped := err
	return reflect.ValueOf(&wrapped).Elem()
}

func TestUnpackExecuteReturnsBody(t *testing.T) {
	t.Parallel()
	want := &v3.UpdateAsset200Response{Id: "asset-1"}
	got, err := unpackExecute([]reflect.Value{reflect.ValueOf(want), errorResult(nil)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestUnpackExecuteSurfacesAPIError(t *testing.T) {
	t.Parallel()
	_, err := unpackExecute([]reflect.Value{
		reflect.ValueOf((*v3.UpdateAsset200Response)(nil)),
		errorResult(errors.New("400 Bad Request")),
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "400 Bad Request") {
		t.Fatalf("expected api error, got %v", err)
	}
}

func TestUnpackExecuteRejectsTypedNilBody(t *testing.T) {
	t.Parallel()
	_, err := unpackExecute([]reflect.Value{
		reflect.ValueOf((*v3.UpdateAsset200Response)(nil)),
		errorResult(nil),
	})
	if err == nil {
		t.Fatal("expected empty-response error")
	}

	if !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("expected empty response, got %v", err)
	}
}

func TestCreateAssetSchemaRequestRoundTrip(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"asset_type":"SCHEMA","name":"solar-system-openapi","fetch":{"url":"https://tester.tools.escape.tech/solar_system/openapi.json"}}`)
	var payload v3.CreateAssetSchemaRequest
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if payload.CreateSchemaViaFetch == nil {
		t.Fatal("expected fetch variant")
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(encoded), `"url":"https://tester.tools.escape.tech/solar_system/openapi.json"`) {
		t.Fatalf("missing fetch url in %s", encoded)
	}
}
