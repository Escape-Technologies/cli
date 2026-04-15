package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/locations"
	"github.com/spf13/cobra"
)

var locationsSearch = ""
var locationsEnabled = false
var locationsLocationTypes = []string{}
var locationsSortType string
var locationsSortDirection string

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
ID                                      NAME                 TYPE     ENABLED  LAST SEEN              LINK
00000000-0000-0000-0000-000000000001    example-location-1   PRIVATE  true     2025-07-24T10:00:00Z  https://app.escape.tech/...
00000000-0000-0000-0000-000000000002    example-location-2   ESCAPE   false                           https://app.escape.tech/...`,
	Example: `escape-cli locations list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Output JSON Schema if requested
		if out.Schema([]v3.LocationSummarized{}) {
			return nil
		}

		filters := &escape.ListLocationsFilters{
			Search:        locationsSearch,
			Enabled:       locationsEnabled,
			LocationTypes: locationsLocationTypes,
			SortType:      locationsSortType,
			SortDirection: locationsSortDirection,
		}
		locations, next, err := escape.ListLocations(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("failed to list locations: %w", err)
		}
		allLocations := locations
		for next != nil && *next != "" {
			locations, next, err = escape.ListLocations(cmd.Context(), *next, filters)
			if err != nil {
				return fmt.Errorf("failed to list locations: %w", err)
			}
			allLocations = append(allLocations, locations...)
		}
		out.Table(allLocations, func() []string {
			res := []string{"ID\tNAME\tTYPE\tENABLED\tLAST SEEN\tLINK"}
			for _, location := range allLocations {
				res = append(
					res,
					fmt.Sprintf(
						"%s\t%s\t%s\t%t\t%s\t%s",
						location.GetId(),
						location.GetName(),
						location.GetType(),
						location.GetEnabled(),
						stringValue(location.AdditionalProperties["lastSeenAt"]),
						location.GetLinks().LocationOverview,
					),
				)
			}
			return res
		})
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
		// Output JSON Schema if requested
		if out.Schema(v3.CreateLocation200Response{}) {
			return nil
		}

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

var (
	locationCreateName         string
	locationCreateSSHPublicKey string
	locationUpdateName         string
	locationUpdateSSHPublicKey string
	locationUpdateEnabled      bool
)

var locationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new private location",
	Long: `Create Private Location - Register a New Scanning Agent

Register a new private location in the Escape platform. After creation,
deploy the agent using 'escape-cli locations start <name>'.`,
	Example: `  # Create a new location
  escape-cli locations create --name "prod-vpc" --ssh-public-key "ssh-ed25519 AAAA..."`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if locationCreateName == "" {
			return errors.New("--name is required")
		}

		id, err := escape.CreateLocation(cmd.Context(), locationCreateName, locationCreateSSHPublicKey)
		if err != nil {
			return fmt.Errorf("failed to create location: %w", err)
		}
		out.Log("Location created: " + id)
		return nil
	},
}

var locationsUpdateCmd = &cobra.Command{
	Use:   "update location-id",
	Short: "Update an existing location",
	Long:  `Update Location - Modify Name or SSH Public Key`,
	Example: `  # Update location name
  escape-cli locations update <location-id> --name "new-name"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var name *string
		var sshPublicKey *string
		var enabled *bool
		if cmd.Flags().Changed("name") {
			name = &locationUpdateName
		}
		if cmd.Flags().Changed("ssh-public-key") {
			sshPublicKey = &locationUpdateSSHPublicKey
		}
		if cmd.Flags().Changed("enabled") {
			enabled = &locationUpdateEnabled
		}
		if name == nil && sshPublicKey == nil && enabled == nil {
			return errors.New("at least one of --name, --ssh-public-key, or --enabled is required")
		}
		err := escape.UpdateLocation(cmd.Context(), args[0], name, sshPublicKey, enabled)
		if err != nil {
			return fmt.Errorf("failed to update location: %w", err)
		}
		out.Log(fmt.Sprintf("Location %s updated", args[0]))
		return nil
	},
}

func init() {
	locationsCmd.AddCommand(locationsListCmd)
	locationsCmd.AddCommand(locationsGetCmd)
	locationsCmd.AddCommand(locationsStartCmd)
	locationsCmd.AddCommand(locationsDeleteCmd)
	locationsCmd.AddCommand(locationsCreateCmd)
	locationsCmd.AddCommand(locationsUpdateCmd)
	rootCmd.AddCommand(locationsCmd)
	locationsListCmd.Flags().StringVarP(&locationsSearch, "search", "s", "", "Search term to filter locations by")
	locationsListCmd.Flags().BoolVarP(&locationsEnabled, "enabled", "e", false, "Filter by enabled locations")
	locationsListCmd.Flags().StringSliceVarP(&locationsLocationTypes, "type", "t", []string{}, "Filter by location type (private, escape, repeater)")
	locationsListCmd.Flags().StringVar(&locationsSortType, "sort-by", "", "sort field")
	locationsListCmd.Flags().StringVar(&locationsSortDirection, "sort-direction", "", "sort direction: asc, desc")
	locationsCreateCmd.Flags().StringVar(&locationCreateName, "name", "", "location name")
	locationsCreateCmd.Flags().StringVar(&locationCreateSSHPublicKey, "ssh-public-key", "", "SSH public key for the location")
	locationsUpdateCmd.Flags().StringVar(&locationUpdateName, "name", "", "new location name")
	locationsUpdateCmd.Flags().StringVar(&locationUpdateSSHPublicKey, "ssh-public-key", "", "new SSH public key")
	locationsUpdateCmd.Flags().BoolVar(&locationUpdateEnabled, "enabled", false, "enable or disable the location")
}
