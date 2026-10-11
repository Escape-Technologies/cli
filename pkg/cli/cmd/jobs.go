package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

// jobStatusFn reads one job. Injectable so a watch test drives statuses without the API.
type jobStatusFn func(ctx context.Context, jobID string) (*v3.GetJob200Response, error)

var validJobsExportBlocks = func() []string {
	blocks := make([]string, 0, len(v3.AllowedENUMPROPERTIESBLOCKSITEMSPROPERTIESKINDEnumValues))
	for _, block := range v3.AllowedENUMPROPERTIESBLOCKSITEMSPROPERTIESKINDEnumValues {
		blocks = append(blocks, string(block))
	}

	return blocks
}()

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage async jobs (reports, exports)",
	Long: `Manage Async Jobs - Trigger and Track Background Operations

Jobs represent long-running background operations such as report generation
and PDF exports. Use trigger-export to start an export, then poll with get.

BETA: job statuses, artefacts and exports are an internal async contract, so this
group can still change.

COMMON WORKFLOWS:
  • Trigger a report export:
    $ escape-cli jobs trigger-export --block SCAN_SUMMARY --block ISSUE_SUMMARY

  • Get job status:
    $ escape-cli jobs get <job-id> [--watch]

  • Rename a finished report, or mail its artefacts:
    $ escape-cli jobs rename <job-id> --name "Q3 penetration test"
    $ escape-cli jobs send-artefacts <job-id> --email security@example.com`,
}

var (
	jobsExportBlocks       []string
	jobsExportScanID       string
	jobsExportName         string
	jobsExportAssetWhere   string
	jobsExportProfileWhere string
	jobsExportIssueWhere   string
	jobsExportDryRun       bool
	jobsExportNoNotify     bool
	jobsWatch              bool
	jobsWatchStream        bool
	jobsWatchTimeout       time.Duration
)

// defaultJobsWatchTimeout bounds `--watch`. A job that never reaches a terminal
// status used to loop forever: the watch now stops and asks for a poll.
const defaultJobsWatchTimeout = 15 * time.Minute

// defaultJobWatchPollInterval is the delay between two job status polls. It is a
// variable so tests drive the wait without real sleeps.
var defaultJobWatchPollInterval = 3 * time.Second

// terminalJobStatuses are the statuses a job can end on, per the job model
// (PENDING, RUNNING, COMPLETED, FAILED). `CANCELLED` is accepted because the
// API contract widens the enum, but the scheduler never sets it today.
var terminalJobStatuses = map[string]bool{
	"COMPLETED": true,
	"FAILED":    true,
	"CANCELLED": true,
}

// failedJobStatuses end the watch with a non-zero exit code.
var failedJobStatuses = map[string]bool{
	"FAILED":    true,
	"CANCELLED": true,
}

var jobsTriggerExportCmd = &cobra.Command{
	Use:   "trigger-export",
	Short: "Trigger a PDF/report export job (Beta)",
	Long: `Trigger Export - Generate Security Reports

Start an async job to generate a PDF or structured security report.
The job runs in the background; use 'jobs get <id> --watch' to poll until complete.

SCOPE:
  Without a filter the export covers the whole organization. --scan-id restricts
  it to one scan, and the --asset-where, --profile-where and --issue-where flags
  take a JSON where-input, the same shape the model uses internally, to narrow
  each resource type. --dry-run validates the selection and schedules no job.

AVAILABLE BLOCKS:
  ` + strings.Join(validJobsExportBlocks, "\n  "),
	Example: `  # Export full report for latest scan
  escape-cli jobs trigger-export --block SCAN_SUMMARY --block ISSUE_SUMMARY

  # Export for a specific scan
  escape-cli jobs trigger-export --block ISSUE_DETAIL --scan-id <scan-id>

  # Name the report and narrow it to the open issues of one asset
  escape-cli jobs trigger-export --block ISSUE_LIST --name "Weekly digest" \
    --issue-where '{"assetIds":["<asset-id>"],"status":["OPEN"]}'

  # Validate the selection without scheduling anything
  escape-cli jobs trigger-export --block ISSUE_LIST --dry-run

  # Trigger and watch until complete
  escape-cli jobs trigger-export --block SCAN_SUMMARY --watch`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.TriggerExport200Response{}) {
			return nil
		}

		if len(jobsExportBlocks) == 0 {
			return errors.New("at least one --block is required")
		}

		for _, block := range jobsExportBlocks {
			if !v3.ENUMPROPERTIESBLOCKSITEMSPROPERTIESKIND(block).IsValid() {
				return fmt.Errorf("invalid block %q; valid values: %s", block, strings.Join(validJobsExportBlocks, ", "))
			}
		}

		if jobsExportDryRun && jobsWatch {
			return errors.New("--dry-run schedules no job, so it cannot be combined with --watch")
		}

		// Validate before the export is triggered, so a bad flag does not leave a job running.
		if jobsWatch {
			if err := validateJobsWatchTimeout(); err != nil {
				return err
			}
		}

		opts := escape.ExportOptions{
			ScanID: jobsExportScanID,
			Name:   jobsExportName,
			Dry:    jobsExportDryRun,
		}
		if jobsExportNoNotify {
			notify := false
			opts.Notify = &notify
		}

		where, err := parseExportWhereFlags(cmd)
		if err != nil {
			return err
		}

		opts.AssetWhere = where.asset
		opts.ProfileWhere = where.profile
		opts.IssueWhere = where.issue

		job, err := escape.TriggerExport(cmd.Context(), jobsExportBlocks, opts)
		if err != nil {
			return fmt.Errorf("failed to trigger export: %w", err)
		}

		if jobsExportDryRun {
			out.Print(job, "Dry run: the export selection is valid, no job was scheduled")

			return nil
		}

		out.Print(job, "Export job started: "+job.GetId())

		if jobsWatch {
			return watchJob(cmd, job.GetId())
		}

		return nil
	},
}

// exportWhereFlags holds the parsed JSON where-inputs of an export request.
type exportWhereFlags struct {
	asset   json.RawMessage
	profile json.RawMessage
	issue   json.RawMessage
}

// parseExportWhereFlags validates the --asset-where, --profile-where and
// --issue-where flags. Each must be a JSON object, the where-input shape the
// API forwards to the internal GraphQL layer.
func parseExportWhereFlags(cmd *cobra.Command) (exportWhereFlags, error) {
	var parsed exportWhereFlags
	for _, flag := range []struct {
		name   string
		value  string
		target *json.RawMessage
	}{
		{"asset-where", jobsExportAssetWhere, &parsed.asset},
		{"profile-where", jobsExportProfileWhere, &parsed.profile},
		{"issue-where", jobsExportIssueWhere, &parsed.issue},
	} {
		if strings.TrimSpace(flag.value) == "" {
			// An empty shell expansion must not silently widen the export.
			if cmd.Flags().Changed(flag.name) {
				return exportWhereFlags{}, fmt.Errorf("--%s must not be blank", flag.name)
			}

			continue
		}

		var decoded any
		if err := json.Unmarshal([]byte(flag.value), &decoded); err != nil {
			return exportWhereFlags{}, fmt.Errorf("--%s must be a JSON object: %w", flag.name, err)
		}

		if _, ok := decoded.(map[string]any); !ok {
			return exportWhereFlags{}, fmt.Errorf("--%s must be a JSON object", flag.name)
		}

		*flag.target = json.RawMessage(flag.value)
	}

	return parsed, nil
}

var jobsGetCmd = &cobra.Command{
	Use:   "get job-id",
	Short: "Get job status and result (Beta)",
	Long: `Get Job Status - Check Async Job Progress

Retrieve the current status and result of an async job.
Use --watch to poll until the job completes.

The watch prints one final job state, stops after --timeout (default 15m, 0 to
disable the bound) and exits non-zero on a FAILED or CANCELLED job. Pass
--stream to keep the older per-poll output.`,
	Example: `  # Check job status
  escape-cli jobs get <job-id>

  # Watch until complete
  escape-cli jobs get <job-id> --watch

  # Watch with a longer budget, printing every poll
  escape-cli jobs get <job-id> --watch --timeout 30m --stream`,
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

var (
	jobsListSearch        string
	jobsListKinds         []string
	jobsListStatuses      []string
	jobsListSortType      string
	jobsListSortDirection string
	jobsListPage          pageFlags

	jobsRenameName string

	jobsArtefactsEmails  []string
	jobsArtefactsSubject string
)

var jobsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the async jobs of the organization (Beta)",
	Long: `List Jobs - Most Recent First

List the async jobs of the organization: exports, report artefacts and the
jobs that follow them. The listing is ordered by creation date, most recent
first; --sort-type and --sort-direction change that. With --size or --cursor
the command returns one page as {items, nextCursor, totalCount}; otherwise it
walks every page.`,
	Example: `  # List every job, most recent first
  escape-cli jobs list

  # Jobs still running
  escape-cli jobs list --status RUNNING

  # Report template jobs only, one page
  escape-cli jobs list --kind SAAS_EXPORT_TEMPLATE --size 20 -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[v3.JobSummary]{}) {
			return nil
		}

		filters := &escape.JobsFilters{
			Search:        jobsListSearch,
			Kinds:         jobsListKinds,
			Statuses:      jobsListStatuses,
			SortType:      jobsListSortType,
			SortDirection: jobsListSortDirection,
		}

		if err := runPagedList(cmd, jobsListPage, func(ctx context.Context, cursor string, size int) ([]v3.JobSummary, *string, int, error) {
			return escape.ListJobs(ctx, cursor, size, filters)
		}, jobsSummaryTable); err != nil {
			return fmt.Errorf("unable to list jobs: %w", err)
		}

		return nil
	},
}

var jobsRenameCmd = &cobra.Command{
	Use:   "rename job-id",
	Short: "Rename an export job (Beta)",
	Long: `Rename an Export - File a Report Under Another Name

Rename an export job. The name is the one shown on the generated report and in
the job listing, so renaming is how a generated export is filed.`,
	Example: `  escape-cli jobs rename <job-id> --name "Q3 penetration test"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.RenameExport200Response{}) {
			return nil
		}

		if jobsRenameName == "" {
			return errors.New("--name is required")
		}

		renamed, err := escape.RenameExport(cmd.Context(), args[0], jobsRenameName)
		if err != nil {
			return fmt.Errorf("unable to rename job %s: %w", args[0], err)
		}

		out.Table(renamed, func() []string {
			return []string{"ID\tNAME", fmt.Sprintf("%s\t%s", renamed.GetId(), renamed.GetName())}
		})

		return nil
	},
}

var jobsSendArtefactsCmd = &cobra.Command{
	Use:     "send-artefacts job-id",
	Aliases: []string{"send-artifacts"},
	Short:   "Email the artefacts a job produced (Beta)",
	Long: `Send Job Artefacts - Mail a Finished Report

Email the artefacts a finished job produced (the PDF or export files) to up to
50 recipients. --subject is optional; without it the platform's default subject
is used. The job must already have finished, so poll 'jobs get --watch' first.`,
	Example: `  escape-cli jobs send-artefacts <job-id> --email security@example.com
  escape-cli jobs send-artefacts <job-id> --email a@example.com --email b@example.com --subject "Q3 report"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.SendJobArtefacts200Response{}) {
			return nil
		}

		if len(jobsArtefactsEmails) == 0 {
			return errors.New("at least one --email is required")
		}

		result, err := escape.SendJobArtefacts(cmd.Context(), args[0], jobsArtefactsEmails, jobsArtefactsSubject)
		if err != nil {
			return fmt.Errorf("unable to send the artefacts of job %s: %w", args[0], err)
		}

		out.Table(result, func() []string {
			return []string{"SENT", strconv.FormatBool(result.GetSent())}
		})

		summary := fmt.Sprintf("Artefacts of job %s were not sent", args[0])
		if result.GetSent() {
			summary = fmt.Sprintf("Artefacts of job %s sent to %d recipient(s)", args[0], len(jobsArtefactsEmails))
		}

		out.Log(summary)

		return nil
	},
}

// jobsSummaryTable lists one line per job. The status is the column an agent
// polls on, and the error only exists on a job that failed.
func jobsSummaryTable(jobs []v3.JobSummary) []string {
	rows := []string{"ID\tKIND\tSTATUS\tNAME\tCREATED AT\tERROR"}
	for _, job := range jobs {
		errorMessage := ""
		if err, ok := job.GetErrorOk(); ok {
			errorMessage = err.GetMessage()
		}

		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
			job.GetId(),
			job.GetKind(),
			job.GetStatus(),
			job.GetName(),
			job.GetCreatedAt(),
			errorMessage,
		))
	}

	return rows
}

func printJob(job *v3.GetJob200Response) {
	out.Table(job, func() []string {
		res := []string{"ID\tSTATUS\tCREATED AT"}
		res = append(res, fmt.Sprintf("%s\t%s\t%s", job.GetId(), job.GetStatus(), job.GetCreatedAt()))

		return res
	})
}

// validateJobsWatchTimeout rejects a negative --timeout: only 0 disables the bound.
func validateJobsWatchTimeout() error {
	if jobsWatchTimeout < 0 {
		return fmt.Errorf("invalid --timeout %s: use 0 to disable the bound", jobsWatchTimeout)
	}

	return nil
}

// watchJob polls the job until it ends on a terminal status, the timeout fires, or the
// context is cancelled. It prints one final job state. --stream keeps the older
// per-poll output for scripts that scrape the progress stream.
func watchJob(cmd *cobra.Command, jobID string) error {
	return runJobWatch(cmd, jobID, escape.GetJob)
}

// runJobWatch is the injectable body of watchJob: getJob is the status source.

func runJobWatch(cmd *cobra.Command, jobID string, getJob jobStatusFn) error {
	if err := validateJobsWatchTimeout(); err != nil {
		return err
	}

	waitCtx := cmd.Context()
	if jobsWatchTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(waitCtx, jobsWatchTimeout)
		defer cancel()
	}

	var last *v3.GetJob200Response
	for {
		// The deadline can fire while a poll was in flight: catch it before the next one.
		// The first poll always runs, so the stop error can carry a status.
		if last != nil && waitCtx.Err() != nil {
			return jobWatchStopError(waitCtx, jobID, last)
		}

		job, err := getJob(waitCtx, jobID)
		if err != nil {
			// The deadline, not the transport, is why the call stopped: a timeout that
			// surfaces as an API error makes the bounded watch look like a crash.
			if waitCtx.Err() != nil {
				return jobWatchStopError(waitCtx, jobID, last)
			}

			return fmt.Errorf("unable to get job: %w", err)
		}

		last = job
		if jobsWatchStream {
			printJob(job)
		}

		status := strings.ToUpper(string(job.GetStatus()))
		if terminalJobStatuses[status] {
			printWatchedJob(job, jobsWatchStream)
			if failedJobStatuses[status] {
				return fmt.Errorf("job ended with status %s", status)
			}

			return nil
		}

		select {
		case <-waitCtx.Done():
			return jobWatchStopError(waitCtx, jobID, last)
		case <-time.After(defaultJobWatchPollInterval):
		}
	}
}

// printWatchedJob prints the single final state, unless --stream already printed every
// poll including this one.
func printWatchedJob(job *v3.GetJob200Response, streamed bool) {
	if streamed || job == nil {
		return
	}

	printJob(job)
}

// jobWatchStopError prints the last known state and classifies why the watch stopped:
// only an expired deadline is a timeout, a canceled context is reported as such.
func jobWatchStopError(ctx context.Context, jobID string, last *v3.GetJob200Response) error {
	printWatchedJob(last, jobsWatchStream)

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return jobWatchTimeoutError(jobID, last)
	}

	return fmt.Errorf("job watch canceled: %w", ctx.Err())
}

func jobWatchTimeoutError(jobID string, last *v3.GetJob200Response) error {
	if last == nil {
		return fmt.Errorf("no job status after %s: poll 'escape-cli jobs get %s'", jobsWatchTimeout, jobID)
	}

	return fmt.Errorf(
		"job %s is still %s after %s: poll 'escape-cli jobs get %s'",
		jobID,
		last.GetStatus(),
		jobsWatchTimeout,
		jobID,
	)
}

func init() {
	jobsCmd.AddCommand(jobsTriggerExportCmd, jobsGetCmd, jobsListCmd, jobsRenameCmd, jobsSendArtefactsCmd)
	jobsRenameCmd.Flags().StringVar(&jobsRenameName, "name", "", "new name of the report (required)")
	markMCPRequired(jobsRenameCmd.Flags(), "name")
	jobsSendArtefactsCmd.Flags().StringSliceVar(&jobsArtefactsEmails, "email", nil, "recipient email address (required; can be specified multiple times)")
	jobsSendArtefactsCmd.Flags().StringVar(&jobsArtefactsSubject, "subject", "", "email subject line")
	markMCPRequired(jobsSendArtefactsCmd.Flags(), "email")
	jobsListCmd.Flags().StringVar(&jobsListSearch, "search", "", "case-insensitive substring match on the job kind (report names are not searchable)")
	jobsListCmd.Flags().StringSliceVar(&jobsListKinds, "kind", []string{}, "filter by job kind (can be specified multiple times)")
	jobsListCmd.Flags().StringSliceVar(&jobsListStatuses, "status", []string{}, "filter by job status: [PENDING RUNNING COMPLETED FAILED]")
	jobsListCmd.Flags().StringVar(&jobsListSortType, "sort-type", "", "rank the jobs by: [CREATED_AT SCHEDULED_AT]")
	// Most recent first: the bare API defaults to ascending, which puts the job an
	// agent just triggered last.
	jobsListCmd.Flags().StringVar(&jobsListSortDirection, "sort-direction", "desc", "sort direction: [asc desc]")
	jobsListPage.bind(jobsListCmd)
	jobsTriggerExportCmd.Flags().StringArrayVar(&jobsExportBlocks, "block", []string{}, "report block to include (can be specified multiple times)")
	jobsTriggerExportCmd.Flags().StringVar(&jobsExportScanID, "scan-id", "", "restrict the export to a scan ID")
	jobsTriggerExportCmd.Flags().StringVar(&jobsExportName, "name", "", "report name shown on the generated document")
	jobsTriggerExportCmd.Flags().StringVar(&jobsExportAssetWhere, "asset-where", "", "JSON AssetWhereInput narrowing exported assets")
	jobsTriggerExportCmd.Flags().StringVar(&jobsExportProfileWhere, "profile-where", "", "JSON ProfilesWhereInput narrowing exported profiles")
	jobsTriggerExportCmd.Flags().StringVar(&jobsExportIssueWhere, "issue-where", "", "JSON IssueWhereInput narrowing exported issues")
	jobsTriggerExportCmd.Flags().BoolVar(&jobsExportDryRun, "dry-run", false, "validate the selection and schedule no job")
	jobsTriggerExportCmd.Flags().BoolVar(&jobsExportNoNotify, "no-notify", false, "skip the email notification when the report is ready")
	jobsTriggerExportCmd.Flags().BoolVarP(&jobsWatch, "watch", "w", false, "watch until job completes")
	jobsTriggerExportCmd.Flags().BoolVar(&jobsWatchStream, "stream", false, "with --watch, print every poll instead of the final state only")
	jobsTriggerExportCmd.Flags().DurationVar(&jobsWatchTimeout, "timeout", defaultJobsWatchTimeout, "maximum time to watch before failing (e.g. 30m; 0 disables the timeout)")
	markMCPSkip(jobsTriggerExportCmd.Flags(), "watch", "stream", "timeout")
	jobsGetCmd.Flags().BoolVarP(&jobsWatch, "watch", "w", false, "watch until job completes")
	jobsGetCmd.Flags().BoolVar(&jobsWatchStream, "stream", false, "with --watch, print every poll instead of the final state only")
	jobsGetCmd.Flags().DurationVar(&jobsWatchTimeout, "timeout", defaultJobsWatchTimeout, "maximum time to watch before failing (e.g. 30m; 0 disables the timeout)")
	markMCPSkip(jobsGetCmd.Flags(), "watch", "stream", "timeout")
	rootCmd.AddCommand(jobsCmd)
}
