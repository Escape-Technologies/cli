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
	Short:   "Interact with integrations",
	Long:    "Interact with your escape integrations",
}

var integrationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List integrations",
	Long:    `List all integrations.

Example output:
ID                                      KIND               NAME                          LOCATION ID
00000000-0000-0000-0000-000000000001    AZURE_DEVOPS       Example Azure Integration     
00000000-0000-0000-0000-000000000002    KUBERNETES         Example K8s Integration      00000000-0000-0000-0000-000000000099`,
	Example: `escape-cli integrations list`,
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
	Short:   "Apply integration config",
	Long:    `Update the integration based on a configuration file (yaml or json).

Example file content:
{
  "data": {
    "kind": "AKAMAI",
    "parameters": {
      "client_secret": "your-secret",
      "host": "your-host",
      "access_token": "your-token",
      "client_token": "your-client-token"
    }
  },
  "name": "Your Integration Name"
}

More information about the integration file format can be found here: https://public.escape.tech/v2/#tag/integrations/POST/integrations`,
	Example: `escape-cli integrations apply integration.yaml
escape-cli integrations apply integration.json`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return escape.UpsertIntegrationFromFile(cmd.Context(), args[0])
	},
}

var integrationsDeleteCmd = &cobra.Command{
	Use:     "delete integration-id",
	Aliases: []string{"del", "remove"},
	Short:   "Delete integration",
	Long:    `Delete an integration
	
Example output:
Integration deleted`,
	Example: `escape-cli integrations delete 00000000-0000-0000-0000-000000000000`,
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
	Long:    `Get detailed information about an integration.

Example output:
Name: example-github-integration
Id: 00000000-0000-0000-0000-000000000001
LocationId: 
Data: {"kind":"GITHUB_API_KEY","parameters":{"api_key":"github_pat_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"}}`,
	Args:    cobra.ExactArgs(1),
	Example: `escape-cli integrations get 00000000-0000-0000-0000-000000000000`,
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
