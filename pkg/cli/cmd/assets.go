package cmd

import (
	"errors"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	assetTypes    = []string{}
	assetStatuses = []string{}
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
	Long:    `List assets of the organization.`,
	Example: `escape-cli asset list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		assets, next, err := escape.ListAssets(cmd.Context(), "", assetTypes, assetStatuses)
		if err != nil {
			return fmt.Errorf("unable to list assets: %w", err)
		}

		// First result
		result := []string{"ID\tTYPE\tNAME\tRISKS\tSTATUS\tLAST SEEN"}
		for _, asset := range assets {
			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", asset.GetId(), asset.GetName(), asset.GetType(), asset.GetRisks(), asset.GetStatus(), asset.GetLastSeenAt()))
		}

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
			for _, asset := range assets {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s", asset.GetId(), asset.GetName(), asset.GetType(), asset.GetCreatedAt()))
			}
		}

		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}

var assetGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get an asset",
	Long:    `Get an asset by ID.`,
	Example: `escape-cli asset get <asset-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("asset ID is required")
		}
		asset, err := escape.GetAsset(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get asset: %w", err)
		}

		result := []string{"ID\tTYPE\tNAME\tRISKS\tSTATUS\tLAST SEEN"}
		result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", asset.GetId(), asset.GetType(), asset.GetName(), asset.GetRisks(), asset.GetStatus(), asset.GetLastSeenAt()))
		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}

var assetDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete an asset",
	Long:    `Delete an asset by ID.`,
	Example: `escape-cli asset delete <asset-id>`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.DeleteAsset(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to delete asset: %w", err)
		}
		fmt.Printf("Asset %s deleted successfully\n", args[0])
		return nil
	},
}

var assetUpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u"},
	Short:   "Update an asset",
	Long:    `Update an asset by ID.`,
	Example: `escape-cli asset update <asset-id>`,
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

		err := escape.UpdateAsset(cmd.Context(), args[0], desc, framework, &assetOwners, status, &assetTagIDs)
		if err != nil {
			return fmt.Errorf("unable to update asset: %w", err)
		}
		fmt.Printf("Asset %s updated successfully\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(assetsCmd)
	assetsCmd.AddCommand(assetsListCmd)
	assetsCmd.AddCommand(assetGetCmd)
	assetsCmd.AddCommand(assetDeleteCmd)
	assetsListCmd.Flags().StringSliceVarP(&assetTypes, "types", "t", []string{}, fmt.Sprintf("Filter by asset types: %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetsListCmd.Flags().StringSliceVarP(&assetStatuses, "statuses", "s", []string{}, fmt.Sprintf("Filter by asset statuses: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUSEnumValues))

	assetsCmd.AddCommand(assetUpdateCmd)
	assetUpdateCmd.Flags().StringVarP(&assetDescription, "description", "d", "", "description of the asset")
	assetUpdateCmd.Flags().StringVarP(&assetFramework, "framework", "f", "", fmt.Sprintf("framework of the asset: %v", v3.AllowedENUMPROPERTIESFRAMEWORKEnumValues))
	assetUpdateCmd.Flags().StringSliceVarP(&assetOwners, "owners", "", []string{}, "list of asset owners (email)")
	assetUpdateCmd.Flags().StringVarP(&assetStatus, "status", "s", "", fmt.Sprintf("status of the asset: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATUSEnumValues))
	assetUpdateCmd.Flags().StringSliceVarP(&assetTagIDs, "tag-ids", "t", []string{}, "list of tag IDs")
}
