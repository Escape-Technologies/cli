package cmd

import (
	"github.com/spf13/cobra"
)

var integrationsCmd = &cobra.Command{
	Use:     "integrations",
	Aliases: []string{"int", "integration"},
	Short:   "Interact with your escape integrations",
}

var integrationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all integrations",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var integrationsCreateCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"create", "update"},
	Short:   "Update the integration based on a configuration file",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

var integrationsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an integration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(quentin@escape.tech): Implement this
	},
}

func init() {
	integrationsCmd.AddCommand(integrationsListCmd)
	integrationsCmd.AddCommand(integrationsCreateCmd)
	integrationsCmd.AddCommand(integrationsDeleteCmd)
	rootCmd.AddCommand(integrationsCmd)
}
