package cmds

import (
	"github.com/spf13/cobra"
)

var locationsCmd = &cobra.Command{
	Use:     "locations",
	Aliases: []string{"loc", "location"},
	Short:   "Interact with your escape locations",
}

var locationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all locations",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var locationsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a location",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a location",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

func init() {
	locationsCmd.AddCommand(locationsListCmd)
	locationsCmd.AddCommand(locationsStartCmd)
	locationsCmd.AddCommand(locationsDeleteCmd)
	RootCmd.AddCommand(locationsCmd)
}
