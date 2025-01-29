package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Integrations commands",
}

var integrationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		res, err := client.GetV1Integrations(context.Background())
		if err != nil {
			return err
		}
		integrations, err := api.ParseGetV1IntegrationsResponse(res)
		if err != nil {
			return err
		}
		switch output {
		case outputJSON:
			json.NewEncoder(os.Stdout).Encode(integrations.JSON200)
		case outputYAML:
			yaml.NewEncoder(os.Stdout).Encode(integrations.JSON200)
		default:
			if integrations.JSON200 == nil {
				fmt.Println("No integrations found")
			} else {
				for _, integration := range *integrations.JSON200 {
					fmt.Println(integration.Name)
				}
			}
		}
		return nil
	},
}
