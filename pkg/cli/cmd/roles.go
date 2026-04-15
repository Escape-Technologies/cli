package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var rolesCmd = &cobra.Command{
	Use:     "roles",
	Aliases: []string{"role"},
	Short:   "Manage organization roles and access bindings",
	Long: `Manage Roles - Control Organization Access

Roles define what actions users can perform in your organization.
Create custom roles, assign them to users, and manage access bindings.

COMMON WORKFLOWS:
  • List all roles:        escape-cli roles list
  • Create a role:         escape-cli roles create < role.json
  • Bind role to user:     escape-cli roles bind --role-id <id> --user-id <id>`,
}

var rolesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all roles",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.ListRoles200ResponseInner{}) {
			return nil
		}

		roles, err := escape.ListRoles(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to list roles: %w", err)
		}

		out.Table(roles, func() []string {
			res := []string{"ID\tNAME\tCREATED AT"}
			for _, r := range roles {
				res = append(res, fmt.Sprintf("%s\t%s\t%s", r.GetId(), r.GetName(), out.GetShortDate(r.GetCreatedAt().String())))
			}
			return res
		})
		return nil
	},
}

var rolesGetCmd = &cobra.Command{
	Use:     "get role-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get a role by ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("role ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateRole200Response{}) {
			return nil
		}

		role, err := escape.GetRole(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get role: %w", err)
		}

		out.Table(role, func() []string {
			res := []string{"ID\tNAME\tCREATED AT"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s", role.GetId(), role.GetName(), role.GetCreatedAt()))
			return res
		})
		return nil
	},
}

var rolesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new role from JSON stdin",
	Long: `Create Role - Define a New Access Role

Create a new role by providing its definition as JSON via stdin.
Use --input-schema to see the expected JSON format.`,
	Example: `  # Create from file
  escape-cli roles create < role.json

  # Create inline
  echo '{"name":"Read Only","permissions":["issues:read"]}' | escape-cli roles create

  # See input schema
  escape-cli roles create --input-schema`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.CreateRoleRequest{}) {
			return nil
		}
		if out.Schema(v3.CreateRole200Response{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		role, err := escape.CreateRole(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}

		out.Table(role, func() []string {
			return []string{
				"ID\tNAME\tCREATED AT",
				fmt.Sprintf("%s\t%s\t%s", role.GetId(), role.GetName(), role.GetCreatedAt()),
			}
		})
		return nil
	},
}

var rolesUpdateCmd = &cobra.Command{
	Use:   "update role-id",
	Short: "Update a role from JSON stdin",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("role ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.UpdateRoleRequest{}) {
			return nil
		}
		if out.Schema(v3.CreateRole200Response{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		role, err := escape.UpdateRole(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}

		out.Table(role, func() []string {
			return []string{
				"ID\tNAME\tCREATED AT",
				fmt.Sprintf("%s\t%s\t%s", role.GetId(), role.GetName(), role.GetCreatedAt()),
			}
		})
		return nil
	},
}

var (
	rolesBindRoleID string
	rolesBindUserID string
)

var rolesBindCmd = &cobra.Command{
	Use:   "bind",
	Short: "Bind a role to a user",
	Example: `  # Bind role to user
  escape-cli roles bind --role-id <role-id> --user-id <user-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if rolesBindRoleID == "" || rolesBindUserID == "" {
			return errors.New("--role-id and --user-id are required")
		}

		bindings, err := escape.CreateRoleBindings(cmd.Context(), rolesBindRoleID, rolesBindUserID)
		if err != nil {
			return fmt.Errorf("failed to bind role: %w", err)
		}

		out.Table(bindings, func() []string {
			res := []string{"BINDING ID\tROLE ID\tUSER ID"}
			for _, b := range bindings {
				res = append(res, fmt.Sprintf("%s\t%s\t%s", b.GetId(), b.GetRoleId(), b.GetUserId()))
			}
			return res
		})
		return nil
	},
}

var rolesUnbindCmd = &cobra.Command{
	Use:   "unbind binding-id",
	Short: "Remove a role binding",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("binding ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := escape.DeleteRoleBinding(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to unbind role: %w", err)
		}
		out.Log(fmt.Sprintf("Role binding %s removed", args[0]))
		return nil
	},
}

var rolesDeleteCmd = &cobra.Command{
	Use:     "delete role-id",
	Aliases: []string{"del", "remove"},
	Short:   "Delete a role",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("role ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := escape.DeleteRole(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}
		out.Log(fmt.Sprintf("Role %s deleted", args[0]))
		return nil
	},
}

func init() {
	rolesCmd.AddCommand(rolesListCmd, rolesGetCmd, rolesCreateCmd, rolesUpdateCmd, rolesDeleteCmd, rolesBindCmd, rolesUnbindCmd)
	rolesBindCmd.Flags().StringVar(&rolesBindRoleID, "role-id", "", "role ID to bind")
	rolesBindCmd.Flags().StringVar(&rolesBindUserID, "user-id", "", "user ID to bind the role to")
	rootCmd.AddCommand(rolesCmd)
}
