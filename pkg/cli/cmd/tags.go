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
	Short:   "Interact with tags",
	Long:    "Interact with your escape tags",
}

var tagsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tags",
	Long: `List tags.

Example output:
ID                                      NAME    COLOR
00000000-0000-0000-0000-000000000001    test    #000000`,
	Example: `escape-cli tags list`, RunE: func(cmd *cobra.Command, _ []string) error {
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
	Aliases: []string{"cr"},
	Args:    cobra.NoArgs,
	Short:   "Create a tag",
	Long:    "Create a tag",
	Example: `escape-cli tags create --name test --color e03d3d`,
	RunE: func(cmd *cobra.Command, _ []string) error {
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

func init() {
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsCreateCmd)

	tagsCreateCmd.Flags().StringP("name", "n", "", "Name of the tag")
	tagsCreateCmd.Flags().StringP("color", "c", "", "Color of the tag")
	_ = tagsCreateCmd.MarkFlagRequired("name")
	_ = tagsCreateCmd.MarkFlagRequired("color")

	rootCmd.AddCommand(tagsCmd)
}
