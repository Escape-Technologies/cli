package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var customRulesCmd = &cobra.Command{
	Use:     "custom-rules",
	Aliases: []string{"cr", "custom-rule"},
	Short:   "Interact with custom rules",
	Long:    "Interact with your escape custom rules",
}

var customRulesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List custom rules",
	Long: `List all custom rules.

Example output:
ID                                      NAME                       SSH PUBLIC KEY
00000000-0000-0000-0000-000000000001    example-custom-rule-1         ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... example1@email.com
00000000-0000-0000-0000-000000000002    example-custom-rule-2         ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... example2@email.com`,
	Example: `escape-cli custom-rules list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		customRules, err := escape.ListCustomRules(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list custom rules: %w", err)
		}
		out.Table(customRules, func() []string {
			res := []string{"ID\tNAME\tSEVERITY\tCREATED AT\tUPDATED AT"}
			for _, customRule := range customRules {
				res = append(
					res,
					fmt.Sprintf(
						"%s\t%s\t%s\t%s\t%s",
						customRule.GetId(),
						customRule.GetName(),
						customRule.GetSeverity(),
						customRule.GetCreatedAt(),
						customRule.GetUpdatedAt(),
					),
				)
			}
			return res
		})
		return nil
	},
}

var customRulesGetCmd = &cobra.Command{
	Use:     "get custom-rule-id",
	Aliases: []string{"g"},
	Short:   "Get a custom rule",
	Long:    `Get a custom rule by its ID.`,
	Args:    cobra.ExactArgs(1),
	Example: `escape-cli custom-rules get 00000000-0000-0000-0000-000000000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		customRule, err := escape.GetCustomRule(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get custom rule: %w", err)
		}
		out.Table(customRule.GetRule(), func() []string {
			return []string{
				"RULE",
				fmt.Sprintf("%v", customRule.GetRule()),
			}
		})
		return nil
	},
}

func init() {
	customRulesCmd.AddCommand(customRulesListCmd)
	customRulesCmd.AddCommand(customRulesGetCmd)
	rootCmd.AddCommand(customRulesCmd)
}
