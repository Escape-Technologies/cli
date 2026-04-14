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

var workflowsCmd = &cobra.Command{
	Use:     "workflows",
	Aliases: []string{"workflow", "wf"},
	Short:   "Manage automation workflows",
	Long: `Manage Workflows - Automate Security Actions

Workflows automate actions triggered by security events such as scan completion.
Create integrations with Jira, Slack, and other tools through configurable workflows.

COMMON WORKFLOWS:
  • List workflows:        escape-cli workflows list
  • Get workflow details:  escape-cli workflows get <id>
  • Create workflow:       escape-cli workflows create < workflow.json`,
}

var workflowsTriggersFlag []string
var workflowsSearchFlag string

var workflowsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all workflows",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]v3.WorkflowSummarized{}) {
			return nil
		}

		filters := &escape.ListWorkflowsFilters{
			Triggers: workflowsTriggersFlag,
			Search:   workflowsSearchFlag,
		}
		workflows, next, err := escape.ListWorkflows(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("unable to list workflows: %w", err)
		}
		all := workflows
		for next != nil && *next != "" {
			workflows, next, err = escape.ListWorkflows(cmd.Context(), *next, filters)
			if err != nil {
				return fmt.Errorf("unable to list workflows: %w", err)
			}
			all = append(all, workflows...)
		}

		out.Table(all, func() []string {
			res := []string{"ID\tNAME\tTRIGGER"}
			for _, w := range all {
				res = append(res, fmt.Sprintf("%s\t%s\t%s", w.GetId(), w.GetName(), w.GetTrigger()))
			}
			return res
		})
		return nil
	},
}

var workflowsGetCmd = &cobra.Command{
	Use:     "get workflow-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get a workflow by ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("workflow ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateWorkflow200Response{}) {
			return nil
		}

		workflow, err := escape.GetWorkflow(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get workflow: %w", err)
		}

		out.Table(workflow, func() []string {
			res := []string{"ID\tNAME\tTRIGGER"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s", workflow.GetId(), workflow.GetName(), workflow.GetTrigger()))
			return res
		})
		return nil
	},
}

var workflowsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workflow from JSON stdin",
	Example: `  # Create from file
  escape-cli workflows create < workflow.json

  # See input schema
  escape-cli workflows create --input-schema`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.CreateWorkflowRequest{}) {
			return nil
		}
		if out.Schema(v3.CreateWorkflow200Response{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		workflow, err := escape.CreateWorkflow(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to create workflow: %w", err)
		}

		out.Table(workflow, func() []string {
			return []string{
				"ID\tNAME\tTRIGGER",
				fmt.Sprintf("%s\t%s\t%s", workflow.GetId(), workflow.GetName(), workflow.GetTrigger()),
			}
		})
		return nil
	},
}

var workflowsUpdateCmd = &cobra.Command{
	Use:   "update workflow-id",
	Short: "Update a workflow from JSON stdin",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("workflow ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.UpdateWorkflowRequest{}) {
			return nil
		}
		if out.Schema(v3.CreateWorkflow200Response{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		workflow, err := escape.UpdateWorkflow(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update workflow: %w", err)
		}

		out.Table(workflow, func() []string {
			return []string{
				"ID\tNAME\tTRIGGER",
				fmt.Sprintf("%s\t%s\t%s", workflow.GetId(), workflow.GetName(), workflow.GetTrigger()),
			}
		})
		return nil
	},
}

var workflowsDeleteCmd = &cobra.Command{
	Use:     "delete workflow-id",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a workflow by ID",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("workflow ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := escape.DeleteWorkflow(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete workflow: %w", err)
		}
		out.Log(fmt.Sprintf("Workflow %s deleted", args[0]))
		return nil
	},
}

func init() {
	workflowsCmd.AddCommand(workflowsListCmd, workflowsGetCmd, workflowsCreateCmd, workflowsUpdateCmd, workflowsDeleteCmd)
	workflowsListCmd.Flags().StringSliceVar(&workflowsTriggersFlag, "trigger", []string{}, "filter by trigger type (e.g. SCAN_FINISHED)")
	workflowsListCmd.Flags().StringVarP(&workflowsSearchFlag, "search", "s", "", "search workflows by name")
	rootCmd.AddCommand(workflowsCmd)
}
