package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
var profileGetExtraAssets bool

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"profile"},
	Short:   "Manage security testing profiles and configurations",
	Long: `Manage Security Profiles - Configure API Testing

Profiles define HOW your APIs and web applications are tested. Each profile
configures authentication, scope, and security checks for a specific asset.

DAST PROFILES:
  create-rest           REST API security testing
  create-graphql        GraphQL API security testing
  create-webapp         Web application security testing

AI PENTEST PROFILES:
  create-pentest-rest       AI-driven penetration testing for REST APIs
  create-pentest-graphql    AI-driven penetration testing for GraphQL APIs
  create-pentest-webapp     AI-driven penetration testing for web applications`,
}

var profilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Short:   "List security testing profiles",
	Long: `List Security Profiles - View Test Configurations

List all security testing profiles in your organization. By default shows DAST
profiles only (REST, GraphQL, WEBAPP). Use --all to include pentest profiles.`,
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
					openIssueCount = strconv.Itoa(*value)
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
  escape-cli profiles get <profile-id> -o json

  # List the extra assets attached to a profile (detailed table)
  escape-cli profiles get <profile-id> --extra-assets

  # List the extra assets attached to a profile as JSON
  escape-cli profiles get <profile-id> -o json | jq '.extraAssets'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileGetExtraAssets {
			if out.Schema([]v3.ProfileExtraAsset{}) {
				return nil
			}
		} else if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

		profileID := args[0]
		profile, err := escape.GetProfile(cmd.Context(), profileID)
		if err != nil || profile == nil {
			return fmt.Errorf("unable to get profile %s: %w", profileID, err)
		}

		if profileGetExtraAssets {
			extraAssets := profile.GetExtraAssets()
			out.Table(extraAssets, func() []string {
				rows := []string{"ID\tCLASS\tTYPE\tSTATUS\tCREATED AT\tNAME"}
				for _, asset := range extraAssets {
					rows = append(rows, fmt.Sprintf(
						"%s\t%s\t%s\t%s\t%s\t%s",
						asset.GetId(),
						asset.GetClass(),
						asset.GetType(),
						asset.GetStatus(),
						asset.GetCreatedAt(),
						asset.GetName(),
					))
				}
				return rows
			})
			return nil
		}

		out.Table(profile, func() []string {
			result := []string{"ID\tCREATED AT\tCRON\tRISKS\tNAME\tEXTRA ASSETS"}
			result = append(result, fmt.Sprintf(
				"%s\t%s\t%s\t%s\t%s\t%s",
				profile.GetId(),
				profile.GetCreatedAt(),
				profile.GetCron(),
				profile.GetRisks(),
				profile.GetName(),
				formatExtraAssets(profile.GetExtraAssets()),
			))
			return result
		})
		return nil
	},
}

func formatExtraAssets(extraAssets []v3.ProfileExtraAsset) string {
	if len(extraAssets) == 0 {
		return ""
	}
	ids := make([]string, 0, len(extraAssets))
	for _, asset := range extraAssets {
		ids = append(ids, asset.GetId())
	}
	return strings.Join(ids, ", ")
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
  --name                 Profile name
  --description          Profile description
  --cron                 Cron schedule (e.g., "0 22 * * *")
  --extra-asset-id       Extra asset IDs to attach, as a comma-separated list.
                         Replaces the full list of extra assets on the profile.
  --clear-extra-assets   Detach every extra asset from the profile.

Alternatively, provide a JSON object via stdin with any combination of fields.`,
	Example: `  # Update name via flag
  escape-cli profiles update <profile-id> --name "New Profile Name"

  # Update cron schedule
  escape-cli profiles update <profile-id> --cron "0 6 * * 1"

  # Attach extra assets (e.g. schemas) to a profile
  escape-cli profiles update <profile-id> \
    --extra-asset-id 11111111-1111-1111-1111-111111111111,22222222-2222-2222-2222-222222222222

  # Detach every extra asset from a profile
  escape-cli profiles update <profile-id> --clear-extra-assets

  # Update via JSON stdin
  echo '{"name": "Updated Name", "cron": "0 22 * * *"}' | escape-cli profiles update <profile-id>

  # Update extra assets via JSON stdin
  echo '{"extraAssetIds": ["11111111-1111-1111-1111-111111111111"]}' | escape-cli profiles update <profile-id>`,
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
		if cmd.Flags().Changed("clear-extra-assets") && profileUpdateClearExtraAssets {
			payload["extraAssetIds"] = []string{}
		} else if cmd.Flags().Changed("extra-asset-id") {
			payload["extraAssetIds"] = parseExtraAssetIDs(profileUpdateExtraAssetIDs)
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
	profileUpdateName             string
	profileUpdateDescription      string
	profileUpdateCron             string
	profileUpdateExtraAssetIDs    string
	profileUpdateClearExtraAssets bool
)

// parseExtraAssetIDs splits the raw --extra-asset-id value on commas and
// trims whitespace. Empty entries are dropped so the payload contains only
// real asset IDs.
func parseExtraAssetIDs(raw string) []string {
	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, id := range parts {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

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

const (
	// profileGetSchemaDefaultTimeout bounds the whole `profiles get-schema`
	// RunE (GetProfile API call + S3 body fetch when -f is used). 10m is
	// generous for large schemas while staying well under a reasonable
	// proxy-cut ceiling; the generated v3 client itself uses
	// http.DefaultClient with no timeout so this is the only bound.
	profileGetSchemaDefaultTimeout = 10 * time.Minute

	// profileUploadSchemaDefaultTimeout wraps the end-to-end upload +
	// schema-build workflow + attach pipeline. The server-side schema-build
	// workflow caps at 10m, so we add ~5m slack for the S3 PUT, workflow
	// polling, and the final profile PUT.
	profileUploadSchemaDefaultTimeout = 15 * time.Minute
)

var (
	profileGetSchemaID      string
	profileGetSchemaOutFile string
	profileGetSchemaTimeout time.Duration

	profileUploadSchemaFile    string
	profileUploadSchemaName    string
	profileUploadSchemaTimeout time.Duration
)

// pickProfileSchema picks a SCHEMA-class entry from the profile's extraAssets.
// With schemaID="" it returns the (single) active schema; otherwise it returns
// the entry whose id matches schemaID. Mirrors the server-side
// pickProfileSchemaAssetId on the already-enriched client payload.
//
// Error semantics (all non-zero exit, no silent fallbacks):
//   - schemaID == "" and no SCHEMA entry is active → "no active schema".
//   - schemaID != "" and no match → "no schema asset with id X".
//   - schemaID == "" and more than one entry is marked isActive → fail fast
//     instead of picking arbitrarily (this should never happen — server
//     invariant is exactly one active SCHEMA — but we assert it rather than
//     paper over a bug).
func pickProfileSchema(profile *v3.GetProfile200Response, schemaID string) (*v3.ProfileExtraAsset, error) {
	if profile == nil {
		return nil, errors.New("profile is nil")
	}

	if schemaID != "" {
		for i, asset := range profile.ExtraAssets {
			if string(asset.Class) == "SCHEMA" && asset.Id == schemaID {
				return &profile.ExtraAssets[i], nil
			}
		}
		return nil, fmt.Errorf("no schema asset with id %s attached to profile", schemaID)
	}

	activeIdx := -1
	activeCount := 0
	for i, asset := range profile.ExtraAssets {
		if string(asset.Class) == "SCHEMA" && asset.IsActive {
			activeIdx = i
			activeCount++
		}
	}
	switch activeCount {
	case 0:
		return nil, errors.New("no active schema attached to this profile")
	case 1:
		return &profile.ExtraAssets[activeIdx], nil
	default:
		return nil, fmt.Errorf(
			"profile has %d active schema entries; refusing to pick arbitrarily (pass --schema-id to disambiguate)",
			activeCount,
		)
	}
}

var profileGetSchemaCmd = &cobra.Command{
	Use:     "get-schema profile-id",
	Aliases: []string{"gs"},
	Short:   "Get the schema attached to a profile",
	Long: `Get Profile Schema - Inspect or Download a Schema Asset

Returns metadata (including a fresh signed URL) for the active schema attached
to the profile. Use --schema-id to pick a specific schema from extraAssets.
Use -f to stream the raw bytes of the schema instead of metadata.

The signed URL is time-limited; fetch the bytes promptly or use -f to
download in one shot.`,
	Example: `  # Print JSON metadata + signed URL for the active schema
  escape-cli profiles get-schema <profile-id>

  # Download the active schema to a file
  escape-cli profiles get-schema <profile-id> -f openapi.json

  # Stream to stdout (pipe into jq, etc.)
  escape-cli profiles get-schema <profile-id> -f -

  # Pick a specific (non-active) schema attached to the profile
  escape-cli profiles get-schema <profile-id> --schema-id <schema-id> -f spec.json

  # Override the end-to-end timeout (default 10m; bounds both API call and body fetch)
  escape-cli profiles get-schema <profile-id> -f big.json --timeout 30m`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.ProfileExtraAsset{}) {
			return nil
		}

		profileID := args[0]

		// A single timeout bounds the entire command: the GetProfile API
		// call (the generated v3 client uses http.DefaultClient with no
		// timeout and can hang indefinitely behind a proxy) AND the
		// subsequent S3 body fetch when -f is used.
		ctx, cancel := context.WithTimeout(cmd.Context(), profileGetSchemaTimeout)
		defer cancel()

		profile, err := escape.GetProfile(ctx, profileID)
		if err != nil || profile == nil {
			return fmt.Errorf("unable to get profile %s: %w", profileID, err)
		}

		schema, err := pickProfileSchema(profile, profileGetSchemaID)
		if err != nil {
			return err
		}

		if profileGetSchemaOutFile == "" {
			out.Table(schema, func() []string {
				url := ""
				if schema.SignedUrl != nil {
					url = *schema.SignedUrl
				}
				active := "false"
				if schema.IsActive {
					active = "true"
				}
				return []string{
					"ID\tACTIVE\tNAME\tCREATED AT\tSIGNED URL",
					fmt.Sprintf("%s\t%s\t%s\t%s\t%s", schema.Id, active, schema.Name, schema.CreatedAt, url),
				}
			})
			return nil
		}

		if schema.SignedUrl == nil || *schema.SignedUrl == "" {
			return errors.New("schema asset has no signed URL; cannot download")
		}

		if profileGetSchemaOutFile == "-" {
			if err := escape.DownloadSignedURL(ctx, *schema.SignedUrl, os.Stdout); err != nil {
				return fmt.Errorf("failed to download schema bytes: %w", err)
			}
			return nil
		}

		// Write to a sibling temp file and rename on success so a failed
		// download (expired signed URL, mid-stream connection drop, context
		// deadline) never leaves a truncated or empty file at the target
		// path. The temp file lives in the same directory as the destination
		// so the rename stays on one filesystem and is atomic on POSIX.
		destDir, destName := filepath.Split(profileGetSchemaOutFile)
		if destDir == "" {
			destDir = "."
		}
		tmp, err := os.CreateTemp(destDir, destName+".*.part")
		if err != nil {
			return fmt.Errorf("unable to create temp file in %s: %w", destDir, err)
		}
		tmpPath := tmp.Name()
		cleanup := func() {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}

		if err := escape.DownloadSignedURL(ctx, *schema.SignedUrl, tmp); err != nil {
			cleanup()
			return fmt.Errorf("failed to download schema bytes: %w", err)
		}
		if err := tmp.Close(); err != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to finalize temp file %s: %w", tmpPath, err)
		}
		if err := os.Rename(tmpPath, profileGetSchemaOutFile); err != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to move %s to %s: %w", tmpPath, profileGetSchemaOutFile, err)
		}

		out.Log(fmt.Sprintf("Schema %s written to %s", schema.Id, profileGetSchemaOutFile))
		return nil
	},
}

var profileUploadSchemaCmd = &cobra.Command{
	Use:     "upload-schema profile-id",
	Aliases: []string{"ups"},
	Short:   "Upload a schema file and attach it to a profile",
	Long: `Upload Schema - One-Shot Upload + Attach

Read schema bytes from --file (or stdin), upload them to Escape, wait for the
schema-build workflow to finish, and attach the resulting asset to the
profile. This collapses the three-step upload/create-asset/attach flow into
a single command.

The schema-build workflow can take up to 10 minutes for very large schemas.
Tune --timeout if the default of 15m is not enough (or lower it to fail
fast in CI).`,
	Example: `  # Upload from a file and attach
  escape-cli profiles upload-schema <profile-id> --file openapi.json

  # Upload from stdin
  cat openapi.json | escape-cli profiles upload-schema <profile-id>

  # Give the new schema asset a name
  escape-cli profiles upload-schema <profile-id> --file spec.json --name "v2 prod"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetProfile200Response{}) {
			return nil
		}

		profileID := args[0]

		var data []byte
		var err error
		if profileUploadSchemaFile != "" {
			data, err = os.ReadFile(profileUploadSchemaFile)
			if err != nil {
				return fmt.Errorf("unable to read %s: %w", profileUploadSchemaFile, err)
			}
		} else {
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
		}
		if len(data) == 0 {
			return errors.New("no schema bytes provided: pass --file or pipe JSON via stdin")
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), profileUploadSchemaTimeout)
		defer cancel()

		upload, err := escape.GetUploadSignedURL(ctx)
		if err != nil {
			return fmt.Errorf("unable to get signed url: %w", err)
		}
		if err := escape.UploadSchema(ctx, upload.GetUrl(), data); err != nil {
			return fmt.Errorf("failed to upload schema bytes: %w", err)
		}

		asset, err := escape.CreateSchemaAsset(ctx, upload.GetId(), profileUploadSchemaName)
		if err != nil {
			return fmt.Errorf("failed to create schema asset: %w", err)
		}

		profile, err := escape.UpdateProfileSchema(ctx, profileID, asset.GetId())
		if err != nil {
			return fmt.Errorf("failed to attach schema to profile: %w", err)
		}

		out.Table(profile, func() []string {
			return []string{
				"PROFILE ID\tSCHEMA ID\tPROFILE NAME\tASSET TYPE",
				fmt.Sprintf("%s\t%s\t%s\t%s", profile.GetId(), asset.GetId(), profile.GetName(), profile.Asset.GetType()),
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
		profileGetSchemaCmd,
		profileUploadSchemaCmd,
		profileDeleteCmd,
	)
	profileGetSchemaCmd.Flags().StringVar(&profileGetSchemaID, "schema-id", "", "specific schema asset ID to pick from extraAssets (default: active schema)")
	profileGetSchemaCmd.Flags().StringVarP(&profileGetSchemaOutFile, "file", "f", "", "write schema bytes to file path (- for stdout); omit to print JSON metadata")
	profileGetSchemaCmd.Flags().DurationVar(&profileGetSchemaTimeout, "timeout", profileGetSchemaDefaultTimeout, "end-to-end timeout bounding both the GetProfile API call and the signed-URL body fetch (when -f is used)")

	profileUploadSchemaCmd.Flags().StringVar(&profileUploadSchemaFile, "file", "", "path to the schema file (reads stdin when omitted)")
	profileUploadSchemaCmd.Flags().StringVar(&profileUploadSchemaName, "name", "", "optional name for the created schema asset")
	profileUploadSchemaCmd.Flags().DurationVar(&profileUploadSchemaTimeout, "timeout", profileUploadSchemaDefaultTimeout, "end-to-end timeout for upload + schema-build workflow + attach")
	profileUpdateCmd.Flags().StringVar(&profileUpdateName, "name", "", "profile name")
	profileUpdateCmd.Flags().StringVar(&profileUpdateDescription, "description", "", "profile description")
	profileUpdateCmd.Flags().StringVar(&profileUpdateCron, "cron", "", "cron schedule (e.g., \"0 22 * * *\")")
	profileUpdateCmd.Flags().StringVar(&profileUpdateExtraAssetIDs, "extra-asset-id", "", "extra asset ID(s) to attach, as a comma-separated list; replaces the full list of extra assets on the profile")
	profileUpdateCmd.Flags().BoolVar(&profileUpdateClearExtraAssets, "clear-extra-assets", false, "detach every extra asset from the profile")
	profileUpdateCmd.MarkFlagsMutuallyExclusive("extra-asset-id", "clear-extra-assets")
	profileGetCmd.Flags().BoolVar(&profileGetExtraAssets, "extra-assets", false, "list the extra assets attached to the profile (detailed table)")
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
