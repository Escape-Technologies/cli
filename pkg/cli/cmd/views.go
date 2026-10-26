package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var viewTargets = []string{"ASSETS", "EVENTS", "INTEGRATIONS", "ISSUES", "LOCATIONS", "PROFILES", "SCANS", "TARGETS"}

var (
	viewsTarget    string
	viewsProjectID string
	viewsListPage  pageFlags

	viewsSortType      string
	viewsSortDirection string
)

var viewsCmd = &cobra.Command{
	Use:     "views",
	Aliases: []string{"view"},
	Short:   "Manage saved views",
	Long: `Manage Saved Views - Reuse a Filter, Widget and Column Layout

A saved view is a named, reusable query over one resource: the filters that
select it, the widgets and columns that display it, and how it is grouped and
sorted. Create one once and every later listing of that resource reuses it.

COMMON WORKFLOWS:
  • List the views of a target:   escape-cli views list --target ISSUES
  • Read the issues a view picks: escape-cli views issues <view-id>
  • Reorder the sidebar:          escape-cli views reorder < reorder.json`,
}

var viewsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the saved views of a target",
	Long: `List the saved views of one resource.

Escape ships managed views (their key field is set). Editing or deleting one
through the API is rejected: the platform owns them.`,
	Example: `  escape-cli views list --target ISSUES
  escape-cli views list --target ASSETS --project-id 00000000-0000-0000-0000-000000000000`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.ViewDetail{}) {
			return nil
		}

		if viewsTarget == "" {
			return errors.New("--target is required")
		}

		views, err := escape.ListViews(cmd.Context(), viewsTarget, viewsProjectID)
		if err != nil {
			return fmt.Errorf("unable to list views: %w", err)
		}

		out.Table(views, func() []string {
			res := []string{"ID\tNAME\tKEY\tORDER"}
			for _, view := range views {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%d", view.GetId(), view.GetName(), view.GetKey(), view.GetOrder()))
			}

			return res
		})

		return nil
	},
}

var viewsIssuesCmd = &cobra.Command{
	Use:   "issues view-id",
	Short: "List the issues a saved view selects",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]v3.IssueSummarized{}) {
			return nil
		}

		filters := &escape.ListViewIssuesFilters{
			SortType:      viewsSortType,
			SortDirection: viewsSortDirection,
		}
		if err := runPagedList(cmd, viewsListPage, func(ctx context.Context, cursor string, size int) ([]v3.IssueSummarized, *string, int, error) {
			return escape.ListViewIssues(ctx, args[0], cursor, size, filters)
		}, func(issues []v3.IssueSummarized) []string {
			res := []string{"ID\tNAME\tSEVERITY\tSTATUS"}
			for _, issue := range issues {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s", issue.GetId(), issue.GetName(), issue.GetSeverity(), issue.GetStatus()))
			}

			return res
		}); err != nil {
			return fmt.Errorf("unable to list view issues: %w", err)
		}

		return nil
	},
}

var viewsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a saved view from JSON stdin",
	Example: `  # Create from a file
  escape-cli views create < view.json

  # See the input schema
  escape-cli views create --input-schema`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.CreateViewRequest{}) {
			return nil
		}

		if out.Schema(v3.CreateView200Response{}) {
			return nil
		}

		body, err := readViewBody(cmd)
		if err != nil {
			return err
		}

		view, err := escape.CreateView(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create view: %w", err)
		}

		printView(view.GetId(), view.GetName())

		return nil
	},
}

var viewsUpdateCmd = &cobra.Command{
	Use:     "update view-id",
	Aliases: []string{"edit"},
	Short:   "Update a saved view from JSON stdin",
	Long: `Update the fields of a saved view. Only the fields present in the
body change; pass an empty array or null to clear one.

Escape-managed views (a non-empty key) cannot be updated.`,
	Example: `  escape-cli views update 00000000-0000-0000-0000-000000000000 < view.json`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.UpdateViewRequest{}) {
			return nil
		}

		if out.Schema(v3.CreateView200Response{}) {
			return nil
		}

		body, err := readViewBody(cmd)
		if err != nil {
			return err
		}

		view, err := escape.UpdateView(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update view: %w", err)
		}

		printView(view.GetId(), view.GetName())

		return nil
	},
}

var viewsReorderCmd = &cobra.Command{
	Use:   "reorder",
	Short: "Reorder the saved views of a target from JSON stdin",
	Long: `Reorder the views of a target. The order field lists the view IDs in
their new order; views omitted from the list keep their relative order after
the provided ones.`,
	Example: `  echo '{"target":"ISSUES","order":["00000000-0000-0000-0000-000000000000","00000000-0000-0000-0000-000000000001"]}' \
  | escape-cli views reorder`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.ReorderViewsRequest{}) {
			return nil
		}

		if out.Schema(out.Message{}) {
			return nil
		}

		body, err := readViewBody(cmd)
		if err != nil {
			return err
		}

		if _, err := escape.ReorderViews(cmd.Context(), body); err != nil {
			return fmt.Errorf("failed to reorder views: %w", err)
		}

		out.Log("Views reordered")

		return nil
	},
}

var viewsDeleteCmd = &cobra.Command{
	Use:     "delete view-id",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a saved view",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(out.Message{}) {
			return nil
		}

		proceed, err := confirmDestructive(cmd, "delete view "+args[0], nil)
		if err != nil || !proceed {
			return err
		}

		if _, err := escape.DeleteView(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete view: %w", err)
		}

		out.Log(fmt.Sprintf("View %s deleted", args[0]))

		return nil
	},
}

// readViewBody reads the JSON body a view command forwards. A terminal stdin is
// read anyway: unlike a flag-only command, there is no other way to express a
// view's filters, and an empty body would fail on the server instead of here.
func readViewBody(cmd *cobra.Command) ([]byte, error) {
	data, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, errors.New("a JSON body is required on stdin; run with --input-schema to see its shape")
	}

	return data, nil
}

func printView(id, name string) {
	out.Table(v3.CreateView200Response{Id: id, Name: name}, func() []string {
		return []string{"ID\tNAME", fmt.Sprintf("%s\t%s", id, name)}
	})
}

func init() {
	viewsCmd.AddCommand(viewsListCmd)
	viewsCmd.AddCommand(viewsIssuesCmd)
	viewsCmd.AddCommand(viewsCreateCmd)
	viewsCmd.AddCommand(viewsUpdateCmd)
	viewsCmd.AddCommand(viewsReorderCmd)
	viewsCmd.AddCommand(viewsDeleteCmd)

	viewsListCmd.Flags().StringVar(&viewsTarget, "target", "", "resource the views apply to: "+strings.Join(viewTargets, ", "))
	if err := viewsListCmd.Flags().SetAnnotation("target", flagEnumAnnotation, viewTargets); err != nil {
		panic(err)
	}

	_ = viewsListCmd.MarkFlagRequired("target")
	viewsListCmd.Flags().StringVar(&viewsProjectID, "project-id", "", "only return views scoped to this project ID")
	viewsListPage.bind(viewsIssuesCmd)
	bindDestructiveGate(viewsDeleteCmd, true)

	viewsIssuesCmd.Flags().StringVar(&viewsSortType, "sort-by", "", "sort field: ASSET, LAST_SEEN, SEVERITY")
	viewsIssuesCmd.Flags().StringVar(&viewsSortDirection, "sort-direction", "", "sort direction: asc, desc")

	rootCmd.AddCommand(viewsCmd)
}
