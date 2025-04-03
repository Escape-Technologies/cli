package cli

import (
	"fmt"

	v1 "github.com/Escape-Technologies/cli/pkg/api/v1"
	"github.com/Escape-Technologies/cli/pkg/locations"
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
		client, err := v1.NewAPIClient()
		if err != nil {
			return err
		}
		data, pretty, err := locations.List(cmd.Context(), client)
		if err != nil {
			return err
		}
		return print(data, pretty)
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:     "del",
	Aliases: []string{"delete"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete location",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := v1.NewAPIClient()
		if err != nil {
			return err
		}
		id, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		data, pretty, err := locations.Delete(cmd.Context(), client, id)
		if err != nil {
			return err
		}
		return print(data, pretty)
	},
}

var locationsGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"read"},
	Args:    cobra.ExactArgs(1),
	Short:   "Get location",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := v1.NewAPIClient()
		if err != nil {
			return err
		}
		id, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		data, pretty, err := locations.Get(cmd.Context(), client, id)
		if err != nil {
			return err
		}
		return print(data, pretty)
	},
}

var locationsCreateInput = locations.LocationSchemaInput{}
var locationsCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Args:    cobra.ExactArgs(1),
	Short:   "Create a new location with the given name",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := v1.NewAPIClient()
		if err != nil {
			return err
		}
		locationsCreateInput.Name = args[0]
		data, pretty, err := locations.Create(cmd.Context(), client, locationsCreateInput)
		if err != nil {
			return err
		}
		return print(data, pretty)
	},
}

var locationsUpsertInput = locations.LocationSchemaInput{}
var locationsUpsertCmd = &cobra.Command{
	Use:   "upsert",
	Args:  cobra.ExactArgs(1),
	Short: "Get or create a location by it's name",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := v1.NewAPIClient()
		if err != nil {
			return err
		}
		locationsUpsertInput.Name = args[0]
		data, pretty, err := locations.Upsert(cmd.Context(), client, locationsUpsertInput)
		if err != nil {
			return err
		}
		return print(data, pretty)
	},
}

var locationsStartCmd = &cobra.Command{
	Use:   "start",
	Args:  cobra.ExactArgs(1),
	Short: "Start the private location",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := v1.NewAPIClient()
		if err != nil {
			return err
		}
		setupTerminalLog()
		return locations.Start(cmd.Context(), client, args[0])
	},
}
