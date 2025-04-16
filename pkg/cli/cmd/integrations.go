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
	RunE: func(cmd *cobra.Command, args []string) error {
		integrations, err := escape.ListIntegrations(cmd.Context())
		if err != nil {
			return err
		}
		out.Table(integrations, func() []string {
			res := []string{"ID\tKIND\tNAME\tLOCATION ID"}
			strPtr := ""
			for _, integration := range integrations {
				if integration.Id == nil {
					integration.Id = &strPtr
				}
				if integration.Name == nil {
					integration.Name = &strPtr
				}
				if integration.LocationId == nil {
					integration.LocationId = &strPtr
				}
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s", *integration.Id, integration.Kind, *integration.Name, *integration.LocationId))
			}
			return res
		})
		return nil
	},
}

var integrationsCreateCmd = &cobra.Command{
	Use:     "apply",
	Aliases: []string{"create", "update"},
	Short:   "Update the integration based on a configuration file",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return escape.UpsertIntegrationFromFile(cmd.Context(), args[0])
	},
}

var integrationsDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "remove"},
	Short:   "Delete an integration",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.DeleteIntegration(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		out.Print(struct {
			Msg string `json:"msg"`
		}{
			Msg: "Integration deleted",
		}, "Integration deleted")
		return nil
	},
}

var integrationsGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"describe"},
	Short:   "Get integration details",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		integration, err := escape.GetIntegration(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		data, _ := integration.GetData().MarshalJSON()
		out.Print(integration, fmt.Sprintf(
			"Name: %s\nId: %s\nLocationId: %s\n\nData: %s",
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
