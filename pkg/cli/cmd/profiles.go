package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

// The generated API client reuses CreateDastRestProfileRequest for every
// profile-creation endpoint, so these aliases keep each command's input schema
// semantically aligned with the profile it creates.
type (
	createRestProfileInput           = v3.CreateDastRestProfileRequest
	createWebappProfileInput         = v3.CreateDastRestProfileRequest
	createGraphqlProfileInput        = v3.CreateDastRestProfileRequest
	createPentestRestProfileInput    = v3.CreateDastRestProfileRequest
	createPentestGraphqlProfileInput = v3.CreateDastRestProfileRequest
	createPentestWebappProfileInput  = v3.CreateDastRestProfileRequest
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
var profileSortType string
var profileSortDirection string

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
		// Output JSON Schema if requested
		if out.Schema([]v3.ProfileSummarized{}) {
			return nil
		}

		allFlag, _ := cmd.Flags().GetBool("all")
		userKindsProvided := cmd.Flags().Changed("kind")
		kindsToUse := profileKinds
		if allFlag && !userKindsProvided {
			kindsToUse = []string{}
		} else if !allFlag && !userKindsProvided {
			kindsToUse = []string{"BLST_REST", "BLST_GRAPHQL", "FRONTEND_DAST"}
		}

		filters := &escape.ListProfilesFilters{
			AssetIDs:      profileAssetIDs,
			Domains:       profileDomains,
			IssueIDs:      profileIssueIDs,
			TagsIDs:       profileTagsIDs,
			Search:        profileSearch,
			Initiators:    profileInitiators,
			Kinds:         kindsToUse,
			Risks:         profileRisks,
			SortType:      profileSortType,
			SortDirection: profileSortDirection,
		}
		profiles, next, err := escape.ListProfiles(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("unable to list profiles: %w", err)
		}
		allProfiles := profiles
		for next != nil && *next != "" {
			profiles, next, err = escape.ListProfiles(cmd.Context(), *next, filters)
			if err != nil {
				return fmt.Errorf("unable to list profiles: %w", err)
			}
			allProfiles = append(allProfiles, profiles...)
		}
		out.Table(allProfiles, func() []string {
			result := []string{"ID\tCREATED AT\tASSET TYPE\tINITIATORS\tSCORE\tCOVERAGE\tOPEN ISSUES\tLAST SCAN STATUS\tNAME"}
			for _, profile := range allProfiles {
				score := ""
				if value, ok := profile.GetScoreOk(); ok {
					score = fmt.Sprintf("%.2f", *value)
				}
				coverage := ""
				if value, ok := profile.GetCoverageOk(); ok {
					coverage = fmt.Sprintf("%.2f", *value)
				}
				openIssueCount := ""
				if value, ok := profile.GetOpenIssueCountOk(); ok {
					openIssueCount = fmt.Sprintf("%d", *value)
				}
				lastScanStatus := ""
				if value, ok := profile.GetLastScanStatusOk(); ok {
					lastScanStatus = *value
				}
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", profile.GetId(), profile.GetCreatedAt(), profile.Asset.GetType(), profile.GetInitiators(), score, coverage, openIssueCount, lastScanStatus, profile.GetName()))
			}
			return result
		})

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
		// Output JSON Schema if requested
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

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
		// Output JSON Schema for input format if requested
		if out.InputSchema(createRestProfileInput{}) {
			return nil
		}
		// Output JSON Schema if requested
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

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
			if profileResponse, ok := response.(*v3.GetProfile200Response); ok {
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
		// Output JSON Schema for input format if requested
		if out.InputSchema(createWebappProfileInput{}) {
			return nil
		}
		// Output JSON Schema if requested
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

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
			if profileResponse, ok := response.(*v3.GetProfile200Response); ok {
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
		// Output JSON Schema for input format if requested
		if out.InputSchema(createGraphqlProfileInput{}) {
			return nil
		}
		// Output JSON Schema if requested
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

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
			if profileResponse, ok := response.(*v3.GetProfile200Response); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", profileResponse.GetId(), profileResponse.GetCreatedAt(), profileResponse.GetName(), profileResponse.Asset.GetType()))
			}
			return result
		})
		return nil
	},
}

var profileCreatePentestRestCmd = &cobra.Command{
	Use:     "create-pentest-rest",
	Aliases: []string{"cpr"},
	Short:   "Create an AI Pentest REST API profile",
	Long: `Create AI Pentest REST Profile - Configure Automated Penetration Testing

Create a new AI-powered penetration testing profile for REST APIs.
Provide configuration via JSON through stdin.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(createPentestRestProfileInput{}) {
			return nil
		}
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		response, err := escape.CreateProfilePentestRest(cmd.Context(), b)
		if err != nil {
			return fmt.Errorf("failed to create pentest REST profile: %w", err)
		}

		out.Table(response, func() []string {
			result := []string{"ID\tCREATED AT\tNAME\tASSET TYPE"}
			if profileResponse, ok := response.(*v3.GetProfile200Response); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", profileResponse.GetId(), profileResponse.GetCreatedAt(), profileResponse.GetName(), profileResponse.Asset.GetType()))
			}
			return result
		})
		return nil
	},
}

var profileCreatePentestGraphqlCmd = &cobra.Command{
	Use:     "create-pentest-graphql",
	Aliases: []string{"cpg"},
	Short:   "Create an AI Pentest GraphQL profile",
	Long: `Create AI Pentest GraphQL Profile - Configure Automated Penetration Testing

Create a new AI-powered penetration testing profile for GraphQL APIs.
Provide configuration via JSON through stdin.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(createPentestGraphqlProfileInput{}) {
			return nil
		}
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		response, err := escape.CreateProfilePentestGraphql(cmd.Context(), b)
		if err != nil {
			return fmt.Errorf("failed to create pentest GraphQL profile: %w", err)
		}

		out.Table(response, func() []string {
			result := []string{"ID\tCREATED AT\tNAME\tASSET TYPE"}
			if profileResponse, ok := response.(*v3.GetProfile200Response); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", profileResponse.GetId(), profileResponse.GetCreatedAt(), profileResponse.GetName(), profileResponse.Asset.GetType()))
			}
			return result
		})
		return nil
	},
}

var profileCreatePentestWebappCmd = &cobra.Command{
	Use:     "create-pentest-webapp",
	Aliases: []string{"cpw"},
	Short:   "Create an AI Pentest WebApp profile",
	Long: `Create AI Pentest WebApp Profile - Configure Automated Penetration Testing

Create a new AI-powered penetration testing profile for web applications.
Provide configuration via JSON through stdin.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(createPentestWebappProfileInput{}) {
			return nil
		}
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		response, err := escape.CreateProfilePentestWebapp(cmd.Context(), b)
		if err != nil {
			return fmt.Errorf("failed to create pentest WebApp profile: %w", err)
		}

		out.Table(response, func() []string {
			result := []string{"ID\tCREATED AT\tNAME\tASSET TYPE"}
			if profileResponse, ok := response.(*v3.GetProfile200Response); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", profileResponse.GetId(), profileResponse.GetCreatedAt(), profileResponse.GetName(), profileResponse.Asset.GetType()))
			}
			return result
		})
		return nil
	},
}

var profileProblemsCmd = &cobra.Command{
	Use:     "problems",
	Aliases: []string{"pb"},
	Short:   "List profiles with scan problems",
	Long: `List Profiles with Scan Problems - Identify Failing Configurations

Display profiles whose latest scan encountered errors or failures. Useful for
identifying configuration issues, broken authentication, or unreachable targets.`,
	Example: `  # List all profiles with problems
  escape-cli profiles problems

  # Filter by asset
  escape-cli profiles problems --asset-id <asset-id>

  # Export to JSON
  escape-cli profiles problems -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.ProfileScanProblemsRow{}) {
			return nil
		}

		filters := &escape.ListProblemsFilters{
			AssetIDs:   profileAssetIDs,
			Domains:    profileDomains,
			IssueIDs:   profileIssueIDs,
			TagsIDs:    profileTagsIDs,
			Search:     profileSearch,
			Initiators: profileInitiators,
			Kinds:      profileKinds,
			Risks:      profileRisks,
		}
		problems, next, err := escape.ListProblems(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("unable to list problems: %w", err)
		}
		all := problems
		for next != nil && *next != "" {
			problems, next, err = escape.ListProblems(cmd.Context(), *next, filters)
			if err != nil {
				return fmt.Errorf("unable to list problems: %w", err)
			}
			all = append(all, problems...)
		}
		out.Table(all, func() []string {
			res := []string{"PROFILE ID\tNAME\tLAST SCAN ID\tSTATUS\tSCORE\tCREATED AT"}
			for _, row := range all {
				scanID, status, score, createdAt := "-", "-", "-", "-"
				if row.LastScan != nil {
					scanID = row.LastScan.GetId()
					status = row.LastScan.GetStatus()
					if s := row.LastScan.GetScore(); s != 0 {
						score = fmt.Sprintf("%.0f", s)
					}
					createdAt = row.LastScan.GetCreatedAt()
				}
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", row.GetId(), row.GetName(), scanID, status, score, createdAt))
			}
			return res
		})
		return nil
	},
}

var profileUpdateCmd = &cobra.Command{
	Use:     "update profile-id",
	Aliases: []string{"u", "edit"},
	Short:   "Update profile metadata",
	Long: `Update Profile - Modify Name, Description, Cron Schedule

Update a profile's metadata fields. Provide updates via JSON through stdin
or use flags for individual fields.

UPDATABLE FIELDS:
  --name          Profile name
  --description   Profile description
  --cron          Cron schedule (e.g., "0 22 * * *")

Alternatively, provide a JSON object via stdin with any combination of fields.`,
	Example: `  # Update name via flag
  escape-cli profiles update <profile-id> --name "New Profile Name"

  # Update cron schedule
  escape-cli profiles update <profile-id> --cron "0 6 * * 1"

  # Update via JSON stdin
  echo '{"name": "Updated Name", "cron": "0 22 * * *"}' | escape-cli profiles update <profile-id>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}
		if out.InputSchema(v3.UpdateProfileRequest{}) {
			return nil
		}

		profileID := args[0]

		var payload map[string]interface{}

		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			if len(b) > 0 {
				if err := json.Unmarshal(b, &payload); err != nil {
					return fmt.Errorf("invalid JSON: %w", err)
				}
			}
		}

		if payload == nil {
			payload = make(map[string]interface{})
		}

		// Flags override stdin values
		if cmd.Flags().Changed("name") {
			payload["name"] = profileUpdateName
		}
		if cmd.Flags().Changed("description") {
			payload["description"] = profileUpdateDescription
		}
		if cmd.Flags().Changed("cron") {
			payload["cron"] = profileUpdateCron
		}

		if len(payload) == 0 {
			return errors.New("no updates provided: use flags or pipe JSON via stdin")
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		profile, err := escape.UpdateProfile(cmd.Context(), profileID, data)
		if err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}

		out.Table(profile, func() []string {
			return []string{
				"ID\tNAME\tCREATED AT\tASSET TYPE",
				fmt.Sprintf("%s\t%s\t%s\t%s", profile.GetId(), profile.GetName(), profile.GetCreatedAt(), profile.Asset.GetType()),
			}
		})
		return nil
	},
}

var (
	profileUpdateName        string
	profileUpdateDescription string
	profileUpdateCron        string
)

var profileUpdateConfigurationCmd = &cobra.Command{
	Use:     "update-configuration profile-id",
	Aliases: []string{"uc", "update-config"},
	Short:   "Update profile configuration",
	Long: `Update Profile Configuration - Modify Auth, Scope, and Scanner Settings

Update a profile's scan configuration via JSON through stdin. The JSON must
contain a "configuration" object with the fields to update.

IMPORTANT: This is a full replace, not a merge. Any configuration section not
included in the JSON will be reset to defaults. Always send the complete
configuration. Use "profiles get <id> -o json" to retrieve the current
configuration before updating.

CONFIGURABLE SECTIONS:
  authentication    - Users, credentials, browser login procedures
  frontend_dast     - Crawling mode, agentic instructions, hotstart URLs, scope
  scope             - Domain allowlist/blocklist, URL filtering
  security_tests    - Enable/disable specific security checks
  network           - Proxy and network settings`,
	Example: `  # Update agentic crawling instructions
  cat <<'EOF' | escape-cli profiles update-configuration <profile-id>
  {
    "configuration": {
      "frontend_dast": {
        "agentic_crawling": {
          "enabled": true,
          "instructions": "Navigate to Accounts, Policies, Claims, Billing sections."
        }
      }
    }
  }
  EOF

  # Update authentication
  cat auth.json | escape-cli profiles update-configuration <profile-id>

  # Update hotstart URLs
  echo '{"configuration":{"frontend_dast":{"hotstart":["https://app.example.com/#/accounts"]}}}' | escape-cli profiles update-configuration <profile-id>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.UpdateProfileConfigurationRequest{}) {
			return nil
		}

		profileID := args[0]

		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		if len(b) == 0 {
			return errors.New("no input provided: pipe JSON configuration via stdin")
		}

		var tmp map[string]interface{}
		if err := json.Unmarshal(b, &tmp); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		result, err := escape.UpdateProfileConfiguration(cmd.Context(), profileID, b)
		if err != nil {
			return fmt.Errorf("failed to update configuration: %w", err)
		}

		out.Print(result, "Configuration updated successfully")
		return nil
	},
}

var profileUpdateSchemaCmd = &cobra.Command{
	Use:     "update-schema profile-id schema-id",
	Aliases: []string{"us"},
	Short:   "Update the schema attached to a profile",
	Long: `Update Profile Schema - Replace the API Schema

Replace the API schema (OpenAPI, Postman, GraphQL) attached to a profile.
First upload the schema file using "upload schema", then pass the returned
upload ID as the schema-id argument.

WORKFLOW:
  1. Upload the schema:    SCHEMA_ID=$(escape-cli upload schema < spec.json -o json | jq -r '.id')
  2. Attach to profile:    escape-cli profiles update-schema <profile-id> $SCHEMA_ID`,
	Example: `  # Upload and attach in one pipeline
  SCHEMA_ID=$(escape-cli upload schema < openapi.json -o json | jq -r '.id')
  escape-cli profiles update-schema <profile-id> $SCHEMA_ID

  # Or step by step
  escape-cli upload schema < openapi.json -o json
  # Copy the returned ID
  escape-cli profiles update-schema <profile-id> <schema-id>`,
	Args: cobra.ExactArgs(2), //nolint:mnd
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

		profileID := args[0]
		schemaID := args[1]

		profile, err := escape.UpdateProfileSchema(cmd.Context(), profileID, schemaID)
		if err != nil {
			return fmt.Errorf("failed to update schema: %w", err)
		}

		out.Table(profile, func() []string {
			return []string{
				"ID\tNAME\tCREATED AT\tASSET TYPE",
				fmt.Sprintf("%s\t%s\t%s\t%s", profile.GetId(), profile.GetName(), profile.GetCreatedAt(), profile.Asset.GetType()),
			}
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
		profileProblemsCmd,
		profileCreateRestCmd,
		profileCreateWebappCmd,
		profileCreateGraphqlCmd,
		profileCreatePentestRestCmd,
		profileCreatePentestGraphqlCmd,
		profileCreatePentestWebappCmd,
		profileUpdateCmd,
		profileUpdateConfigurationCmd,
		profileUpdateSchemaCmd,
		profileDeleteCmd,
	)
	profileUpdateCmd.Flags().StringVar(&profileUpdateName, "name", "", "profile name")
	profileUpdateCmd.Flags().StringVar(&profileUpdateDescription, "description", "", "profile description")
	profileUpdateCmd.Flags().StringVar(&profileUpdateCron, "cron", "", "cron schedule (e.g., \"0 22 * * *\")")
	profilesListCmd.Flags().Bool("all", false, "Show profiles for all asset types")
	profilesListCmd.Flags().StringSliceVarP(&profileAssetIDs, "asset-id", "a", []string{}, "asset ID")
	profilesListCmd.Flags().StringSliceVarP(&profileDomains, "domain", "d", []string{}, "domain")
	profilesListCmd.Flags().StringSliceVarP(&profileIssueIDs, "issue-id", "i", []string{}, "issue ID")
	profilesListCmd.Flags().StringSliceVarP(&profileTagsIDs, "tag-id", "t", []string{}, "tag ID")
	profilesListCmd.Flags().StringVarP(&profileSearch, "search", "s", "", "search")
	profilesListCmd.Flags().StringSliceVarP(&profileInitiators, "initiator", "n", []string{}, "initiator")
	profilesListCmd.Flags().StringSliceVarP(&profileKinds, "kind", "k", []string{}, "kind")
	profilesListCmd.Flags().StringSliceVarP(&profileRisks, "risk", "r", []string{}, "risk")
	profilesListCmd.Flags().StringVar(&profileSortType, "sort-by", "", "sort field")
	profilesListCmd.Flags().StringVar(&profileSortDirection, "sort-direction", "", "sort direction: asc, desc")
	rootCmd.AddCommand(profilesCmd)
}
