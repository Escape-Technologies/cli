package escape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestListRolesDrainsCursorPages drives ListRoles against a server that pages
// roles behind the X-Next-Cursor header. One request used to cap the listing at
// the default page size, so the full list is the contract under test.
func TestListRolesDrainsCursorPages(t *testing.T) {
	var cursors []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/roles" {
			http.NotFound(w, r)

			return
		}

		cursors = append(cursors, r.URL.Query().Get("cursor"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("cursor") {
		case "":
			w.Header().Set("X-Next-Cursor", "page-2")
			_, _ = w.Write([]byte(`[{"id":"00000000-0000-0000-0000-000000000001","name":"admin","permissions":[],"createdAt":"2026-01-01T00:00:00Z","bindings":[]}]`))
		case "page-2":
			w.Header().Set("X-Next-Cursor", "page-2")
			_, _ = w.Write([]byte(`[{"id":"00000000-0000-0000-0000-000000000002","name":"viewer","permissions":[],"createdAt":"2026-01-01T00:00:00Z","bindings":[]}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	// The repeated page-2 cursor is a server fault: failing beats returning a
	// truncated list as if it were complete.
	_, err := ListRoles(context.Background())
	if err == nil || !strings.Contains(err.Error(), "did not advance") {
		t.Fatalf("err = %v, want a pagination error", err)
	}

	if len(cursors) != 2 {
		t.Fatalf("cursor sequence = %v, want the repeated cursor to stop the loop", cursors)
	}
}

// TestListCustomRulesDrainsCursorPages is the same drain for custom rules,
// whose listing is also a bare array behind the cursor header.
func TestListCustomRulesDrainsCursorPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/custom-rules" {
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "" {
			w.Header().Set("X-Next-Cursor", "next")
			_, _ = w.Write([]byte(`[{"id":"00000000-0000-0000-0000-000000000001","name":"rule-1","context":"probe","severity":"INFO","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","content":{"rule":{}}}]`))

			return
		}

		_, _ = w.Write([]byte(`[{"id":"00000000-0000-0000-0000-000000000002","name":"rule-2","context":"probe","severity":"INFO","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","content":{"rule":{}}}]`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	rules, err := ListCustomRules(context.Background())
	if err != nil {
		t.Fatalf("list custom rules: %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("rules = %d, want 2 (every page drained): %#v", len(rules), rules)
	}
}
