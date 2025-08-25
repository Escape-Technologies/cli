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
	assetTypes    = []string{}
	assetStatuses = []string{}
	assetActivities = false
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
	Short:   "Interact with assets",
	Long:    "Interact with your assets",
}

var assetsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List assets",
	Example: `escape-cli asset list`,
	Long: `List assets of the organization.
Example output:
ID                                      CREATED AT                            TYPE                            NAME                            RISKS                           STATUS       LAST SEEN
00000000-0000-0000-0000-000000000001    2025-07-22T15:42:12.127Z              KUBERNETES_CLUSTER              private-location                []                              MONITORED    2025-07-22T15:42:12.127Z
00000000-0000-0000-0000-000000000002    2025-07-22T15:52:41.697Z              WEBAPP                          https://escape.tech             [EXPOSED UNAUTHENTICATED]       MONITORED    2025-07-22T15:52:41.697Z`,

	RunE: func(cmd *cobra.Command, _ []string) error {
		assets, next, err := escape.ListAssets(cmd.Context(), "", assetTypes, assetStatuses)
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
				assetTypes,
				assetStatuses,
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
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get an asset",
	Example: `escape-cli asset get <asset-id>`,
	Long: `Get an asset by ID.
Example output:
ID                                      CREATED AT                            TYPE                            NAME                            RISKS                           STATUS       LAST SEEN
00000000-0000-0000-0000-000000000001    2025-07-22T15:52:41.697Z              WEBAPP                          https://escape.tech             [EXPOSED UNAUTHENTICATED]       MONITORED    2025-07-22T15:52:41.697Z`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("asset ID is required")
		}
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
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete an asset",
	Example: `escape-cli asset delete <asset-id>`,
	Args:    cobra.ExactArgs(1),
	Long: `Delete an asset by ID.
Example output:
Asset 00000000-0000-0000-0000-000000000001 successfully deleted`,
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
	Use:     "update",
	Aliases: []string{"u"},
	Short:   "Update an asset",
	Example: `escape-cli asset update <asset-id> -s MONITORED -f KUBERNETES_CLUSTER -d "My Kubernetes Cluster" -o "owner1-id,owner2-id" -t "tag1-id,tag2-id"`,
	Long: `Update an asset by ID.
Example output:
Asset 00000000-0000-0000-0000-000000000001 successfully updated`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("asset ID is required")
		}
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
	Aliases: []string{"c"},
	Short:   "Create an asset",
	Example: `escape-cli asset create <test.json`,
	Long: `Create an asset by JSON.
Example output:
ID                                    TYPE  NAME                                     STATUS
8163b58c-5413-4224-bdae-a0d395c4a766  IPV6  2001:0db8:85a3:0000:0000:8a2e:0370:7334  MONITORED
		
for more examples see required fields at https://public.escape.tech/v3/#tag/assets`,

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
	assetsListCmd.Flags().StringSliceVarP(&assetTypes, "types", "t", []string{}, fmt.Sprintf("Filter by asset types: %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetsListCmd.Flags().StringSliceVarP(&assetStatuses, "statuses", "s", []string{}, fmt.Sprintf("Filter by asset statuses: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUSEnumValues))
	
	assetsCmd.AddCommand(assetGetCmd)
	assetGetCmd.Flags().BoolVarP(&assetActivities, "activities", "a", false, "list of activities attached to the issues of the asset")
	assetsCmd.AddCommand(assetDeleteCmd)
	

	assetsCmd.AddCommand(assetUpdateCmd)
	assetUpdateCmd.Flags().StringVarP(&assetDescription, "description", "d", "", "description of the asset")
	assetUpdateCmd.Flags().StringVarP(&assetFramework, "framework", "f", "", fmt.Sprintf("framework of the asset: %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetUpdateCmd.Flags().StringSliceVarP(&assetOwners, "owners", "", []string{}, "list of asset owners (email)")
	assetUpdateCmd.Flags().StringVarP(&assetStatus, "status", "s", "", fmt.Sprintf("status of the asset: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUSEnumValues))
	assetUpdateCmd.Flags().StringSliceVarP(&assetTagIDs, "tag-ids", "t", []string{}, "list of tag IDs")

	assetsCmd.AddCommand(createAssetCmd)
}
