package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:     "users",
	Aliases: []string{"user", "us"},
	Short:   "Manage organization users",
	Long: `Manage Organization Users - Invite, View, and Administer Members

Manage the users in your Escape organization. List existing members,
view individual profiles, or invite new users by email.

COMMON WORKFLOWS:
  • View current user context:
    $ escape-cli users me

  • List all users:
    $ escape-cli users list

  • Invite new members:
    $ escape-cli users invite --email alice@example.com`,
}

func runUsersMe(cmd *cobra.Command) error {
	if out.Schema(v3.GetMe200Response{}) {
		return nil
	}

	user, err := escape.GetMe(cmd.Context())
	if err != nil {
		return fmt.Errorf("unable to get current user: %w", err)
	}

	u := user.GetUser()
	org := user.GetOrganization()
	name := stringValue(u.AdditionalProperties["name"])
	out.Table(user, func() []string {
		res := []string{"ID\tNAME\tEMAIL\tORG ID\tORG NAME"}
		res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", u.GetId(), name, u.GetEmail(), org.GetId(), org.GetName()))
		return res
	})
	return nil
}

var usersMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current authenticated user",
	Long: `Get Current User - View Your Profile

Display information about the currently authenticated user,
including ID, name, email, and organization details.`,
	Example: `  # Get current user info
  escape-cli users me

  # Get in JSON format
  escape-cli users me -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runUsersMe(cmd)
	},
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current authenticated user",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runUsersMe(cmd)
	},
}

var usersListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all organization users",
	Long: `List Users - View All Organization Members

Display all users in your organization including their IDs,
email addresses, and role information.`,
	Example: `  # List all users
  escape-cli users list

  # Export to JSON
  escape-cli users list -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.ListUsers200ResponseInner{}) {
			return nil
		}

		users, err := escape.ListUsers(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to list users: %w", err)
		}

		out.Table(users, func() []string {
			res := []string{"ID\tEMAIL\tNAME\tROLE\tCREATED AT"}
			for _, u := range users {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", u.GetId(), u.GetEmail(), stringValue(u.AdditionalProperties["name"]), strings.Join(roleIDs(u.GetRoleBindings()), ","), out.GetShortDate(u.GetCreatedAt().String())))
			}
			return res
		})
		return nil
	},
}

var usersGetCmd = &cobra.Command{
	Use:     "get user-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get a specific user by ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("user ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetUser200Response{}) {
			return nil
		}

		user, err := escape.GetUser(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get user: %w", err)
		}

		out.Table(user, func() []string {
			res := []string{"ID\tEMAIL\tNAME\tROLES\tCREATED AT"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", user.GetId(), user.GetEmail(), stringValue(user.AdditionalProperties["name"]), strings.Join(roleNames(user.GetRoleBindings()), ","), out.GetShortDate(user.GetCreatedAt().String())))
			return res
		})
		return nil
	},
}

var usersInviteEmails []string

func roleIDs(bindings []v3.ListProjects200ResponseDataInnerBindingsInner) []string {
	ids := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if binding.GetRoleId() != "" {
			ids = append(ids, binding.GetRoleId())
		}
	}
	return ids
}

func roleNames(bindings []v3.CreateProject200ResponseBindingsInner) []string {
	names := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		role := binding.GetRole()
		if role.GetName() != "" {
			names = append(names, role.GetName())
		}
	}
	return names
}

var usersInviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite users to the organization",
	Long: `Invite Users - Send Organization Invitations

Invite one or more users to join your Escape organization by email.
They will receive an invitation to set up their account.`,
	Example: `  # Invite a single user
  escape-cli users invite --email alice@example.com

  # Invite multiple users
  escape-cli users invite --email alice@example.com --email bob@example.com`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if len(usersInviteEmails) == 0 {
			return errors.New("at least one --email is required")
		}

		if out.Schema([]v3.ListUsers200ResponseInner{}) {
			return nil
		}

		users, err := escape.InviteUsers(cmd.Context(), usersInviteEmails)
		if err != nil {
			return fmt.Errorf("unable to invite users: %w", err)
		}

		out.Table(users, func() []string {
			res := []string{"ID\tEMAIL"}
			for _, u := range users {
				res = append(res, fmt.Sprintf("%s\t%s", u.GetId(), u.GetEmail()))
			}
			return res
		})
		return nil
	},
}

func init() {
	usersCmd.AddCommand(usersMeCmd, usersListCmd, usersGetCmd, usersInviteCmd)
	usersInviteCmd.Flags().StringArrayVar(&usersInviteEmails, "email", []string{}, "email address to invite (can be specified multiple times)")
	rootCmd.AddCommand(meCmd)
	rootCmd.AddCommand(usersCmd)
}
