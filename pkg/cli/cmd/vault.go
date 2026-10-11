package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var vaultVariableTypes = []string{"EXTERNAL", "PLAINTEXT", "SECRET"}

var vaultCmd = &cobra.Command{
	Use:     "vault",
	Aliases: []string{"vaults"},
	Short:   "Manage the variables vault (Beta)",
	Long: `Manage the Variables Vault - Secret References for Scan Configurations
(Beta)

The vault stores the secret values scan configurations reference as {{KEY}}.
A variable is either a plaintext value, a masked secret, or a pointer to an
entry of a provider vault (Azure Key Vault, AWS Secrets Manager, and the other
linked vaults).

BETA: the variables vault is gated by the VARIABLES_VAULT feature flag and has
not been announced, so this group can still change.

COMMON WORKFLOWS:
  • List the variables:   escape-cli vault variables list
  • Store a secret:       escape-cli vault variables create --key API_TOKEN --type SECRET --content <value>
  • Rotate a secret:      escape-cli vault variables rotate <variable-id> --content <new-value>
  • List linked vaults:   escape-cli vault list`,
}

var vaultVariablesCmd = &cobra.Command{
	Use:   "variables",
	Short: "Manage the vault variables (Beta)",
}

var (
	vaultVariablesSearch    string
	vaultVariablesListPage  pageFlags
	vaultVariablesCreateKey string
	vaultVariableType       string
	vaultVariableContent    string
	vaultVariableVaultID    string
)

var vaultVariablesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the vault variables (Beta)",
	Long: `List the vault variables of the organization.

Values are write-only: the API never returns the content of a variable, only
its key and type.`,
	Example: `  escape-cli vault variables list
  escape-cli vault variables list --search API`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.VaultVariableSummary{}) {
			return nil
		}

		filters := &escape.ListVaultVariablesFilters{Search: vaultVariablesSearch}
		if err := runPagedList(cmd, vaultVariablesListPage, func(ctx context.Context, cursor string, size int) ([]v3.VaultVariableSummary, *string, int, error) {
			return escape.ListVaultVariables(ctx, cursor, size, filters)
		}, func(variables []v3.VaultVariableSummary) []string {
			res := []string{"ID\tKEY\tTYPE\tVAULT\tUPDATED AT"}
			for _, variable := range variables {
				vault := ""
				if linked := variable.GetVault(); linked.GetId() != "" {
					vault = linked.GetId()
				}

				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
					variable.GetId(),
					variable.GetKey(),
					variable.GetType(),
					vault,
					variable.GetUpdatedAt().String(),
				))
			}

			return res
		}); err != nil {
			return fmt.Errorf("unable to list vault variables: %w", err)
		}

		return nil
	},
}

var vaultVariablesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a vault variable (Beta)",
	Long: `Create a vault variable.

The content is the value the variable holds: a plaintext or secret value, or,
for an EXTERNAL variable, the name of the entry inside the linked vault
(--vault-id is then required). The content is never returned by the API
afterwards.`,
	Example: `  escape-cli vault variables create --key API_TOKEN --type SECRET --content tok_123
  escape-cli vault variables create --key DB_PASSWORD --type EXTERNAL --vault-id <vault-id> --content production/db`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.CreateVaultVariable200Response{}) {
			return nil
		}

		key := strings.TrimSpace(vaultVariablesCreateKey)
		if key == "" {
			return errors.New("--key is required")
		}

		if !v3.ENUMPROPERTIESDATAITEMSPROPERTIESTYPE(vaultVariableType).IsValid() {
			return fmt.Errorf("invalid --type %q; valid values: %s", vaultVariableType, strings.Join(vaultVariableTypes, ", "))
		}

		if vaultVariableContent == "" {
			return errors.New("--content is required: the value is write-only and cannot be read back")
		}

		if vaultVariableType == "EXTERNAL" && vaultVariableVaultID == "" {
			return errors.New("--vault-id is required for an EXTERNAL variable")
		}

		variable, err := escape.CreateVaultVariable(cmd.Context(), key, vaultVariableType, vaultVariableContent, vaultVariableVaultID)
		if err != nil {
			return fmt.Errorf("failed to create vault variable: %w", err)
		}

		printVaultVariable(variable)

		return nil
	},
}

var vaultVariablesDeleteCmd = &cobra.Command{
	Use:     "delete variable-id",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a vault variable (Beta)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.DeleteProfile200Response{}) {
			return nil
		}

		proceed, err := confirmDestructive(cmd, "delete vault variable "+args[0], nil)
		if err != nil || !proceed {
			return err
		}

		if _, err := escape.DeleteVaultVariable(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete vault variable: %w", err)
		}

		out.Log(fmt.Sprintf("Vault variable %s deleted", args[0]))

		return nil
	},
}

var vaultVariablesRotateCmd = &cobra.Command{
	Use:   "rotate variable-id",
	Short: "Rotate the value of a vault variable (Beta)",
	Long: `Rotate the value of a vault variable.

The new content replaces the stored one. Scan configurations keep
referencing the variable by its key, so a rotation needs no profile change.`,
	Example: `  escape-cli vault variables rotate <variable-id> --content tok_456`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateVaultVariable200Response{}) {
			return nil
		}

		if vaultVariableContent == "" {
			return errors.New("--content is required: the value is write-only and cannot be read back")
		}

		variable, err := escape.RotateVaultVariable(cmd.Context(), args[0], vaultVariableContent)
		if err != nil {
			return fmt.Errorf("failed to rotate vault variable: %w", err)
		}

		printVaultVariable(variable)

		return nil
	},
}

var (
	vaultListPage          pageFlags
	vaultLinkIntegrationID string
	vaultLinkExternalID    string
)

var vaultListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the linked vaults (Beta)",
	Long: `List the provider vaults linked to the organization, reachable through
their integration.`,
	Example: `  escape-cli vault list`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.VaultSummary1{}) {
			return nil
		}

		if err := runPagedList(cmd, vaultListPage, func(ctx context.Context, cursor string, size int) ([]v3.VaultSummary1, *string, int, error) {
			return escape.ListVaults(ctx, cursor, size)
		}, func(vaults []v3.VaultSummary1) []string {
			res := []string{"ID\tNAME\tTYPE\tEXTERNAL ID\tINTEGRATION ID"}
			for _, vault := range vaults {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
					vault.GetId(),
					vault.GetName(),
					vault.GetType(),
					vault.GetExternalId(),
					vault.GetIntegrationId(),
				))
			}

			return res
		}); err != nil {
			return fmt.Errorf("unable to list vaults: %w", err)
		}

		return nil
	},
}

var vaultLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link a provider vault (Beta)",
	Long: `Link a vault of a provider integration (Azure Key Vault, AWS Secrets
Manager, and the others) to the organization, so EXTERNAL variables can
reference its entries.`,
	Example: `  escape-cli vault link --integration-id <integration-id> --external-id <vault-id-at-the-provider>`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.LinkVault200Response{}) {
			return nil
		}

		if vaultLinkIntegrationID == "" || vaultLinkExternalID == "" {
			return errors.New("--integration-id and --external-id are required")
		}

		vault, err := escape.LinkVault(cmd.Context(), vaultLinkIntegrationID, vaultLinkExternalID)
		if err != nil {
			return fmt.Errorf("failed to link vault: %w", err)
		}

		out.Table(vault, func() []string {
			return []string{
				"ID\tNAME\tTYPE\tEXTERNAL ID\tINTEGRATION ID",
				fmt.Sprintf("%s\t%s\t%s\t%s\t%s", vault.GetId(), vault.GetName(), vault.GetType(), vault.GetExternalId(), vault.GetIntegrationId()),
			}
		})

		return nil
	},
}

var vaultEntriesCmd = &cobra.Command{
	Use:     "entries vault-id",
	Aliases: []string{"entry"},
	Short:   "List the entries of a linked vault (Beta)",
	Long: `List the entries of a linked vault, as the provider names them. An
EXTERNAL variable references one of these names as its content.`,
	Example: `  escape-cli vault entries <vault-id>`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]v3.VaultEntry{}) {
			return nil
		}

		entries, err := escape.ListVaultEntries(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to list vault entries: %w", err)
		}

		out.Table(entries, func() []string {
			res := []string{"NAME"}
			for _, entry := range entries {
				res = append(res, entry.GetName())
			}

			return res
		})

		return nil
	},
}

func printVaultVariable(variable *v3.CreateVaultVariable200Response) {
	vault := ""
	if linked := variable.GetVault(); linked.GetId() != "" {
		vault = linked.GetId()
	}

	out.Table(variable, func() []string {
		return []string{
			"ID\tKEY\tTYPE\tVAULT",
			fmt.Sprintf("%s\t%s\t%s\t%s", variable.GetId(), variable.GetKey(), variable.GetType(), vault),
		}
	})
}

func init() {
	vaultVariablesCmd.AddCommand(vaultVariablesListCmd, vaultVariablesCreateCmd, vaultVariablesDeleteCmd, vaultVariablesRotateCmd)
	vaultCmd.AddCommand(vaultVariablesCmd)
	vaultCmd.AddCommand(vaultListCmd, vaultLinkCmd, vaultEntriesCmd)

	vaultVariablesListPage.bind(vaultVariablesListCmd)
	vaultListPage.bind(vaultListCmd)
	vaultVariablesListCmd.Flags().StringVar(&vaultVariablesSearch, "search", "", "search vault variables by key")

	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariablesCreateKey, "key", "", "variable key, referenced from scan configs as {{KEY}}")
	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariableType, "type", "", "variable kind: "+strings.Join(vaultVariableTypes, ", "))
	if err := vaultVariablesCreateCmd.Flags().SetAnnotation("type", flagEnumAnnotation, vaultVariableTypes); err != nil {
		panic(err)
	}

	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariableContent, "content", "", "value of the variable (write-only: the API never returns it)")
	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariableVaultID, "vault-id", "", "linked vault the variable points into (required for EXTERNAL)")
	_ = vaultVariablesCreateCmd.MarkFlagRequired("key")
	_ = vaultVariablesCreateCmd.MarkFlagRequired("type")

	vaultVariablesRotateCmd.Flags().StringVar(&vaultVariableContent, "content", "", "new value of the variable (write-only: the API never returns it)")

	vaultLinkCmd.Flags().StringVar(&vaultLinkIntegrationID, "integration-id", "", "integration the vault is reachable through")
	vaultLinkCmd.Flags().StringVar(&vaultLinkExternalID, "external-id", "", "the vault's own coordinate at the provider")
	_ = vaultLinkCmd.MarkFlagRequired("integration-id")
	_ = vaultLinkCmd.MarkFlagRequired("external-id")

	bindDestructiveGate(vaultVariablesDeleteCmd, true)

	rootCmd.AddCommand(vaultCmd)
}
