package cli

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
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
		locations, err := escape.GetLocations(cmd.Context())
		if err != nil {
			return err
		}
		print(
			locations,
			func() {
				for _, location := range locations {
					fmt.Println(location.String())
				}
			},
		)
		return nil
	},
}

var locationsGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"read"},
	Args:    cobra.ExactArgs(1),
	Short:   "Get location",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		location, err := escape.GetLocation(cmd.Context(), id)
		if err != nil {
			return err
		}
		return print(location, func() {
			fmt.Println(location.String())
		})
	},
}

var locationsCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new"},
	Args:    cobra.ExactArgs(2),
	Short:   "Create a new location with the given name and SSH public key",
	RunE: func(cmd *cobra.Command, args []string) error {
		location := escape.Location{
			Name:          args[0],
			SSHPublicKey:  args[1],
		}
		createdLocation, err := escape.CreateLocation(cmd.Context(), location)
		if err != nil {
			return err
		}
		return print(createdLocation, func() {
			fmt.Println(createdLocation.String())
		})
	},
}

var locationsStartCmd = &cobra.Command{
	Use:   "start",
	Args:  cobra.ExactArgs(1),
	Short: "Start the private location",
	RunE: func(cmd *cobra.Command, args []string) error {
		return locations.Start(cmd.Context(), args[0])
	},
}
