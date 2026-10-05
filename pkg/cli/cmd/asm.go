package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var asmAssetIDs []string

var asmCmd = &cobra.Command{
	Use:   "asm",
	Short: "Attack Surface Management operations",
	Long:  `Trigger and manage ASM discovery scans on your attack surface.`,
}

var asmTriggerCmd = &cobra.Command{
	Use:     "trigger",
	Aliases: []string{"scan", "start"},
	Short:   "Trigger ASM scans on assets",
	Long:    `Trigger Attack Surface Management scans. Optionally filter by asset IDs to scan specific assets.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.TriggerAsmScans200Response{}) {
			return nil
		}

		var where *v3.TriggerAsmScansRequestWhere
		if len(asmAssetIDs) > 0 {
			where = &v3.TriggerAsmScansRequestWhere{
				AssetIds: asmAssetIDs,
			}
		}

		result, err := escape.TriggerAsmScans(cmd.Context(), where)
		if err != nil {
			return fmt.Errorf("unable to trigger ASM scans: %w", err)
		}

		out.Print(result, "ASM scans triggered successfully")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(asmCmd)
	asmCmd.AddCommand(asmTriggerCmd)
	asmTriggerCmd.Flags().StringSliceVar(&asmAssetIDs, "asset-id", nil, "filter by asset ID(s)")
	// Without an asset id the API scans every asset. The CLI keeps that as an
	// explicit operator action. The MCP tool must name the assets.
	markMCPRequired(asmTriggerCmd.Flags(), "asset-id")
}
