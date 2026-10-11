package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	securityTestsSearch       string
	securityTestsEnabled      string
	securityTestsScannerKinds []string
	securityTestsAssetTypes   []string
	securityTestsSortType     string
	securityTestsSortOrder    string
	securityTestsListPage     pageFlags
)

// securityTestsSortTypes and securityTestsSortOrders are the values the catalog
// accepts. Both are checked here so a typo fails before a request is sent.
var (
	securityTestsSortTypes  = []string{"ISSUES_COUNT", "SEVERITY"}
	securityTestsSortOrders = []string{"asc", "desc"}
)

var securityTestsCmd = &cobra.Command{
	Use:     "security-tests",
	Aliases: []string{"security-test"},
	Short:   "Browse the security-test catalog of the organization",
	Long: fmt.Sprintf(`Security Tests - The Catalog Every Scan Draws From

The security-test catalog lists every check the scanners can run, with the
organization configuration layered on top: whether a test is enabled and the
severity override its issues report. A test also carries the compliance items
it maps to, so the catalog is the bridge between a scan result and a
regulatory framework.

SCANNER KINDS: %v
ASSET TYPES: %v

COMMON WORKFLOWS:
  • Find tests that fire the most issues:
    $ escape-cli security-tests list --sort-type ISSUES_COUNT

  • Read one test before changing its configuration:
    $ escape-cli security-tests get <uid>`,
		v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSCANNERSITEMSEnumValues,
		v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPEEnumValues,
	),
}

var securityTestsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the security tests of the organization catalog",
	Long: `List Security Tests - Most Impactful First

List the security tests of the organization catalog with their configuration
and the issue counts still open under them. With --size or --cursor the
command returns one page as {items, nextCursor, totalCount}; otherwise it
walks every page.`,
	Example: `  # List the whole catalog
  escape-cli security-tests list

  # Only the tests that are switched off
  escape-cli security-tests list --enabled false

  # Tests matching a name, for one scanner
  escape-cli security-tests list --search injection --scanner-kind BLST_REST

  # Tests ranked by the number of issues they still report
  escape-cli security-tests list --sort-type ISSUES_COUNT`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[v3.SecurityTest]{}) {
			return nil
		}

		if securityTestsSortType != "" && !slices.Contains(securityTestsSortTypes, securityTestsSortType) {
			return fmt.Errorf("invalid --sort-type %q; valid values: %s", securityTestsSortType, strings.Join(securityTestsSortTypes, ", "))
		}

		if securityTestsSortOrder != "" && !slices.Contains(securityTestsSortOrders, securityTestsSortOrder) {
			return fmt.Errorf("invalid --sort-direction %q; valid values: %s", securityTestsSortOrder, strings.Join(securityTestsSortOrders, ", "))
		}

		filters := &escape.SecurityTestsFilters{
			Search:        securityTestsSearch,
			Enabled:       securityTestsEnabled,
			ScannerKinds:  securityTestsScannerKinds,
			AssetTypes:    securityTestsAssetTypes,
			SortType:      securityTestsSortType,
			SortDirection: securityTestsSortOrder,
		}

		if err := runPagedList(cmd, securityTestsListPage, func(ctx context.Context, cursor string, size int) ([]v3.SecurityTest, *string, int, error) {
			return escape.ListSecurityTests(ctx, cursor, size, filters)
		}, securityTestsTable); err != nil {
			return fmt.Errorf("unable to list security tests: %w", err)
		}

		return nil
	},
}

var securityTestsGetCmd = &cobra.Command{
	Use:     "get security-test-uid",
	Aliases: []string{"describe", "show"},
	Short:   "Get a security test and its organization configuration",
	Long: `Get Security Test - Read Configuration and Open Issues

Fetch one security test by uid: its description, the scanner kinds and asset
types it applies to, the enabled state and severity the organization uses, and
the issue counts still open under it. The compliance section lists the
framework items the test satisfies.`,
	Example: `  escape-cli security-tests get 00000000-0000-0000-0000-000000000000
  escape-cli security-tests get <uid> -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetSecurityTest200Response{}) {
			return nil
		}

		test, err := escape.GetSecurityTest(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get security test: %w", err)
		}

		out.Table(test, func() []string {
			return []string{
				"UID\tNAME\tCATEGORY\tENABLED\tSEVERITY\tCVSS\tISSUES",
				formatSecurityTestRow(
					test.GetId(),
					test.GetName(),
					string(test.GetCategory()),
					test.GetEnabled(),
					string(test.GetSeverity()),
					test.GetCvssScore(),
					test.GetIssueCounts(),
				),
			}
		})

		if !out.IsPretty() {
			return nil
		}

		fmt.Printf("\n%s\n\n", test.GetDescription())

		mappings := test.GetCompliance()
		if len(mappings) == 0 {
			return nil
		}

		out.Table(mappings, func() []string {
			rows := []string{"COMPLIANCE FRAMEWORK\tCOMPLIANCE ITEM"}
			for _, mapping := range mappings {
				rows = append(rows, fmt.Sprintf("%s\t%s", mapping.GetFramework(), mapping.GetItem()))
			}

			return rows
		})

		return nil
	},
}

func securityTestsTable(tests []v3.SecurityTest) []string {
	rows := []string{"UID\tNAME\tCATEGORY\tENABLED\tSEVERITY\tCVSS\tISSUES"}
	for _, test := range tests {
		rows = append(rows, formatSecurityTestRow(
			test.GetId(),
			test.GetName(),
			string(test.GetCategory()),
			test.GetEnabled(),
			string(test.GetSeverity()),
			test.GetCvssScore(),
			test.GetIssueCounts(),
		))
	}

	return rows
}

// formatSecurityTestRow renders one catalog entry. severity is the effective
// severity, so an empty value means the organization did not override the
// scanner default; the same holds for a CVSS score of zero.
func formatSecurityTestRow(
	id, name, category string,
	enabled bool,
	severity string,
	cvss float32,
	counts v3.SecurityTestIssueCounts,
) string {
	severityLabel := severity
	if severityLabel == "" {
		severityLabel = "default"
	}

	cvssLabel := "-"
	if cvss > 0 {
		cvssLabel = fmt.Sprintf("%.1f", cvss)
	}

	return fmt.Sprintf("%s\t%s\t%s\t%t\t%s\t%s\t%s",
		id,
		name,
		category,
		enabled,
		severityLabel,
		cvssLabel,
		formatSecurityTestIssueCounts(counts),
	)
}

// formatSecurityTestIssueCounts renders the open issue counts per severity.
// A quiet test reads 0 rather than an empty cell, so "no issues" is visible.
func formatSecurityTestIssueCounts(counts v3.SecurityTestIssueCounts) string {
	return fmt.Sprintf("C%d H%d M%d L%d I%d",
		counts.GetCritical(),
		counts.GetHigh(),
		counts.GetMedium(),
		counts.GetLow(),
		counts.GetInfo(),
	)
}

func init() {
	securityTestsListCmd.Flags().StringVar(&securityTestsSearch, "search", "", "search term to filter security tests by name")
	securityTestsListCmd.Flags().StringVar(&securityTestsEnabled, "enabled", "", "filter by enabled state (true or false)")
	securityTestsListCmd.Flags().StringSliceVar(&securityTestsScannerKinds, "scanner-kind", []string{}, fmt.Sprintf("filter by scanner kind: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSCANNERSITEMSEnumValues))
	securityTestsListCmd.Flags().StringSliceVar(&securityTestsAssetTypes, "asset-type", []string{}, fmt.Sprintf("filter by asset type: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPEEnumValues))
	securityTestsListCmd.Flags().StringVar(&securityTestsSortType, "sort-type", "", fmt.Sprintf("rank the tests by: %v", securityTestsSortTypes))
	securityTestsListCmd.Flags().StringVar(&securityTestsSortOrder, "sort-direction", "", fmt.Sprintf("sort direction: %v", securityTestsSortOrders))
	securityTestsListPage.bind(securityTestsListCmd)

	securityTestsCmd.AddCommand(securityTestsListCmd, securityTestsGetCmd)
	rootCmd.AddCommand(securityTestsCmd)
}
