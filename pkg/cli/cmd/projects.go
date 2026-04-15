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

var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project", "proj"},
	Short:   "Manage organization projects",
	Long: `Manage Projects - Organize Your Security Work

Projects let you group profiles, assets, and scans for teams or business units.
Use them to scope access control with role bindings.

COMMON WORKFLOWS:
  • List all projects:     escape-cli projects list
  • Create a project:      escape-cli projects create < project.json
  • Get project details:   escape-cli projects get <project-id>`,
}

var projectsSearchFlag string

var projectsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all projects",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.ListProjects200ResponseDataInner{}) {
			return nil
		}

		filters := &escape.ListProjectsFilters{Search: projectsSearchFlag}
		projects, next, err := escape.ListProjects(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("unable to list projects: %w", err)
		}
		all := projects
		for next != nil && *next != "" {
			projects, next, err = escape.ListProjects(cmd.Context(), *next, filters)
			if err != nil {
				return fmt.Errorf("unable to list projects: %w", err)
			}
			all = append(all, projects...)
		}

		out.Table(all, func() []string {
			res := []string{"ID\tNAME\tCREATED AT"}
			for _, p := range all {
				res = append(res, fmt.Sprintf("%s\t%s\t%s", p.GetId(), p.GetName(), out.GetShortDate(p.GetCreatedAt().String())))
			}
			return res
		})
		return nil
	},
}

var projectsGetCmd = &cobra.Command{
	Use:     "get project-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get a project by ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("project ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateProject200Response{}) {
			return nil
		}

		project, err := escape.GetProject(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get project: %w", err)
		}

		out.Table(project, func() []string {
			res := []string{"ID\tNAME\tCREATED AT"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s", project.GetId(), project.GetName(), project.GetCreatedAt()))
			return res
		})
		return nil
	},
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project from JSON stdin",
	Example: `  # Create a project
  echo '{"name":"My Project"}' | escape-cli projects create

  # See input schema
  escape-cli projects create --input-schema`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.CreateProjectRequest{}) {
			return nil
		}
		if out.Schema(v3.CreateProject200Response{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		project, err := escape.CreateProject(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		out.Table(project, func() []string {
			return []string{
				"ID\tNAME\tCREATED AT",
				fmt.Sprintf("%s\t%s\t%s", project.GetId(), project.GetName(), project.GetCreatedAt()),
			}
		})
		return nil
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update project-id",
	Short: "Update a project from JSON stdin",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("project ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.UpdateProjectRequest{}) {
			return nil
		}
		if out.Schema(v3.CreateProject200Response{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		project, err := escape.UpdateProject(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}

		out.Table(project, func() []string {
			return []string{
				"ID\tNAME\tCREATED AT",
				fmt.Sprintf("%s\t%s\t%s", project.GetId(), project.GetName(), project.GetCreatedAt()),
			}
		})
		return nil
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:     "delete project-id",
	Aliases: []string{"del", "remove"},
	Short:   "Delete a project",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("project ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := escape.DeleteProject(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}
		out.Log(fmt.Sprintf("Project %s deleted", args[0]))
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsListCmd, projectsGetCmd, projectsCreateCmd, projectsUpdateCmd, projectsDeleteCmd)
	projectsListCmd.Flags().StringVarP(&projectsSearchFlag, "search", "s", "", "search projects by name")
	rootCmd.AddCommand(projectsCmd)
}
