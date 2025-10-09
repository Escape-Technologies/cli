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

var customRulesGetContent bool

var customRulesCmd = &cobra.Command{
	Use:     "custom-rules",
	Aliases: []string{"cr", "custom-rule", "rules"},
	Short:   "Manage custom security testing rules",
	Long: `Manage Custom Security Rules - Extend Security Testing

Custom rules allow you to define organization-specific security checks beyond
the standard OWASP and industry tests. Create rules for:
  • Business logic vulnerabilities
  • Company-specific security policies
  • Custom compliance requirements
  • API-specific security patterns

RULE CONTEXTS:
  • ASM (API Security Management)  - Attack surface monitoring
  • DAST (Dynamic Testing)         - Active security testing

RULE SEVERITY:
  CRITICAL, HIGH, MEDIUM, LOW, INFO`,
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
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("custom rule ID is required")
		}
		return nil
	},
	Example: `escape-cli custom-rules get 00000000-0000-0000-0000-000000000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		customRule, err := escape.GetCustomRule(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get custom rule: %w", err)
		}
		if customRulesGetContent {
			b, _ := json.Marshal(customRule.GetContent())
			out.Print(customRule.GetContent(), string(b))
			return nil
		}

		out.Table(customRule, func() []string {
			var tagNames []string
			for _, t := range customRule.GetTags() {
				tagNames = append(tagNames, out.TagText(t.GetName(), t.GetColor()))
			}
			tagsPretty := strings.Join(tagNames, ", ")
			return []string{
				"ID\tNAME\tCONTEXT\tSEVERITY\tASM ENABLED\tDAST ENABLED\tCREATED AT\tUPDATED AT\tTAGS",
				fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\t%s",
					customRule.GetId(),
					customRule.GetName(),
					customRule.GetContext(),
					customRule.GetSeverity(),
					customRule.GetAsmEnabled(),
					customRule.GetDastEnabled(),
					customRule.GetCreatedAt(),
					customRule.GetUpdatedAt(),
					tagsPretty,
				),
			}
		})
		return nil
	},
}

var customRulesDeleteCmd = &cobra.Command{
	Use:     "delete custom-rule-id",
	Aliases: []string{"d"},
	Short:   "Delete a custom rule",
	Long:    `Delete a custom rule by its ID.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("custom rule ID is required")
		}
		return nil
	},
	Example: `escape-cli custom-rules delete 00000000-0000-0000-0000-000000000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		res, err := escape.DeleteCustomRule(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to delete custom rule: %w", err)
		}
		if res.GetDeleted() {
			out.Log(fmt.Sprintf("Custom rule %s successfully deleted", id))
			return nil
		}
		return fmt.Errorf("failed to delete custom rule %s", id)
	},
}

var customRulesCreateCmd = &cobra.Command{
	Use:     "create <custom-rule.json",
	Aliases: []string{"c"},
	Short:   "Create a custom rule",
	Long:    `Create a custom rule.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			_ = cmd.Help()
			return errors.New("this command does not accept any arguments, it reads from stdin")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		// validate JSON early
		var tmp map[string]interface{}
		if err := json.Unmarshal(b, &tmp); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		res, err := escape.CreateCustomRule(cmd.Context(), b)
		if err != nil {
			return fmt.Errorf("failed to create custom rule: %w", err)
		}
		out.Table(res, func() []string {
			var tagNames []string
			for _, t := range res.GetTags() {
				tagNames = append(tagNames, out.TagText(t.GetName(), t.GetColor()))
			}
			tagsPretty := strings.Join(tagNames, ", ")
			return []string{
				"ID\tNAME\tCONTEXT\tSEVERITY\tASM ENABLED\tDAST ENABLED\tCREATED AT\tUPDATED AT\tTAGS",
				fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\t%s",
					res.GetId(),
					res.GetName(),
					res.GetContext(),
					res.GetSeverity(),
					res.GetAsmEnabled(),
					res.GetDastEnabled(),
					res.GetCreatedAt(),
					res.GetUpdatedAt(),
					tagsPretty,
				),
			}
		})
		return nil
	},
}

var customRulesUpdateCmd = &cobra.Command{
	Use:     "update custom-rule-id <custom-rule.json",
	Aliases: []string{"u"},
	Short:   "Update a custom rule",
	Long:    `Update a custom rule by its ID.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("custom rule ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		// validate JSON early
		var tmp v3.UpdateCustomRuleRequest
		if err := json.Unmarshal(b, &tmp); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		res, err := escape.UpdateCustomRule(cmd.Context(), id, b)
		if err != nil {
			return fmt.Errorf("failed to update custom rule: %w", err)
		}
		out.Table(res, func() []string {
			var tagNames []string
			for _, t := range res.GetTags() {
				tagNames = append(tagNames, out.TagText(t.GetName(), t.GetColor()))
			}
			tagsPretty := strings.Join(tagNames, ", ")
			return []string{
				"ID\tNAME\tCONTEXT\tSEVERITY\tASM ENABLED\tDAST ENABLED\tCREATED AT\tUPDATED AT\tTAGS",
				fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\t%s",
					res.GetId(),
					res.GetName(),
					res.GetContext(),
					res.GetSeverity(),
					res.GetAsmEnabled(),
					res.GetDastEnabled(),
					res.GetCreatedAt(),
					res.GetUpdatedAt(),
					tagsPretty,
				),
			}
		})
		return nil
	},
}

func init() {
	customRulesCmd.AddCommand(customRulesListCmd)
	customRulesCmd.AddCommand(customRulesGetCmd)
	customRulesGetCmd.Flags().BoolVarP(&customRulesGetContent, "content", "c", false, "print only the content JSON")
	customRulesCmd.AddCommand(customRulesCreateCmd)
	customRulesCmd.AddCommand(customRulesUpdateCmd)
	customRulesCmd.AddCommand(customRulesDeleteCmd)
	rootCmd.AddCommand(customRulesCmd)
}
