package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var scanStartCmdWatch bool
var scanStartCmd = &cobra.Command{
	Use:   "start",
	Args:  cobra.ExactArgs(1),
	Short: "Start a new scan of an application",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var scanEventsCmdWatch bool
var scanEventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"ev", "event"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all events of a scan",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var scanResultsCmd = &cobra.Command{
	Use:     "results",
	Aliases: []string{"res", "result"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all results (issues) of a scan",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var scanDownloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"dl", "zip"},
	Args:    cobra.ExactArgs(2),
	Short:   "Download a scan result exchange archive (zip export)",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

func init() {
	scansCmd.AddCommand(scansListCmd)
	scanStartCmd.PersistentFlags().BoolVarP(&scanStartCmdWatch, "watch", "w", false, "watch for events")
	scansCmd.AddCommand(scanStartCmd)
	scanEventsCmd.PersistentFlags().BoolVarP(&scanEventsCmdWatch, "watch", "w", false, "watch for events")
	scansCmd.AddCommand(scanEventsCmd)
	scansCmd.AddCommand(scanResultsCmd)
	scansCmd.AddCommand(scanDownloadCmd)
	rootCmd.AddCommand(scansCmd)
}
