package cli

import (
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var locationsCmd = &cobra.Command{
	Use:     "locations",
	Aliases: []string{"loc"},
	Short:   "Locations commands",
}

var locationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		locations, err := client.ListLocationsWithResponse(cmd.Context())
		if err != nil {
			return err
		}
		print(locations.JSON200, func() {
			if locations.JSON200 == nil {
				fmt.Println("No locations found")
			} else {
				for _, location := range *locations.JSON200 {
					size := len(*location.Type)
					fmt.Printf("%s%s%s\n", *location.Type, strings.Repeat(" ", 10-size), *location.Name)
				}
			}
		})
		return nil
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:     "del",
	Aliases: []string{"delete"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete location",
	RunE: func(cmd *cobra.Command, args []string) error {
		idString := args[0]
		id, err := uuid.Parse(idString)
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		log.Info("Deleting location %s", args[0])
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		deleted, err := client.DeleteLocationWithResponse(cmd.Context(), id)
		if err != nil {
			return err
		}
		if deleted.JSON200 != nil {
			print(deleted.JSON200, func() {
				fmt.Println("Location deleted")
			})
		} else if deleted.JSON404 != nil {
			print(deleted.JSON404, func() {
				fmt.Printf("Location %s not found: %s", idString, deleted.JSON404.Message)
			})
		} else if deleted.JSON400 != nil {
			print(deleted.JSON400, func() {
				fmt.Printf("Unable to delete location: %s", deleted.JSON400.Message)
			})
		} else {
			print(unknowError, func() {
				fmt.Printf("Unable to delete location: Unknown error")
			})
		}
		return nil
	},
}

var locationsGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"read"},
	Args:    cobra.ExactArgs(1),
	Short:   "Get location",
	RunE: func(cmd *cobra.Command, args []string) error {
		idString := args[0]
		id, err := uuid.Parse(idString)
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		log.Info("Getting location %s", args[0])
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		location, err := client.GetLocationWithResponse(cmd.Context(), id)
		if err != nil {
			return err
		}
		if location.JSON200 != nil {
			print(location.JSON200, func() {
				fmt.Println("Location:", *location.JSON200.Name)
				fmt.Println("Type:", *location.JSON200.Type)
				fmt.Println("Id:", *location.JSON200.Id)
			})
		} else if location.JSON404 != nil {
			print(location.JSON404, func() {
				fmt.Printf("Location %s not found: %s", idString, location.JSON404.Message)
			})
		} else {
			print(unknowError, func() {
				fmt.Printf("Unable to get location: Unknown error")
			})
		}
		return nil
	},
}

var locationsCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Args:    cobra.ExactArgs(1),
	Short:   "Create a new location with the given name",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		log.Info("Deleting location %s", args[0])
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		location, err := client.CreateLocationWithResponse(cmd.Context(), api.CreateLocationJSONRequestBody{
			Name: name,
		})
		if err != nil {
			return err
		}
		if location.JSON200 != nil {
			print(location.JSON200, func() {
				fmt.Println("Location:", *location.JSON200.Name)
				fmt.Println("Type:", *location.JSON200.Type)
				fmt.Println("Id:", *location.JSON200.Id)
			})
		} else if location.JSON400 != nil {
			print(location.JSON400, func() {
				fmt.Printf("Unable to create location: %s", location.JSON400.Message)
			})
		} else if location.JSON500 != nil {
			print(location.JSON500, func() {
				fmt.Printf("Unable to create location: %s", location.JSON500.Message)
			})
		} else {
			print(unknowError, func() {
				fmt.Printf("Unable to upsert location: Unknown error")
			})
		}
		return nil
	},
}

var locationsUpsertCmd = &cobra.Command{
	Use:   "upsert",
	Args:  cobra.ExactArgs(1),
	Short: "Get or create a location by it's name",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		log.Info("Upserting location %s", args[0])
		client, err := api.NewAPIClient()
		if err != nil {
			return err
		}
		location, err := client.UpsertLocationWithResponse(cmd.Context(), api.UpsertLocationJSONRequestBody{
			Name: name,
		})
		if err != nil {
			return err
		}
		if location.JSON200 != nil {
			print(location.JSON200, func() {
				fmt.Println("Location:", *location.JSON200.Name)
				fmt.Println("Type:", *location.JSON200.Type)
				fmt.Println("Id:", *location.JSON200.Id)
			})
		} else if location.JSON400 != nil {
			print(location.JSON400, func() {
				fmt.Printf("Unable to upsert location: %s", location.JSON400.Message)
			})
		} else if location.JSON500 != nil {
			print(location.JSON500, func() {
				fmt.Printf("Unable to upsert location: %s", location.JSON500.Message)
			})
		} else {
			print(unknowError, func() {
				fmt.Printf("Unable to upsert location: Unknown error")
			})
		}
		return nil
	},
}
