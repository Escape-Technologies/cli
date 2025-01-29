package cli

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/spf13/cobra"
)

var integrationsKubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Kubernetes integration command",
}

var integrationsKubernetesList = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(0),
	Short:   "List integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		integrations, err := client.GetV1IntegrationsKubernetesWithResponse(context.Background())
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

var integrationsKubernetesDelete = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete integration given an id",
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		res, err := client.DeleteV1IntegrationsKubernetesIdWithResponse(context.Background(), id)
		if err != nil {
			return err
		}
		if res.JSON200 != nil {
			print(res.JSON200, func() {
				fmt.Println("Integration deleted")
			})
		} else if res.JSON404 != nil {
			print(res.JSON404, func() {
				fmt.Println("Integration not found")
			})
		} else if res.JSON500 != nil {
			print(res.JSON500, func() {
				fmt.Println("Error deleting integration", res.JSON500.Message)
			})
		}
		return nil
	},
}
