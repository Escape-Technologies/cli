package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	assetTypes      = []string{}
	assetStatuses   = []string{}
	assetActivities = false
	manuallyCreated = false
)

var (
	assetDescription string
	assetFramework   string
	assetOwners      []string
	assetStatus      string
	assetTagIDs      []string
)

var assetsCmd = &cobra.Command{
	Use:     "assets",
	Aliases: []string{"asset"},
	Short:   "Manage your API and application inventory",
	Long: `Manage Assets - Track Your API and Application Portfolio

Assets represent your APIs, applications, and infrastructure components that are
being monitored and tested by Escape. Each asset can have multiple profiles
(test configurations) and accumulates security findings over time.

ASSET TYPES:
  • WEBAPP              - Web applications and SPAs
  • REST_API            - RESTful APIs
  • GRAPHQL_API         - GraphQL endpoints
  • SOAP_API            - SOAP/XML web services
  • GRPC_API            - gRPC services
  • KUBERNETES_CLUSTER  - K8s clusters (private locations)
  • IPV4, IPV6          - Network endpoints
  • DOMAIN              - DNS domains

ASSET STATUS:
  • MONITORED       - Active monitoring and scanning
  • DEPRECATED      - Deprecated asset (kept for tracking)
  • OUT_OF_SCOPE    - Excluded from scope
  • PERMANENT       - Permanent asset status
  • THIRD_PARTY     - Third-party owned/managed asset
  • FALSE_POSITIVE  - Incorrectly flagged asset status

COMMON WORKFLOWS:
  • List all monitored assets:
    $ escape-cli assets list --statuses MONITORED

  • View asset details and risks:
    $ escape-cli assets get <asset-id>

  • Create a new asset for monitoring:
    $ echo '{"asset_type": "WEBAPP", "url": "https://api.example.com"}' | escape-cli assets create

  • Update asset metadata:
    $ escape-cli assets update <asset-id> --status MONITORED --description "Production API"`,
}

var assetsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List assets with filtering options",
	Long: `List Assets - View Your API Inventory

List all assets in your organization with flexible filtering. Assets represent
the APIs, applications, and infrastructure being monitored by Escape.

FILTER OPTIONS:
  -t, --types            Filter by asset types (WEBAPP, REST_API, GRAPHQL_API, etc.)
  --statuses             Filter by monitoring status (DEPRECATED, FALSE_POSITIVE, MONITORED, OUT_OF_SCOPE, PERMANENT, THIRD_PARTY)
  -s, --search           Free-text search across asset names and URLs
  -m, --manually-created Filter assets created manually vs auto-discovered

RISK INDICATORS:
  • EXPOSED          - Publicly accessible on the internet
  • UNAUTHENTICATED  - No authentication required
  • HIGH_RISK        - Contains high/critical vulnerabilities
  • EXTERNAL         - Third-party or partner APIs

ASSET LIFECYCLE:
  1. Discovery    - Asset found through scanning or manual creation
  2. Profiling    - Security profiles configured
  3. Monitoring   - Regular security scanning active
  4. Remediation  - Issues being fixed
  5. Archiving    - Asset deprecated or decommissioned

Example output:
ID                                      CREATED AT                TYPE                NAME                   RISKS                      STATUS       LAST SEEN
00000000-0000-0000-0000-000000000001    2025-07-22T15:42:12.127Z  KUBERNETES_CLUSTER  private-location       []                         MONITORED    2025-07-22T15:42:12.127Z
00000000-0000-0000-0000-000000000002    2025-07-22T15:52:41.697Z  WEBAPP              https://escape.tech    [EXPOSED UNAUTHENTICATED]  MONITORED    2025-07-22T15:52:41.697Z`,
	Example: `  # List all monitored assets
  escape-cli assets list --statuses MONITORED

  # List only web applications
  escape-cli assets list --types WEBAPP

  # Search for specific assets
  escape-cli assets list --search "api.example.com"

  # List manually created assets
  escape-cli assets list --manually-created

  # Export asset inventory to JSON
  escape-cli assets list -o json > asset-inventory.json`,

	RunE: func(cmd *cobra.Command, _ []string) error {
		assets, next, err := escape.ListAssets(cmd.Context(), "", &escape.ListAssetsFilters{
			AssetTypes:      assetTypes,
			AssetStatuses:   assetStatuses,
			Search:          search,
			ManuallyCreated: manuallyCreated,
		})
		if err != nil {
			return fmt.Errorf("unable to list assets: %w", err)
		}

		out.Table(assets, func() []string {
			res := []string{"ID\tCREATED AT\tTYPE\tSTATUS\tLAST SEEN\tRISKS\tNAME"}
			for _, asset := range assets {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", asset.GetId(), asset.GetCreatedAt(), asset.GetType(), asset.GetStatus(), asset.GetLastSeenAt(), asset.GetRisks(), asset.GetName()))
			}
			return res
		})

		for next != nil && *next != "" {
			assets, next, err = escape.ListAssets(
				cmd.Context(),
				*next,
				&escape.ListAssetsFilters{
					AssetTypes:      assetTypes,
					AssetStatuses:   assetStatuses,
					Search:          search,
					ManuallyCreated: manuallyCreated,
				},
			)
			if err != nil {
				return fmt.Errorf("unable to list assets: %w", err)
			}
			out.Table(assets, func() []string {
				res := []string{"ID\tCREATED AT\tTYPE\tSTATUS\tLAST SEEN\tRISKS\tNAME"}
				for _, asset := range assets {
					res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", asset.GetId(), asset.GetCreatedAt(), asset.GetType(), asset.GetStatus(), asset.GetLastSeenAt(), asset.GetRisks(), asset.GetName()))
				}
				return res
			})
		}

		return nil
	},
}

var assetGetCmd = &cobra.Command{
	Use:     "get asset-id",
	Aliases: []string{"g", "show", "describe"},
	Short:   "Get detailed information about an asset",
	Long: `Get Asset Details - View Complete Asset Information

Retrieve comprehensive information about a specific asset including its type,
status, risk indicators, and last seen timestamp.

DISPLAYED INFORMATION:
  • ID          - Unique asset identifier
  • CREATED AT  - When asset was first discovered or created
  • TYPE        - Asset classification (WEBAPP, REST_API, etc.)
  • NAME        - Asset name or primary URL
  • RISKS       - Security risk indicators
  • STATUS      - Current monitoring status
  • LAST SEEN   - Most recent scan or check

ADDITIONAL OPTIONS:
  -a, --activities   Show related issue activities for this asset

USE CASES:
  • Review asset security posture
  • Check when asset was last scanned
  • Verify asset configuration
  • Export asset data for reports

Example output:
ID                                      CREATED AT                TYPE    NAME                RISKS                      STATUS      LAST SEEN
00000000-0000-0000-0000-000000000001    2025-07-22T15:52:41.697Z  WEBAPP  https://escape.tech [EXPOSED UNAUTHENTICATED]  MONITORED   2025-07-22T15:52:41.697Z`,
	Example: `  # Get asset details
  escape-cli assets get 00000000-0000-0000-0000-000000000000

  # Get asset with related activities
  escape-cli assets get <asset-id> --activities

  # Export asset details to JSON
  escape-cli assets get <asset-id> -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("asset ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		asset, err := escape.GetAsset(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get asset: %w", err)
		}

		if assetActivities {
			issues, _, err := escape.ListIssues(cmd.Context(), "", &escape.ListIssuesFilters{
				AssetIDs: []string{args[0]},
			})
			if err != nil {
				return fmt.Errorf("unable to list issues: %w", err)
			}

			allActivities := []v3.ActivitySummarized{}
			for _, issue := range issues {
				activities, err := escape.ListIssueActivities(cmd.Context(), issue.GetId())
				if err != nil {
					return fmt.Errorf("unable to list activities: %w", err)
				}
				allActivities = append(allActivities, activities...)
			}

			out.Table(allActivities, func() []string {
				res := []string{"ID\tCREATED AT\tKIND"}
				for _, activity := range allActivities {
					res = append(res, fmt.Sprintf("%s\t%s\t%s", activity.GetId(), activity.GetCreatedAt(), activity.GetKind()))
				}
				return res
			})
		} else {

			out.Table(asset, func() []string {
				res := []string{"ID\tCREATED AT\tTYPE\tSTATUS\tLAST SEEN\tRISKS\tNAME"}
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", asset.GetId(), asset.GetCreatedAt(), asset.GetType(), asset.GetStatus(), asset.GetLastSeenAt(), asset.GetRisks(), asset.GetName()))
				return res
			})
		}

		return nil
	},
}

var assetDeleteCmd = &cobra.Command{
	Use:     "delete asset-id",
	Aliases: []string{"d", "rm", "remove"},
	Short:   "Delete an asset from your inventory",
	Long: `Delete Asset - Remove from Monitoring

Permanently delete an asset from your inventory. This will also remove:
  • All associated security profiles
  • Historical scan results
  • Issue findings linked to this asset
  • Activity logs and events

WARNING: This action is IRREVERSIBLE!

ALTERNATIVES TO DELETION:
  Instead of deleting, consider:
  • Changing status: Use 'escape-cli assets update <id> --status <STATUS>'
    (valid STATUS values: DEPRECATED, FALSE_POSITIVE, MONITORED, OUT_OF_SCOPE, PERMANENT, THIRD_PARTY)

WHEN TO DELETE:
  • Test/temporary assets no longer needed
  • Duplicate asset entries
  • Assets created by mistake
  • Complete decommissioning (after exporting data)

BEFORE DELETING:
  • Export asset data if needed: escape-cli assets get <id> -o json
  • Review associated issues: escape-cli issues list --asset-id <id>
  • Consider archiving instead for audit trails`,
	Example: `  # Delete an asset
  escape-cli assets delete 00000000-0000-0000-0000-000000000000

  # Export data before deleting
  escape-cli assets get <asset-id> -o json > asset-backup.json
  escape-cli assets delete <asset-id>

  # Delete multiple test assets
  escape-cli assets list --search "test" -o json | jq -r '.[].id' | xargs -I {} escape-cli assets delete {}`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("asset ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.DeleteAsset(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to delete asset: %w", err)
		}
		fmt.Printf("Asset %s successfully deleted\n", args[0])
		return nil
	},
}

var assetUpdateCmd = &cobra.Command{
	Use:     "update asset-id",
	Aliases: []string{"u", "modify", "edit"},
	Short:   "Update asset metadata and configuration",
	Long: `Update Asset - Modify Asset Information

Update an existing asset's metadata including status, description, owners, tags,
and framework classification. Use this to maintain accurate asset inventory.

UPDATABLE FIELDS:
  -d, --description    Human-readable description
  -f, --framework      Asset framework/type classification
  -s, --status         Monitoring status (DEPRECATED, FALSE_POSITIVE, MONITORED, OUT_OF_SCOPE, PERMANENT, THIRD_PARTY)
  --owners             Asset owners (email addresses)
  -t, --tag-ids        Tag IDs for organization

USE CASES:
  • Update asset description for clarity
  • Change monitoring status
  • Assign ownership for accountability
  • Add tags for organization and filtering
  • Mark assets as DEPRECATED / OUT_OF_SCOPE`,
	Example: `  # Update asset description
  escape-cli assets update <asset-id> --description "Production REST API"

  # Change monitoring status
  escape-cli assets update <asset-id> --status MONITORED

  # Assign owners
  escape-cli assets update <asset-id> --owners "security@example.com,devops@example.com"

  # Add tags for organization
  escape-cli assets update <asset-id> --tag-ids "00000000-0000-0000-0000-000000000000,00000000-0000-0000-0000-000000000001"

  # Mark deprecated asset
  escape-cli assets update <asset-id> --status DEPRECATED --description "Deprecated - removed 2025-10-01"

  # Update multiple fields at once
  escape-cli assets update <asset-id> \
    --status MONITORED \
    --description "Customer API v2" \
    --owners "api-team@example.com" \
    --tag-ids "tag-external,tag-production"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("asset ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var framework *v3.ENUMPROPERTIESFRAMEWORK
		if assetFramework != "" {
			f := v3.ENUMPROPERTIESFRAMEWORK(assetFramework)
			framework = &f
		}

		var status *v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUS
		if assetStatus != "" {
			s := v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUS(assetStatus)
			status = &s
		}

		var desc *string
		if assetDescription != "" {
			desc = &assetDescription
		}

		var owners *[]string
		if len(assetOwners) > 0 {
			owners = &assetOwners
		}

		var tagIDs *[]string
		if len(assetTagIDs) > 0 {
			tagIDs = &assetTagIDs
		}

		err := escape.UpdateAsset(cmd.Context(), args[0], desc, framework, owners, status, tagIDs)
		if err != nil {
			return fmt.Errorf("unable to update asset: %w", err)
		}
		fmt.Printf("Asset %s successfully updated\n", args[0])
		return nil
	},
}

var createAssetCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c", "add", "new"},
	Short:   "Create a new asset for security monitoring",
	Long: `Create Asset - Add New API or Application to Inventory

Create a new asset to begin security monitoring. Provide asset details via JSON
input through stdin. Once created, you can configure security profiles for scanning.

REQUIRED FIELDS:
  • asset_type   - Asset classification (see types below)
  • url or name  - Identifier (depends on asset type)

COMMON ASSET TYPES & EXAMPLES:
  WEBAPP:
    {"asset_type": "WEBAPP", "url": "https://app.example.com"}
  
  REST_API:
    {"asset_type": "REST_API", "url": "https://api.example.com"}
  
  GRAPHQL_API:
    {"asset_type": "GRAPHQL_API", "url": "https://api.example.com/graphql"}
  
  IPV4/IPV6:
    {"asset_type": "IPV4", "ip": "192.168.1.1"}
    {"asset_type": "IPV6", "ip": "2001:0db8:85a3::8a2e:0370:7334"}
  
  DOMAIN:
    {"asset_type": "DOMAIN", "name": "example.com"}

OPTIONAL FIELDS:
  • description  - Human-readable description
  • status       - Initial status (default: MONITORED)
  • tags         - Array of tag IDs for organization

For complete schema and all asset types:
https://public.escape.tech/v3/#tag/assets

Example output:
ID                                    TYPE    NAME                  STATUS
8163b58c-5413-4224-bdae-a0d395c4a766  WEBAPP  https://example.com   MONITORED`,
	Example: `  # Create a web application asset
  echo '{"asset_type": "WEBAPP", "url": "https://app.example.com"}' | escape-cli assets create

  # Create from a file
  escape-cli assets create < asset-config.json

  # Create REST API asset
  cat <<EOF | escape-cli assets create
  {
    "asset_type": "REST_API",
    "url": "https://api.example.com",
    "description": "Production API",
    "status": "MONITORED"
  }
  EOF

  # Create and capture ID for further use
  ASSET_ID=$(echo '{"asset_type": "WEBAPP", "url": "https://example.com"}' | escape-cli assets create -o json | jq -r '.id')`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			_ = cmd.Help()
			return errors.New("this command does not accept any arguments, it reads from stdin")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		var data []byte

		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		data = b

		var asset map[string]interface{}
		if err := json.Unmarshal(data, &asset); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		typeVal, _ := asset["asset_type"].(string)
		if strings.TrimSpace(typeVal) == "" {
			return errors.New("invalid JSON: missing 'asset_type'")
		}

		response, err := escape.CreateAsset(cmd.Context(), data, strings.ToUpper(typeVal))
		if err != nil {
			return fmt.Errorf("failed to create asset: %w", err)
		}

		out.Table(response, func() []string {
			result := []string{"ID\tTYPE\tNAME\tSTATUS"}
			if assetResponse, ok := response.(*v3.AssetDetailed); ok {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s",
					assetResponse.GetId(),
					assetResponse.GetType(),
					assetResponse.GetName(),
					assetResponse.GetStatus(),
				))
			}
			return result
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(assetsCmd)
	assetsCmd.AddCommand(assetsListCmd)
	assetsListCmd.Flags().StringSliceVarP(&assetTypes, "types", "t", []string{}, fmt.Sprintf("filter by asset types (comma-separated): %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetsListCmd.Flags().StringSliceVarP(&assetStatuses, "statuses", "", []string{}, fmt.Sprintf("filter by monitoring status: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUSEnumValues))
	assetsListCmd.Flags().StringVarP(&search, "search", "s", "", "free-text search across asset names and URLs")
	assetsListCmd.Flags().BoolVarP(&manuallyCreated, "manually-created", "m", false, "show only manually created assets (exclude auto-discovered)")

	assetsCmd.AddCommand(assetGetCmd)
	assetGetCmd.Flags().BoolVarP(&assetActivities, "activities", "a", false, "include issue activity timeline for this asset")
	assetsCmd.AddCommand(assetDeleteCmd)

	assetsCmd.AddCommand(assetUpdateCmd)
	assetUpdateCmd.Flags().StringVarP(&assetDescription, "description", "d", "", "human-readable description of the asset")
	assetUpdateCmd.Flags().StringVarP(&assetFramework, "framework", "f", "", fmt.Sprintf("asset framework/type classification: %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetUpdateCmd.Flags().StringSliceVarP(&assetOwners, "owners", "", []string{}, "comma-separated list of owner email addresses")
	assetUpdateCmd.Flags().StringVarP(&assetStatus, "status", "s", "", fmt.Sprintf("monitoring status: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUSEnumValues))
	assetUpdateCmd.Flags().StringSliceVarP(&assetTagIDs, "tag-ids", "t", []string{}, "comma-separated list of tag IDs for organization")

	assetsCmd.AddCommand(createAssetCmd)
}
