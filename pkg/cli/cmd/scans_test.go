package cmd

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

func TestResolveKinds_Defaults(t *testing.T) {
	cmd := newTestCommand(t)

	got := resolveScanKinds(cmd)

	if !reflect.DeepEqual(got, defaultScanKinds) {
		t.Fatalf("expected default scan kinds %v, got %v", defaultScanKinds, got)
	}
}

func TestResolveKinds_AllKinds(t *testing.T) {
	cmd := newTestCommand(t)
	if err := cmd.Flags().Set("all-kinds", "true"); err != nil {
		t.Fatalf("failed to set all-kinds flag: %v", err)
	}

	got := resolveScanKinds(cmd)

	if got != nil {
		t.Fatalf("expected nil kinds for all-kinds, got %v", got)
	}

	if filter := scanKindsFilter(got); filter != nil {
		t.Fatalf("expected nil filter for all-kinds, got %v", *filter)
	}
}

func TestResolveKinds_KindOverrides(t *testing.T) {
	cmd := newTestCommand(t)
	if err := cmd.Flags().Set("all-kinds", "true"); err != nil {
		t.Fatalf("failed to set all-kinds flag: %v", err)
	}

	if err := cmd.Flags().Set("kind", "ASM_REST"); err != nil {
		t.Fatalf("failed to set kind flag: %v", err)
	}

	got := resolveScanKinds(cmd)
	want := []string{"ASM_REST"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected explicit kinds %v, got %v", want, got)
	}
}

// TestScansListPrintsProfileAndAssetIDs guards the scan list output contract and
// the new filters wired to GET /scans.
func TestScansListPrintsProfileAndAssetIDs(t *testing.T) {
	const (
		scanID    = "00000000-0000-0000-0000-0000000000ce"
		profileID = "11111111-1111-1111-1111-111111111111"
		assetID   = "22222222-2222-2222-2222-222222222222"
	)

	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/scans" {
			http.NotFound(w, r)

			return
		}

		query := r.URL.Query()
		for key, want := range map[string]string{
			"search": "checkout",
			"noTags": "true",
			"dnf":    `{"operator":"AND"}`,
			"tagIds": "tag-1,tag-2",
		} {
			if got := query.Get(key); got != want {
				t.Errorf("query %s = %q, want %q", key, got, want)
			}
		}

		if got := query["assetTypes"]; len(got) != 1 || got[0] != "WEBAPP" {
			t.Errorf("assetTypes = %v", got)
		}

		writeJSON(t, w, map[string]any{
			"data": []map[string]any{{
				"id":            scanID,
				"profileId":     profileID,
				"assetId":       assetID,
				"status":        "FINISHED",
				"createdAt":     "2026-01-01T00:00:00Z",
				"duration":      1.5,
				"progressRatio": 1,
				"initiator":     "MANUAL",
				"kind":          "BLST_REST",
				"links":         map[string]any{"scanIssues": "https://app.escape.tech/x"},
			}},
			"nextCursor": nil,
			"totalCount": 1,
		})
	})

	prevSearch := scanSearch
	prevTagIDs := scanTagIDs
	prevAssetTypes := scanAssetTypes
	prevNoTags := scanNoTags
	prevDnf := scanDnf
	t.Cleanup(func() {
		scanSearch = prevSearch
		scanTagIDs = prevTagIDs
		scanAssetTypes = prevAssetTypes
		scanNoTags = prevNoTags
		scanDnf = prevDnf
	})

	scanSearch = "checkout"
	scanTagIDs = []string{"tag-1", "tag-2"}
	scanAssetTypes = []string{"WEBAPP"}
	scanNoTags = "true"
	scanDnf = `{"operator":"AND"}`

	scansListCmd.SetContext(context.Background())

	stdout, stderr, err := captureOutput(t, "pretty", func() error {
		return scansListCmd.RunE(scansListCmd, nil)
	})
	if err != nil {
		t.Fatalf("list: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "PROFILE ID") || !strings.Contains(stdout, "ASSET ID") {
		t.Fatalf("stdout missing profile/asset headers:\n%s", stdout)
	}

	if !strings.Contains(stdout, profileID) || !strings.Contains(stdout, assetID) {
		t.Fatalf("stdout missing profile/asset ids:\n%s", stdout)
	}
}

func TestScanConfigurationPrintsDeclaredDocument(t *testing.T) {
	const scanID = "00000000-0000-0000-0000-0000000000cc"
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/scans/"+scanID+"/configuration" {
			http.NotFound(w, r)

			return
		}

		writeJSON(t, w, map[string]any{
			"scanId":        scanID,
			"configuration": map[string]any{"mode": "read_only"},
		})
	})

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return scanConfigurationCmd.RunE(scanConfigurationCmd, []string{scanID})
	})
	if err != nil {
		t.Fatalf("configuration: %v\nstderr: %s", err, stderr)
	}

	var configuration v3.GetScanConfiguration200Response
	mustUnmarshalOne(t, stdout, &configuration)
	if configuration.GetScanId() != scanID {
		t.Fatalf("configuration = %#v", configuration)
	}
}

func TestScanStatisticsPrintsDeclaredDocument(t *testing.T) {
	const scanID = "00000000-0000-0000-0000-0000000000cd"
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/scans/"+scanID+"/statistics" {
			http.NotFound(w, r)

			return
		}

		writeJSON(t, w, map[string]any{
			"scanId": scanID,
			"issue": map[string]any{
				"severities": []map[string]any{{"severity": "HIGH", "count": 2}},
				"categories": []map[string]any{},
				"compliance": []map[string]any{},
				"coverage":   []string{},
			},
			"events":            map[string]any{"eventsOverTime": []map[string]any{}},
			"requestTypeCounts": []map[string]any{{"operation": "GET", "count": 3}},
		})
	})

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return scanStatisticsCmd.RunE(scanStatisticsCmd, []string{scanID})
	})
	if err != nil {
		t.Fatalf("statistics: %v\nstderr: %s", err, stderr)
	}

	var statistics v3.GetScanStatistics200Response
	mustUnmarshalOne(t, stdout, &statistics)
	if statistics.GetScanId() != scanID || len(statistics.GetRequestTypeCounts()) != 1 {
		t.Fatalf("statistics = %#v", statistics)
	}
}

func newTestCommand(t *testing.T) *cobra.Command {
	t.Helper()

	prevKinds := scanKinds
	prevAllKinds := scanListAllKinds
	t.Cleanup(func() {
		scanKinds = prevKinds
		scanListAllKinds = prevAllKinds
	})

	scanKinds = []string{}
	scanListAllKinds = false

	cmd := &cobra.Command{}
	cmd.SetErr(io.Discard)
	cmd.Flags().StringSliceVarP(&scanKinds, "kind", "k", []string{}, "")
	cmd.Flags().BoolVar(&scanListAllKinds, "all-kinds", false, "")

	return cmd
}
