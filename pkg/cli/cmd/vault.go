package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var vaultVariableTypes = []string{"EXTERNAL", "PLAINTEXT", "SECRET"}

var vaultCmd = &cobra.Command{
	Use:     "vault",
	Aliases: []string{"vaults"},
	Short:   "Manage the variables vault",
	Long: `Manage the Variables Vault - Secret References for Scan Configurations

The vault stores the secret values scan configurations reference as ${NAME}.
A variable is either a plaintext value, a masked secret, or a pointer to an
entry of a provider vault.

COMMON WORKFLOWS:
  • List the variables:   escape-cli vault variables list
  • Store a secret:       escape-cli vault variables create --name API_TOKEN --type SECRET --value <value>
  • Point at a vault:     escape-cli vault variables create --name DB_PASSWORD --type EXTERNAL --external-id <vault-uri> --external-entry <entry>
  • Replace a value:      escape-cli vault variables update <variable-id> --value <new-value>`,
}

var vaultVariablesCmd = &cobra.Command{
	Use:   "variables",
	Short: "Manage the vault variables",
}

var (
	vaultVariablesSearch       string
	vaultVariablesListPage     pageFlags
	vaultVariablesCreateName   string
	vaultVariableType          string
	vaultVariableValue         string
	vaultVariableExternalID    string
	vaultVariableExternalEntry string
	vaultVariableUpdateValue   string
	vaultVariableUpdateEntry   string
)

var vaultVariablesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the vault variables",
	Long: `List the vault variables of the organization.

Values are write-only: the API never returns the value of a variable, only
its name and type.`,
	Example: `  escape-cli vault variables list
  escape-cli vault variables list --search API`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]escape.VariableContent{}) {
			return nil
		}

		filters := &escape.ListVaultVariablesFilters{Search: vaultVariablesSearch}
		if err := runPagedList(cmd, vaultVariablesListPage, func(ctx context.Context, cursor string, size int) ([]escape.VariableContent, *string, int, error) {
			return escape.ListVaultVariables(ctx, cursor, size, filters)
		}, func(variables []escape.VariableContent) []string {
			res := []string{"ID\tNAME\tTYPE"}
			for _, variable := range variables {
				res = append(res, fmt.Sprintf("%s\t%s\t%s", variable.ID, variable.Name, variable.Type))
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
	Short: "Create a vault variable",
	Long: `Create a vault variable.

PLAINTEXT and SECRET take --value. EXTERNAL takes --external-id (the vault's
own coordinate at the provider) and --external-entry (the name of the entry
inside that vault). Escape looks the vault and the entry up before it stores
the variable. The value is never returned by the API afterwards.`,
	Example: `  escape-cli vault variables create --name API_TOKEN --type SECRET --value tok_123
  escape-cli vault variables create --name DB_PASSWORD --type EXTERNAL --external-id https://my-vault.vault.azure.net/ --external-entry production-db`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(escape.VariableContent{}) {
			return nil
		}

		name := strings.TrimSpace(vaultVariablesCreateName)
		if name == "" {
			return errors.New("--name is required")
		}

		if !validVaultVariableType(vaultVariableType) {
			return fmt.Errorf("invalid --type %q; valid values: %s", vaultVariableType, strings.Join(vaultVariableTypes, ", "))
		}

		variable, err := createVaultVariable(cmd.Context(), name)
		if err != nil {
			return err
		}

		printVaultVariable(variable)

		return nil
	},
}

var vaultVariablesUpdateCmd = &cobra.Command{
	Use:   "update variable-id",
	Short: "Update a vault variable",
	Long: `Replace the value of a plaintext or secret variable, or the entry
name of an external variable.

Pass --value or --external-entry, not both. The name and the type stay as
they are. The new value is write-only.`,
	Example: `  escape-cli vault variables update <variable-id> --value tok_456
  escape-cli vault variables update <variable-id> --external-entry production-db`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(escape.VariableContent{}) {
			return nil
		}

		valueSet := cmd.Flags().Changed("value")
		entrySet := cmd.Flags().Changed("external-entry")
		if valueSet == entrySet {
			return errors.New("pass either --value or --external-entry")
		}

		variable, err := escape.UpdateVaultVariable(cmd.Context(), args[0], vaultVariableUpdateValue, vaultVariableUpdateEntry, entrySet)
		if err != nil {
			return fmt.Errorf("failed to update vault variable: %w", err)
		}

		printVaultVariable(variable)

		return nil
	},
}

var vaultVariablesDeleteCmd = &cobra.Command{
	Use:     "delete variable-id",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a vault variable",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(out.Message{}) {
			return nil
		}

		proceed, err := confirmDestructive(cmd, "delete vault variable "+args[0], nil)
		if err != nil || !proceed {
			return err
		}

		if err := escape.DeleteVaultVariable(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete vault variable: %w", err)
		}

		out.Log(fmt.Sprintf("Vault variable %s deleted", args[0]))

		return nil
	},
}

func validVaultVariableType(value string) bool {
	for _, candidate := range vaultVariableTypes {
		if value == candidate {
			return true
		}
	}

	return false
}

func createVaultVariable(ctx context.Context, name string) (*escape.VariableContent, error) {
	if vaultVariableType == "EXTERNAL" {
		if vaultVariableValue != "" {
			return nil, errors.New("--value is not used for an EXTERNAL variable; pass --external-id and --external-entry")
		}

		if strings.TrimSpace(vaultVariableExternalID) == "" || strings.TrimSpace(vaultVariableExternalEntry) == "" {
			return nil, errors.New("--external-id and --external-entry are required for an EXTERNAL variable")
		}

		variable, err := escape.CreateExternalVaultVariable(ctx, name, vaultVariableExternalID, vaultVariableExternalEntry)
		if err != nil {
			return nil, fmt.Errorf("failed to create vault variable: %w", err)
		}

		return variable, nil
	}

	if vaultVariableExternalID != "" || vaultVariableExternalEntry != "" {
		return nil, errors.New("--external-id and --external-entry are only used for an EXTERNAL variable")
	}

	if vaultVariableValue == "" {
		return nil, errors.New("--value is required: the value is write-only and cannot be read back")
	}

	var (
		variable *escape.VariableContent
		err      error
	)
	if vaultVariableType == "SECRET" {
		variable, err = escape.CreateSecretVaultVariable(ctx, name, vaultVariableValue)
	} else {
		variable, err = escape.CreatePlaintextVaultVariable(ctx, name, vaultVariableValue)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create vault variable: %w", err)
	}

	return variable, nil
}

func printVaultVariable(variable *escape.VariableContent) {
	out.Table(variable, func() []string {
		return []string{
			"ID\tNAME\tTYPE",
			fmt.Sprintf("%s\t%s\t%s", variable.ID, variable.Name, variable.Type),
		}
	})
}

func init() {
	vaultVariablesCmd.AddCommand(vaultVariablesListCmd, vaultVariablesCreateCmd, vaultVariablesUpdateCmd, vaultVariablesDeleteCmd)
	vaultCmd.AddCommand(vaultVariablesCmd)

	vaultVariablesListPage.bind(vaultVariablesListCmd)
	vaultVariablesListCmd.Flags().StringVar(&vaultVariablesSearch, "search", "", "search vault variables by name")

	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariablesCreateName, "name", "", "variable name, referenced from scan configs as ${NAME}")
	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariableType, "type", "", "variable kind: "+strings.Join(vaultVariableTypes, ", "))
	if err := vaultVariablesCreateCmd.Flags().SetAnnotation("type", flagEnumAnnotation, vaultVariableTypes); err != nil {
		panic(err)
	}

	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariableValue, "value", "", "value of a plaintext or secret variable (write-only: the API never returns it)")
	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariableExternalID, "external-id", "", "the provider vault's own coordinate (required for EXTERNAL)")
	vaultVariablesCreateCmd.Flags().StringVar(&vaultVariableExternalEntry, "external-entry", "", "entry name inside the provider vault (required for EXTERNAL)")
	_ = vaultVariablesCreateCmd.MarkFlagRequired("name")
	_ = vaultVariablesCreateCmd.MarkFlagRequired("type")

	vaultVariablesUpdateCmd.Flags().StringVar(&vaultVariableUpdateValue, "value", "", "new value of a plaintext or secret variable (write-only)")
	vaultVariablesUpdateCmd.Flags().StringVar(&vaultVariableUpdateEntry, "external-entry", "", "new entry name of an external variable")

	bindDestructiveGate(vaultVariablesDeleteCmd, true)

	rootCmd.AddCommand(vaultCmd)
}
