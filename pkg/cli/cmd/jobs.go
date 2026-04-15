package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage async jobs (reports, exports)",
	Long: `Manage Async Jobs - Trigger and Track Background Operations

Jobs represent long-running background operations such as report generation
and PDF exports. Use trigger-export to start an export, then poll with get.

COMMON WORKFLOWS:
  • Trigger a report export:
    $ escape-cli jobs trigger-export --block EXECUTIVE_SUMMARY --block VULNERABILITIES

  • Get job status:
    $ escape-cli jobs get <job-id> [--watch]`,
}

var (
	jobsExportBlocks []string
	jobsExportScanID string
	jobsWatch        bool
)

var jobsTriggerExportCmd = &cobra.Command{
	Use:   "trigger-export",
	Short: "Trigger a PDF/report export job",
	Long: `Trigger Export - Generate Security Reports

Start an async job to generate a PDF or structured security report.
The job runs in the background; use 'jobs get <id> --watch' to poll until complete.

AVAILABLE BLOCKS:
  EXECUTIVE_SUMMARY    - High-level overview for management
  VULNERABILITIES      - Detailed vulnerability listing
  COMPLIANCE           - Compliance status report
  METHODOLOGY          - Testing methodology description`,
	Example: `  # Export full report for latest scan
  escape-cli jobs trigger-export --block EXECUTIVE_SUMMARY --block VULNERABILITIES

  # Export for a specific scan
  escape-cli jobs trigger-export --block VULNERABILITIES --scan-id <scan-id>

  # Trigger and watch until complete
  escape-cli jobs trigger-export --block EXECUTIVE_SUMMARY --watch`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.TriggerExport200Response{}) {
			return nil
		}

		if len(jobsExportBlocks) == 0 {
			return errors.New("at least one --block is required")
		}

		job, err := escape.TriggerExport(cmd.Context(), jobsExportBlocks, jobsExportScanID)
		if err != nil {
			return fmt.Errorf("failed to trigger export: %w", err)
		}

		out.Print(job, "Export job started: "+job.GetId())

		if jobsWatch {
			return watchJob(cmd, job.GetId())
		}
		return nil
	},
}

var jobsGetCmd = &cobra.Command{
	Use:   "get job-id",
	Short: "Get job status and result",
	Long: `Get Job Status - Check Async Job Progress

Retrieve the current status and result of an async job.
Use --watch to poll until the job completes.`,
	Example: `  # Check job status
  escape-cli jobs get <job-id>

  # Watch until complete
  escape-cli jobs get <job-id> --watch`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetJob200Response{}) {
			return nil
		}
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("job ID is required")
		}

		if jobsWatch {
			return watchJob(cmd, args[0])
		}

		job, err := escape.GetJob(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get job: %w", err)
		}
		printJob(job)
		return nil
	},
}

func printJob(job *v3.GetJob200Response) {
	out.Table(job, func() []string {
		res := []string{"ID\tSTATUS\tCREATED AT"}
		res = append(res, fmt.Sprintf("%s\t%s\t%s", job.GetId(), job.GetStatus(), job.GetCreatedAt()))
		return res
	})
}

func watchJob(cmd *cobra.Command, jobID string) error {
	terminalStatuses := map[string]bool{
		"FINISHED": true,
		"FAILED":   true,
		"CANCELED": true,
	}
	for {
		job, err := escape.GetJob(cmd.Context(), jobID)
		if err != nil {
			return fmt.Errorf("unable to get job: %w", err)
		}
		printJob(job)
		status := strings.ToUpper(string(job.GetStatus()))
		if terminalStatuses[status] {
			if status == "FAILED" || status == "CANCELED" {
				return fmt.Errorf("job ended with status %s", status)
			}
			return nil
		}
		time.Sleep(3 * time.Second) //nolint:mnd
	}
}

func init() {
	jobsCmd.AddCommand(jobsTriggerExportCmd, jobsGetCmd)
	jobsTriggerExportCmd.Flags().StringArrayVar(&jobsExportBlocks, "block", []string{}, "report block to include (can be specified multiple times)")
	jobsTriggerExportCmd.Flags().StringVar(&jobsExportScanID, "scan-id", "", "scan ID to export (defaults to latest)")
	jobsTriggerExportCmd.Flags().BoolVarP(&jobsWatch, "watch", "w", false, "watch until job completes")
	jobsGetCmd.Flags().BoolVarP(&jobsWatch, "watch", "w", false, "watch until job completes")
	rootCmd.AddCommand(jobsCmd)
}
