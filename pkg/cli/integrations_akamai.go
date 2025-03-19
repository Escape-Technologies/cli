package cli

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/spf13/cobra"
)

var integrationsAkamaiCmd = &cobra.Command{
	Use:   "akamai",
	Short: "Akamai integration command",
}

var integrationsAkamaiList = &cobra.Command{
	Use:   "list",
	Short: "List integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		integrations, err := client.GetAkamaiIntegrationsWithResponse(cmd.Context())
		if err != nil {
			return err
		}
		print(
			integrations.JSON200,
			func() {
				if integrations.JSON200 == nil {
					fmt.Println("No integrations found")
				} else {
					for _, integration := range *integrations.JSON200 {
						fmt.Printf("%s\t%s\t%s\n", integration.Id, integration.Name, integration.LocationId)
					}
				}
			},
		)
		return nil
	},
}
