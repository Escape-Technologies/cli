package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/locations"
	"github.com/spf13/cobra"
)

var locationsCmd = &cobra.Command{
	Use:     "locations",
	Aliases: []string{"loc", "location"},
	Short:   "Interact with locations",
	Long:    "Interact with your escape locations",
}

var locationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List locations",
	Long:    `List all locations.

Example output:
ID                                      NAME                       SSH PUBLIC KEY
00000000-0000-0000-0000-000000000001    example-location-1         ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... example1@email.com
00000000-0000-0000-0000-000000000002    example-location-2         ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... example2@email.com`,
	Example: `escape-cli locations list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		locations, next, err := escape.ListLocations(cmd.Context(), "")
		if err != nil {
			return fmt.Errorf("failed to list locations: %w", err)
		}
		out.Table(locations, func() []string {
			res := []string{"ID\tNAME\tSSH PUBLIC KEY"}
			for _, location := range locations {
				res = append(
					res,
					fmt.Sprintf(
						"%s\t%s\t%s",
						location.GetId(),
						location.GetName(),
						location.GetSshPublicKey(),
					),
				)
			}
			return res
		})
		for next != "" {
			locations, next, err = escape.ListLocations(cmd.Context(), next)
			if err != nil {
				return fmt.Errorf("failed to list locations: %w", err)
			}
			out.Table(locations, func() []string {
				res := []string{}
				for _, location := range locations {
					res = append(res, fmt.Sprintf("%s\t%s\t%s", location.GetId(), location.GetName(), location.GetSshPublicKey()))
				}
				return res
			})
		}
		return nil
	},
}

var locationsStartCmd = &cobra.Command{
	Use:     "start location-name",
	Short:   "Start a location",
	Long:    `Start a location by its name. This will establish a connection to the Escape Platform.

Example output:
[info] Verbose mode: 0
[info] escape-cli version: Version: 0.1.14, Commit: 05ffe67, BuildDate: 2025-04-25T15:30:54Z
[info] Private location testdoc in sync with Escape Platform, starting location...
[info] Private location ready to accept connections
[info] Connected to k8s API
...
[The command will continue running until interrupted with Ctrl+C]`,
	Args:    cobra.ExactArgs(1),
	Example: `escape-cli locations start my-location`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if rootCmdVerbose == 0 {
			out.SetupTerminalLog()
			defer out.StopTerminalLog()
		}
		return locations.Start(cmd.Context(), args[0])
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:     "delete location-id",
	Aliases: []string{"del", "remove"},
	Short:   "Delete a location",
	Long:    `Delete a location by its ID.

Example output:
Location deleted`,
	Args:    cobra.ExactArgs(1),
	Example: `escape-cli locations delete 00000000-0000-0000-0000-000000000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.DeleteLocation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to delete location: %w", err)
		}
		out.Log("Location deleted")
		return nil
	},
}

func init() {
	locationsCmd.AddCommand(locationsListCmd)
	locationsCmd.AddCommand(locationsStartCmd)
	locationsCmd.AddCommand(locationsDeleteCmd)
	rootCmd.AddCommand(locationsCmd)
}
