package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"statistics", "dashboard"},
	Short:   "Show organization security posture statistics",
	Long: `Organization Statistics - Security Posture at a Glance

Get a high-level overview of your organization's security posture:
application count, monitored asset count, and open issue counts by severity.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetStatistics200Response{}) {
			return nil
		}
		stats, err := escape.GetStatistics(cmd.Context())
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
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
