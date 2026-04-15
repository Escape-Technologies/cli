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

var tagsCmd = &cobra.Command{
	Use:     "tags",
	Aliases: []string{"tag"},
	Short:   "Manage tags for organizing assets and profiles",
	Long: `Manage Tags - Organize Your Security Resources

Tags help you organize and filter assets, profiles, issues, and other resources.
Create custom tags with colors for visual organization in the platform.

COMMON USE CASES:
  • Environment labels (production, staging, development)
  • Team ownership (frontend-team, backend-team, security-team)
  • Criticality (critical, high-priority, low-priority)
  • Compliance (pci-dss, hipaa, gdpr)
  • Project grouping (project-alpha, project-beta)`,
}

var tagsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all organization tags",
	Long: `List Tags - View All Available Tags

Display all tags in your organization with their IDs, names, and color codes.
Use these tag IDs when filtering or updating assets, profiles, and issues.`,
	Example: `  # List all tags
  escape-cli tags list

  # Export tags to JSON
  escape-cli tags list -o json`, RunE: func(cmd *cobra.Command, _ []string) error {
		// Output JSON Schema if requested
		if out.Schema([]v3.TagDetail{}) {
			return nil
		}

		tags, err := escape.ListTags(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to list tags: %w", err)
		}

		result := make([]*v3.TagDetail, 0, len(tags))
		fields := []string{"ID\tNAME\tCOLOR"}

		for _, tag := range tags {
			result = append(result, &tag)
		}
		out.Table(result, func() []string {
			for _, tag := range result {
				fields = append(fields, fmt.Sprintf("%s\t%s\t%s", tag.GetId(), tag.GetName(), tag.GetColor()))
			}
			return fields
		})

		return nil
	},
}

var tagsCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"cr", "add", "new"},
	Args:    cobra.NoArgs,
	Short:   "Create a new tag",
	Long: `Create Tag - Add New Organization Label

Create a new tag with a custom name and color. Use hex color codes without the # prefix.`,
	Example: `  # Create a production tag (red)
  escape-cli tags create --name production --color e03d3d

  # Create a staging tag (yellow)
  escape-cli tags create --name staging --color f5a623

  # Create a team tag (blue)
  escape-cli tags create --name backend-team --color 4a90e2`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Output JSON Schema if requested
		if out.Schema(v3.TagDetail{}) {
			return nil
		}

		name, _ := cmd.Flags().GetString("name")
		color, _ := cmd.Flags().GetString("color")
		if strings.TrimSpace(name) == "" || strings.TrimSpace(color) == "" {
			return errors.New("name and color are required: use --name and --color flags")
		}
		tag, err := escape.CreateTag(cmd.Context(), name, color)
		if err != nil {
			return fmt.Errorf("unable to create tag: %w", err)
		}
		out.Print(tag, "Tag created")
		return nil
	},
}

var tagsGetCmd = &cobra.Command{
	Use:     "get tag-id",
	Aliases: []string{"g", "show", "describe"},
	Args:    cobra.ExactArgs(1),
	Short:   "Get a tag by ID",
	Long:    `Get a tag by its ID.`,
	Example: `  # Get a tag
  escape-cli tags get 00000000-0000-0000-0000-000000000000

  # Get tag as JSON
  escape-cli tags get <tag-id> -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateTag200Response{}) {
			return nil
		}

		tag, err := escape.GetTag(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get tag: %w", err)
		}

		out.Table(tag, func() []string {
			return []string{
				"ID\tNAME\tCOLOR",
				fmt.Sprintf("%s\t%s\t%s", tag.GetId(), tag.GetName(), tag.GetColor()),
			}
		})
		return nil
	},
}

var (
	tagUpdateName  string
	tagUpdateColor string
)

var tagsUpdateCmd = &cobra.Command{
	Use:     "update tag-id",
	Aliases: []string{"u", "edit"},
	Args:    cobra.ExactArgs(1),
	Short:   "Update a tag name or color",
	Example: `  # Rename a tag
  escape-cli tags update <tag-id> --name new-name

  # Change color
  escape-cli tags update <tag-id> --color ff0000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateTag200Response{}) {
			return nil
		}

		var name, color *string
		if cmd.Flags().Changed("name") {
			name = &tagUpdateName
		}
		if cmd.Flags().Changed("color") {
			color = &tagUpdateColor
		}
		if name == nil && color == nil {
			return errors.New("at least one of --name or --color is required")
		}

		tag, err := escape.UpdateTag(cmd.Context(), args[0], name, color)
		if err != nil {
			return fmt.Errorf("unable to update tag: %w", err)
		}
		out.Table(tag, func() []string {
			return []string{
				"ID\tNAME\tCOLOR",
				fmt.Sprintf("%s\t%s\t%s", tag.GetId(), tag.GetName(), tag.GetColor()),
			}
		})
		return nil
	},
}

var tagsDeleteCmd = &cobra.Command{
	Use:     "delete tag-id",
	Aliases: []string{"del", "rm", "remove"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete a tag",
	Long: `Delete Tag - Remove tag from organization

Permanently delete a tag from your organization`,
	Example: `  # Delete a tag
  escape-cli tags delete 00000000-0000-0000-0000-000000000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		err := escape.DeleteTag(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("unable to delete tag: %w", err)
		}
		out.Log("Tag deleted")
		return nil
	},
}

func init() {
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsGetCmd)
	tagsCmd.AddCommand(tagsCreateCmd)
	tagsCmd.AddCommand(tagsUpdateCmd)
	tagsCmd.AddCommand(tagsDeleteCmd)

	tagsCreateCmd.Flags().StringP("name", "n", "", "Name of the tag")
	tagsCreateCmd.Flags().StringP("color", "c", "", "Color of the tag")
	_ = tagsCreateCmd.MarkFlagRequired("name")
	_ = tagsCreateCmd.MarkFlagRequired("color")

	tagsUpdateCmd.Flags().StringVarP(&tagUpdateName, "name", "n", "", "new tag name")
	tagsUpdateCmd.Flags().StringVarP(&tagUpdateColor, "color", "c", "", "new tag color (hex without #)")

	rootCmd.AddCommand(tagsCmd)
}
