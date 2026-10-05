package cmd

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/spf13/cobra"
)

// problemsCommandOutput is the schema for `escape-cli problems`.
// Default output is one page of ProblemSummary. --all expands each application
// into ProblemDetail rows. Both are pages because MCP requests one page.
var problemsCommandOutput = providedSchema{schema: problemsOutputSchema()}

var (
	problemsDetailed   bool
	problemsAssetIDs   []string
	problemsDomains    []string
	problemsIssueIDs   []string
	problemsTagIDs     []string
	problemsSearch     string
	problemsInitiators []string
	problemsKinds      []string
	problemsRisks      []string
	problemsPage       pageFlags
)

var problemsCmd = &cobra.Command{
	Use:     "problems",
	Aliases: []string{"problem"},
	Short:   "List scan problems across all applications",
	Long: `List Scan Problems - View All Applications with Issues

Display all applications that have at least one scan problem, with optional information
about each problem. Problems represent issues encountered during scanning
that prevented successful completion or indicate configuration problems.

PROBLEM TYPES:
  • SCAN_FAILED     - Scan could not complete due to technical issues
  • CONFIG_ERROR    - Profile configuration problems
  • AUTH_FAILED     - Authentication or authorization issues
  • NETWORK_ERROR   - Network connectivity problems
  • TIMEOUT         - Scan exceeded time limits
  • RESOURCE_ERROR  - Insufficient resources or quotas

FILTER OPTIONS:
  -a, --all         Show detailed problem information (code, message, severity)
  --asset-ids       Filter by specific asset IDs
  --domains         Filter by domain names
  --issue-ids       Filter by issue IDs
  --tag-ids         Filter by tag IDs
  -s, --search      Search across application names and descriptions
  --initiators      Filter by scan initiators
  --kinds           Filter by scan kinds
  --risks           Filter by risk types

OUTPUT FORMATS:
  • Basic: Shows application ID, name, scan status, and problem count
  • Detailed (-a): Includes problem codes, messages, and severity levels

Example output (basic):
ID                                      NAME                    SCAN STATUS    PROBLEMS
00000000-0000-0000-0000-000000000001    My Application         FAILED         2
00000000-0000-0000-0000-000000000002    Another App            ERROR          1

Example output (detailed with -a):
ID                                      NAME                    SCAN STATUS    PROBLEM CODE    SEVERITY    MESSAGE
00000000-0000-0000-0000-000000000001    My Application         FAILED         AUTH_FAILED     HIGH        Authentication failed: invalid credentials
00000000-0000-0000-0000-000000000001    My Application         FAILED         TIMEOUT         MEDIUM      Scan exceeded maximum duration limit
00000000-0000-0000-0000-000000000002    Another App            ERROR          CONFIG_ERROR    LOW         Profile configuration is invalid`,
	Example: `  # List all applications with scan problems
  escape-cli problems

  # Show detailed problem information
  escape-cli problems --all

  # Filter by specific assets
  escape-cli problems --asset-ids "asset1,asset2"

  # Search for problems in specific applications
  escape-cli problems --search "production"

  # Export problems to JSON
  escape-cli problems --all -o json > scan-problems.json

  # Filter by scan initiators
  escape-cli problems --initiators "scheduled,manual"`,

	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(problemsCommandOutput) {
			return nil
		}

		filters := &escape.ListProblemsFilters{
			AssetIDs:   problemsAssetIDs,
			Domains:    problemsDomains,
			IssueIDs:   problemsIssueIDs,
			TagsIDs:    problemsTagIDs,
			Search:     problemsSearch,
			Initiators: problemsInitiators,
			Kinds:      problemsKinds,
			Risks:      problemsRisks,
		}
		fetch := func(ctx context.Context, cursor string, size int) ([]v3.ProfileScanProblemsRow, *string, int, error) {
			rows, next, total, err := escape.ListProblems(ctx, cursor, filters, size)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("unable to list problems: %w", err)
			}

			return rows, next, total, nil
		}
		// --all expands each application into one row per problem. Both modes
		// still walk API pages through runPagedList, so MCP's size default
		// bounds the request.
		if problemsDetailed {
			return runPagedList(cmd, problemsPage, func(ctx context.Context, cursor string, size int) ([]ProblemDetail, *string, int, error) {
				rows, next, total, err := fetch(ctx, cursor, size)
				if err != nil {
					return nil, nil, 0, err
				}

				return problemDetails(rows), next, total, nil
			}, problemDetailTable)
		}

		return runPagedList(cmd, problemsPage, func(ctx context.Context, cursor string, size int) ([]ProblemSummary, *string, int, error) {
			rows, next, total, err := fetch(ctx, cursor, size)
			if err != nil {
				return nil, nil, 0, err
			}

			return problemSummaries(rows), next, total, nil
		}, problemSummaryTable)
	},
}

func appsWithProblems(rows []v3.ProfileScanProblemsRow) []v3.ProfileScanProblemsRow {
	apps := []v3.ProfileScanProblemsRow{}
	for _, app := range rows {
		if !app.HasLastResourceScan() {
			continue
		}

		scan := app.GetLastResourceScan()
		if len(scan.GetProblems()) == 0 {
			continue
		}

		apps = append(apps, app)
	}

	return apps
}

func problemSummaries(rows []v3.ProfileScanProblemsRow) []ProblemSummary {
	summaries := []ProblemSummary{}
	for _, app := range appsWithProblems(rows) {
		scan := app.GetLastResourceScan()
		summaries = append(summaries, ProblemSummary{
			AppID:        app.GetId(),
			AppName:      app.GetName(),
			ScanStatus:   scan.GetStatus(),
			ProblemCount: len(scan.GetProblems()),
		})
	}

	return summaries
}

func problemDetails(rows []v3.ProfileScanProblemsRow) []ProblemDetail {
	details := []ProblemDetail{}
	for _, app := range appsWithProblems(rows) {
		scan := app.GetLastResourceScan()
		for _, problem := range scan.GetProblems() {
			details = append(details, ProblemDetail{
				AppID:      app.GetId(),
				AppName:    app.GetName(),
				ScanStatus: scan.GetStatus(),
				Code:       problem.GetCode(),
				Severity:   problem.GetSeverity(),
				Message:    problem.GetMessage(),
			})
		}
	}

	return details
}

func problemSummaryTable(summaries []ProblemSummary) []string {
	res := []string{"ID\tNAME\tSCAN STATUS\tPROBLEMS"}
	for _, summary := range summaries {
		res = append(res, fmt.Sprintf("%s\t%s\t%s\t%d",
			summary.AppID,
			summary.AppName,
			summary.ScanStatus,
			summary.ProblemCount,
		))
	}

	return res
}

func problemDetailTable(details []ProblemDetail) []string {
	res := []string{"ID\tNAME\tSCAN STATUS\tPROBLEM CODE\tSEVERITY\tMESSAGE"}
	for _, problem := range details {
		res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
			problem.AppID,
			problem.AppName,
			problem.ScanStatus,
			problem.Code,
			problem.Severity,
			problem.Message,
		))
	}

	return res
}

// problemsOutputSchema is the page object whose items are either summaries or
// details. --all changes the element type; the page fields stay the same.
// MCP requires a top-level object, so the two shapes live under items.
func problemsOutputSchema() *clischema.JSONSchema {
	summary := clischema.Generate(Page[ProblemSummary]{})
	detail := clischema.Generate(Page[ProblemDetail]{})
	summaryItems := summary.Properties["items"]
	detailItems := detail.Properties["items"]
	summaryItems.Description = "One ProblemSummary per application. Printed when --all is omitted."
	detailItems.Description = "One ProblemDetail per problem. Printed when --all is set."
	summary.Description = "Without --all, items are ProblemSummary. With --all, items are ProblemDetail. A full listing (no --size or --cursor) prints the items array by itself; one page prints this object."
	summary.Properties["items"] = &clischema.JSONSchema{
		Description: "ProblemSummary rows, or ProblemDetail rows when --all is set.",
		OneOf:       []*clischema.JSONSchema{summaryItems, detailItems},
	}

	return summary
}

// ProblemSummary represents a summary of problems for an application
type ProblemSummary struct {
	AppID        string
	AppName      string
	ScanStatus   string
	ProblemCount int
}

// ProblemDetail represents detailed information about a specific problem
type ProblemDetail struct {
	AppID      string
	AppName    string
	ScanStatus string
	Code       string
	Severity   string
	Message    string
}

func init() {
	problemsCmd.Flags().BoolVarP(&problemsDetailed, "all", "a", false, "show detailed problem information (code, message, severity)")
	problemsCmd.Flags().StringSliceVarP(&problemsAssetIDs, "asset-ids", "", []string{}, "filter by asset IDs (comma-separated)")
	problemsCmd.Flags().StringSliceVarP(&problemsDomains, "domains", "", []string{}, "filter by domain names (comma-separated)")
	problemsCmd.Flags().StringSliceVarP(&problemsIssueIDs, "issue-ids", "", []string{}, "filter by issue IDs (comma-separated)")
	problemsCmd.Flags().StringSliceVarP(&problemsTagIDs, "tag-ids", "", []string{}, "filter by tag IDs (comma-separated)")
	problemsCmd.Flags().StringVarP(&problemsSearch, "search", "s", "", "search across application names and descriptions")
	problemsCmd.Flags().StringSliceVarP(&problemsInitiators, "initiators", "", []string{}, "filter by scan initiators (comma-separated)")
	problemsCmd.Flags().StringSliceVarP(&problemsKinds, "kinds", "", []string{}, "filter by scan kinds (comma-separated)")
	problemsCmd.Flags().StringSliceVarP(&problemsRisks, "risks", "", []string{}, "filter by risk types (comma-separated)")
	problemsPage.bind(problemsCmd)
	rootCmd.AddCommand(problemsCmd)
}
