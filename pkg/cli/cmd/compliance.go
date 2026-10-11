package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	complianceProfileIDs string
	complianceListPage   pageFlags

	complianceEnableFrameworks []string
)

var complianceCmd = &cobra.Command{
	Use:     "compliance",
	Aliases: []string{"compliances"},
	Short:   "Report on, and choose, the compliance frameworks of the organization",
	Long: `Compliance - From Scan Findings to Framework Items

A compliance framework is a list of items; each item is satisfied by a set of
security tests. This group reports where the organization stands: which
frameworks are enabled, which of their items every application still fails,
and the same picture for a single scan. enable replaces the list of frameworks
the organization measures.

` + fmt.Sprintf("FRAMEWORKS: %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues) + `

COMMON WORKFLOWS:
  • Organization posture, with the tests behind each item:
    $ escape-cli compliance list

  • Applications that still fail a framework:
    $ escape-cli compliance per-application --profile-ids <profile-id>

  • Compliance of one scan, next to its statistics:
    $ escape-cli scans compliance <scan-id>`,
}

var complianceListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the compliance posture of every framework",
	Long: `List Compliance Posture - Frameworks, Items and Mapped Tests

List the compliance posture of the organization: every framework, whether it
is enabled, and the security tests mapped to each of its compliance items.`,
	Example: `  escape-cli compliance list
  escape-cli compliance list -o json | jq '.data[0].items[0].tests'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetCompliancePosture200Response{}) {
			return nil
		}

		posture, err := escape.GetCompliancePosture(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to get compliance posture: %w", err)
		}

		out.Table(posture, func() []string {
			rows := []string{"FRAMEWORK\tENABLED\tITEMS\tMAPPED TESTS"}
			for _, framework := range posture.GetData() {
				items := framework.GetItems()
				mapped := 0
				for _, item := range items {
					mapped += len(item.GetTests())
				}

				rows = append(rows, fmt.Sprintf("%s\t%t\t%d\t%d",
					framework.GetFramework(),
					framework.GetEnabled(),
					len(items),
					mapped,
				))
			}

			return rows
		})

		return nil
	},
}

var complianceEnabledCmd = &cobra.Command{
	Use:     "enabled",
	Aliases: []string{"frameworks"},
	Short:   "List the compliance frameworks enabled for the organization",
	Long: `List Enabled Frameworks - The Ones Being Measured

List the compliance frameworks the organization turned on. A framework that
is not in this list is not measured, so its items never appear as failures.`,
	Example: `  escape-cli compliance enabled
  escape-cli compliance enabled -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetEnabledCompliances200Response{}) {
			return nil
		}

		enabled, err := escape.GetEnabledCompliances(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to get enabled compliances: %w", err)
		}

		printEnabledCompliances(enabled)

		return nil
	},
}

var complianceEnableCmd = &cobra.Command{
	Use:     "enable",
	Aliases: []string{"set", "update"},
	Short:   "Set the compliance frameworks the organization measures (Beta)",
	Long: `Enable Compliance Frameworks - Choose What Is Measured

Replace the list of compliance frameworks the organization measures. The body
is the full list to enable, so a framework left out is disabled: an empty
--framework turns every framework off. Requires admin permissions.`,
	Example: `  # Measure one framework only
  escape-cli compliance enable --framework OWASP

  # Measure several
  escape-cli compliance enable --framework OWASP --framework PCI_DSS

  # Turn every framework off
  escape-cli compliance enable --framework=`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetEnabledCompliances200Response{}) {
			return nil
		}

		frameworks := dedupeComplianceFrameworks(complianceEnableFrameworks)
		for _, framework := range frameworks {
			if !v3.ENUMPROPERTIESISSUEPROPERTIESCOMPLIANCEITEMSPROPERTIESCOMPLIANCE(framework).IsValid() {
				return fmt.Errorf("invalid framework %q; valid values: %s", framework, strings.Join(complianceFrameworkNames(), ", "))
			}
		}

		proceed, err := confirmDestructive(cmd, "replace the compliance frameworks measured by the whole organization", func(ctx context.Context) (any, error) {
			return escape.GetEnabledCompliances(ctx)
		})
		if err != nil || !proceed {
			return err
		}

		enabled, err := escape.UpdateEnabledCompliances(cmd.Context(), frameworks)
		if err != nil {
			return fmt.Errorf("unable to update enabled compliances: %w", err)
		}

		printEnabledCompliances(enabled)

		return nil
	},
}

// dedupeComplianceFrameworks keeps the first occurrence of each value. The
// server deduplicates too, but a dry answer here keeps the confirmation prompt
// and the request the operator approved identical.
func dedupeComplianceFrameworks(frameworks []string) []string {
	seen := make(map[string]struct{}, len(frameworks))
	unique := make([]string, 0, len(frameworks))
	for _, framework := range frameworks {
		if _, exists := seen[framework]; exists {
			continue
		}

		seen[framework] = struct{}{}
		unique = append(unique, framework)
	}

	return unique
}

// complianceFrameworkNames is the framework list PUT /compliance/enabled
// accepts. It is narrower than the framework enum the posture listing prints:
// that enum also names frontend stacks, which are not compliance frameworks.
func complianceFrameworkNames() []string {
	values := v3.AllowedENUMPROPERTIESISSUEPROPERTIESCOMPLIANCEITEMSPROPERTIESCOMPLIANCEEnumValues
	names := make([]string, len(values))
	for i, value := range values {
		names[i] = string(value)
	}

	return names
}

func printEnabledCompliances(enabled *v3.GetEnabledCompliances200Response) {
	out.Table(enabled, func() []string {
		rows := []string{"FRAMEWORK"}
		for _, framework := range enabled.GetData() {
			rows = append(rows, string(framework))
		}

		return rows
	})
}

var compliancePerApplicationCmd = &cobra.Command{
	Use:     "per-application",
	Aliases: []string{"applications", "apps"},
	Short:   "List compliance statistics per application",
	Long: `Compliance per Application - Who Still Fails What

List, for each application, how many compliance items of each enabled
framework it fails. Filter to a few applications with --profile-ids. With
--size or --cursor the command returns one page as {items, nextCursor,
totalCount}; otherwise it walks every page.`,
	Example: `  # Every application, every framework it fails
  escape-cli compliance per-application

  # Only the applications named on the command line
  escape-cli compliance per-application --profile-ids <profile-id>,<profile-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[v3.GetCompliancePerApplication200ResponseDataInner]{}) {
			return nil
		}

		if err := runPagedList(cmd, complianceListPage, func(ctx context.Context, cursor string, size int) ([]v3.GetCompliancePerApplication200ResponseDataInner, *string, int, error) {
			return escape.ListCompliancePerApplication(ctx, cursor, size, complianceProfileIDs)
		}, compliancePerApplicationTable); err != nil {
			return fmt.Errorf("unable to list compliance per application: %w", err)
		}

		return nil
	},
}

var scansComplianceCmd = &cobra.Command{
	Use:     "compliance scan-id",
	Aliases: []string{"compliances"},
	Short:   "Get the compliance posture of a scan",
	Long: `Get Scan Compliance - Framework Items a Scan Failed

Fetch the compliance posture of one scan: per framework, whether the
organization enables it and which of its compliance items the scan failed.`,
	Example: `  # Failed items of one scan
  escape-cli scans compliance <scan-id>

  escape-cli scans compliance <scan-id> -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetScanCompliance200Response{}) {
			return nil
		}

		compliance, err := escape.GetScanCompliance(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get scan compliance: %w", err)
		}

		out.Table(compliance, func() []string {
			rows := []string{"FRAMEWORK\tENABLED\tFAILED ITEM COUNT\tFAILED ITEMS"}
			for _, framework := range compliance.GetData() {
				failed := framework.GetFailedItems()
				failedLabel := "-"
				if len(failed) > 0 {
					failedLabel = strings.Join(failed, ", ")
				}

				rows = append(rows, fmt.Sprintf("%s\t%t\t%d\t%s",
					framework.GetFramework(),
					framework.GetEnabled(),
					len(failed),
					failedLabel,
				))
			}

			return rows
		})

		return nil
	},
}

func compliancePerApplicationTable(applications []v3.GetCompliancePerApplication200ResponseDataInner) []string {
	rows := []string{"APPLICATION\tAPPLICATION ID\tFRAMEWORK\tFAILED\tTOTAL"}
	for _, application := range applications {
		applicationData := application.GetApplication()
		for _, compliance := range application.GetCompliances() {
			rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%d\t%d",
				applicationData.GetName(),
				applicationData.GetId(),
				compliance.GetFramework(),
				compliance.GetFailed(),
				compliance.GetTotal(),
			))
		}
	}

	return rows
}

func init() {
	compliancePerApplicationCmd.Flags().StringVar(&complianceProfileIDs, "profile-ids", "", "filter by profile IDs (comma-separated)")
	complianceListPage.bind(compliancePerApplicationCmd)

	complianceEnableCmd.Flags().StringSliceVar(&complianceEnableFrameworks, "framework", nil, "framework to measure; the full list replaces the current one: "+strings.Join(complianceFrameworkNames(), ", "))
	if err := complianceEnableCmd.Flags().SetAnnotation("framework", flagEnumAnnotation, complianceFrameworkNames()); err != nil {
		panic(err)
	}

	markMCPRequired(complianceEnableCmd.Flags(), "framework")
	// An empty list is the request that disables every framework. Required
	// still rejects an omitted property; allow-empty accepts [].
	markMCPAllowEmpty(complianceEnableCmd.Flags(), "framework")
	bindDestructiveGate(complianceEnableCmd, true)

	complianceCmd.AddCommand(complianceListCmd, complianceEnabledCmd, complianceEnableCmd, compliancePerApplicationCmd)
	scansCmd.AddCommand(scansComplianceCmd)
	rootCmd.AddCommand(complianceCmd)
}
