package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
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
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		res, err := client.GetV1Locations(context.Background())
		if err != nil {
			return err
		}
		locations, err := api.ParseGetV1LocationsResponse(res)
		if err != nil {
			return err
		}
		switch output {
		case outputJSON:
			json.NewEncoder(os.Stdout).Encode(locations.JSON200)
		case outputYAML:
			yaml.NewEncoder(os.Stdout).Encode(locations.JSON200)
		default:
			if locations.JSON200 == nil {
				fmt.Println("No locations found")
			} else {
				for _, location := range *locations.JSON200 {
					fmt.Println(location.Name)
				}
			}
		}
		return nil
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete location",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("location id is required")
		}
		log.Info("Deleting location %s", args[0])
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		res, err := client.DeleteV1LocationsId(context.Background(), args[0])
		log.Debug("Deleting response %s", res.Status)
		if err != nil {
			return err
		}
		deleted, err := api.ParseDeleteV1LocationsIdResponse(res)
		if err != nil {
			return err
		}
		ok := true
		var data interface{}
		if deleted.JSON200 != nil {
			data = deleted.JSON200
		} else if deleted.JSON404 != nil {
			ok = false
			data = deleted.JSON404
			log.Info("Location %s not found: %s", args[0], deleted.JSON404.Message)
		} else {
			ok = false
			data = deleted.JSON400
			log.Info("Error deleting location %s: %s", args[0], deleted.JSON400.Message)
		}
		print(data, func() {
			if ok {
				fmt.Println("Location deleted")
			} else {
				fmt.Println("unable to delete location")
			}
		})
		return nil
	},
}
