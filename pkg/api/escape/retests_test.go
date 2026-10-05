package escape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRetestsSendsRequestedSizeAndTotalCount(t *testing.T) {
	var gotSize string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/retests" {
			t.Errorf("path = %s", r.URL.Path)
		}

		gotSize = r.URL.Query().Get("size")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"retest-1","profileId":"profile-1","status":"FINISHED","progressRatio":1,"createdAt":"2026-01-01T00:00:00Z","agents":[]}],"nextCursor":"next","totalCount":7}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_AUTHORIZATION", "Bearer test")

	rows, next, total, err := ListRetests(context.Background(), "", &ListRetestsFilters{ProfileIDs: []string{"profile-1"}}, 5)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if gotSize != "5" || total != 7 || next == nil || *next != "next" || len(rows) != 1 {
		t.Fatalf("size=%s total=%d next=%v rows=%d", gotSize, total, next, len(rows))
	}

	if _, _, _, err := ListRetests(context.Background(), "", nil, 0); err != nil {
		t.Fatalf("default page: %v", err)
	}

	if gotSize != "100" {
		t.Fatalf("size 0 requested page size %s, want the historical maximum 100", gotSize)
	}
}
