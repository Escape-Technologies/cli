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
	Aliases: []string{"profile"},
	Short:   "Manage security testing profiles and configurations",
	Long: `Manage Security Profiles - Configure API Testing

Profiles define HOW your APIs are tested. Each profile configures test settings,
authentication, and security checks for a specific asset. One asset can have
multiple profiles for different testing scenarios.

PROFILE TYPES:
  • BLST_REST         - REST API security testing
  • BLST_GRAPHQL      - GraphQL API security testing
  • FRONTEND_DAST     - Web application security testing

KEY FEATURES:
  • Authentication configuration (API keys, OAuth, JWT, etc.)
  • Test scope and endpoint selection
  • Custom security rules and checks
  • Scheduled scanning (cron expressions)
  • CI/CD integration settings

COMMON WORKFLOWS:
  • List all profiles:       escape-cli profiles list
  • Get profile details:     escape-cli profiles get <profile-id>
  • Create REST profile:     escape-cli profiles create-rest < config.json
  • Start scan on profile:   escape-cli scans start <profile-id>`,
}

var profilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Short:   "List security testing profiles",
	Long: `List Security Profiles - View Test Configurations

List all security testing profiles in your organization. By default, shows REST,
GraphQL, and WEBAPP profiles. Use --all to include all profile types.

FILTER OPTIONS:
  -a, --asset-id     Filter by asset ID
  -d, --domain       Filter by domain
  -i, --issue-id     Filter by issues found
  -t, --tag-id       Filter by tags
  -s, --search       Free-text search
  -k, --kind         Filter by profile type
  -r, --risk         Filter by risk level
  --all              Show all profile types (default: REST, GraphQL, WEBAPP only)

Example output:
ID                                      CREATED AT              ASSET TYPE    INITIATORS  NAME
00000000-0000-0000-0000-000000000001    2025-02-21T11:15:07Z    WEBAPP        [API]       Example-App-1
00000000-0000-0000-0000-000000000002    2025-03-12T19:19:08Z    REST_API      [CI]        Example-API-2`,
	Example: `  # List all standard profiles
  escape-cli profiles list

  # List all profile types
  escape-cli profiles list --all

  # List profiles for a specific asset
  escape-cli profiles list --asset-id <asset-id>

  # Search for profiles
  escape-cli profiles list --search "production"`,
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
			AssetIDs:   profileAssetIDs,
			Domains:    profileDomains,
			IssueIDs:   profileIssueIDs,
			TagsIDs:    profileTagsIDs,
			Search:     profileSearch,
			Initiators: profileInitiators,
			Kinds:      kindsToUse,
			Risks:      profileRisks,
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
				AssetIDs:   profileAssetIDs,
				Domains:    profileDomains,
				IssueIDs:   profileIssueIDs,
				TagsIDs:    profileTagsIDs,
				Search:     profileSearch,
				Initiators: profileInitiators,
				Kinds:      kindsToUse,
				Risks:      profileRisks,
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
	Aliases: []string{"describe", "show"},
	Short:   "Get detailed profile information",
	Long: `Get Profile Details - View Complete Configuration

Retrieve comprehensive information about a security testing profile including
schedule, risks, and configuration details.`,
	Example: `  # Get profile details
  escape-cli profiles get 00000000-0000-0000-0000-000000000001

  # Export profile configuration
  escape-cli profiles get <profile-id> -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileID := args[0]
		profile, err := escape.GetProfile(cmd.Context(), profileID)
		if err != nil || profile == nil {
			return fmt.Errorf("unable to get profile %s: %w", profileID, err)
		}

		out.Table(profile, func() []string {
			result := []string{"ID\tCREATED AT\tCRON\tRISKS\tNAME"}
			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetCreatedAt(), profile.GetCron(), profile.GetRisks(), profile.GetName()))
			return result
		})
		return nil
	},
}

var profileCreateRestCmd = &cobra.Command{
	Use:     "create-rest",
	Aliases: []string{"cr"},
	Short:   "Create a REST API security testing profile",
	Long: `Create REST Profile - Configure REST API Security Testing

Create a new profile for testing REST APIs. Provide configuration via JSON through stdin.
See https://public.escape.tech/v3/#tag/profiles for complete schema.`,
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
	Use:     "create-webapp",
	Aliases: []string{"cw"},
	Short:   "Create a web application security testing profile",
	Long: `Create WEBAPP Profile - Configure Web Application Security Testing

Create a new profile for testing web applications. Provide configuration via JSON through stdin.`,
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
	Use:     "create-graphql",
	Aliases: []string{"cg"},
	Short:   "Create a GraphQL API security testing profile",
	Long: `Create GraphQL Profile - Configure GraphQL API Security Testing

Create a new profile for testing GraphQL APIs. Provide configuration via JSON through stdin.`,
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
	Aliases: []string{"del", "rm"},
	Short:   "Delete a security testing profile",
	Long: `Delete Profile - Remove Test Configuration

Permanently delete a security testing profile. This will also remove scan history
and scheduled scans. The asset itself is NOT deleted.`,
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
