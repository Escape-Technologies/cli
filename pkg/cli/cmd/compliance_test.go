package cmd

import (
	"context"
	"net/http"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

const (
	complianceScanID    = "00000000-0000-0000-0000-0000000000b0"
	complianceProfileID = "00000000-0000-0000-0000-0000000000b1"
)

// TestComplianceCommandsPrintDeclaredDocument walks every compliance reader
// and checks each prints its declared response.
func TestComplianceCommandsPrintDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/compliance":
			writeJSON(t, w, map[string]any{"data": []map[string]any{{
				"framework": "PCI_DSS",
				"enabled":   true,
				"items":     []map[string]any{{"name": "6.5.1", "tests": []map[string]any{{"id": securityTestUID, "name": "SQL injection"}}}},
			}}})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/compliance/enabled":
			writeJSON(t, w, map[string]any{"data": []string{"PCI_DSS", "OWASP"}})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/compliance/per-application":
			if got := r.URL.Query().Get("profileIds"); got != complianceProfileID {
				t.Errorf("profileIds = %q, want %s", got, complianceProfileID)
			}

			writeJSON(t, w, map[string]any{"data": []map[string]any{{
				"application": map[string]any{"id": complianceProfileID, "name": "Checkout"},
				"compliances": []map[string]any{{"framework": "PCI_DSS", "failed": 2, "total": 10}},
			}}, "totalCount": 1})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/scans/"+complianceScanID+"/compliance":
			writeJSON(t, w, map[string]any{"data": []map[string]any{{
				"framework":   "PCI_DSS",
				"enabled":     true,
				"failedItems": []string{"6.5.1"},
			}}})
		default:
			http.NotFound(w, r)
		}
	})

	complianceProfileIDs = complianceProfileID
	t.Cleanup(func() { complianceProfileIDs = "" })

	tests := []struct {
		name  string
		cmd   *cobra.Command
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "posture",
			cmd:  complianceListCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var posture v3.GetCompliancePosture200Response
				mustUnmarshalOne(t, stdout, &posture)
				data := posture.GetData()
				if len(data) != 1 || data[0].GetFramework() != "PCI_DSS" || !data[0].GetEnabled() {
					t.Fatalf("posture = %#v", posture)
				}

				items := data[0].GetItems()
				if len(items) != 1 || len(items[0].GetTests()) != 1 {
					t.Fatalf("items = %#v", items)
				}
			},
		},
		{
			name: "enabled",
			cmd:  complianceEnabledCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var enabled v3.GetEnabledCompliances200Response
				mustUnmarshalOne(t, stdout, &enabled)
				if len(enabled.GetData()) != 2 || string(enabled.GetData()[0]) != "PCI_DSS" {
					t.Fatalf("enabled = %#v", enabled)
				}
			},
		},
		{
			name: "per application",
			cmd:  compliancePerApplicationCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var applications []v3.GetCompliancePerApplication200ResponseDataInner
				mustUnmarshalOne(t, stdout, &applications)
				if len(applications) != 1 {
					t.Fatalf("applications = %#v", applications)
				}

				compliances := applications[0].GetCompliances()
				if len(compliances) != 1 || compliances[0].GetFailed() != 2 || compliances[0].GetTotal() != 10 {
					t.Fatalf("compliances = %#v", compliances)
				}
			},
		},
		{
			name: "scan",
			cmd:  scansComplianceCmd,
			args: []string{complianceScanID},
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var scan v3.GetScanCompliance200Response
				mustUnmarshalOne(t, stdout, &scan)
				data := scan.GetData()
				if len(data) != 1 || len(data[0].GetFailedItems()) != 1 || data[0].GetFailedItems()[0] != "6.5.1" {
					t.Fatalf("scan compliance = %#v", scan)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetContext(context.Background())

			stdout, stderr, err := captureJSONCommand(t, func() error {
				return tt.cmd.RunE(tt.cmd, tt.args)
			})
			if err != nil {
				t.Fatalf("%s: %v\nstderr: %s", tt.name, err, stderr)
			}

			tt.check(t, stdout)
		})
	}
}
