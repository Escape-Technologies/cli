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

var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Locations commands",
}

var locationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := initClient()
		if err != nil {
			return err
		}
		res, err := client.GetPrivateLocations(context.Background())
		if err != nil {
			return err
		}
		locations, err := api.ParseGetPrivateLocationsResponse(res)
		if err != nil {
			return err
		}
		switch output {
		case outputJSON:
			json.NewEncoder(os.Stdout).Encode(locations.JSON200.Locations)
		case outputYAML:
			yaml.NewEncoder(os.Stdout).Encode(locations.JSON200.Locations)
		default:
			for _, location := range locations.JSON200.Locations {
				fmt.Println(location.Name)
			}
		}
		return nil
	},
}
