package cmd

import (
	"context"
	"net/http"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func TestPermissionsListPrintsDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/permissions" {
			http.NotFound(w, r)

			return
		}

		writeJSON(t, w, []map[string]any{
			{"id": "VIEW_ONLY", "name": "View only", "description": "Read every resource"},
		})
	})

	permissionsListCmd.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return permissionsListCmd.RunE(permissionsListCmd, nil)
	})
	if err != nil {
		t.Fatalf("list: %v\nstderr: %s", err, stderr)
	}

	var permissions []v3.PermissionCatalogEntry
	mustUnmarshalOne(t, stdout, &permissions)
	if len(permissions) != 1 || permissions[0].GetId() != "VIEW_ONLY" {
		t.Fatalf("permissions = %#v", permissions)
	}

	if permissions[0].GetName() != "View only" || permissions[0].GetDescription() == "" {
		t.Fatalf("permission = %#v", permissions[0])
	}
}
