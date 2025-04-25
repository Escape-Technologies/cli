package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
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
	RunE: func(cmd *cobra.Command, _ []string) error {
		integrations, err := escape.ListIntegrations(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list integrations: %w", err)
		}
		out.Table(integrations, func() []string {
			res := []string{"ID\tKIND\tNAME\tLOCATION ID"}
			for _, integration := range integrations {
				res = append(
					res,
					fmt.Sprintf(
						"%s\t%s\t%s\t%s",
						integration.GetId(),
						integration.GetKind(),
						integration.GetName(),
						integration.GetLocationId(),
					),
				)
			}
			return res
		})
		return nil
	},
}

var integrationsCreateCmd = &cobra.Command{
	Use:     "apply integration-path",
	Aliases: []string{"create", "update"},
	Short:   "Update the integration based on a configuration file",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return escape.UpsertIntegrationFromFile(cmd.Context(), args[0])
	},
}

var integrationsDeleteCmd = &cobra.Command{
	Use:     "delete integration-id",
	Aliases: []string{"del", "remove"},
	Short:   "Delete an integration",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.DeleteIntegration(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}
		out.Log("Integration deleted")
		return nil
	},
}

var integrationsGetCmd = &cobra.Command{
	Use:     "get integration-id",
	Aliases: []string{"describe"},
	Short:   "Get integration details",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		integration, err := escape.GetIntegration(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get integration: %w", err)
		}
		data, _ := integration.GetData().MarshalJSON()
		out.Print(integration, fmt.Sprintf(
			"Name: %s\nId: %s\nLocationId: %s\nData: %s",
			integration.GetName(),
			integration.GetId(),
			integration.GetLocationId(),
			string(data),
		))
		return nil
	},
}

func init() {
	integrationsCmd.AddCommand(integrationsListCmd)
	integrationsCmd.AddCommand(integrationsCreateCmd)
	integrationsCmd.AddCommand(integrationsDeleteCmd)
	integrationsCmd.AddCommand(integrationsGetCmd)
	rootCmd.AddCommand(integrationsCmd)
}
