package escape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListUsersSendsTheSearchToTheAPI pins the search to the endpoint: the
// public API matches the email only, and filtering names client-side promised
// a match the endpoint never performs.
func TestListUsersSendsTheSearchToTheAPI(t *testing.T) {
	var query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/users" {
			http.NotFound(w, r)

			return
		}

		query = r.URL.Query().Get("search")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"00000000-0000-0000-0000-000000000001","email":"alice@example.com","roleBindings":[],"createdAt":"2026-01-01T00:00:00Z"}]`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	users, err := ListUsers(context.Background(), "alice")
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if query != "alice" {
		t.Fatalf("search query = %q, want it sent to the API", query)
	}

	if len(users) != 1 || users[0].GetEmail() != "alice@example.com" {
		t.Fatalf("users = %#v, want the API response unfiltered by the client", users)
	}
}

// TestListUsersWithoutSearchSendsNoParameter keeps an empty search from
// narrowing the listing to nothing.
func TestListUsersWithoutSearchSendsNoParameter(t *testing.T) {
	var query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query().Get("search")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	if _, err := ListUsers(context.Background(), ""); err != nil {
		t.Fatalf("list users: %v", err)
	}

	if query != "" {
		t.Fatalf("search query = %q, want it omitted", query)
	}
}
