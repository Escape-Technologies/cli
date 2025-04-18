package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var scansCmd = &cobra.Command{
	Use:     "scans",
	Aliases: []string{"sc", "scan"},
	Short:   "View scans results",
}

var scansListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all scans of an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		applicationID := args[0]
		scans, next, err := escape.ListScans(cmd.Context(), applicationID, "")
		if err != nil {
			return fmt.Errorf("unable to list scans: %w", err)
		}
		out.Table(scans, func() []string {
			res := []string{"ID\tSTATUS\tCREATED AT\tPROGRESS RATIO"}
			for _, scan := range scans {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%f", scan.GetId(), scan.GetStatus(), scan.GetCreatedAt(), scan.GetProgressRatio()))
			}
			return res
		})
		for next != "" {
			scans, next, err = escape.ListScans(cmd.Context(), applicationID, next)
			if err != nil {
				return fmt.Errorf("unable to list scans: %w", err)
			}
			out.Table(scans, func() []string {
				res := []string{}
				for _, scan := range scans {
					res = append(res, fmt.Sprintf("%s\t%s\t%s\t%f", scan.GetId(), scan.GetStatus(), scan.GetCreatedAt(), scan.GetProgressRatio()))
				}
				return res
			})
		}
		return nil
	},
}

var scanStartCmdWatch bool
var scanStartCmd = &cobra.Command{
	Use:   "start",
	Args:  cobra.ExactArgs(1),
	Short: "Start a new scan of an application",
	Run: func(_ *cobra.Command, _ []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var scanEventsCmdWatch bool
var scanEventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"ev", "event"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all events of a scan",
	Run: func(_ *cobra.Command, _ []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var scanGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"describe", "results", "res", "result", "issues", "iss"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all results (issues) of a scan",
	Run: func(_ *cobra.Command, _ []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var scanDownloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"dl", "zip"},
	Args:    cobra.ExactArgs(2), //nolint:mnd
	Short:   "Download a scan result exchange archive (zip export)",
	Run: func(_ *cobra.Command, _ []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

func init() {
	scansCmd.AddCommand(scansListCmd)
	scanStartCmd.PersistentFlags().BoolVarP(&scanStartCmdWatch, "watch", "w", false, "watch for events")
	scansCmd.AddCommand(scanStartCmd)
	scanEventsCmd.PersistentFlags().BoolVarP(&scanEventsCmdWatch, "watch", "w", false, "watch for events")
	scansCmd.AddCommand(scanEventsCmd)
	scansCmd.AddCommand(scanGetCmd)
	scansCmd.AddCommand(scanDownloadCmd)
	rootCmd.AddCommand(scansCmd)
}
