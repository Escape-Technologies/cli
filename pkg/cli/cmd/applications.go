package cmd

import (
	"github.com/spf13/cobra"
)

var applicationsCmd = &cobra.Command{
	Use:     "applications",
	Aliases: []string{"app", "application"},
	Short:   "Interact with your escape applications",
}

var applicationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all applications",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var applicationGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"describe"},
	Short:   "Get details about an application current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var applicationUpdateSchemaCmd = &cobra.Command{
	Use:   "update-schema",
	Short: "Update the schema of an application",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var applicationUpdateConfigCmd = &cobra.Command{
	Use:   "update-config",
	Short: "Update the configuration of an application",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsListCmd)
	applicationsCmd.AddCommand(applicationGetCmd)
	applicationsCmd.AddCommand(applicationUpdateSchemaCmd)
	applicationsCmd.AddCommand(applicationUpdateConfigCmd)
	rootCmd.AddCommand(applicationsCmd)
}
