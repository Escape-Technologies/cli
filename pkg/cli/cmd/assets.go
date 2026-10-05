package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	assetTypes         = []string{}
	assetStatuses      = []string{}
	assetProjectIDs    = []string{}
	assetActivities    = false
	manuallyCreated    = false
	assetSortType      string
	assetSortDirection string
	assetListPage      pageFlags
)

var (
	assetDescription      string
	assetFramework        string
	assetName             string
	assetOwners           []string
	assetStatus           string
	assetTagIDs           []string
	assetUpdateProjectIDs []string
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
  • IPV4, IPV6          - Network endpoints
  • DOMAIN              - DNS domains

ASSET STATUS:
  • MONITORED    - Active monitoring and scanning
  • UNMONITORED  - Discovered but not actively tested
  • ARCHIVED     - No longer in use

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
  --statuses             Filter by monitoring status (MONITORED, UNMONITORED, ARCHIVED)
  --project-id           Filter by project ID
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
00000000-0000-0000-0000-000000000001    2025-07-22T15:42:12.127Z  GITHUB_REPOSITORY   escape-api             []                         MONITORED    2025-07-22T15:42:12.127Z
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
		// Output JSON Schema if requested
		if out.Schema([]v3.AssetSummarized{}) {
			return nil
		}

		filters := &escape.ListAssetsFilters{
			AssetTypes:      assetTypes,
			AssetStatuses:   assetStatuses,
			ProjectIDs:      assetProjectIDs,
			Search:          search,
			ManuallyCreated: manuallyCreated,
			SortType:        assetSortType,
			SortDirection:   assetSortDirection,
		}
		if err := runPagedList(cmd, assetListPage, func(ctx context.Context, cursor string, size int) ([]v3.AssetSummarized, *string, int, error) {
			return escape.ListAssets(ctx, cursor, filters, size)
		}, func(assets []v3.AssetSummarized) []string {
			res := []string{"ID\tCREATED AT\tTYPE\tSTATUS\tLAST SEEN\tRISKS\tTAGS\tOWNERS\tPROJECTS\tNAME"}
			for _, asset := range assets {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s", asset.GetId(), asset.GetCreatedAt(), asset.GetType(), asset.GetStatus(), asset.GetLastSeenAt(), asset.GetRisks(), joinTags(asset.GetTags()), ownersColumn(asset.AdditionalProperties), len(asset.GetProjectIds()), asset.GetName()))
			}

			return res
		}); err != nil {
			return fmt.Errorf("unable to list assets: %w", err)
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
		// Output JSON Schema if requested
		if out.Schema(v3.AssetDetailed{}) {
			return nil
		}

		asset, err := escape.GetAsset(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get asset: %w", err)
		}

		if assetActivities {
			issues, _, _, err := escape.ListIssues(cmd.Context(), "", &escape.ListIssuesFilters{
				AssetIDs: []string{args[0]},
			}, "", "", 0)
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
				res := []string{"ID\tCREATED AT\tTYPE\tSTATUS\tLAST SEEN\tRISKS\tTAGS\tOWNERS\tPROJECTS\tNAME"}
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s", asset.GetId(), asset.GetCreatedAt(), asset.GetType(), asset.GetStatus(), asset.GetLastSeenAt(), asset.GetRisks(), joinTags(asset.GetTags()), ownersColumn(asset.AdditionalProperties), len(asset.GetProjectIds()), asset.GetName()))

				return res
			})
		}

		return nil
	},
}

func joinTags(tags []v3.Tag) string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag.GetName() != "" {
			names = append(names, tag.GetName())
		}
	}

	return strings.Join(names, ",")
}

func ownersColumn(additionalProperties map[string]interface{}) string {
	value, ok := additionalProperties["owners"]
	if !ok {
		return ""
	}

	items, ok := value.([]interface{})
	if !ok {
		return stringValue(value)
	}

	owners := make([]string, 0, len(items))
	for _, item := range items {
		owners = append(owners, stringValue(item))
	}

	return strings.Join(owners, ",")
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

⚠️  WARNING: This action is IRREVERSIBLE!

ALTERNATIVES TO DELETION:
  Instead of deleting, consider:
  • Archiving: Use 'escape-cli assets update <id> --status ARCHIVED'
  • Unmonitoring: Use 'escape-cli assets update <id> --status UNMONITORED'

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
		if out.Schema(v3.DeleteProfile200Response{}) {
			return nil
		}

		result, err := escape.DeleteAsset(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to delete asset: %w", err)
		}

		out.Print(result, fmt.Sprintf("Asset %s successfully deleted", args[0]))

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
  --name               Custom asset name (defaults to the discovered name)
  --project-id         Project IDs to assign the asset to
  -s, --status         Monitoring status (MONITORED, UNMONITORED, ARCHIVED)
  --owners             Asset owners (email addresses)
  -t, --tag-ids        Tag IDs for organization

STATUS TRANSITIONS:
  • MONITORED    → UNMONITORED   Stop active scanning
  • UNMONITORED  → MONITORED     Resume security testing
  • Any          → ARCHIVED      Mark as decommissioned

USE CASES:
  • Update asset description for clarity
  • Change monitoring status
  • Assign ownership for accountability
  • Add tags for organization and filtering
  • Archive deprecated APIs`,
	Example: `  # Update asset description
  escape-cli assets update <asset-id> --description "Production REST API"

  # Change monitoring status
  escape-cli assets update <asset-id> --status MONITORED

  # Assign owners
  escape-cli assets update <asset-id> --owners "security@example.com,devops@example.com"

  # Add tags for organization
  escape-cli assets update <asset-id> --tag-ids "tag-prod,tag-critical"

  # Assign to projects and rename
  escape-cli assets update <asset-id> --project-id "proj-1,proj-2" --name "Payments API"

  # Archive decommissioned asset
  escape-cli assets update <asset-id> --status ARCHIVED --description "Deprecated - removed 2025-10-01"

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
		if out.Schema(v3.UpdateAsset200Response{}) {
			return nil
		}

		var framework *v3.ENUMPROPERTIESFRAMEWORK
		if assetFramework != "" {
			f := v3.ENUMPROPERTIESFRAMEWORK(assetFramework)
			framework = &f
		}

		var status *v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS
		if assetStatus != "" {
			s := v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS(assetStatus)
			status = &s
		}

		var desc *string
		if assetDescription != "" {
			desc = &assetDescription
		}

		var name *string
		if assetName != "" {
			name = &assetName
		}

		var owners *[]string
		if len(assetOwners) > 0 {
			owners = &assetOwners
		}

		var tagIDs *[]string
		if len(assetTagIDs) > 0 {
			tagIDs = &assetTagIDs
		}

		var projectIDs *[]string
		if len(assetUpdateProjectIDs) > 0 {
			projectIDs = &assetUpdateProjectIDs
		}

		asset, err := escape.UpdateAsset(cmd.Context(), args[0], desc, framework, owners, status, tagIDs, projectIDs, name)
		if err != nil {
			return fmt.Errorf("unable to update asset: %w", err)
		}

		out.Print(asset, "Asset "+args[0]+" successfully updated")

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
  
  REST:
    {"asset_type": "REST", "url": "https://api.example.com"}
  
  GRAPHQL:
    {"asset_type": "GRAPHQL", "url": "https://api.example.com/graphql"}
  
  IPV4/IPV6:
    {"asset_type": "IPV4", "ip": "192.168.1.1"}
    {"asset_type": "IPV6", "ip": "2001:0db8:85a3::8a2e:0370:7334"}
  
  DNS:
    {"asset_type": "DNS", "name": "example.com"}

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
    "asset_type": "REST",
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
		if rootCmdInputSchema {
			schema, err := assetCreateInputSchema()
			if err != nil {
				return fmt.Errorf("input schema: %w", err)
			}

			out.SetInputSchema(true)
			if out.InputSchema(providedSchema{schema: schema}) {
				return nil
			}
		}

		// Output JSON Schema if requested
		if out.Schema(v3.AssetDetailed{}) {
			return nil
		}

		data, err := readPipedStdin(cmd.InOrStdin())
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return errors.New("JSON body is required on stdin")
		}

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

var (
	bulkAssetIDs      []string
	bulkAssetTypes    []string
	bulkAssetStatuses []string
	bulkAssetTagIDs   []string
	bulkAssetProjIDs  []string
	bulkAssetStatus   string
	assetCommentMsg   string
)

var assetCommentCmd = &cobra.Command{
	Use:   "comment asset-id",
	Short: "Add a comment to an asset",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("asset ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateAssetComment200Response{}) {
			return nil
		}

		if assetCommentMsg == "" {
			return errors.New("--message is required")
		}

		comment, err := escape.CommentAsset(cmd.Context(), args[0], assetCommentMsg)
		if err != nil {
			return fmt.Errorf("unable to add comment: %w", err)
		}

		out.Print(comment, "Comment added to asset "+args[0])

		return nil
	},
}

var assetListActivitiesCmd = &cobra.Command{
	Use:     "list-activities asset-id",
	Aliases: []string{"activities"},
	Short:   "List activities for an asset",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("asset ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]v3.ActivitySummarized{}) {
			return nil
		}

		activities, err := escape.ListAssetActivities(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to list activities: %w", err)
		}

		out.Table(activities, func() []string {
			res := []string{"ID\tCREATED AT\tKIND\tAUTHOR EMAIL"}
			for _, a := range activities {
				email := ""
				if a.Author != nil {
					email = a.Author.Email
				}

				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s", a.GetId(), a.GetCreatedAt(), a.GetKind(), email))
			}

			return res
		})

		return nil
	},
}

var assetBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Update multiple assets matching a filter",
	Long:  `Bulk update tags, projects, or status of assets matching a filter predicate.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.BulkUpdateAssets200Response{}) {
			return nil
		}

		if err := requireAssetBulkSelection(); err != nil {
			return err
		}

		where := v3.BulkUpdateAssetsRequestWhere{}
		if len(bulkAssetIDs) > 0 {
			where.AssetIds = bulkAssetIDs
		}

		if len(bulkAssetTypes) > 0 {
			types := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE, len(bulkAssetTypes))
			for i, t := range bulkAssetTypes {
				types[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE(t)
			}

			where.Types = types
		}

		if len(bulkAssetStatuses) > 0 {
			statuses := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS, len(bulkAssetStatuses))
			for i, s := range bulkAssetStatuses {
				statuses[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS(s)
			}

			where.Statuses = statuses
		}

		var status *v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS
		if bulkAssetStatus != "" {
			s := v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS(bulkAssetStatus)
			status = &s
		}

		result, err := escape.BulkUpdateAssets(cmd.Context(), where, bulkAssetTagIDs, bulkAssetProjIDs, status)
		if err != nil {
			return fmt.Errorf("unable to bulk update assets: %w", err)
		}

		out.Print(result, "Bulk update completed")

		return nil
	},
}

var assetBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Delete multiple assets matching a filter",
	Long:  `Schedule multiple assets matching a filter predicate for deletion.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.BulkUpdateAssets200Response{}) {
			return nil
		}

		if err := requireAssetBulkSelection(); err != nil {
			return err
		}

		where := v3.BulkUpdateAssetsRequestWhere{}
		if len(bulkAssetIDs) > 0 {
			where.AssetIds = bulkAssetIDs
		}

		if len(bulkAssetTypes) > 0 {
			types := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE, len(bulkAssetTypes))
			for i, t := range bulkAssetTypes {
				types[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPE(t)
			}

			where.Types = types
		}

		if len(bulkAssetStatuses) > 0 {
			statuses := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS, len(bulkAssetStatuses))
			for i, s := range bulkAssetStatuses {
				statuses[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUS(s)
			}

			where.Statuses = statuses
		}

		result, err := escape.BulkDeleteAssets(cmd.Context(), where)
		if err != nil {
			return fmt.Errorf("unable to bulk delete assets: %w", err)
		}

		out.Print(result, "Bulk delete scheduled")

		return nil
	},
}

// requireAssetBulkSelection rejects a bulk asset mutation with an empty where.
// The public API accepts that payload, and deleteAssets/updateAssets then match
// every asset the caller can see.
func requireAssetBulkSelection() error {
	if len(bulkAssetIDs) > 0 || len(bulkAssetTypes) > 0 || len(bulkAssetStatuses) > 0 {
		return nil
	}

	return errors.New("at least one of --asset-id, --type, or --asset-status is required")
}

func init() {
	rootCmd.AddCommand(assetsCmd)
	assetsCmd.AddCommand(assetsListCmd)
	assetsListCmd.Flags().StringSliceVarP(&assetTypes, "types", "t", []string{}, fmt.Sprintf("filter by asset types (comma-separated): %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetsListCmd.Flags().StringSliceVarP(&assetStatuses, "statuses", "", []string{}, fmt.Sprintf("filter by monitoring status: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUSEnumValues))
	assetsListCmd.Flags().StringSliceVar(&assetProjectIDs, "project-id", []string{}, "filter by project IDs")
	assetsListCmd.Flags().StringVarP(&search, "search", "s", "", "free-text search across asset names and URLs")
	assetsListCmd.Flags().BoolVarP(&manuallyCreated, "manually-created", "m", false, "show only manually created assets (exclude auto-discovered)")
	assetsListCmd.Flags().StringVar(&assetSortType, "sort-by", "", "sort field (e.g., LAST_SEEN, CREATED_AT)")
	assetsListCmd.Flags().StringVar(&assetSortDirection, "sort-direction", "", "sort direction: asc, desc")
	assetListPage.bind(assetsListCmd)

	assetsCmd.AddCommand(assetGetCmd)
	assetGetCmd.Flags().BoolVarP(&assetActivities, "activities", "a", false, "include issue activity timeline for this asset")
	assetsCmd.AddCommand(assetDeleteCmd)

	assetsCmd.AddCommand(assetUpdateCmd)
	assetUpdateCmd.Flags().StringVarP(&assetDescription, "description", "d", "", "human-readable description of the asset")
	assetUpdateCmd.Flags().StringVarP(&assetFramework, "framework", "f", "", fmt.Sprintf("asset framework/type classification: %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetUpdateCmd.Flags().StringVar(&assetName, "name", "", "custom asset name (falls back to the discovered name when empty)")
	assetUpdateCmd.Flags().StringSliceVar(&assetUpdateProjectIDs, "project-id", nil, "project IDs to assign the asset to")
	assetUpdateCmd.Flags().StringSliceVarP(&assetOwners, "owners", "", []string{}, "comma-separated list of owner email addresses")
	assetUpdateCmd.Flags().StringVarP(&assetStatus, "status", "s", "", fmt.Sprintf("monitoring status: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUSEnumValues))
	assetUpdateCmd.Flags().StringSliceVarP(&assetTagIDs, "tag-ids", "t", []string{}, "comma-separated list of tag IDs for organization")

	assetsCmd.AddCommand(createAssetCmd)

	assetsCmd.AddCommand(assetCommentCmd)
	assetCommentCmd.Flags().StringVar(&assetCommentMsg, "message", "", "comment text (max 512 characters)")

	assetsCmd.AddCommand(assetListActivitiesCmd)

	assetsCmd.AddCommand(assetBulkUpdateCmd)
	assetBulkUpdateCmd.Flags().StringSliceVar(&bulkAssetIDs, "asset-id", nil, "filter by asset ID(s)")
	assetBulkUpdateCmd.Flags().StringSliceVar(&bulkAssetTypes, "type", nil, "filter by asset type(s)")
	assetBulkUpdateCmd.Flags().StringSliceVar(&bulkAssetStatuses, "asset-status", nil, "filter by current status")
	assetBulkUpdateCmd.Flags().StringSliceVar(&bulkAssetTagIDs, "tag-id", nil, "tag IDs to assign")
	assetBulkUpdateCmd.Flags().StringSliceVar(&bulkAssetProjIDs, "project-id", nil, "project IDs to assign")
	assetBulkUpdateCmd.Flags().StringVar(&bulkAssetStatus, "status", "", "new status to apply")

	assetsCmd.AddCommand(assetBulkDeleteCmd)
	assetBulkDeleteCmd.Flags().StringSliceVar(&bulkAssetIDs, "asset-id", nil, "filter by asset ID(s)")
	assetBulkDeleteCmd.Flags().StringSliceVar(&bulkAssetTypes, "type", nil, "filter by asset type(s)")
	assetBulkDeleteCmd.Flags().StringSliceVar(&bulkAssetStatuses, "asset-status", nil, "filter by current status")
}
