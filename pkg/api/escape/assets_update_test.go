package escape

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

const assetUpdateFixture = `{
  "id": "00000000-0000-0000-0000-00000000000a",
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
}`

func captureAssetBody(t *testing.T, method, path, response string) *[]byte {
	t.Helper()

	raw := []byte{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method || r.URL.Path != path {
			t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, method, path)
		}

		raw, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	return &raw
}

func TestUpdateAssetSendsEmptyClearsAndOmitsUnsetFields(t *testing.T) {
	const assetID = "00000000-0000-0000-0000-00000000000a"
	raw := captureAssetBody(t, http.MethodPut, "/v3/assets/"+assetID, assetUpdateFixture)

	empty := ""
	none := []string{}
	status := v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS_MONITORED
	_, err := UpdateAsset(context.Background(), assetID, UpdateAssetInput{
		Description:  &empty,
		Name:         &empty,
		BaseURL:      &empty,
		Address:      &empty,
		BestSchemaID: &empty,
		Owners:       &none,
		TagIDs:       &none,
		ProjectIDs:   &none,
		SchemaIDs:    &none,
		Status:       &status,
	})
	if err != nil {
		t.Fatalf("update asset: %v", err)
	}

	wire := string(*raw)
	for _, want := range []string{
		`"description":""`,
		`"name":""`,
		`"baseUrl":""`,
		`"address":""`,
		`"bestSchemaId":""`,
		`"owners":[]`,
		`"tagIds":[]`,
		`"projectIds":[]`,
		`"schemaIds":[]`,
		`"status":"MONITORED"`,
	} {
		if !strings.Contains(wire, want) {
			t.Fatalf("body missing %s: %s", want, wire)
		}
	}

	for _, absent := range []string{`"framework"`, `"environment"`} {
		if strings.Contains(wire, absent) {
			t.Fatalf("unset %s was sent: %s", absent, wire)
		}
	}
}

func TestBulkUpdateAssetsSendsEmptyListsAndOmitsUnsetOnes(t *testing.T) {
	raw := captureAssetBody(t, http.MethodPost, "/v3/assets/bulk-update", `{"success":true}`)
	status := v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS_MONITORED
	_, err := BulkUpdateAssets(context.Background(), v3.BulkUpdateAssetsRequestWhere{
		AssetIds: []string{"00000000-0000-0000-0000-00000000000a"},
	}, []string{}, nil, &status)
	if err != nil {
		t.Fatalf("bulk update: %v", err)
	}

	wire := string(*raw)
	if !strings.Contains(wire, `"tagIds":[]`) {
		t.Fatalf("body missing empty tagIds: %s", wire)
	}

	if strings.Contains(wire, `"projectIds"`) {
		t.Fatalf("unset projectIds was sent: %s", wire)
	}

	if !strings.Contains(wire, `"status":"MONITORED"`) {
		t.Fatalf("body missing status: %s", wire)
	}
}
