package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"profile", "profiles"},
	Short:   "Interact with profiles",
	Long:    "Interact with your escape profiles",
}

var profilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List profiles",
	Long: `List all profiles.

Example output:
ID                                      TYPE         NAME                                                           CREATED AT              HAS CI    CRON
00000000-0000-0000-0000-000000000001    REST         Example-Application-1                                          2025-02-21T11:15:07Z    false     0 11 * * 5
00000000-0000-0000-0000-000000000002    REST         Example-Application-2                                          2025-03-12T19:19:08Z    false     0 19 * * 3`,
	Example: `escape-cli profiles list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		profiles, next, err := escape.ListProfiles(cmd.Context(), "")
		if err != nil {
			return fmt.Errorf("unable to list profiles: %w", err)
		}

		// First result
		result := []string{"ID\tNAME\tCREATED AT\tINITIATORS"}
		for _, profile := range profiles {
			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetName(), profile.GetCreatedAt(), profile.GetInitiators(), profile.GetLinks().ProfileSummary))
		}

		for next != nil && *next != "" {
			profiles, next, err = escape.ListProfiles(cmd.Context(), *next)
			if err != nil {
				return fmt.Errorf("unable to list profiles: %w", err)
			}
			for _, profile := range profiles {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetName(), profile.GetCreatedAt(), profile.GetInitiators(), profile.GetLinks().ProfileSummary))
			}
		}

		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}

var profileGetCmd = &cobra.Command{
	Use:     "get profile-id",
	Aliases: []string{"describe"},
	Short:   "Get a profile",
	Long: `Get a profile by ID.

Example output:
ID                                      TYPE         NAME                                                           CREATED AT              HAS CI    CRON
00000000-0000-0000-0000-000000000001    REST         Example-Application-1                                          2025-02-21T11:15:07Z    false     0 11 * * 5`,
	Example: `escape-cli profiles get 00000000-0000-0000-0000-000000000001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profileID := args[0]
		profile, err := escape.GetProfile(cmd.Context(), profileID)
		if err != nil || profile == nil {
			return fmt.Errorf("unable to get profile %s: %w", profileID, err)
		}

		// TODO : better display
		result := []string{"ID\tNAME\tCREATED AT\tUPDATED AT\tLINK"}
		result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t", profile.GetId(), profile.GetName(), profile.GetCreatedAt(), profile.GetUpdatedAt(), profile.GetLinks().ProfileSummary))

		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}

func init() {
	profilesCmd.AddCommand(
		profilesListCmd,
		profileGetCmd,
	)
	rootCmd.AddCommand(profilesCmd)
}
