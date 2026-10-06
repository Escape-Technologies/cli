package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
)

func TestRegistryToolsAdvertiseConsistentOutputSchema(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	byPath := make(map[string]string, len(specs))
	rawByPath := make(map[string]json.RawMessage, len(specs))
	for _, spec := range specs {
		byPath[spec.Path] = spec.Name
		rawByPath[spec.Path] = spec.Tool.RawOutputSchema
	}

	for path, schemas := range CommandSchemaRegistry() {
		if schemas.Output == nil {
			continue
		}

		raw, ok := rawByPath[path]
		if !ok {
			// Parent commands (for example `audit`) are registered for their
			// output type but are not tools: MCP exposes the leaf command.
			continue
		}

		// A bounded emails wait can return a pending payload that is not a
		// ScanEmailDetails value. The catalog leaves that schema unset so the
		// pending result still validates.
		if path == "escape-cli emails wait" {
			if len(raw) != 0 {
				t.Errorf("%s must not advertise ScanEmailDetails; a pending result would fail it", path)
			}

			continue
		}

		if len(raw) == 0 {
			t.Errorf("%s (%s) has no output schema", path, byPath[path])
			continue
		}

		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Errorf("%s: unmarshal output schema: %v", path, err)
			continue
		}

		if got["type"] != "object" {
			t.Errorf("%s: output schema type = %#v, want object", path, got["type"])
		}

		declared := schemaFor(schemas.Output)
		if declared.Type == "array" {
			assertWrappedArraySchema(t, path, got, declared)
			continue
		}

		want := schemaMap(t, declared)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s output schema\n got %s\nwant %s", path, raw, mustJSON(t, want))
		}
	}
}

func TestArrayToolsAdvertiseWrappedItemsSchema(t *testing.T) {
	rawByName := toolOutputSchemas(t)
	for _, name := range []string{
		"issues_list",
		"locations_list",
		"scans_list",
		"users_list",
		"assets_list",
		"events_list",
		"profiles_list",
		"tags_list",
	} {
		raw, ok := rawByName[name]
		if !ok {
			t.Errorf("missing MCP tool %s", name)
			continue
		}

		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}

		if doc["type"] != "object" {
			t.Errorf("%s type = %#v", name, doc["type"])
		}

		props, _ := doc["properties"].(map[string]any)
		items, _ := props["items"].(map[string]any)
		if items["type"] != "array" {
			t.Errorf("%s properties.items = %#v", name, props["items"])
		}

		if items["items"] == nil {
			t.Errorf("%s properties.items.items is missing", name)
		}
	}
}

func TestNamedToolsDeclareTypedOutputSchema(t *testing.T) {
	rawByName := toolOutputSchemas(t)
	cases := []struct {
		name string
		keys []string
	}{
		{name: "issues_update", keys: []string{"ids"}},
		{name: "issues_bulk_update", keys: []string{"ids"}},
		{name: "locations_create", keys: []string{"links"}},
		{name: "locations_update", keys: []string{"links"}},
		{name: "issues_comment", keys: []string{"id", "createdAt", "kind"}},
	}
	for _, testCase := range cases {
		raw, ok := rawByName[testCase.name]
		if !ok {
			t.Errorf("missing MCP tool %s", testCase.name)
			continue
		}

		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Errorf("%s: %v", testCase.name, err)
			continue
		}

		props, _ := doc["properties"].(map[string]any)
		if _, msgOnly := props["msg"]; msgOnly && len(props) == 1 {
			t.Errorf("%s still advertises a msg-only payload", testCase.name)
		}

		for _, key := range testCase.keys {
			if _, exists := props[key]; !exists {
				t.Errorf("%s output schema missing %q in %s", testCase.name, key, raw)
			}
		}
	}
}

func TestIssueUpdateJSONIsOneTypedDocument(t *testing.T) {
	const issueID = "00000000-0000-0000-0000-000000000001"
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v3/issues/"+issueID {
			http.NotFound(w, r)
			return
		}

		writeJSON(t, w, map[string]any{"ids": []string{issueID}})
	})

	restoreIssueUpdateFlags(t)
	issueUpdateStatusStr = "OPEN"

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return issueUpdateStatusCmd.RunE(issueUpdateStatusCmd, []string{issueID})
	})
	if err != nil {
		t.Fatalf("update: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stderr, "Updated issue "+issueID) {
		t.Fatalf("status line missing from stderr: %q", stderr)
	}

	var updated v3.UpdateIssue200Response
	mustUnmarshalOne(t, stdout, &updated)
	if len(updated.GetIds()) != 1 || updated.GetIds()[0] != issueID {
		t.Fatalf("updated = %#v", updated)
	}
}

func TestIssueBulkUpdateJSONIsOneTypedDocument(t *testing.T) {
	const issueID = "00000000-0000-0000-0000-000000000002"
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/issues/bulk-update" {
			http.NotFound(w, r)
			return
		}

		writeJSON(t, w, map[string]any{"ids": []string{issueID}, "count": 1, "dryRun": false})
	})

	saveBulkUpdateFlags(t)
	bulkIssueStatus = "RESOLVED"
	bulkIssueIDs = []string{issueID}

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return issueBulkUpdateCmd.RunE(issueBulkUpdateCmd, nil)
	})
	if err != nil {
		t.Fatalf("bulk update: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stderr, "Updated 1 issues") {
		t.Fatalf("status line missing from stderr: %q", stderr)
	}

	var updated v3.BulkUpdateIssues200Response
	mustUnmarshalOne(t, stdout, &updated)
	if len(updated.GetIds()) != 1 || updated.GetIds()[0] != issueID {
		t.Fatalf("updated = %#v", updated)
	}
}

func TestIssueCommentJSONMatchesDeclaredOutput(t *testing.T) {
	const issueID = "00000000-0000-0000-0000-000000000003"
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/issues/"+issueID+"/activities" {
			http.NotFound(w, r)
			return
		}

		writeJSON(t, w, map[string]any{
			"id":        "activity-1",
			"createdAt": "2026-01-01T00:00:00Z",
			"kind":      "COMMENT",
			"comment":   "ship the fix",
		})
	})
	t.Cleanup(func() {
		_ = issueCommentCmd.Flags().Set("message", "")
	})
	if err := issueCommentCmd.Flags().Set("message", "ship the fix"); err != nil {
		t.Fatal(err)
	}

	issueCommentCmd.SetIn(bytes.NewReader(nil))

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return issueCommentCmd.RunE(issueCommentCmd, []string{issueID})
	})
	if err != nil {
		t.Fatalf("comment: %v\nstderr: %s", err, stderr)
	}

	var comment v3.CreateAssetComment200Response
	mustUnmarshalOne(t, stdout, &comment)
	if comment.GetId() != "activity-1" || comment.GetComment() != "ship the fix" {
		t.Fatalf("comment = %#v", comment)
	}

	if comment.GetKind() != v3.ENUMITEMSPROPERTIESKIND_COMMENT {
		t.Fatalf("kind = %s", comment.GetKind())
	}
}

func TestLocationCreateAndUpdateJSONMatchDeclaredOutput(t *testing.T) {
	const locationID = "00000000-0000-0000-0000-000000000004"
	locationBody := map[string]any{
		"id":      locationID,
		"name":    "prod-vpc",
		"type":    "PRIVATE",
		"enabled": true,
		"links":   map[string]any{"locationOverview": "https://app.escape.tech/loc"},
	}
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/locations":
			writeJSON(t, w, locationBody)
		case r.Method == http.MethodPut && r.URL.Path == "/v3/locations/"+locationID:
			updated := map[string]any{}
			for key, value := range locationBody {
				updated[key] = value
			}

			updated["name"] = "renamed"
			writeJSON(t, w, updated)
		default:
			http.NotFound(w, r)
		}
	})

	prevName := locationCreateName
	prevKey := locationCreateSSHPublicKey
	t.Cleanup(func() {
		locationCreateName = prevName
		locationCreateSSHPublicKey = prevKey
		resetLocationUpdateFlags()
	})
	locationCreateName = "prod-vpc"
	locationCreateSSHPublicKey = "ssh-ed25519 AAAA"
	locationsCreateCmd.SetIn(bytes.NewReader(nil))

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return locationsCreateCmd.RunE(locationsCreateCmd, nil)
	})
	if err != nil {
		t.Fatalf("create: %v\nstderr: %s", err, stderr)
	}

	var created v3.CreateLocation200Response
	mustUnmarshalOne(t, stdout, &created)
	if created.GetId() != locationID || created.GetName() != "prod-vpc" {
		t.Fatalf("created = %#v", created)
	}

	resetLocationUpdateFlags()
	if err := locationsUpdateCmd.Flags().Set("name", "renamed"); err != nil {
		t.Fatal(err)
	}

	locationsUpdateCmd.SetIn(bytes.NewReader(nil))
	stdout, stderr, err = captureJSONCommand(t, func() error {
		return locationsUpdateCmd.RunE(locationsUpdateCmd, []string{locationID})
	})
	if err != nil {
		t.Fatalf("update: %v\nstderr: %s", err, stderr)
	}

	var updated v3.CreateLocation200Response
	mustUnmarshalOne(t, stdout, &updated)
	if updated.GetId() != locationID || updated.GetName() != "renamed" {
		t.Fatalf("updated = %#v", updated)
	}
}

func TestIssueActivitiesJSONMatchesDeclaredOutput(t *testing.T) {
	const issueID = "00000000-0000-0000-0000-000000000005"
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/issues/"+issueID+"/activities" {
			http.NotFound(w, r)
			return
		}

		writeJSON(t, w, []map[string]any{{
			"id":        "activity-2",
			"createdAt": "2026-01-02T00:00:00Z",
			"kind":      "COMMENT",
		}})
	})

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return issueListActivitiesCmd.RunE(issueListActivitiesCmd, []string{issueID})
	})
	if err != nil {
		t.Fatalf("list activities: %v\nstderr: %s", err, stderr)
	}

	var activities []v3.ActivitySummarized
	mustUnmarshalOne(t, stdout, &activities)
	if len(activities) != 1 || activities[0].GetId() != "activity-2" {
		t.Fatalf("activities = %#v\nstdout: %s", activities, stdout)
	}
}

func assertWrappedArraySchema(t *testing.T, path string, got map[string]any, declared *clischema.JSONSchema) {
	t.Helper()
	props, _ := got["properties"].(map[string]any)
	itemsSchema, _ := props["items"].(map[string]any)
	want := schemaMap(t, &clischema.JSONSchema{
		Type:  "array",
		Items: declared.Items,
	})
	if !reflect.DeepEqual(itemsSchema, want) {
		t.Errorf("%s wrapped items\n got %#v\nwant %#v", path, itemsSchema, want)
	}

	if _, leaked := got["items"]; leaked && got["type"] == "array" {
		t.Errorf("%s advertised a top-level array", path)
	}
}

func toolOutputSchemas(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	rawByName := make(map[string]json.RawMessage, len(specs))
	for _, spec := range specs {
		rawByName[spec.Name] = spec.Tool.RawOutputSchema
	}

	return rawByName
}

func schemaMap(t *testing.T, schema *clischema.JSONSchema) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	return decoded
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	return string(encoded)
}

func serveJSON(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
}

func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func captureJSONCommand(t *testing.T, fn func() error) (string, string, error) {
	t.Helper()
	return captureOutput(t, "json", fn)
}

func captureOutput(t *testing.T, mode string, fn func() error) (string, string, error) {
	t.Helper()
	if err := out.SetOutput(mode); err != nil {
		t.Fatalf("set output: %v", err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}

	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	origOut := os.Stdout
	origErr := os.Stderr
	os.Stdout = outW
	os.Stderr = errW
	defer func() {
		os.Stdout = origOut
		os.Stderr = origErr
	}()

	cmdErr := fn()
	if err := outW.Close(); err != nil {
		t.Fatalf("close stdout: %v", err)
	}

	if err := errW.Close(); err != nil {
		t.Fatalf("close stderr: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, outR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if _, err := io.Copy(&errBuf, errR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	return outBuf.String(), errBuf.String(), cmdErr
}

func restoreIssueUpdateFlags(t *testing.T) {
	t.Helper()
	prevStatus := issueUpdateStatusStr
	prevReason := issueUpdateReason
	prevSeverity := issueUpdateSeverity
	prevReset := issueResetSeverity
	t.Cleanup(func() {
		issueUpdateStatusStr = prevStatus
		issueUpdateReason = prevReason
		issueUpdateSeverity = prevSeverity
		issueResetSeverity = prevReset
	})
	issueUpdateStatusStr = ""
	issueUpdateReason = ""
	issueUpdateSeverity = ""
	issueResetSeverity = false
}

func resetLocationUpdateFlags() {
	for _, name := range []string{"name", "ssh-public-key", "enabled"} {
		flag := locationsUpdateCmd.Flags().Lookup(name)
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	}
}

func mustUnmarshalOne(t *testing.T, stdout string, dest any) {
	t.Helper()
	if err := json.Unmarshal([]byte(stdout), dest); err != nil {
		t.Fatalf("stdout is not one JSON document matching the declared type: %v\n%s", err, stdout)
	}
}
