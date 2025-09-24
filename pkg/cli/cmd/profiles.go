package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var profileKinds = []string{
	"BLST_REST",
	"BLST_GRAPHQL",
	"FRONTEND_DAST",
}

var profileAssetIDs []string
var profileDomains []string
var profileIssueIDs []string
var profileTagsIDs []string
var profileSearch string
var profileInitiators []string
var profileRisks []string

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"profile", "profiles"},
	Short:   "Interact with profiles",
	Long:    "Interact with your escape profiles",
}


var profilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Short:   "List profiles",
	Long: `List REST, WEBAPP and GRAPHQL profiles. (use --all to show all profiles kinds)
	
Example output:
ID                                      CREATED AT              INITIATORS  NAME
00000000-0000-0000-0000-000000000001    2025-02-21T11:15:07Z    [API]       Example-Application-1
00000000-0000-0000-0000-000000000002    2025-03-12T19:19:08Z    [API]       Example-Application-2`,
	Example: `escape-cli profiles list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		allFlag, _ := cmd.Flags().GetBool("all")
		userKindsProvided := cmd.Flags().Changed("kind")
		kindsToUse := profileKinds
		if allFlag && !userKindsProvided {
			kindsToUse = []string{}
		} else if !allFlag && !userKindsProvided {
			kindsToUse = []string{"BLST_REST", "BLST_GRAPHQL", "FRONTEND_DAST"}
		}

		profiles, next, err := escape.ListProfiles(cmd.Context(), "", &escape.ListProfilesFilters{
			AssetIDs: profileAssetIDs,
			Domains: profileDomains,
			IssueIDs: profileIssueIDs,
			TagsIDs: profileTagsIDs,
			Search: profileSearch,
			Initiators: profileInitiators,
			Kinds: kindsToUse,
			Risks: profileRisks,
		})
		if err != nil {
			return fmt.Errorf("unable to list profiles: %w", err)
		}

		out.Table(profiles, func() []string {
			result := []string{"ID\tCREATED AT\tASSET TYPE\tINITIATORS\tNAME"}
			for _, profile := range profiles {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetCreatedAt(), profile.Asset.GetType(), profile.GetInitiators(), profile.GetName()))
			}
			return result
		})

		for next != nil && *next != "" {
			profiles, next, err = escape.ListProfiles(cmd.Context(), *next, &escape.ListProfilesFilters{
				AssetIDs: profileAssetIDs,
				Domains: profileDomains,
				IssueIDs: profileIssueIDs,
				TagsIDs: profileTagsIDs,
				Search: profileSearch,
				Initiators: profileInitiators,
				Kinds: kindsToUse,
				Risks: profileRisks,
			})
			if err != nil {
				return fmt.Errorf("unable to list profiles: %w", err)
			}
			out.Table(profiles, func() []string {
				result := []string{"ID\tCREATED AT\tASSET TYPE\tINITIATORS\tNAME"}
				for _, profile := range profiles {
					result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetCreatedAt(), profile.Asset.GetType(), profile.GetInitiators(), profile.GetName()))
				}
				return result
			})
		}

		return nil
	},
}

var profileGetCmd = &cobra.Command{
	Use:     "get profile-id",
	Aliases: []string{"describe"},
	Short:   "Get a profile",
	Long: `Get a profile by ID.

Example output:
ID                                      CREATED AT              INITIATORS  NAME
00000000-0000-0000-0000-000000000001    2025-02-21T11:15:07Z    [API]       Example-Application-1`,
	Example: `escape-cli profiles get 00000000-0000-0000-0000-000000000001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileID := args[0]
		profile, err := escape.GetProfile(cmd.Context(), profileID)
		if err != nil || profile == nil {
			return fmt.Errorf("unable to get profile %s: %w", profileID, err)
		}

		out.Table(profile, func() []string {
			result := []string{"ID\tCREATED AT\tCRON\tRISKS\tNAME"}
			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetCreatedAt(), profile.GetCron(), profile.GetRisks(),profile.GetName()))
			return result
		})
		return nil
	},
}

var profileCreateRestCmd = &cobra.Command{
	Use:     "create-rest <profile.json",
	Aliases: []string{"cr"},
	Short:   "Create a REST profile",
	Long:    "Create a REST profile",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var data []byte
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		data = b

		var profile map[string]interface{}
		if err := json.Unmarshal(data, &profile); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		response, err := escape.CreateProfileRest(cmd.Context(), data)
		if err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		out.Table(response, func() []string {
			result := []string{"ID\tCREATED AT\tNAME\tASSET TYPE"}
			if profileResponse, ok := response.(*v3.ProfileDetailed); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", profileResponse.GetId(), profileResponse.GetCreatedAt(), profileResponse.GetName(), profileResponse.Asset.GetType()))
			}
			return result
		})
		return nil
	},
}

var profileCreateWebappCmd = &cobra.Command{
	Use:     "create-webapp <profile.json",
	Aliases: []string{"cw"},
	Short:   "Create a WEBAPP profile",
	Long:    "Create a WEBAPP profile",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var data []byte
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		data = b

		var profile map[string]interface{}
		if err := json.Unmarshal(data, &profile); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		response, err := escape.CreateProfileWebapp(cmd.Context(), data)
		if err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		out.Table(response, func() []string {
			result := []string{"ID\tCREATED AT\tNAME\tASSET TYPE"}
			if profileResponse, ok := response.(*v3.ProfileDetailed); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", profileResponse.GetId(), profileResponse.GetCreatedAt(), profileResponse.GetName(), profileResponse.Asset.GetType()))
			}
			return result
		})
		return nil
	},
}
var profileCreateGraphqlCmd = &cobra.Command{
	Use:     "create-graphql <profile.json",
	Aliases: []string{"cg"},
	Short:   "Create a GRAPHQL profile",
	Long:    "Create a GRAPHQL profile",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var data []byte
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		data = b

		var profile map[string]interface{}
		if err := json.Unmarshal(data, &profile); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		response, err := escape.CreateProfileGraphql(cmd.Context(), data)
		if err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		out.Table(response, func() []string {
			result := []string{"ID\tCREATED AT\tNAME\tASSET TYPE"}
			if profileResponse, ok := response.(*v3.ProfileDetailed); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", profileResponse.GetId(), profileResponse.GetCreatedAt(), profileResponse.GetName(), profileResponse.Asset.GetType()))
			}
			return result
		})
		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:     "delete profile-id",
	Aliases: []string{"del"},
	Short:   "Delete a profile",
	Long:    "Delete a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileID := args[0]
		err := escape.DeleteProfile(cmd.Context(), profileID)
		if err != nil {
			return fmt.Errorf("unable to delete profile %s: %w", profileID, err)
		}
		out.Log(fmt.Sprintf("Profile %s successfully deleted", profileID))
		return nil
	},
	Args: cobra.ExactArgs(1),
}
func init() {
	profilesCmd.AddCommand(
		profilesListCmd,
		profileGetCmd,
		profileCreateRestCmd,
		profileCreateWebappCmd,
		profileCreateGraphqlCmd,
		profileDeleteCmd,
	)
	profilesListCmd.Flags().Bool("all", false, "Show profiles for all asset types")
	profilesListCmd.Flags().StringSliceVarP(&profileAssetIDs, "asset-id", "a", []string{}, "asset ID")
	profilesListCmd.Flags().StringSliceVarP(&profileDomains, "domain", "d", []string{}, "domain")
	profilesListCmd.Flags().StringSliceVarP(&profileIssueIDs, "issue-id", "i", []string{}, "issue ID")
	profilesListCmd.Flags().StringSliceVarP(&profileTagsIDs, "tag-id", "t", []string{}, "tag ID")
	profilesListCmd.Flags().StringVarP(&profileSearch, "search", "s", "", "search")
	profilesListCmd.Flags().StringSliceVarP(&profileInitiators, "initiator", "n", []string{}, "initiator")
	profilesListCmd.Flags().StringSliceVarP(&profileKinds, "kind", "k", []string{}, "kind")
	profilesListCmd.Flags().StringSliceVarP(&profileRisks, "risk", "r", []string{}, "risk")
	rootCmd.AddCommand(profilesCmd)
}
