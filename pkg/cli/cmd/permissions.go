package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var permissionsCmd = &cobra.Command{
	Use:     "permissions",
	Aliases: []string{"permission"},
	Short:   "Read the permission catalog a role can be granted",
	Long: `Permissions - The Building Blocks of a Role

List every permission a custom role can carry, with what each one grants.
The permission ids are the accepted values of "escape-cli roles create" and
"escape-cli roles update", so this is the place to look before writing a role
definition.`,
}

var permissionsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List every permission that can be granted to a role",
	Long: `List Permissions - Read the Available Access Controls

List every permission a role can be granted, with its id, name and what it
grants. The permission ids are the accepted values of "escape-cli roles
create" and "escape-cli roles update".`,
	Example: `  escape-cli permissions list
  escape-cli permissions list -o json | jq '.[].id'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.PermissionCatalogEntry{}) {
			return nil
		}

		permissions, err := escape.ListPermissions(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to list permissions: %w", err)
		}

		out.Table(permissions, func() []string {
			rows := []string{"ID\tNAME\tDESCRIPTION"}
			for _, permission := range permissions {
				rows = append(rows, fmt.Sprintf("%s\t%s\t%s",
					permission.GetId(),
					permission.GetName(),
					permission.GetDescription(),
				))
			}

			return rows
		})

		return nil
	},
}

func init() {
	permissionsCmd.AddCommand(permissionsListCmd)
	rootCmd.AddCommand(permissionsCmd)
}
