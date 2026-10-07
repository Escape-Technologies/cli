package cmd

import (
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
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
