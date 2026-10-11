package cmd

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"dashboard"},
	Short:   "Show organization security posture statistics",
	Long: `Organization Statistics - Security Posture at a Glance

Get a high-level overview of your organization's security posture:
application count, monitored asset count, and open issue counts by severity.

Use "escape-cli statistics" for the per-source breakdowns: endpoint coverage,
issue counts, and the trends time series.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return printOrganizationStatistics(cmd.Context())
	},
}

// printOrganizationStatistics renders the organization posture. The `statistics`
// group root runs it too: `statistics` used to be an alias of `stats`, and the
// sub-resources live under it now, so the bare command kept its meaning.
func printOrganizationStatistics(ctx context.Context) error {
	if out.Schema(v3.GetStatistics200Response{}) {
		return nil
	}

	stats, err := escape.GetStatistics(ctx)
	if err != nil {
		return fmt.Errorf("unable to get statistics: %w", err)
	}

	out.Table(stats, func() []string {
		return []string{
			"APPLICATIONS\tASSETS\tHIGH\tMEDIUM\tLOW\tINFO",
			fmt.Sprintf("%.0f\t%.0f\t%.0f\t%.0f\t%.0f\t%.0f",
				stats.Applications.GetCount(),
				stats.Assets.GetTotal(),
				stats.Issues.GetHigh(),
				stats.Issues.GetMedium(),
				stats.Issues.GetLow(),
				stats.Issues.GetInfo(),
			),
		}
	})

	return nil
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
