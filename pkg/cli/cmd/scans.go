package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/spf13/cobra"
)

var scanProfileIDs []string
var scanAfter string
var scanBefore string
var scanIgnored string
var scanInitiator []string
var scanKinds []string
var scanStatus []string

var scansCmd = &cobra.Command{
	Use:     "scans",
	Aliases: []string{"sc", "scan"},
	Short:   "Run and manage security scans on your APIs",
	Long: `Manage Security Scans - Start, Monitor, and Review API Security Tests

Scans are security tests that analyze your APIs for vulnerabilities. Each scan
runs against a profile and produces a detailed security report with discovered issues.

SCAN LIFECYCLE:
  1. STARTING   - Scan initialization
  2. RUNNING    - Active testing in progress
  3. FINISHED   - Scan completed successfully
  4. FAILED     - Scan encountered an error
  5. CANCELED   - Manually stopped

COMMON WORKFLOWS:
  • Start a scan and watch progress:
    $ escape-cli scans start <profile-id> --watch

  • List recent scans for a profile:
    $ escape-cli scans list -p <profile-id>

  • View scan results:
    $ escape-cli scans get <scan-id>
    $ escape-cli scans issues <scan-id>

  • CI/CD Integration:
    $ escape-cli scans start <profile-id> --watch --commit-hash $GITHUB_SHA`,
}

var scansListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List security scans with flexible filtering",
	Long: `List Security Scans - Query and Filter Scan History

List all scans across your organization with powerful filtering capabilities.
Filter by profile, status, date range, scanner type, and more.

FILTER OPTIONS:
  -p, --profile-id    Filter by one or more profile IDs
  -s, --status        Filter by scan status (RUNNING, FINISHED, FAILED, CANCELED)
  -k, --kind          Filter by scanner type (BLST_REST, BLST_GRAPHQL, FRONTEND_DAST)
  -i, --initiator     Filter by who started the scan (MANUAL, API, SCHEDULED, CI)
  --after             Show scans created after this date (RFC3339 format)
  --before            Show scans created before this date (RFC3339 format)
  --ignored           Filter by ignored status (true/false)

SCANNER TYPES:
  • BLST_REST         - REST API security testing
  • BLST_GRAPHQL      - GraphQL API security testing  
  • FRONTEND_DAST     - Web application security testing

Example output:
ID                                      CREATED AT                           KIND           STATUS      PROGRESS    LINK
00000000-0000-0000-0000-000000000001    2025-02-05 08:34:47.541 +0000 UTC    BLST_REST      FINISHED    1.000000    https://...
00000000-0000-0000-0000-000000000002    2025-02-02 08:27:23.919 +0000 UTC    BLST_GRAPHQL   RUNNING     0.453000    https://...`,
	Example: `  # List all scans for a specific profile
  escape-cli scans list -p 00000000-0000-0000-0000-000000000000

  # List only running scans
  escape-cli scans list --status RUNNING

  # List failed scans from the last week
  escape-cli scans list --status FAILED --after 2025-01-01T00:00:00Z

  # List CI-triggered scans for multiple profiles
  escape-cli scans list -p profile-1,profile-2 -i CI

  # Export scan list to JSON for processing
  escape-cli scans list -o json > scans.json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		scans, next, err := escape.ListScans(cmd.Context(), "", &escape.ListScansFilters{
			ProfileIDs: &scanProfileIDs,
			After:      scanAfter,
			Before:     scanBefore,
			Ignored:    scanIgnored,
			Initiator:  &scanInitiator,
			Kinds:      &scanKinds,
			Status:     &scanStatus,
		})
		if err != nil {
			return fmt.Errorf("unable to list scans: %w", err)
		}
		out.Table(scans, func() []string {
			res := []string{"ID\tCREATED AT\tKIND\tSTATUS\tPROGRESS\tLINK"}
			for _, scan := range scans {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%f\t%s", scan.GetId(), scan.GetCreatedAt(), scan.GetKind(), scan.GetStatus(), scan.GetProgressRatio(), scan.GetLinks().ScanIssues))
			}
			return res
		})

		for next != nil && *next != "" {
			scans, next, err = escape.ListScans(cmd.Context(), *next, &escape.ListScansFilters{
				ProfileIDs: &scanProfileIDs,
				After:      scanAfter,
				Before:     scanBefore,
				Ignored:    scanIgnored,
				Initiator:  &scanInitiator,
				Kinds:      &scanKinds,
				Status:     &scanStatus,
			})

			if err != nil {
				return fmt.Errorf("unable to list scans: %w", err)
			}
			out.Table(scans, func() []string {
				res := []string{
					"ID\tCREATED AT\tKIND\tSTATUS\tPROGRESS\tLINK",
				}
				for _, scan := range scans {
					res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%f\t%s", scan.GetId(), scan.GetCreatedAt(), scan.GetKind(), scan.GetStatus(), scan.GetProgressRatio(), scan.GetLinks().ScanIssues))
				}
				return res
			})
		}
		return nil
	},
}

var scanGetCmd = &cobra.Command{
	Use:     "get scan-id",
	Aliases: []string{"describe", "show", "status"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	Short: "Get detailed information about a specific scan",
	Long: `Get Scan Details - View Status and Metadata

Retrieve detailed information about a specific scan including its current status,
progress, creation time, and results link.

USE CASES:
  • Check if a scan is still running
  • Get the scan results URL
  • Verify scan completion in automation scripts
  • Monitor scan progress

Example output:
ID                                      CREATED AT                           KIND          STATUS      PROGRESS    LINK
00000000-0000-0000-0000-000000000001    2024-11-27 08:06:59.576 +0000 UTC    BLST_REST     FINISHED    1.000000    https://app.escape.tech/...`,
	Example: `  # Get scan status
  escape-cli scans get 00000000-0000-0000-0000-000000000000

  # Get scan in JSON format for scripting
  escape-cli scans get 00000000-0000-0000-0000-000000000000 -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		scan, err := escape.GetScan(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get scan: %w", err)
		}
		out.Table(scan, func() []string {
			res := []string{"ID\tCREATED AT\tKIND\tSTATUS\tPROGRESS\tLINK"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%f\t%s", scan.GetId(), scan.GetCreatedAt(), scan.GetKind(), scan.GetStatus(), scan.GetProgressRatio(), scan.GetLinks().ScanIssues))
			return res
		})
		return nil
	},
}

func extractCommitDataFromEnv() {
	log.Trace("Extracting commit data from environment variables")
	if scanStartCmdCommitHash != "" ||
		scanStartCmdCommitLink != "" ||
		scanStartCmdCommitBranch != "" ||
		scanStartCmdCommitAuthor != "" ||
		scanStartCmdCommitAuthorProfilePictureLink != "" {
		log.Info("Commit data already set, skipping environment variables extraction")
		return
	}

	if os.Getenv("GITHUB_SHA") != "" {
		log.Info("Extracting commit data from GitHub environment variables")
		// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
		scanStartCmdCommitHash = os.Getenv("GITHUB_SHA")
		scanStartCmdCommitBranch = os.Getenv("GITHUB_REF_NAME")
		scanStartCmdCommitAuthor = os.Getenv("GITHUB_ACTOR")
		scanStartCmdCommitAuthorProfilePictureLink = "https://avatars.githubusercontent.com/u/" + os.Getenv("GITHUB_ACTOR_ID") + "?v=4"
		scanStartCmdCommitLink = os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + scanStartCmdCommitHash
		return
	}
	if os.Getenv("GITLAB_CI") != "" {
		log.Info("Extracting commit data from GitLab environment variables")
		// https://docs.gitlab.com/ci/variables/predefined_variables/
		scanStartCmdCommitHash = os.Getenv("CI_COMMIT_SHA")
		scanStartCmdCommitBranch = os.Getenv("CI_COMMIT_REF_NAME")
		scanStartCmdCommitAuthor = os.Getenv("GITLAB_USER_EMAIL")
		scanStartCmdCommitLink = os.Getenv("CI_PROJECT_URL") + "/-/commit/" + scanStartCmdCommitHash
		return
	}
	if os.Getenv("CIRCLE_SHA1") != "" {
		log.Info("Extracting commit data from CircleCI environment variables")
		// https://circleci.com/docs/variables/#built-in-environment-variables
		scanStartCmdCommitHash = os.Getenv("CIRCLE_SHA1")
		scanStartCmdCommitBranch = os.Getenv("CIRCLE_BRANCH")
		scanStartCmdCommitAuthor = os.Getenv("CIRCLE_USERNAME")
		return
	}
	if os.Getenv("COMMIT_HASH") != "" {
		log.Info("Extracting commit data from local environment variables")
		scanStartCmdCommitHash = os.Getenv("COMMIT_HASH")
		scanStartCmdCommitLink = os.Getenv("COMMIT_LINK")
		scanStartCmdCommitBranch = os.Getenv("COMMIT_BRANCH")
		scanStartCmdCommitAuthor = os.Getenv("COMMIT_AUTHOR")
		return
	}

	log.Info("No commit data found in environment variables")
}

func debugCommitData() {
	log.Debug("Commit Hash: %s", scanStartCmdCommitHash)
	log.Debug("Commit Link: %s", scanStartCmdCommitLink)
	log.Debug("Commit Branch: %s", scanStartCmdCommitBranch)
	log.Debug("Commit Author: %s", scanStartCmdCommitAuthor)
	log.Debug("Commit AuthorProfilePictureLink: %s", scanStartCmdCommitAuthorProfilePictureLink)
}

var scanStartCmdCommitHash = ""
var scanStartCmdCommitLink = ""
var scanStartCmdCommitBranch = ""
var scanStartCmdCommitAuthor = ""
var scanStartCmdCommitAuthorProfilePictureLink = ""
var scanStartCmdConfigurationOverride = ""
var scanStartCmdAdditionalProperties = ""
var scanStartCmdWatch bool
var scanStartCmd = &cobra.Command{
	Use: "start profile-id",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("profile ID is required")
		}
		return nil
	},
	Short: "Start a new security scan on a profile",
	Long: `Start Security Scan - Trigger API Security Testing

Launch a new security scan on a configured profile. The scan will analyze your
API for security vulnerabilities, misconfigurations, and potential threats.

COMMIT TRACKING:
  Link scans to your git commits for full traceability. Commit info is auto-detected
  from CI/CD environments (GitHub Actions, GitLab CI, CircleCI) or can be manually specified:
    --commit-hash      Git commit SHA
    --commit-branch    Branch name
    --commit-author    Author name/email
    --commit-link      Link to commit in your VCS

CONFIGURATION OVERRIDE:
  Temporarily override profile settings for a single scan using --override:
    '{"scan": {"read_only": true}}'                    # Non-destructive testing only
    '{"scan": {"timeout": 3600}}'                      # Custom timeout
    '{"scan": {"max_attack_surface": 1000}}'          # Limit endpoints tested

WATCH MODE:
  Use --watch to monitor scan progress in real-time. The command will:
    • Display progress updates as they happen
    • Show final results when complete
    • Exit with appropriate status code for CI/CD

CI/CD INTEGRATION:
  Perfect for automated security testing in your pipeline. The CLI automatically
  detects and uses environment variables from popular CI/CD platforms.`,
	Example: `  # Start a scan and return immediately
  escape-cli scans start 00000000-0000-0000-0000-000000000000

  # Start and watch progress (recommended for CI/CD)
  escape-cli scans start 00000000-0000-0000-0000-000000000000 --watch

  # Start with manual commit tracking
  escape-cli scans start <profile-id> \
    --commit-hash abc123 \
    --commit-branch main \
    --commit-author "john@example.com"

  # Start with configuration override (read-only mode)
  escape-cli scans start <profile-id> \
    --override '{"scan": {"read_only": true}}'

  # GitHub Actions example
  escape-cli scans start $PROFILE_ID \
    --watch \
    --commit-hash $GITHUB_SHA \
    --commit-branch $GITHUB_REF_NAME

  # Start and save scan ID for later use
  SCAN_ID=$(escape-cli scans start <profile-id> -o json | jq -r '.id')`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configurationOverride := map[string]interface{}{}
		if scanStartCmdConfigurationOverride != "" {
			err := json.Unmarshal([]byte(scanStartCmdConfigurationOverride), &configurationOverride)
			if err != nil {
				return fmt.Errorf("unable to unmarshal configuration override: %w", err)
			}
		}
		additionalProperties := map[string]interface{}{}
		if scanStartCmdAdditionalProperties != "" {
			err := json.Unmarshal([]byte(scanStartCmdAdditionalProperties), &additionalProperties)
			if err != nil {
				return fmt.Errorf("unable to unmarshal additional properties: %w", err)
			}
		}
		extractCommitDataFromEnv()
		debugCommitData()
		scan, err := escape.StartScan(
			cmd.Context(),
			args[0],
			scanStartCmdCommitHash,
			scanStartCmdCommitLink,
			scanStartCmdCommitBranch,
			scanStartCmdCommitAuthor,
			scanStartCmdCommitAuthorProfilePictureLink,
			configurationOverride,
			additionalProperties,
			v3.ENUMPROPERTIESDATAITEMSPROPERTIESINITIATORSITEMS_MANUAL,
		)
		if err != nil {
			return fmt.Errorf("unable to start scan: %w", err)
		}
		out.Print(scan, "Scan started: "+scan.GetId())
		if scanStartCmdWatch {
			err := watchScan(cmd.Context(), scan.GetId())
			if err != nil {
				return fmt.Errorf("unable to watch scan: %w", err)
			}
		}
		return nil
	},
}

var scanCancelCmd = &cobra.Command{
	Use: "cancel scan-id",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	Short: "Cancel a running scan",
	Long: `Cancel Running Scan - Stop Scan Execution

Stop a scan that is currently in STARTING or RUNNING state. The scan will be
immediately terminated and marked as CANCELED.

IMPORTANT:
  • Only running scans can be canceled
  • Partial results may be available
  • The scan will still appear in your scan history
  • Cannot be undone - you'll need to start a new scan

USE CASES:
  • Stop a scan that's taking too long
  • Cancel a scan started by mistake
  • Abort scans during emergency situations
  • Clean up stuck scans`,
	Example: `  # Cancel a running scan
  escape-cli scans cancel 00000000-0000-0000-0000-000000000000

  # Cancel multiple scans in a script
  for scan_id in $(escape-cli scans list --status RUNNING -o json | jq -r '.[].id'); do
    escape-cli scans cancel $scan_id
  done`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := escape.CancelScan(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to cancel scan: %w", err)
		}
		out.Log("Scan canceled")
		return nil
	},
}

var scanIgnoreCmd = &cobra.Command{
	Use: "ignore scan-id",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	Short: "Mark a scan as ignored",
	Long: `Ignore Scan - Exclude from Reports and Metrics

Mark a scan as ignored to exclude it from reports, metrics, and trends analysis.
Ignored scans are hidden by default in listings but remain in the system.

WHEN TO IGNORE:
  • Test scans during development
  • Scans with known configuration issues
  • Duplicate or invalid scans
  • Scans that don't represent your production API

NOTE: You can filter ignored scans in listings with --ignored flag`,
	Example: `  # Ignore a test scan
  escape-cli scans ignore 00000000-0000-0000-0000-000000000000

  # View only ignored scans
  escape-cli scans list --ignored true`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := escape.IgnoreScan(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to ignore scan: %w", err)
		}
		out.Log("Scan ignored")
		return nil
	},
}

func watchScan(ctx context.Context, scanID string) error {
	ch, err := escape.WatchScan(ctx, scanID)
	if err != nil {
		return fmt.Errorf("unable to watch scan: %w", err)
	}
	var status *v3.ScanDetailed1
	for event := range ch {
		if event == nil {
			continue
		}
		status = event
		out.Table(event, func() []string {
			res := []string{}
			res = append(res, "STATUS\tPROGRESS")
			res = append(
				res,
				fmt.Sprintf("%s\t%d%%", event.Status, int(event.ProgressRatio*100)), //nolint:mnd
			)
			return res
		})
	}
	if status == nil {
		return errors.New("unable to watch scan")
	} else if status.Status == "CANCELED" {
		out.Log("Scan canceled")
	} else if status.Status == "FAILED" {
		out.Log("Scan failed")
	} else {
		out.Print(status, "Scan completed")
		err := printScanIssues(ctx, scanID)
		if err != nil {
			return fmt.Errorf("unable to fetch scan issues: %w", err)
		}
	}
	return nil
}

var scanWatchCmd = &cobra.Command{
	Use: "watch scan-id",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	Short: "Watch scan progress in real-time",
	Long: `Watch Scan Progress - Monitor Security Scan Execution

Attach to a running scan and monitor its progress in real-time. The command will
display status updates as they occur and exit when the scan completes.

BEHAVIOR:
  • Shows real-time progress percentage
  • Updates as scan progresses through testing phases
  • Displays final results upon completion
  • Exits with status code 0 on success, non-zero on failure

USE CASES:
  • Monitor long-running scans
  • Wait for scan completion in scripts
  • Get immediate feedback on scan progress
  • CI/CD pipelines that need to block until scan completes

The watch will continue until the scan reaches a terminal state:
  FINISHED, FAILED, or CANCELED`,
	Example: `  # Watch a running scan
  escape-cli scans watch 00000000-0000-0000-0000-000000000000

  # Start and watch in one command (recommended)
  escape-cli scans start <profile-id> --watch

  # CI/CD example: Start, watch, and fail if scan fails
  escape-cli scans watch $(escape-cli scans start <profile-id> -o json | jq -r '.id')`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return watchScan(cmd.Context(), args[0])
	},
}

func printScanIssues(ctx context.Context, scanID string) error {
	issues, err := escape.GetScanIssues(ctx, scanID)
	if err != nil {
		return fmt.Errorf("unable to fetch scan issues: %w", err)
	}
	out.Table(issues, func() []string {
		res := []string{"ID\tSEVERITY\tCATEGORY\tNAME\tLINK"}
		for _, i := range issues {
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
				i.GetId(),
				i.GetSeverity(),
				i.GetCategory(),
				i.GetName(),
				i.GetLinks().IssueOverview,
			))
		}
		return res
	})
	return nil
}

var scanIssuesCmd = &cobra.Command{
	Use:     "issues scan-id",
	Aliases: []string{"results", "res", "result", "iss"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	Short: "View security issues found in a scan",
	Long: `View Scan Issues - Review Discovered Vulnerabilities

Display all security issues discovered during a scan. Each issue represents a
potential security vulnerability, misconfiguration, or compliance violation.

ISSUE INFORMATION:
  • ID          - Unique identifier
  • SEVERITY    - CRITICAL, HIGH, MEDIUM, LOW, INFO
  • CATEGORY    - Issue classification (e.g., INJECTION, AUTH, CRYPTO)
  • NAME        - Human-readable description
  • LINK        - Direct URL to detailed analysis

SEVERITY LEVELS:
  🔴 CRITICAL   - Immediate action required, easily exploitable
  🟠 HIGH       - Significant risk, should be fixed soon
  🟡 MEDIUM     - Moderate risk, plan remediation
  🔵 LOW        - Minor issue, fix when convenient
  ⚪ INFO       - Informational, no immediate risk

NEXT STEPS:
  1. Review issues in order of severity
  2. Click the LINK to see detailed remediation steps
  3. Use 'escape-cli issues update' to track progress

Example output:
ID                                      SEVERITY    CATEGORY                  NAME                                LINK
00000000-0000-0000-0000-000000000001    MEDIUM      PROTOCOL                  Insecure Security Policy header     https://...
00000000-0000-0000-0000-000000000002    LOW         INFORMATION_DISCLOSURE    Debug mode enabled                  https://...`,
	Example: `  # View all issues from a scan
  escape-cli scans issues 00000000-0000-0000-0000-000000000000

  # Export issues to JSON for processing
  escape-cli scans issues <scan-id> -o json > issues.json

  # Count critical issues in a scan
  escape-cli scans issues <scan-id> -o json | jq '[.[] | select(.severity == "CRITICAL")] | length'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		err := printScanIssues(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get scan issues: %w", err)
		}
		return nil
	},
}

func init() {
	scansCmd.AddCommand(scansListCmd)
	scansListCmd.PersistentFlags().StringSliceVarP(&scanProfileIDs, "profile-id", "p", []string{}, "filter by profile ID(s) - comma-separated for multiple")
	scansListCmd.PersistentFlags().StringVar(&scanAfter, "after", "", "show scans created after this date (RFC3339 format, e.g., 2025-01-01T00:00:00Z)")
	scansListCmd.PersistentFlags().StringVar(&scanBefore, "before", "", "show scans created before this date (RFC3339 format)")
	scansListCmd.PersistentFlags().StringVar(&scanIgnored, "ignored", "", "filter by ignored status (true/false)")
	scansListCmd.PersistentFlags().StringSliceVarP(&scanInitiator, "initiator", "i", []string{}, "filter by initiator: MANUAL, API, SCHEDULED, CI")
	scansListCmd.PersistentFlags().StringSliceVarP(&scanKinds, "kind", "k", []string{}, "filter by scanner type: BLST_REST, BLST_GRAPHQL, FRONTEND_DAST")
	scansListCmd.PersistentFlags().StringSliceVarP(&scanStatus, "status", "s", []string{}, "filter by status: STARTING, RUNNING, FINISHED, FAILED, CANCELED")
	scanStartCmd.PersistentFlags().BoolVarP(&scanStartCmdWatch, "watch", "w", false, "watch scan progress in real-time until completion")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitHash, "commit-hash", "", "git commit SHA for traceability (auto-detected in CI/CD)")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitLink, "commit-link", "", "URL to commit in your VCS")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitBranch, "commit-branch", "", "git branch name (auto-detected in CI/CD)")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitAuthor, "commit-author", "", "commit author name or email")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitAuthorProfilePictureLink, "profile-picture", "", "URL to author's profile picture")
	scanStartCmd.PersistentFlags().StringVarP(&scanStartCmdConfigurationOverride, "override", "c", "", "JSON configuration override for this scan")
	scansCmd.AddCommand(scanStartCmd)
	scansCmd.AddCommand(scanGetCmd)
	scansCmd.AddCommand(scanIssuesCmd)
	scansCmd.AddCommand(scanWatchCmd)
	scansCmd.AddCommand(scanCancelCmd)
	scansCmd.AddCommand(scanIgnoreCmd)
	rootCmd.AddCommand(scansCmd)
}
