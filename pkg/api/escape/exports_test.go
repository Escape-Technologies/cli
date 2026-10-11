package escape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSynchronousExportsSendPathQueryAndAccept(t *testing.T) {
	type hit struct {
		path   string
		accept string
		where  string
	}

	hits := []hit{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}

		if got := r.Header.Get("Authorization"); got != "Key "+prd2312APIKey {
			t.Errorf("authorization = %q", got)
		}

		where := ""
		if _, present := r.URL.Query()["where"]; present {
			where = r.URL.Query().Get("where")
		}

		hits = append(hits, hit{
			path:   r.URL.Path,
			accept: r.Header.Get("Accept"),
			where:  where,
		})

		switch r.URL.Path {
		case "/v3/exports/issues/csv", "/v3/exports/assets/csv":
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
			_, _ = w.Write([]byte("id,name\n1,a"))
		case "/v3/exports/issues/sarif":
			w.Header().Set("Content-Type", "application/sarif+json")
			_, _ = w.Write([]byte(`{"version":"2.1.0","runs":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	ctx := context.Background()
	where := `{"status":"OPEN"}`
	csvDoc, err := ExportIssuesCSV(ctx, where)
	if err != nil {
		t.Fatalf("issues csv: %v", err)
	}

	if csvDoc.ContentType != csvAccept || string(csvDoc.Body) != "id,name\n1,a" {
		t.Fatalf("issues csv document = %#v", csvDoc)
	}

	sarifDoc, err := ExportIssuesSARIF(ctx, "")
	if err != nil {
		t.Fatalf("issues sarif: %v", err)
	}

	if sarifDoc.ContentType != sarifAccept || string(sarifDoc.Body) != `{"version":"2.1.0","runs":[]}` {
		t.Fatalf("issues sarif document = %#v", sarifDoc)
	}

	assetDoc, err := ExportAssetsCSV(ctx, `{"statuses":["MONITORED"]}`)
	if err != nil {
		t.Fatalf("assets csv: %v", err)
	}

	if assetDoc.ContentType != csvAccept || string(assetDoc.Body) != "id,name\n1,a" {
		t.Fatalf("assets csv document = %#v", assetDoc)
	}

	want := []hit{
		{path: "/v3/exports/issues/csv", accept: csvAccept, where: where},
		{path: "/v3/exports/issues/sarif", accept: sarifAccept, where: ""},
		{path: "/v3/exports/assets/csv", accept: csvAccept, where: `{"statuses":["MONITORED"]}`},
	}
	if len(hits) != len(want) {
		t.Fatalf("hits = %#v", hits)
	}

	for i, got := range hits {
		if got != want[i] {
			t.Errorf("hit %d = %#v, want %#v", i, got, want[i])
		}
	}
}

func TestSynchronousExportOmitsABlankWhere(t *testing.T) {
	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("id\n"))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	if _, err := ExportIssuesCSV(context.Background(), ""); err != nil {
		t.Fatal(err)
	}

	if rawQuery != "" {
		t.Fatalf("query = %q, want no where parameter", rawQuery)
	}
}

func TestSynchronousExportSurfacesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "too many issues", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	_, err := ExportAssetsCSV(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "400") || !strings.Contains(err.Error(), "too many issues") {
		t.Fatalf("err = %v", err)
	}
}
