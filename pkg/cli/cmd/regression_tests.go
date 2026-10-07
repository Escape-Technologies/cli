package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	regressionTestsSearch      string
	regressionTestsListPage    pageFlags
	regressionTestsHistoryPage pageFlags
)

var regressionTestsCmd = &cobra.Command{
	Use:     "regression-tests",
	Aliases: []string{"regression-test", "regtests"},
	Short:   "Manage regression tests",
	Long: `Manage Regression Tests - Replay a Report as a Repeatable Test

A regression test is built from a source report (PDF or Markdown) uploaded
beforehand. Run it to have the agent reproduce the reported findings, follow
the run status, and answer the clarification question when the agent asks one.

COMMON WORKFLOWS:
  • Create a test from an uploaded report:
    $ escape-cli regression-tests create < body.json

  • Run it and follow the history:
    $ escape-cli regression-tests run <test-id>
    $ escape-cli regression-tests history <test-id>

  • Answer a clarification:
    $ echo '{"content":"Use the v2 checkout API"}' | \
        escape-cli regression-tests answer <test-id> <run-id>`,
}

var regressionTestsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List regression tests",
	Long: `List Regression Tests - Most Recent First

List the regression tests of the organization. With --size or --cursor the
command returns one page as {items, nextCursor, totalCount}; otherwise it walks
every page.`,
	Example: `  # List every regression test
  escape-cli regression-tests list

  # Search by name
  escape-cli regression-tests list --search checkout

  # Export one page
  escape-cli regression-tests list --size 50 -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[v3.RegressionTestSummary]{}) {
			return nil
		}

		if err := runPagedList(cmd, regressionTestsListPage, func(ctx context.Context, cursor string, size int) ([]v3.RegressionTestSummary, *string, int, error) {
			return escape.ListRegressionTests(ctx, cursor, regressionTestsSearch, size)
		}, func(tests []v3.RegressionTestSummary) []string {
			rows := []string{"ID\tNAME\tSTATUS\tINPUT FORMAT\tINPUT FILE\tCREATED AT\tUPDATED AT"}
			for _, test := range tests {
				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s",
					test.GetId(),
					test.GetName(),
					test.GetStatus(),
					test.GetInputFormat(),
					test.GetInputFilename(),
					test.GetCreatedAt(),
					test.GetUpdatedAt(),
				))
			}

			return rows
		}); err != nil {
			return fmt.Errorf("unable to list regression tests: %w", err)
		}

		return nil
	},
}

var regressionTestsGetCmd = &cobra.Command{
	Use:     "get regression-test-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get a regression test and its recent runs",
	Long: `Get Regression Test - Read Status and Recent Runs

Fetch a regression test by ID. The response embeds the most recent runs; use
history to page through every run.`,
	Example: `  escape-cli regression-tests get 00000000-0000-0000-0000-000000000000
  escape-cli regression-tests get <test-id> -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateRegressionTest200Response{}) {
			return nil
		}

		test, err := escape.GetRegressionTest(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get regression test: %w", err)
		}

		printRegressionTest(test)

		return nil
	},
}

var regressionTestsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a regression test from an uploaded report",
	Long: `Create Regression Test - From an Uploaded Report

Upload the source report first with "escape-cli upload" (POST /upload/signed-url),
then send a JSON body on stdin with the name, context, file name and the
returned temporary object key.

BODY:
  {
    "name": "Checkout regression test",
    "additionalContext": "Focus on the checkout flow.",
    "inputFilename": "report.pdf",
    "inputFormat": "PDF",                       // optional, inferred from the name
    "temporaryObjectKey": "00000000-0000-0000-0000-000000000000"
  }`,
	Example: `  cat <<'EOF' | escape-cli regression-tests create
  {"name":"Checkout","additionalContext":"Focus on checkout","inputFilename":"report.pdf","temporaryObjectKey":"00000000-0000-0000-0000-000000000000"}
  EOF`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.CreateRegressionTestInput{}) {
			return nil
		}

		if out.Schema(v3.CreateRegressionTest200Response{}) {
			return nil
		}

		body, err := readRequiredBody(cmd)
		if err != nil {
			return err
		}

		test, err := escape.CreateRegressionTest(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("unable to create regression test: %w", err)
		}

		printRegressionTest(test)

		return nil
	},
}

var regressionTestsUpdateCmd = &cobra.Command{
	Use:     "update regression-test-id",
	Aliases: []string{"u", "edit"},
	Short:   "Update a regression test name or context",
	Long: `Update Regression Test - Rename or Edit the Context

Update the name and/or the additional context of a regression test. Send a JSON
body on stdin with at least one of the two fields.`,
	Example: `  echo '{"name":"Checkout v2"}' | escape-cli regression-tests update <test-id>
  echo '{"additionalContext":"Focus on refunds"}' | escape-cli regression-tests update <test-id>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.UpdateRegressionTestRequest{}) {
			return nil
		}

		if out.Schema(v3.CreateRegressionTest200Response{}) {
			return nil
		}

		body, err := readRequiredBody(cmd)
		if err != nil {
			return err
		}

		test, err := escape.UpdateRegressionTest(cmd.Context(), args[0], body)
		if err != nil {
			return fmt.Errorf("unable to update regression test: %w", err)
		}

		printRegressionTest(test)

		return nil
	},
}

var regressionTestsDeleteCmd = &cobra.Command{
	Use:     "delete regression-test-id",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a regression test",
	Long: `Delete Regression Test - Remove the Test and Its Runs

Permanently delete a regression test and its run history.`,
	Example: `  escape-cli regression-tests delete <test-id>`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.DeleteProfile200Response{}) {
			return nil
		}

		result, err := escape.DeleteRegressionTest(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to delete regression test: %w", err)
		}

		out.Print(result, fmt.Sprintf("Regression test %s deleted", args[0]))

		return nil
	},
}

var regressionTestsRunCmd = &cobra.Command{
	Use:     "run regression-test-id",
	Aliases: []string{"start"},
	Short:   "Run a regression test",
	Long: `Run Regression Test - Reproduce the Reported Findings

Start a run of the regression test. The run is asynchronous: poll the test or
its history to track the status. It fails when a run is already in progress.`,
	Example: `  escape-cli regression-tests run <test-id>
  escape-cli regression-tests run <test-id> -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateRegressionTest200Response{}) {
			return nil
		}

		test, err := escape.RunRegressionTest(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to run regression test: %w", err)
		}

		printRegressionTest(test)

		return nil
	},
}

var regressionTestsStopCmd = &cobra.Command{
	Use:   "stop regression-test-id",
	Short: "Stop the run in progress",
	Long: `Stop Regression Test - Cancel the Run in Progress

Stop the run in progress for a regression test. A run that already finished is
left unchanged.`,
	Example: `  escape-cli regression-tests stop <test-id>`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.CreateRegressionTest200Response{}) {
			return nil
		}

		test, err := escape.StopRegressionTest(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to stop regression test: %w", err)
		}

		printRegressionTest(test)

		return nil
	},
}

var regressionTestsHistoryCmd = &cobra.Command{
	Use:     "history regression-test-id",
	Aliases: []string{"runs"},
	Short:   "List the runs of a regression test",
	Long: `List Regression Test History - Runs, Most Recent First

List the runs of a regression test. With --size or --cursor the command returns
one page as {items, nextCursor, totalCount}; otherwise it walks every page.`,
	Example: `  escape-cli regression-tests history <test-id>
  escape-cli regression-tests history <test-id> --size 50 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(Page[v3.RegressionTestRun]{}) {
			return nil
		}

		if err := runPagedList(cmd, regressionTestsHistoryPage, func(ctx context.Context, cursor string, size int) ([]v3.RegressionTestRun, *string, int, error) {
			return escape.ListRegressionTestHistory(ctx, args[0], cursor, size)
		}, regressionTestRunsTable); err != nil {
			return fmt.Errorf("unable to list regression test history: %w", err)
		}

		return nil
	},
}

var regressionTestsAnswerCmd = &cobra.Command{
	Use:     "answer regression-test-id run-id",
	Aliases: []string{"clarify", "reply"},
	Short:   "Answer a run clarification question",
	Long: `Answer Clarification - Reply to the Agent's Question

Answer the clarification question asked by the agent during a run. The run must
be waiting for a clarification. Send a JSON body on stdin with the answer.`,
	Example: `  echo '{"content":"The checkout flow now uses the v2 API."}' | \
    escape-cli regression-tests answer <test-id> <run-id>`,
	Args: cobra.ExactArgs(2), //nolint:mnd
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.ReplyRegressionTestClarificationInput{}) {
			return nil
		}

		if out.Schema(v3.ReplyRegressionTestClarification200Response{}) {
			return nil
		}

		body, err := readRequiredBody(cmd)
		if err != nil {
			return err
		}

		run, err := escape.ReplyRegressionTestClarification(cmd.Context(), args[0], args[1], body)
		if err != nil {
			return fmt.Errorf("unable to answer clarification: %w", err)
		}

		out.Table(run, func() []string {
			return []string{
				"ID\tSTATUS\tCLARIFICATION QUESTION\tCLARIFICATION ANSWER\tVALIDATION ID\tUPDATED AT",
				fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
					run.GetId(),
					run.GetStatus(),
					run.GetClarificationQuestion(),
					run.GetClarificationAnswer(),
					run.GetValidationId(),
					run.GetUpdatedAt(),
				),
			}
		})

		return nil
	},
}

// readRequiredBody reads a JSON body from stdin and fails when it is empty.
func readRequiredBody(cmd *cobra.Command) ([]byte, error) {
	body, err := readPipedStdin(cmd.InOrStdin())
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, errors.New("no input provided: pipe a JSON body via stdin")
	}

	return body, nil
}

func printRegressionTest(test *v3.CreateRegressionTest200Response) {
	out.Table(test, func() []string {
		return []string{
			"ID\tNAME\tSTATUS\tINPUT FORMAT\tINPUT FILE\tRUNS\tMORE RUNS\tCREATED AT",
			fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%t\t%s",
				test.GetId(),
				test.GetName(),
				test.GetStatus(),
				test.GetInputFormat(),
				test.GetInputFilename(),
				len(test.GetRuns()),
				test.GetHasMoreRuns(),
				test.GetCreatedAt(),
			),
		}
	})

	if !out.IsPretty() {
		return
	}

	runs := test.GetRuns()
	out.Table(runs, func() []string {
		return regressionTestRunsTable(runs)
	})
}

func regressionTestRunsTable(runs []v3.RegressionTestRun) []string {
	rows := []string{"ID\tSTATUS\tCLARIFICATION QUESTION\tVALIDATION ID\tCREATED AT\tUPDATED AT"}
	for _, run := range runs {
		rows = append(rows, formatRegressionTestRun(run))
	}

	return rows
}

func formatRegressionTestRun(run v3.RegressionTestRun) string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
		run.GetId(),
		run.GetStatus(),
		run.GetClarificationQuestion(),
		run.GetValidationId(),
		run.GetCreatedAt(),
		run.GetUpdatedAt(),
	)
}

func init() {
	regressionTestsCmd.AddCommand(
		regressionTestsListCmd,
		regressionTestsGetCmd,
		regressionTestsCreateCmd,
		regressionTestsUpdateCmd,
		regressionTestsDeleteCmd,
		regressionTestsRunCmd,
		regressionTestsStopCmd,
		regressionTestsHistoryCmd,
		regressionTestsAnswerCmd,
	)

	regressionTestsListCmd.Flags().StringVar(&regressionTestsSearch, "search", "", "search term to filter regression tests by name")
	regressionTestsListPage.bind(regressionTestsListCmd)
	regressionTestsHistoryPage.bind(regressionTestsHistoryCmd)

	rootCmd.AddCommand(regressionTestsCmd)
}
