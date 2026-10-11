package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestParseAssetLinks(t *testing.T) {
	t.Parallel()

	links, err := parseAssetLinks("")
	if err != nil || links != nil {
		t.Fatalf("empty links = %+v, err = %v", links, err)
	}

	links, err = parseAssetLinks(`{"create":[{"targetId":"parent-1","verb":"USES"}]}`)
	if err != nil {
		t.Fatalf("parse links: %v", err)
	}

	if len(links.Create) != 1 || links.Create[0].GetTargetId() != "parent-1" || links.Create[0].GetVerb() != "USES" {
		t.Fatalf("create links = %+v", links.Create)
	}

	if _, err := parseAssetLinks("{not json"); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestFormatAssetRelationships(t *testing.T) {
	t.Parallel()

	parents := []v3.AssetRelationship{{
		Id:    "p1",
		Name:  "escape-api",
		Class: v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESCLASS("API"),
		Type:  v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE("REST"),
	}}
	children := []v3.AssetRelationship{{
		Id:    "c1",
		Name:  "docs",
		Class: v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESCLASS("FRONTEND"),
		Type:  v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE("WEBAPP"),
	}}

	got := formatAssetRelationships(parents, children)
	want := "parents: escape-api (REST); children: docs (WEBAPP)"
	if got != want {
		t.Fatalf("relationships = %q, want %q", got, want)
	}

	if got := formatAssetRelationships(nil, nil); got != "" {
		t.Fatalf("empty relationships = %q", got)
	}
}

func TestFormatAssetTechnology(t *testing.T) {
	t.Parallel()

	version := "4.17.21"
	technology := &v3.AssetTechnologyDetailed{
		Type:          v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESTECHNOLOGYPROPERTIESTYPE("PACKAGE"),
		TechnologyKey: "npm/lodash",
		Version:       &version,
	}

	if got, want := formatAssetTechnology(technology), "PACKAGE npm/lodash@4.17.21"; got != want {
		t.Fatalf("technology = %q, want %q", got, want)
	}

	if got := formatAssetTechnology(nil); got != "" {
		t.Fatalf("nil technology = %q", got)
	}
}

func TestFormatAssetPorts(t *testing.T) {
	t.Parallel()

	asset := &v3.AssetDetailed1{
		Host: &v3.AssetHostDetailed{
			Ports: []v3.AssetHostDetailedPortsInner{
				{Port: 80, Protocols: []string{"tcp"}},
				{Port: 443, Protocols: []string{"tcp", "udp"}},
			},
		},
	}

	if got, want := formatAssetPorts(asset), "80/tcp,443/tcp+udp"; got != want {
		t.Fatalf("ports = %q, want %q", got, want)
	}

	if got := formatAssetPorts(&v3.AssetDetailed1{}); got != "" {
		t.Fatalf("hostless ports = %q", got)
	}
}

func TestAssetBulkImportRequiresAssets(t *testing.T) {
	assetBulkImportCmd.SetIn(strings.NewReader("   "))

	err := assetBulkImportCmd.RunE(assetBulkImportCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "JSON body is required") {
		t.Fatalf("empty stdin error = %v", err)
	}

	assetBulkImportCmd.SetIn(strings.NewReader(`{"assets":[]}`))

	err = assetBulkImportCmd.RunE(assetBulkImportCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "at least one asset") {
		t.Fatalf("empty assets error = %v", err)
	}
}

const assetCommandFixture = `{
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

func captureCommandBody(t *testing.T, wantMethod, wantPath, response string) *[]byte {
	t.Helper()

	raw := []byte{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != wantMethod || r.URL.Path != wantPath {
			t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, wantMethod, wantPath)
		}

		raw, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
	t.Setenv("ESCAPE_AUTHORIZATION", "")
	if err := out.SetOutput("json"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })

	return &raw
}

func prepareCommand(t *testing.T, command *cobra.Command) {
	t.Helper()
	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetContext(context.Background())
}

func setSliceFlag(t *testing.T, command *cobra.Command, name string, values []string) {
	t.Helper()

	flag := command.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing flag %s", name)
	}

	slice, ok := flag.Value.(pflag.SliceValue)
	if !ok {
		t.Fatalf("%s is %T, want a slice flag", name, flag.Value)
	}

	previous := append([]string{}, slice.GetSlice()...)
	previousChanged := flag.Changed
	t.Cleanup(func() {
		_ = slice.Replace(previous)
		flag.Changed = previousChanged
	})
	if err := slice.Replace(values); err != nil {
		t.Fatalf("replace %s: %v", name, err)
	}

	flag.Changed = true
}

// unsetSliceFlag leaves a non-nil empty slice with Changed false. owners and
// tag-ids default to that shape, which must not be sent as a clear.
func unsetSliceFlag(t *testing.T, command *cobra.Command, name string) {
	t.Helper()

	flag := command.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing flag %s", name)
	}

	slice, ok := flag.Value.(pflag.SliceValue)
	if !ok {
		t.Fatalf("%s is %T, want a slice flag", name, flag.Value)
	}

	previous := append([]string{}, slice.GetSlice()...)
	previousChanged := flag.Changed
	t.Cleanup(func() {
		_ = slice.Replace(previous)
		flag.Changed = previousChanged
	})
	if err := slice.Replace([]string{}); err != nil {
		t.Fatalf("replace %s: %v", name, err)
	}

	flag.Changed = false
}

func TestAssetUpdateSendsExplicitEmptiesAndOmitsUnsetFlags(t *testing.T) {
	const assetID = "00000000-0000-0000-0000-00000000000a"
	raw := captureCommandBody(t, http.MethodPut, "/v3/assets/"+assetID, assetCommandFixture)
	prepareCommand(t, assetUpdateCmd)

	for _, name := range []string{"description", "name", "base-url", "address", "best-schema-id", "framework", "status", "environment"} {
		resetFlag(t, assetUpdateCmd, name)
	}

	for _, name := range []string{"description", "name", "base-url", "address", "best-schema-id", "framework"} {
		if err := assetUpdateCmd.Flags().Set(name, ""); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}

	for _, name := range []string{"owners", "tag-ids", "project-id", "schema-ids"} {
		setSliceFlag(t, assetUpdateCmd, name, []string{})
	}

	if err := assetUpdateCmd.RunE(assetUpdateCmd, []string{assetID}); err != nil {
		t.Fatalf("update: %v", err)
	}

	wire := string(*raw)
	for _, want := range []string{
		`"description":""`,
		`"owners":[]`,
		`"tagIds":[]`,
		`"projectIds":[]`,
		`"schemaIds":[]`,
	} {
		if !strings.Contains(wire, want) {
			t.Fatalf("body missing %s: %s", want, wire)
		}
	}

	// name is nullable, but "" is not null, so an explicit empty flag is omitted.
	// address, base URL, and best schema id are plain optional strings: "" is
	// not a clear the API accepts. framework is an enum. status and environment
	// were left unset.
	for _, absent := range []string{`"name"`, `"baseUrl"`, `"address"`, `"bestSchemaId"`, `"framework"`, `"status"`, `"environment"`} {
		if strings.Contains(wire, absent) {
			t.Fatalf("omitted %s was sent: %s", absent, wire)
		}
	}
}

func TestAssetUpdateSendsNonEmptyNameAddressBaseURLAndBestSchema(t *testing.T) {
	const assetID = "00000000-0000-0000-0000-00000000000a"
	raw := captureCommandBody(t, http.MethodPut, "/v3/assets/"+assetID, assetCommandFixture)
	prepareCommand(t, assetUpdateCmd)

	for _, name := range []string{"name", "base-url", "address", "best-schema-id", "description"} {
		resetFlag(t, assetUpdateCmd, name)
	}

	sets := map[string]string{
		"name":           "Payments API",
		"base-url":       "https://example.com",
		"address":        "10.0.0.1",
		"best-schema-id": "00000000-0000-0000-0000-00000000000b",
	}
	for name, value := range sets {
		if err := assetUpdateCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}

	if err := assetUpdateCmd.RunE(assetUpdateCmd, []string{assetID}); err != nil {
		t.Fatalf("update: %v", err)
	}

	wire := string(*raw)
	for _, want := range []string{
		`"name":"Payments API"`,
		`"baseUrl":"https://example.com"`,
		`"address":"10.0.0.1"`,
		`"bestSchemaId":"00000000-0000-0000-0000-00000000000b"`,
	} {
		if !strings.Contains(wire, want) {
			t.Fatalf("body missing %s: %s", want, wire)
		}
	}
}

func TestAssetUpdateOmitsUnsetDefaultEmptySlices(t *testing.T) {
	const assetID = "00000000-0000-0000-0000-00000000000a"
	raw := captureCommandBody(t, http.MethodPut, "/v3/assets/"+assetID, assetCommandFixture)
	prepareCommand(t, assetUpdateCmd)

	for _, name := range []string{"description", "name", "base-url", "address", "best-schema-id", "framework", "status", "environment", "manually-set-status", "parent-links", "child-links"} {
		resetFlag(t, assetUpdateCmd, name)
	}

	if err := assetUpdateCmd.Flags().Set("description", ""); err != nil {
		t.Fatalf("set description: %v", err)
	}

	for _, name := range []string{"owners", "tag-ids", "project-id", "schema-ids"} {
		unsetSliceFlag(t, assetUpdateCmd, name)
	}

	if err := assetUpdateCmd.RunE(assetUpdateCmd, []string{assetID}); err != nil {
		t.Fatalf("update: %v", err)
	}

	wire := string(*raw)
	if !strings.Contains(wire, `"description":""`) {
		t.Fatalf("body missing empty description: %s", wire)
	}

	for _, absent := range []string{`"owners"`, `"tagIds"`, `"projectIds"`, `"schemaIds"`, `"name"`} {
		if strings.Contains(wire, absent) {
			t.Fatalf("unset %s was sent: %s", absent, wire)
		}
	}
}

func TestAssetBulkUpdateSendsEmptyTagListAndOmitsUnsetProjects(t *testing.T) {
	raw := captureCommandBody(t, http.MethodPost, "/v3/assets/bulk-update", `{"success":true}`)
	prepareCommand(t, assetBulkUpdateCmd)

	for _, name := range []string{"asset-id", "type", "asset-status", "status", "yes", "dry-run"} {
		resetFlag(t, assetBulkUpdateCmd, name)
	}

	setSliceFlag(t, assetBulkUpdateCmd, "asset-id", []string{"00000000-0000-0000-0000-00000000000a"})
	setSliceFlag(t, assetBulkUpdateCmd, "tag-id", []string{})
	resetFlag(t, assetBulkUpdateCmd, "project-id")
	if err := assetBulkUpdateCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes: %v", err)
	}

	if err := assetBulkUpdateCmd.RunE(assetBulkUpdateCmd, nil); err != nil {
		t.Fatalf("bulk update: %v", err)
	}

	wire := string(*raw)
	if !strings.Contains(wire, `"tagIds":[]`) {
		t.Fatalf("body missing empty tagIds: %s", wire)
	}

	if strings.Contains(wire, `"projectIds"`) {
		t.Fatalf("unset projectIds was sent: %s", wire)
	}
}
