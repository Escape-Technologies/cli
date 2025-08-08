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
	Long:    `List all profiles.

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
		out.Table(profiles, func() []string {
			result := []string{"ID\tNAME\tCREATED AT\tINITIATORS"}
			for _, profile := range profiles {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetName(), profile.GetCreatedAt(), profile.GetInitiators(), profile.GetLinks().ProfileSummary))
			}
			return result
		})
		for next != "" {
			profiles, next, err = escape.ListProfiles(cmd.Context(), next)
			if err != nil {
				return fmt.Errorf("unable to list profiles: %w", err)
			}
			out.Table(profiles, func() []string {
				result := []string{}
				for _, profile := range profiles {
					result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetName(), profile.GetCreatedAt(), profile.GetInitiators(), profile.GetLinks().ProfileSummary))
				}
				return result
			})
		}
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesListCmd)
	rootCmd.AddCommand(profilesCmd)
}
