package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	statisticsEndpointsSource    string
	statisticsEndpointsProjectID string
	statisticsIssuesSource       string
	statisticsIssuesProjectID    string
	statisticsTrendsStart        string
	statisticsTrendsEnd          string
	statisticsTrendsInterval     string
)

// statisticsSourceUsage names the scanner sources the sub-resources accept.
// Two commands share it, so the list stays one line.
var statisticsSourceUsage = fmt.Sprintf("SOURCE: [%s %s] (default %s)", escape.StatisticsSourceDAST, escape.StatisticsSourcePentest, escape.StatisticsSourceDAST)

var statisticsCmd = &cobra.Command{
	Use:   "statistics",
	Short: "Read the detailed security statistics of the organization",
	Long: `Statistics - Coverage, Issue Counts and Trends

The organization overview is "escape-cli stats", which the bare
"escape-cli statistics" also runs. The subcommands break it down: how much of
the attack surface the scans reached, how the issues of one scanner source
are spread, and how both moved over time.

` + statisticsSourceUsage + `

COMMON WORKFLOWS:
  • How much of the attack surface is covered:
    $ escape-cli statistics endpoints

  • How many issues of one scanner source are still open:
    $ escape-cli statistics issues --source PENTEST

  • Whether the posture is improving:
    $ escape-cli statistics trends --start 2026-07-01T00:00:00Z --end 2026-10-01T00:00:00Z`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return printOrganizationStatistics(cmd.Context())
	},
}

var statisticsEndpointsCmd = &cobra.Command{
	Use:   "endpoints",
	Short: "Count the API routes, GraphQL resolvers and web pages reached",
	Long: `Endpoint Coverage - What the Scans Actually Reached

Count the API routes, GraphQL resolvers and web pages the scans reached, for
one scanner source. An application whose routes are only partly covered has
gaps the scanner never explored.`,
	Example: `  escape-cli statistics endpoints
  escape-cli statistics endpoints --source PENTEST --project-id <project-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetStatisticsEndpoints200Response{}) {
			return nil
		}

		endpoints, err := escape.GetStatisticsEndpoints(cmd.Context(), escape.StatisticsSource(statisticsEndpointsSource), statisticsEndpointsProjectID)
		if err != nil {
			return fmt.Errorf("unable to get endpoint statistics: %w", err)
		}

		out.Table(endpoints, func() []string {
			return []string{
				"API ROUTES\tGRAPHQL RESOLVERS\tWEB PAGES",
				fmt.Sprintf("%d\t%d\t%d",
					endpoints.GetApiRoutes(),
					endpoints.GetGraphqlResolvers(),
					endpoints.GetWebPages(),
				),
			}
		})

		return nil
	},
}

var statisticsIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "Count open and recently resolved issues of one scanner source",
	Long: `Issue Statistics - Open, Resolved, Grouped by Status and Severity

Count the issues of one scanner source: how many are open, how many were
resolved in the last 30 days, the average time it took to resolve them, and
the same counts grouped by status and by severity.`,
	Example: `  escape-cli statistics issues
  escape-cli statistics issues --source PENTEST -o json | jq '.bySeverity'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetStatisticsIssues200Response{}) {
			return nil
		}

		issues, err := escape.GetStatisticsIssues(cmd.Context(), escape.StatisticsSource(statisticsIssuesSource), statisticsIssuesProjectID)
		if err != nil {
			return fmt.Errorf("unable to get issue statistics: %w", err)
		}

		mttr, _ := issues.GetAverageTimeToResolveOk()
		out.Table(issues, func() []string {
			return []string{
				"OPEN\tREMEDIATED (30D)\tMTTR (DAYS)",
				fmt.Sprintf("%d\t%d\t%s",
					issues.GetOpen(),
					issues.GetRemediated(),
					formatStatisticsDays(mttr),
				),
			}
		})

		// JSON and YAML already carry the whole document: the breakdowns are
		// for the human reader, so they only print in pretty mode.
		if !out.IsPretty() {
			return nil
		}

		if statuses := issues.GetByStatus(); len(statuses) > 0 {
			out.Table(statuses, func() []string {
				rows := []string{"STATUS\tCOUNT"}
				for _, status := range statuses {
					rows = append(rows, fmt.Sprintf("%s\t%d", status.GetStatus(), status.GetCount()))
				}

				return rows
			})
		}

		if severities := issues.GetBySeverity(); len(severities) > 0 {
			out.Table(severities, func() []string {
				rows := []string{"SEVERITY\tCOUNT"}
				for _, severity := range severities {
					rows = append(rows, fmt.Sprintf("%s\t%d", severity.GetSeverity(), severity.GetCount()))
				}

				return rows
			})
		}

		return nil
	},
}

var statisticsTrendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Get the time-bucketed security trends of the organization",
	Long: `Trends - Open versus Resolved Issues, and API Health over Time

Return the time series of the organization over the half-open window
[start, end): the open and resolved issue counts of each bucket, and the
average API health score of each bucket. Both series share the same bucket
grid, so row i of one is row i of the other.

The resolved count of a bucket is cumulative: once an issue is resolved it is
counted in every later bucket, so a value repeats. Difference two consecutive
buckets to get the resolutions of that bucket.`,
	Example: `  escape-cli statistics trends --start 2026-07-01T00:00:00Z --end 2026-10-01T00:00:00Z
  escape-cli statistics trends --start 2026-07-01T00:00:00Z --end 2026-10-01T00:00:00Z --interval 1 week`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetStatisticsTrends200Response{}) {
			return nil
		}

		trends, err := escape.GetStatisticsTrends(
			cmd.Context(),
			statisticsTrendsStart,
			statisticsTrendsEnd,
			statisticsTrendsInterval,
		)
		if err != nil {
			return fmt.Errorf("unable to get statistics trends: %w", err)
		}

		out.Table(trends, func() []string {
			rows := []string{"BUCKET\tOPEN\tRESOLVED\tAVERAGE API HEALTH"}
			health := trends.GetAverageApiHealth()
			for index, bucket := range trends.GetOpenResolved() {
				rows = append(rows, fmt.Sprintf("%s\t%d\t%d\t%s",
					bucket.GetX(),
					bucket.GetOpen(),
					bucket.GetResolved(),
					formatStatisticsScore(apiHealthAt(health, index)),
				))
			}

			return rows
		})

		return nil
	},
}

// apiHealthAt reads the health score sharing the bucket at index. Both series
// are on the same grid, so a shorter one simply has no score for that bucket.
func apiHealthAt(health []v3.GetStatisticsTrends200ResponseAverageApiHealthInner, index int) *float32 {
	if index >= len(health) {
		return nil
	}

	score, _ := health[index].GetYOk()

	return score
}

func formatStatisticsDays(days *float32) string {
	if days == nil {
		return "-"
	}

	return fmt.Sprintf("%.1f", *days)
}

func formatStatisticsScore(score *float32) string {
	if score == nil {
		return "-"
	}

	return fmt.Sprintf("%.1f", *score)
}

func init() {
	statisticsEndpointsCmd.Flags().StringVar(&statisticsEndpointsSource, "source", "", statisticsSourceUsage)
	statisticsEndpointsCmd.Flags().StringVar(&statisticsEndpointsProjectID, "project-id", "", "filter by project ID")
	statisticsIssuesCmd.Flags().StringVar(&statisticsIssuesSource, "source", "", statisticsSourceUsage)
	statisticsIssuesCmd.Flags().StringVar(&statisticsIssuesProjectID, "project-id", "", "filter by project ID")
	statisticsTrendsCmd.Flags().StringVar(&statisticsTrendsStart, "start", "", "window start (ISO 8601, required)")
	statisticsTrendsCmd.Flags().StringVar(&statisticsTrendsEnd, "end", "", "window end (ISO 8601, required)")
	statisticsTrendsCmd.Flags().StringVar(&statisticsTrendsInterval, "interval", "", `bucket width (e.g. "1 day", "1 week")`)
	_ = statisticsTrendsCmd.MarkFlagRequired("start")
	_ = statisticsTrendsCmd.MarkFlagRequired("end")

	statisticsCmd.AddCommand(statisticsEndpointsCmd, statisticsIssuesCmd, statisticsTrendsCmd)
	rootCmd.AddCommand(statisticsCmd)
}
