package cmd

import (
	"context"
	"net/http"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

const securityTestUID = "00000000-0000-0000-0000-0000000000a0"

func securityTestFixture() map[string]any {
	return map[string]any{
		"id":                securityTestUID,
		"name":              "SQL injection",
		"category":          "INJECTION",
		"description":       "Injects SQL payloads into every parameter.",
		"documentationPath": "/docs/tests/sql-injection",
		"scanners":          []string{"BLST_REST"},
		"assetTypes":        []string{"REST"},
		"enabled":           true,
		"defaultSeverity":   "HIGH",
		"customSeverity":    "CRITICAL",
		"severity":          "CRITICAL",
		"cvssScore":         9.8,
		"compliance":        []map[string]any{{"framework": "PCI_DSS", "item": "6.5.1"}},
		"issueCounts":       map[string]any{"critical": 1, "high": 2, "medium": 3, "low": 4, "info": 5},
	}
}

// TestSecurityTestsCommandsPrintDeclaredDocument walks list and get and
// checks each prints its declared response.
func TestSecurityTestsCommandsPrintDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/security-tests":
			if got := r.URL.Query().Get("search"); got != "sql" {
				t.Errorf("search = %q, want sql", got)
			}

			if got := r.URL.Query().Get("sortType"); got != "ISSUES_COUNT" {
				t.Errorf("sortType = %q, want ISSUES_COUNT", got)
			}

			writeJSON(t, w, map[string]any{"data": []map[string]any{securityTestFixture()}, "totalCount": 1})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/security-tests/"+securityTestUID:
			writeJSON(t, w, securityTestFixture())
		default:
			http.NotFound(w, r)
		}
	})

	securityTestsSearch = "sql"
	securityTestsSortType = "ISSUES_COUNT"
	t.Cleanup(func() {
		securityTestsSearch = ""
		securityTestsSortType = ""
	})

	tests := []struct {
		name  string
		cmd   *cobra.Command
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "list",
			cmd:  securityTestsListCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var tests []v3.SecurityTest
				mustUnmarshalOne(t, stdout, &tests)
				if tests[0].GetId() != securityTestUID {
					t.Fatalf("list = %#v", tests)
				}

				counts := tests[0].GetIssueCounts()
				if counts.GetCritical() != 1 || counts.GetHigh() != 2 {
					t.Fatalf("issue counts = %#v", counts)
				}
			},
		},
		{
			name: "get",
			cmd:  securityTestsGetCmd,
			args: []string{securityTestUID},
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var test v3.GetSecurityTest200Response
				mustUnmarshalOne(t, stdout, &test)
				if test.GetId() != securityTestUID || !test.GetEnabled() {
					t.Fatalf("get = %#v", test)
				}

				mappings := test.GetCompliance()
				if len(mappings) != 1 || mappings[0].GetFramework() != "PCI_DSS" {
					t.Fatalf("compliance = %#v", mappings)
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

func TestSecurityTestsListRejectsUnknownSortBeforeTheAPI(t *testing.T) {
	securityTestsSortType = "NOT_A_SORT"
	t.Cleanup(func() { securityTestsSortType = "" })

	err := securityTestsListCmd.RunE(securityTestsListCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--sort-type") {
		t.Fatalf("err = %v", err)
	}
}

func TestSecurityTestsUpdateDryRunSaysNoWrite(t *testing.T) {
	usage := securityTestsUpdateCmd.Flags().Lookup("dry-run").Usage
	if !strings.Contains(usage, "sends no write request") {
		t.Fatalf("usage = %q", usage)
	}
}
