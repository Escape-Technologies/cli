package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	problemsDetailed bool
	problemsAssetIDs []string
	problemsDomains  []string
	problemsIssueIDs []string
	problemsTagIDs   []string
	problemsSearch   string
	problemsInitiators []string
	problemsKinds    []string
	problemsRisks    []string
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
		problems, next, err := escape.ListProblems(cmd.Context(), "", &escape.ListProblemsFilters{
			AssetIDs:   problemsAssetIDs,
			Domains:    problemsDomains,
			IssueIDs:   problemsIssueIDs,
			TagsIDs:    problemsTagIDs,
			Search:     problemsSearch,
			Initiators: problemsInitiators,
			Kinds:      problemsKinds,
			Risks:      problemsRisks,
		})
		if err != nil {
			return fmt.Errorf("unable to list problems: %w", err)
		}

		// Filter out applications without problems
		appsWithProblems := []v3.LastScanStatusSummarized{}
		for _, app := range problems {
			if app.HasLastResourceScan() {
				scan := app.GetLastResourceScan()
				if len(scan.GetProblems()) > 0 {
					appsWithProblems = append(appsWithProblems, app)
				}
			}
		}

		if problemsDetailed {
			// Show detailed problem information
			allProblems := []ProblemDetail{}
			for _, app := range appsWithProblems {
				scan := app.GetLastResourceScan()
				for _, problem := range scan.GetProblems() {
					allProblems = append(allProblems, ProblemDetail{
						AppID:      app.GetId(),
						AppName:    app.GetName(),
						ScanStatus: scan.GetStatus(),
						Code:       problem.GetCode(),
						Severity:   problem.GetSeverity(),
						Message:    problem.GetMessage(),
					})
				}
			}

			out.Table(allProblems, func() []string {
				res := []string{"ID\tNAME\tSCAN STATUS\tPROBLEM CODE\tSEVERITY\tMESSAGE"}
				for _, problem := range allProblems {
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
			})
		} else {
			// Show basic problem summary
			problemSummaries := []ProblemSummary{}
			for _, app := range appsWithProblems {
				scan := app.GetLastResourceScan()
				problemCount := len(scan.GetProblems())
				problemSummaries = append(problemSummaries, ProblemSummary{
					AppID:        app.GetId(),
					AppName:      app.GetName(),
					ScanStatus:   scan.GetStatus(),
					ProblemCount: problemCount,
				})
			}

			out.Table(problemSummaries, func() []string {
				res := []string{"ID\tNAME\tSCAN STATUS\tPROBLEMS"}
				for _, summary := range problemSummaries {
					res = append(res, fmt.Sprintf("%s\t%s\t%s\t%d",
						summary.AppID,
						summary.AppName,
						summary.ScanStatus,
						summary.ProblemCount,
					))
				}
				return res
			})
		}

		// Handle pagination
		for next != nil && *next != "" {
			problems, next, err = escape.ListProblems(
				cmd.Context(),
				*next,
				&escape.ListProblemsFilters{
					AssetIDs:   problemsAssetIDs,
					Domains:    problemsDomains,
					IssueIDs:   problemsIssueIDs,
					TagsIDs:    problemsTagIDs,
					Search:     problemsSearch,
					Initiators: problemsInitiators,
					Kinds:      problemsKinds,
					Risks:      problemsRisks,
				},
			)
			if err != nil {
				return fmt.Errorf("unable to list problems: %w", err)
			}

			// Filter out applications without problems
			appsWithProblems = []v3.LastScanStatusSummarized{}
			for _, app := range problems {
				if app.HasLastResourceScan() {
					scan := app.GetLastResourceScan()
					if len(scan.GetProblems()) > 0 {
						appsWithProblems = append(appsWithProblems, app)
					}
				}
			}

			if problemsDetailed {
				// Show detailed problem information
				allProblems := []ProblemDetail{}
				for _, app := range appsWithProblems {
					scan := app.GetLastResourceScan()
					for _, problem := range scan.GetProblems() {
						allProblems = append(allProblems, ProblemDetail{
							AppID:      app.GetId(),
							AppName:    app.GetName(),
							ScanStatus: scan.GetStatus(),
							Code:       problem.GetCode(),
							Severity:   problem.GetSeverity(),
							Message:    problem.GetMessage(),
						})
					}
				}

				out.Table(allProblems, func() []string {
					res := []string{"ID\tNAME\tSCAN STATUS\tPROBLEM CODE\tSEVERITY\tMESSAGE"}
					for _, problem := range allProblems {
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
				})
			} else {
				// Show basic problem summary
				problemSummaries := []ProblemSummary{}
				for _, app := range appsWithProblems {
					scan := app.GetLastResourceScan()
					problemCount := len(scan.GetProblems())
					problemSummaries = append(problemSummaries, ProblemSummary{
						AppID:        app.GetId(),
						AppName:      app.GetName(),
						ScanStatus:   scan.GetStatus(),
						ProblemCount: problemCount,
					})
				}

				out.Table(problemSummaries, func() []string {
					res := []string{"ID\tNAME\tSCAN STATUS\tPROBLEMS"}
					for _, summary := range problemSummaries {
						res = append(res, fmt.Sprintf("%s\t%s\t%s\t%d",
							summary.AppID,
							summary.AppName,
							summary.ScanStatus,
							summary.ProblemCount,
						))
					}
					return res
				})
			}
		}

		return nil
	},
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
	rootCmd.AddCommand(problemsCmd)
}
