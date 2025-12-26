package cmd

import (
	"fmt"
	"strconv"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/locations"
	"github.com/spf13/cobra"
)

var locationsSearch = ""
var locationsEnabled = false
var locationsLocationTypes = []string{}

var locationsCmd = &cobra.Command{
	Use:     "locations",
	Aliases: []string{"loc", "location"},
	Short:   "Manage private scanning locations for internal APIs",
	Long: `Manage Private Locations - Scan Internal and Private APIs

Private locations allow you to scan APIs behind firewalls, in private networks,
or on-premises infrastructure. Deploy agents in your environment to enable
security testing without exposing your APIs to the internet.

LOCATION TYPES:
  • PRIVATE   - Self-hosted agents in your infrastructure
  • ESCAPE    - Escape-managed cloud scanners
  • REPEATER  - Proxy/repeater configurations

USE CASES:
  • Scan APIs in private VPCs
  • Test pre-production environments
  • Kubernetes cluster scanning
  • On-premises application security

SETUP:
  1. Create location in Escape platform
  2. Deploy agent: escape-cli locations start <location-name>
  3. Configure profiles to use the location
  4. Start scans`,
}

var locationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List locations",
	Long: `List all locations.

Example output:
ID                                      NAME                       SSH PUBLIC KEY
00000000-0000-0000-0000-000000000001    example-location-1         ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... example1@email.com
00000000-0000-0000-0000-000000000002    example-location-2         ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... example2@email.com`,
	Example: `escape-cli locations list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		locations, next, err := escape.ListLocations(cmd.Context(), "", &escape.ListLocationsFilters{
			Search:        locationsSearch,
			Enabled:       locationsEnabled,
			LocationTypes: locationsLocationTypes,
		})
		if err != nil {
			return fmt.Errorf("failed to list locations: %w", err)
		}
		out.Table(locations, func() []string {
			res := []string{"ID\tNAME\tTYPE\tENABLED\tLINK"}
			for _, location := range locations {
				res = append(
					res,
					fmt.Sprintf(
						"%s\t%s\t%s\t%t\t%s",
						location.GetId(),
						location.GetName(),
						location.GetType(),
						location.GetEnabled(),
						location.GetLinks().LocationOverview,
					),
				)
			}
			return res
		})
		for next != nil && *next != "" {
			locations, next, err = escape.ListLocations(cmd.Context(), *next, &escape.ListLocationsFilters{
				Search:        locationsSearch,
				Enabled:       locationsEnabled,
				LocationTypes: locationsLocationTypes,
			})
			if err != nil {
				return fmt.Errorf("failed to list locations: %w", err)
			}
			out.Table(locations, func() []string {
				res := []string{"ID\tNAME\tTYPE\tENABLED\tLINK"}
				for _, location := range locations {
					res = append(
						res,
						fmt.Sprintf(
							"%s\t%s\t%s\t%t\t%s",
							location.GetId(),
							location.GetName(),
							location.GetType(),
							location.GetEnabled(),
							location.GetLinks().LocationOverview,
						),
					)
				}
				return res
			})
		}
		return nil
	},
}

var locationsGetCmd = &cobra.Command{
	Use:     "get location-id",
	Aliases: []string{"g"},
	Short:   "Get a location",
	Long:    `Get a location by its ID.`,
	Args:    cobra.ExactArgs(1),
	Example: `escape-cli locations get 00000000-0000-0000-0000-000000000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		location, err := escape.GetLocation(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get location: %w", err)
		}
		out.Table(location, func() []string {
			return []string{
				"ID\tNAME\tTYPE\tENABLED\tLINK",
				fmt.Sprintf("%s\t%s\t%s\t%s\t%s", location.GetId(), location.GetName(), location.GetType(), strconv.FormatBool(location.GetEnabled()), location.GetLinks().LocationOverview),
			}
		})
		return nil
	},
}
var locationsStartCmd = &cobra.Command{
	Use:   "start location-name",
	Short: "Start private location agent",
	Long: `Start Private Location Agent - Enable Internal API Scanning

Start the private location agent to establish a secure connection to the Escape Platform.
The agent will run continuously, processing scan requests for your internal APIs.

DEPLOYMENT:
  • Run on infrastructure with access to your private APIs
  • Requires network connectivity to Escape Platform
  • Supports Kubernetes, Docker, or bare metal deployment

OPERATION:
  The agent will:
  1. Authenticate with Escape Platform
  2. Establish secure tunnel
  3. Wait for scan requests
  4. Execute scans on internal APIs
  5. Report results back to platform

Run with -v for detailed logging. Use Ctrl+C to stop gracefully.`,
	Args: cobra.ExactArgs(1),
	Example: `  # Start location agent
  escape-cli locations start my-private-location

  # Start with verbose logging
  escape-cli locations start my-location -v

  # Run as systemd service
  sudo systemctl start escape-location@my-location`,
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
	Long: `Delete a location by its ID.

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
	locationsCmd.AddCommand(locationsGetCmd)
	locationsCmd.AddCommand(locationsStartCmd)
	locationsCmd.AddCommand(locationsDeleteCmd)
	rootCmd.AddCommand(locationsCmd)
	locationsListCmd.Flags().StringVarP(&locationsSearch, "search", "s", "", "Search term to filter locations by")
	locationsListCmd.Flags().BoolVarP(&locationsEnabled, "enabled", "e", false, "Filter by enabled locations")
	locationsListCmd.Flags().StringSliceVarP(&locationsLocationTypes, "type", "t", []string{}, "Filter by location type (private, escape, repeater)")
}
