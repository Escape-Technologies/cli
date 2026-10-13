package cmd

import (
	"context"
	"errors"
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

	securityTestsUpdateEnabled       string
	securityTestsUpdateSeverity      string
	securityTestsUpdateResetSeverity bool

	securityTestsBulkEnabled       string
	securityTestsBulkSeverity      string
	securityTestsBulkResetSeverity bool
	securityTestsBulkIDs           []string
	securityTestsBulkSearch        string
	securityTestsBulkWhereEnabled  string
	securityTestsBulkScannerKinds  []string
	securityTestsBulkCategories    []string
	securityTestsBulkCompliances   []string
	securityTestsBulkAssetTypes    []string
	securityTestsBulkAll           bool
	securityTestsBulkDryRun        bool
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
	Short:   "Browse and tune the security-test catalog of the organization",
	Long: fmt.Sprintf(`Security Tests - The Catalog Every Scan Draws From

The security-test catalog lists every check the scanners can run, with the
organization configuration layered on top: whether a test is enabled and the
severity override its issues report. A test also carries the compliance items
it maps to, so the catalog is the bridge between a scan result and a
regulatory framework.

list and get read the catalog; update and bulk-update change the organization
configuration and need admin rights.

SCANNER KINDS: %v
ASSET TYPES: %v

COMMON WORKFLOWS:
  • Find tests that fire the most issues:
    $ escape-cli security-tests list --sort-type ISSUES_COUNT

  • Read one test before changing its configuration:
    $ escape-cli security-tests get <uid>

  • Turn a noisy test off, or the whole REST suite:
    $ escape-cli security-tests update <uid> --enabled false
    $ escape-cli security-tests bulk-update --scanner-kind BLST_REST --enabled false`,
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

var securityTestsUpdateCmd = &cobra.Command{
	Use:     "update security-test-uid",
	Aliases: []string{"u", "edit"},
	Short:   "Change one security test's organization configuration (Beta)",
	Long: `Update Security Test - Enable, Disable or Re-rank One Check

Change whether a single security test runs for the organization, and the
severity override the issues it reports carry. --reset-severity drops the
override and returns the test to the scanner default. At least one of
--enabled, --set-severity or --reset-severity is required.`,
	Example: `  # Turn one check off
  escape-cli security-tests update 00000000-0000-0000-0000-000000000000 --enabled false

  # Report its issues as HIGH, then back to the scanner default
  escape-cli security-tests update <uid> --set-severity HIGH
  escape-cli security-tests update <uid> --reset-severity`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetSecurityTest200Response{}) {
			return nil
		}

		body, err := buildUpdateSecurityTestRequest(
			securityTestsUpdateEnabled,
			securityTestsUpdateSeverity,
			securityTestsUpdateResetSeverity,
		)
		if err != nil {
			return err
		}

		proceed, err := confirmDestructive(cmd, "update security test "+args[0]+" for the whole organization", func(ctx context.Context) (any, error) {
			return escape.GetSecurityTest(ctx, args[0])
		})
		if err != nil || !proceed {
			return err
		}

		test, err := escape.UpdateSecurityTest(cmd.Context(), args[0], *body)
		if err != nil {
			return fmt.Errorf("unable to update security test %s: %w", args[0], err)
		}

		printSecurityTest(test)

		return nil
	},
}

var securityTestsBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Change the configuration of every security test matching a filter (Beta)",
	Long: `Bulk Update Security Tests - Retune a Whole Suite at Once

Change whether the security tests matching a filter run for the organization,
and the severity override their issues report. The API requires a non-empty
filter or --all, and rejects --all combined with a filter.

--dry-run forwards the API's own preview: it returns the matching security test
IDs without changing anything.`,
	Example: `  # Preview which tests a change would touch
  escape-cli security-tests bulk-update --scanner-kind BLST_REST --enabled false --dry-run

  # Switch off every test of one category
  escape-cli security-tests bulk-update --category CONFIGURATION --enabled false`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.BulkUpdateSecurityTests200Response{}) {
			return nil
		}

		body, err := buildBulkUpdateSecurityTestsRequest()
		if err != nil {
			return err
		}

		// --dry-run only previews the matching IDs, so it never prompts.
		if !securityTestsBulkDryRun {
			proceed, err := confirmDestructive(cmd, "update every security test matching the given filters", nil)
			if err != nil || !proceed {
				return err
			}
		}

		result, err := escape.BulkUpdateSecurityTests(cmd.Context(), *body)
		if err != nil {
			return fmt.Errorf("unable to bulk update security tests: %w", err)
		}

		changed := "UPDATED IDS"
		summary := fmt.Sprintf("Updated %d security tests", result.GetCount())
		if result.GetDryRun() {
			changed = "MATCHING IDS (DRY RUN)"
			summary = fmt.Sprintf("Dry run: %d security tests would be updated", result.GetCount())
		}

		out.Table(result, func() []string {
			res := []string{changed}
			res = append(res, result.GetIds()...)

			return res
		})
		out.Log(summary)

		return nil
	},
}

// nullSecurityTestSeverity sends customSeverity as JSON null. The generated
// setter can only emit a severity or omit the field, and the API treats an
// omission as "leave the override unchanged". An empty string is not a
// severity, so the null rides in the additional-properties map.
func nullSecurityTestSeverity(props *map[string]interface{}) {
	if *props == nil {
		*props = map[string]interface{}{}
	}

	(*props)["customSeverity"] = nil
}

// buildUpdateSecurityTestRequest assembles the single-test payload. enabled
// and the severity change are independent: the API accepts both in one body.
// A reset carries an explicit null so the override is dropped rather than
// left out, which the API reads as "no change".
func buildUpdateSecurityTestRequest(enabled, severity string, reset bool) (*v3.UpdateSecurityTestRequest, error) {
	if reset && severity != "" {
		return nil, errors.New("--reset-severity and --set-severity are mutually exclusive")
	}

	body := v3.NewUpdateSecurityTestRequest()
	changed := false

	if enabled != "" {
		parsed, err := parseSecurityTestEnabled(enabled)
		if err != nil {
			return nil, err
		}

		body.SetEnabled(parsed)
		changed = true
	}

	if reset {
		nullSecurityTestSeverity(&body.AdditionalProperties)
		changed = true
	} else if severity != "" {
		parsed, err := parseSecurityTestSeverity(severity)
		if err != nil {
			return nil, err
		}

		body.SetCustomSeverity(parsed)
		changed = true
	}

	if !changed {
		return nil, errors.New("at least one of --enabled, --set-severity, or --reset-severity is required")
	}

	return body, nil
}

func buildBulkUpdateSecurityTestsRequest() (*v3.BulkUpdateSecurityTestsRequest, error) {
	if securityTestsBulkResetSeverity && securityTestsBulkSeverity != "" {
		return nil, errors.New("--reset-severity and --set-severity are mutually exclusive")
	}

	body := v3.NewBulkUpdateSecurityTestsRequestWithDefaults()
	changed := false

	if securityTestsBulkEnabled != "" {
		parsed, err := parseSecurityTestEnabled(securityTestsBulkEnabled)
		if err != nil {
			return nil, err
		}

		body.SetEnabled(parsed)
		changed = true
	}

	if securityTestsBulkResetSeverity {
		nullSecurityTestSeverity(&body.AdditionalProperties)
		changed = true
	} else if securityTestsBulkSeverity != "" {
		parsed, err := parseSecurityTestSeverity(securityTestsBulkSeverity)
		if err != nil {
			return nil, err
		}

		body.SetCustomSeverity(parsed)
		changed = true
	}

	if !changed {
		return nil, errors.New("at least one of --enabled, --set-severity, or --reset-severity is required")
	}

	if securityTestsBulkAll && securityTestsBulkFilterFlagsSet() {
		return nil, errors.New("--all cannot be combined with a filter; pass --all alone to target every security test")
	}

	if !securityTestsBulkAll && !securityTestsBulkFilterFlagsSet() {
		return nil, errors.New("at least one of --security-test-id, --search, --where-enabled, --scanner-kind, --category, --compliance, --asset-type, or --all is required")
	}

	// An empty where is only sent with a filter: with --all the API rejects a
	// non-empty where, and an empty one is noise.
	if securityTestsBulkFilterFlagsSet() {
		where, err := buildSecurityTestsBulkWhere()
		if err != nil {
			return nil, err
		}

		body.SetWhere(where)
	}

	// All and DryRun are pointers with omitempty, so they are only forwarded
	// when set: an explicit false would be accepted by the API but adds noise.
	if securityTestsBulkAll {
		body.SetAll(true)
	}

	if securityTestsBulkDryRun {
		body.SetDryRun(true)
	}

	return body, nil
}

func buildSecurityTestsBulkWhere() (v3.BulkUpdateSecurityTestsRequestWhere, error) {
	where := v3.BulkUpdateSecurityTestsRequestWhere{}

	if len(securityTestsBulkIDs) > 0 {
		where.Ids = securityTestsBulkIDs
	}

	if securityTestsBulkSearch != "" {
		where.SetSearch(securityTestsBulkSearch)
	}

	if securityTestsBulkWhereEnabled != "" {
		parsed, err := parseSecurityTestEnabled(securityTestsBulkWhereEnabled)
		if err != nil {
			return where, err
		}

		where.SetEnabled(parsed)
	}

	if len(securityTestsBulkCategories) > 0 {
		categories := make([]v3.ENUMPROPERTIESISSUEPROPERTIESCATEGORIESITEMSPROPERTIESCATEGORY, len(securityTestsBulkCategories))
		for i, category := range securityTestsBulkCategories {
			categories[i] = v3.ENUMPROPERTIESISSUEPROPERTIESCATEGORIESITEMSPROPERTIESCATEGORY(category)
		}

		where.Categories = categories
	}

	if len(securityTestsBulkCompliances) > 0 {
		compliances := make([]v3.ENUMPROPERTIESISSUEPROPERTIESCOMPLIANCEITEMSPROPERTIESCOMPLIANCE, len(securityTestsBulkCompliances))
		for i, compliance := range securityTestsBulkCompliances {
			compliances[i] = v3.ENUMPROPERTIESISSUEPROPERTIESCOMPLIANCEITEMSPROPERTIESCOMPLIANCE(compliance)
		}

		where.Compliances = compliances
	}

	if len(securityTestsBulkScannerKinds) > 0 {
		kinds := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESSCANNERSITEMS, len(securityTestsBulkScannerKinds))
		for i, kind := range securityTestsBulkScannerKinds {
			kinds[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESSCANNERSITEMS(kind)
		}

		where.ScannerKinds = kinds
	}

	if len(securityTestsBulkAssetTypes) > 0 {
		types := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE, len(securityTestsBulkAssetTypes))
		for i, assetType := range securityTestsBulkAssetTypes {
			types[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE(assetType)
		}

		where.AssetTypes = types
	}

	return where, nil
}

func securityTestsBulkFilterFlagsSet() bool {
	return len(securityTestsBulkIDs) > 0 ||
		securityTestsBulkSearch != "" ||
		securityTestsBulkWhereEnabled != "" ||
		len(securityTestsBulkScannerKinds) > 0 ||
		len(securityTestsBulkCategories) > 0 ||
		len(securityTestsBulkCompliances) > 0 ||
		len(securityTestsBulkAssetTypes) > 0
}

func parseSecurityTestEnabled(value string) (bool, error) {
	switch value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid enabled state %q; use true or false", value)
	}
}

func parseSecurityTestSeverity(value string) (v3.ENUMPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY, error) {
	severity := v3.ENUMPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY(value)
	if !severity.IsValid() {
		return "", fmt.Errorf("invalid severity %q; valid values: %v", value, v3.AllowedENUMPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues)
	}

	return severity, nil
}

// printSecurityTest renders one updated test with the same columns the catalog
// listing uses, so a before/after read lines up.
func printSecurityTest(test *v3.GetSecurityTest200Response) {
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

	securityTestsUpdateCmd.Flags().StringVar(&securityTestsUpdateEnabled, "enabled", "", "switch the test on or off for the organization (true or false)")
	securityTestsUpdateCmd.Flags().StringVar(&securityTestsUpdateSeverity, "set-severity", "", fmt.Sprintf("severity its issues report: %v", v3.AllowedENUMPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues))
	securityTestsUpdateCmd.Flags().BoolVar(&securityTestsUpdateResetSeverity, "reset-severity", false, "drop the severity override and use the scanner default")

	securityTestsBulkUpdateCmd.Flags().StringVar(&securityTestsBulkEnabled, "enabled", "", "switch the matched tests on or off (true or false)")
	securityTestsBulkUpdateCmd.Flags().StringVar(&securityTestsBulkSeverity, "set-severity", "", fmt.Sprintf("severity the matched tests report: %v", v3.AllowedENUMPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues))
	securityTestsBulkUpdateCmd.Flags().BoolVar(&securityTestsBulkResetSeverity, "reset-severity", false, "drop the severity overrides and use the scanner defaults")
	securityTestsBulkUpdateCmd.Flags().StringSliceVar(&securityTestsBulkIDs, "security-test-id", nil, "filter by security test ID(s)")
	securityTestsBulkUpdateCmd.Flags().StringVar(&securityTestsBulkSearch, "search", "", "filter by name substring")
	securityTestsBulkUpdateCmd.Flags().StringVar(&securityTestsBulkWhereEnabled, "where-enabled", "", "filter by the current enabled state of the matched tests (true or false)")
	securityTestsBulkUpdateCmd.Flags().StringSliceVar(&securityTestsBulkScannerKinds, "scanner-kind", nil, fmt.Sprintf("filter by scanner kind: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSCANNERSITEMSEnumValues))
	securityTestsBulkUpdateCmd.Flags().StringSliceVar(&securityTestsBulkCategories, "category", nil, fmt.Sprintf("filter by category: %v", v3.AllowedENUMPROPERTIESISSUEPROPERTIESCATEGORIESITEMSPROPERTIESCATEGORYEnumValues))
	securityTestsBulkUpdateCmd.Flags().StringSliceVar(&securityTestsBulkCompliances, "compliance", nil, fmt.Sprintf("filter by compliance framework: %v", v3.AllowedENUMPROPERTIESISSUEPROPERTIESCOMPLIANCEITEMSPROPERTIESCOMPLIANCEEnumValues))
	securityTestsBulkUpdateCmd.Flags().StringSliceVar(&securityTestsBulkAssetTypes, "asset-type", nil, fmt.Sprintf("filter by asset type: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPEEnumValues))
	securityTestsBulkUpdateCmd.Flags().BoolVar(&securityTestsBulkAll, "all", false, "target every security test; cannot be combined with filter flags")
	securityTestsBulkUpdateCmd.Flags().BoolVar(&securityTestsBulkDryRun, "dry-run", false, "print the matching security test IDs without updating anything")
	bindDestructiveGate(securityTestsBulkUpdateCmd, false)

	securityTestsCmd.AddCommand(securityTestsListCmd, securityTestsGetCmd, securityTestsUpdateCmd, securityTestsBulkUpdateCmd)
	bindDestructiveGate(securityTestsUpdateCmd, true)
	// The shared gate says the dry run sends no request. This command GETs the
	// current test first, so the help has to say no write is sent.
	securityTestsUpdateCmd.Flags().Lookup("dry-run").Usage = "report what the command would change and sends no write request"
	rootCmd.AddCommand(securityTestsCmd)
}
