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
	Short:   "Interact with your escape locations",
}

var locationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all locations",
	RunE: func(cmd *cobra.Command, _ []string) error {
		locations, err := escape.ListLocations(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list locations: %w", err)
		}
		out.Table(locations, func() []string {
			res := []string{"ID\tNAME\tSSH PUBLIC KEY"}
			strPtr := ""
			for _, location := range locations {
				if location.Id == nil {
					location.Id = &strPtr
				}
				if location.Name == nil {
					location.Name = &strPtr
				}
				if location.SshPublicKey == nil {
					location.SshPublicKey = &strPtr
				}
				res = append(res, fmt.Sprintf("%s\t%s\t%s", *location.Id, *location.Name, *location.SshPublicKey))
			}
			return res
		})
		return nil
	},
}

var locationsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return locations.Start(cmd.Context(), args[0])
	},
}

var locationsDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "remove"},
	Short:   "Delete a location",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.DeleteLocation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to delete location: %w", err)
		}
		out.Print(struct {
			Msg string `json:"msg"`
		}{
			Msg: "Location deleted",
		}, "Location deleted")
		return nil
	},
}

func init() {
	locationsCmd.AddCommand(locationsListCmd)
	locationsCmd.AddCommand(locationsStartCmd)
	locationsCmd.AddCommand(locationsDeleteCmd)
	rootCmd.AddCommand(locationsCmd)
}
