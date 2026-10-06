package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/spf13/cobra"
)

var scanProfileIDs []string
var scanProjectIDs []string
var scanAssetIDs []string
var scanAfter string
var scanBefore string
var scanIgnored string
var scanInitiator []string
var scanKinds []string
var scanStatus []string
var scanSortType string
var scanSortDirection string
var scanListAllKinds bool
var scanListLimit int

var (
	scanKindUsage      = fmt.Sprintf("filter by scanner type: %v", v3.AllowedENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMSEnumValues)
	scanStatusUsage    = fmt.Sprintf("filter by status: %v", v3.AllowedENUMPROPERTIESSTATUSEnumValues)
	scanInitiatorUsage = fmt.Sprintf("filter by initiator: %v", v3.AllowedENUMPROPERTIESINITIATOREnumValues)
)

var defaultScanKinds = []string{
	"BLST_REST",
	"BLST_GRAPHQL",
	"FRONTEND_DAST",
	"AUTOMATED_PENTEST",
}

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
By default DAST and AI Pentesting scan kinds are returned (REST, GraphQL,
WebApp, and automated pentests). Use --all-kinds to include ASM and other kinds.

Filter by profile, status, date range, scanner type, and more.

FILTER OPTIONS:
  -p, --profile-id    Filter by one or more profile IDs
  -s, --status        Filter by scan status (RUNNING, FINISHED, FAILED, CANCELED)
  -k, --kind          Filter by scanner type (overrides the default)
  --all-kinds         Include ASM and all scan kinds (default: DAST + AI Pentesting)
  --limit             Cap how many scans are fetched (0 = no limit)
  -i, --initiator     Filter by who started the scan (MANUAL, API, SCHEDULED, CI)
  --after             Show scans created after this date (RFC3339 format)
  --before            Show scans created before this date (RFC3339 format)
  --ignored           Filter by ignored status (true/false)

SCANNER TYPES:
  • BLST_REST                  - REST API security testing
  • BLST_GRAPHQL               - GraphQL API security testing
  • FRONTEND_DAST              - Web application security testing
  • AUTOMATED_PENTEST          - AI Pentesting

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
		// Output JSON Schema if requested
		if out.Schema([]v3.ScanSummarized{}) {
			return nil
		}

		kinds := resolveScanKinds(cmd)
		filters := &escape.ListScansFilters{
			ProfileIDs:    &scanProfileIDs,
			ProjectIDs:    &scanProjectIDs,
			AssetIDs:      &scanAssetIDs,
			After:         scanAfter,
			Before:        scanBefore,
			Ignored:       scanIgnored,
			Initiator:     &scanInitiator,
			Kinds:         scanKindsFilter(kinds),
			Status:        &scanStatus,
			SortType:      scanSortType,
			SortDirection: scanSortDirection,
		}
		allScans, err := fetchAllScans(cmd.Context(), filters, scanListLimit, isPrettyOutput())
		if err != nil {
			return err
		}

		if len(scanProfileIDs) > 0 && len(allScans) == 0 {
			return fmt.Errorf(
				"no scans found for profile ID(s) %s; verify the profile exists with profiles get",
				strings.Join(scanProfileIDs, ", "),
			)
		}

		out.Table(allScans, func() []string {
			res := []string{"ID\tCREATED AT\tKIND\tSTATUS\tPROGRESS\tLINK"}
			for _, scan := range allScans {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%f\t%s", scan.GetId(), scan.GetCreatedAt(), scan.GetKind(), scan.GetStatus(), scan.GetProgressRatio(), scan.GetLinks().ScanIssues))
			}

			return res
		})

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
		// Output JSON Schema if requested
		if out.Schema(v3.StartScan200Response{}) {
			return nil
		}

		scan, err := escape.GetScan(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get scan: %w", err)
		}

		out.Table(scan, func() []string {
			res := []string{"ID\tCREATED AT\tFINISHED AT\tKIND\tSTATUS\tPROGRESS\tSCORE\tCOVERAGE\tDURATION\tPROFILE ID\tORG ID\tCOMMIT BRANCH\tCOMMIT HASH\tCOMMIT AUTHOR\tLINK"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d%%\t%.0f\t%.0f%%\t%.0fs\t%s\t%s\t%s\t%s\t%s\t%s",
				scan.GetId(),
				scan.GetCreatedAt(),
				scan.GetFinishedAt(),
				scan.GetKind(),
				scan.GetStatus(),
				int(scan.GetProgressRatio()*100), //nolint:mnd
				scan.GetScore(),
				scan.GetCoverage()*100, //nolint:mnd
				scan.GetDuration(),
				scan.GetProfileId(),
				scan.GetOrganizationId(),
				scan.GetCommitBranch(),
				scan.GetCommitHash(),
				scan.GetCommitAuthor(),
				scan.GetLinks().ScanIssues,
			))

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

// scanFailOnSeverity is the canonical issue severity for the watch gate.
// Empty means the gate is off. scans start and scans watch share it because
// one process runs one command.
var scanFailOnSeverity string

// issueSeverityOrder is least to most severe. The generated Allowed list is
// alphabetical, so it is not a rank. A test checks that every generated value
// has a place here.
var issueSeverityOrder = []v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY{
	v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY_INFO,
	v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY_LOW,
	v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY_MEDIUM,
	v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY_HIGH,
	v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY_CRITICAL,
}

// severityLadder renders issueSeverityOrder as flag help, least to most
// severe. The generated Allowed list is alphabetical and would read as a rank.
func severityLadder() string {
	levels := make([]string, len(issueSeverityOrder))
	for i, level := range issueSeverityOrder {
		levels[i] = string(level)
	}

	return strings.Join(levels, " < ")
}

var failOnSeverityUsage = "exit non-zero when a finished scan has an open issue at or above this severity (requires --watch on scans start): " + severityLadder()

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
  Temporarily override profile settings for a single scan using --override.
  JSON must match the Scanner Next config schema (top-level fields, not a "scan" wrapper):
    '{"mode": "read_only"}'                           # Non-destructive testing only
    '{"max_duration": 3600}'                          # Custom max duration (seconds)

WATCH MODE:
  Use --watch to block until the scan ends. Progress updates print as they
  happen and the exit code follows the scans watch contract: 0 only when
  the scan finishes and no severity gate fails; non-zero when the scan
  fails ("scan <scan-id> failed"), is canceled ("scan <scan-id> was
  canceled"), or never reaches a terminal status ("scan <scan-id> did not
  finish (last status X)").

  --fail-on-severity is rejected unless --watch is set. It exits non-zero
  when a finished scan has an open issue at or above LEVEL (severity rules
  in scans watch).

  The process error is written to stderr and the exit code is 1. JSON mode
  does not add an error document. --watch -o json prints one document, the
  final scan; issues are read for the severity gate and are not printed.

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
    --override '{"mode": "read_only"}'

  # GitHub Actions example
  escape-cli scans start $PROFILE_ID \
    --watch \
    --commit-hash $GITHUB_SHA \
    --commit-branch $GITHUB_REF_NAME

  # Fail the pipeline on a high or critical finding
  escape-cli scans start <profile-id> --watch --fail-on-severity HIGH

  # Start and save scan ID for later use
  SCAN_ID=$(escape-cli scans start <profile-id> -o json | jq -r '.id')`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output JSON Schema if requested
		if out.Schema(v3.ScanDetailed1{}) {
			return nil
		}

		if err := validateScanFailOnSeverity(scanStartCmdWatch); err != nil {
			return err
		}

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

		started := "Scan started\n  ID:   " + scan.GetId()
		if link := scan.GetLinks().ScanIssues; link != "" {
			started += "\n  View: " + out.LinkURL(link)
		}

		// --watch -o json prints the finished scan once. The create response
		// is still STARTING, and a second document would break jq. A failed,
		// canceled, or unfinished scan returns the watch error after that
		// document.
		if scanStartCmdWatch && out.IsJSON() {
			return watchScan(cmd.Context(), scan.GetId(), watchJSONStatus)
		}

		out.Print(scan, started)
		if scanStartCmdWatch {
			return watchScan(cmd.Context(), scan.GetId(), watchJSONSilent)
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

		out.Print(out.Message{Msg: "Scan canceled"}, "Scan canceled")

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

		out.Print(out.Message{Msg: "Scan ignored"}, "Scan ignored")

		return nil
	},
}

// watchJSONResult selects the JSON document written when a watch ends.
// Pretty mode ignores it and keeps the progress tables.
type watchJSONResult int

const (
	// watchJSONSilent writes nothing. Pretty `scans start --watch` already
	// printed the create response and uses the tables for progress.
	watchJSONSilent watchJSONResult = iota
	// watchJSONStatus writes one document: the final scan. That is the
	// `scans start --watch -o json` contract.
	watchJSONStatus
	// watchJSONStatusAndIssues writes the final scan, then the issue list
	// when the scan finished. That is the `scans watch -o json` contract.
	watchJSONStatusAndIssues
)

// watchScan follows a scan until it reaches a terminal status.
func watchScan(ctx context.Context, scanID string, jsonResult watchJSONResult) error {
	ch, err := escape.WatchScan(ctx, scanID)
	if err != nil {
		return fmt.Errorf("unable to watch scan: %w", err)
	}

	return followWatch(ctx, scanID, ch, jsonResult)
}

// followWatch consumes WatchScan events until the stream closes, printing
// progress in pretty mode, then hands the last status to finishWatch. The
// stream can close without a terminal status when the API stops answering;
// finishWatch turns that into an error instead of success.
func followWatch(ctx context.Context, scanID string, ch <-chan *v3.StartScan200Response, jsonResult watchJSONResult) error {
	jsonMode := out.IsJSON()
	var status *v3.StartScan200Response
	for event := range ch {
		if event == nil {
			continue
		}

		status = event
		if jsonMode {
			continue
		}

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
	}

	return finishWatch(ctx, scanID, status, jsonResult)
}

// validateScanFailOnSeverity checks --fail-on-severity before any scan work.
// scans start accepts the flag only together with --watch. Input is matched
// against the generated issue severity enum without case and stored in
// canonical form.
func validateScanFailOnSeverity(watch bool) error {
	scanFailOnSeverity = strings.TrimSpace(scanFailOnSeverity)
	if scanFailOnSeverity == "" {
		return nil
	}

	if !watch {
		return errors.New("--fail-on-severity requires --watch")
	}

	level, err := parseIssueSeverity(scanFailOnSeverity)
	if err != nil {
		return err
	}

	scanFailOnSeverity = string(level)

	return nil
}

func parseIssueSeverity(value string) (v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY, error) {
	return parseStringEnum(
		strings.ToUpper(strings.TrimSpace(value)),
		"severity",
		v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues,
	)
}

func severityRank(level v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY) (int, bool) {
	for rank, candidate := range issueSeverityOrder {
		if candidate == level {
			return rank, true
		}
	}

	return 0, false
}

// issueCountsTowardGate reports whether an issue can fail --fail-on-severity.
// OPEN and MANUAL_REVIEW are still open. IGNORED, RESOLVED, and FALSE_POSITIVE
// are not. An unknown status still counts, so a new status cannot hide a finding.
func issueCountsTowardGate(issue v3.IssueSummarized) bool {
	switch issue.GetStatus() {
	case v3.ENUMPROPERTIESFILTERPROPERTIESSTATUSITEMS_OPEN,
		v3.ENUMPROPERTIESFILTERPROPERTIESSTATUSITEMS_MANUAL_REVIEW:
		return true
	case v3.ENUMPROPERTIESFILTERPROPERTIESSTATUSITEMS_IGNORED,
		v3.ENUMPROPERTIESFILTERPROPERTIESSTATUSITEMS_RESOLVED,
		v3.ENUMPROPERTIESFILTERPROPERTIESSTATUSITEMS_FALSE_POSITIVE:
		return false
	default:
		return true
	}
}

// severityGate fails when an open issue is at least as severe as the flag.
// validateScanFailOnSeverity has already stored a canonical level, so the
// floor is always a known rank. The message keeps the literal "issue(s)" so
// callers can match it.
func severityGate(issues []v3.IssueSummarized) error {
	if scanFailOnSeverity == "" {
		return nil
	}

	floor, _ := severityRank(v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY(scanFailOnSeverity))
	count := 0
	for _, issue := range issues {
		if !issueCountsTowardGate(issue) {
			continue
		}

		rank, known := severityRank(issue.GetSeverity())
		if !known || rank >= floor {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	return fmt.Errorf("%d issue(s) at or above %s", count, scanFailOnSeverity)
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
  • Exits 0 when the scan finishes and no severity gate fails
  • Exits non-zero when the scan fails ("scan <scan-id> failed"), is
    canceled ("scan <scan-id> was canceled"), or the stream ends before a
    terminal status ("scan <scan-id> did not finish (last status X)")
  • --fail-on-severity LEVEL also exits non-zero when a finished scan has an
    open issue at or above LEVEL. The error is "N issue(s) at or above LEVEL".
    LEVEL is INFO, LOW, MEDIUM, HIGH, CRITICAL, matched without case.
    OPEN and MANUAL_REVIEW issues count. Resolved, ignored, and false-positive
    issues do not.

JSON output keeps the existing documents. The error is written to stderr and
is not another JSON document. -o json prints the final scan, then the issue
list, when the scan finishes. A failed, canceled, or unfinished scan prints
only the scan document. scans start --watch -o json prints only the final
scan; issues are read for --fail-on-severity and are not printed.

USE CASES:
  • Monitor long-running scans
  • Wait for scan completion in scripts
  • Get immediate feedback on scan progress
  • CI/CD pipelines that need to block until scan completes

The watch polls until the scan reaches a terminal state (FINISHED, FAILED,
or CANCELED) or the API stops answering.`,
	Example: `  # Watch a running scan
  escape-cli scans watch 00000000-0000-0000-0000-000000000000

  # Start and watch in one command (recommended)
  escape-cli scans start <profile-id> --watch

  # CI/CD example: Start, watch, and fail if the scan fails
  escape-cli scans watch $(escape-cli scans start <profile-id> -o json | jq -r '.id')

  # Fail when the finished scan has a high or critical issue
  escape-cli scans watch <scan-id> --fail-on-severity HIGH`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output JSON Schema if requested
		if out.Schema(v3.ScanDetailed1{}) {
			return nil
		}

		if err := validateScanFailOnSeverity(true); err != nil {
			return err
		}

		return watchScan(cmd.Context(), args[0], watchJSONStatusAndIssues)
	},
}

var scanIssuesPage pageFlags
var scanProblemsPage pageFlags

// finishWatch prints the terminal scan and returns the process error.
// FAILED and CANCELED fail in pretty and JSON mode. JSON prints the scan
// document first and does not print issues for those statuses. Only a
// finished scan succeeds: any other last status (STARTING, RUNNING, PENDING,
// or an unknown value — the stream can close without a terminal status when
// the API stops answering) fails the run the same way.
// A finished scan prints issues: a pretty table, or the second JSON document
// for `scans watch`. `scans start --watch -o json` stays one scan document.
// --fail-on-severity then returns "N issue(s) at or above LEVEL".
// main writes the error to stderr and exits 1, so stdout stays the documents.
func finishWatch(ctx context.Context, scanID string, status *v3.StartScan200Response, jsonResult watchJSONResult) error {
	switch v3.ENUMPROPERTIESSTATUS(status.Status) {
	case v3.ENUMPROPERTIESSTATUS_FINISHED, v3.ENUMPROPERTIESSTATUS_COMPLETED:
		return finishSuccessfulWatch(ctx, scanID, status, jsonResult)
	case v3.ENUMPROPERTIESSTATUS_FAILED:
		writeWatchFailure(status, jsonResult, "Scan failed")
		return fmt.Errorf("scan %s failed", scanID)
	case v3.ENUMPROPERTIESSTATUS_CANCELED:
		writeWatchFailure(status, jsonResult, "Scan canceled")
		return fmt.Errorf("scan %s was canceled", scanID)
	case v3.ENUMPROPERTIESSTATUS_PENDING, v3.ENUMPROPERTIESSTATUS_RUNNING, v3.ENUMPROPERTIESSTATUS_STARTING:
		// The stream closed while the scan was still queued or running.
		fallthrough
	default:
		// An unknown status is not a terminal one either.
		writeWatchFailure(status, jsonResult, fmt.Sprintf("Scan did not finish (status %s)", status.Status))
		return fmt.Errorf("scan %s did not finish (last status %s)", scanID, status.Status)
	}
}

// writeWatchFailure prints the terminal failure without an issue list.
// JSON silent mode leaves the caller's document alone. Pretty mode logs a
// status line. The returned error is what makes the process exit non-zero.
func writeWatchFailure(status *v3.StartScan200Response, jsonResult watchJSONResult, prettyLine string) {
	if out.IsJSON() {
		if jsonResult == watchJSONSilent {
			return
		}

		out.Print(status, "")

		return
	}

	out.Log(prettyLine)
}

// finishSuccessfulWatch prints a finished scan, then applies the severity gate.
// Issues are fetched once. They are printed in pretty mode and for
// `scans watch -o json`. `scans start --watch -o json` reads them only when
// the gate is set, and still prints just the scan document.
func finishSuccessfulWatch(ctx context.Context, scanID string, status *v3.StartScan200Response, jsonResult watchJSONResult) error {
	printIssues := !out.IsJSON() || jsonResult == watchJSONStatusAndIssues
	if out.IsJSON() {
		if jsonResult != watchJSONSilent {
			out.Print(status, "")
		}
	} else {
		out.Print(status, "Scan completed")
	}

	if !printIssues && scanFailOnSeverity == "" {
		return nil
	}

	issues, _, _, err := fetchScanIssues(ctx, scanID, false, "", 0)
	if err != nil {
		return err
	}

	if printIssues {
		emitScanIssues(false, issues, nil, 0)
	}

	return severityGate(issues)
}

func fetchScanIssues(ctx context.Context, scanID string, single bool, cursor string, size int) ([]v3.IssueSummarized, *string, int, error) {
	issues, next, total, err := resolveList(ctx, single, cursor, size, func(ctx context.Context, cursor string, size int) ([]v3.IssueSummarized, *string, int, error) {
		return escape.GetScanIssues(ctx, scanID, cursor, size)
	})
	if err != nil {
		return nil, nil, 0, fmt.Errorf("unable to fetch scan issues: %w", err)
	}

	return issues, next, total, nil
}

func emitScanIssues(single bool, issues []v3.IssueSummarized, next *string, total int) {
	emitList(single, issues, next, total, func() []string {
		res := []string{"ID\tSEVERITY\tCATEGORY\tNAME\tLINK"}
		for _, issue := range issues {
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
				issue.GetId(),
				issue.GetSeverity(),
				issue.GetCategory(),
				issue.GetName(),
				issue.GetLinks().IssueOverview,
			))
		}

		return res
	})
}

func printScanIssues(ctx context.Context, scanID string, single bool, cursor string, size int) error {
	issues, next, total, err := fetchScanIssues(ctx, scanID, single, cursor, size)
	if err != nil {
		return err
	}

	emitScanIssues(single, issues, next, total)

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
		// Output JSON Schema if requested
		if out.Schema([]v3.IssueSummarized{}) {
			return nil
		}

		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}

		if err := validatePageFlags(cmd, scanIssuesPage.size); err != nil {
			return err
		}

		err := printScanIssues(cmd.Context(), args[0], singlePageRequested(cmd), scanIssuesPage.cursor, scanIssuesPage.size)
		if err != nil {
			return fmt.Errorf("unable to get scan issues: %w", err)
		}

		return nil
	},
}

var (
	scanTargetsType string
	scanTargetsSize int
)

var scanTargetsCmd = &cobra.Command{
	Use:     "targets scan-id",
	Aliases: []string{"target"},
	Short:   "List API targets discovered during a scan. For per-user request success, use scans coverage.",
	Long: `List Scan Targets - View Discovered Endpoints

Display all API endpoints and GraphQL resolvers discovered and tested
during a scan. Useful for coverage analysis and CI/CD quality gates.

FILTER OPTIONS:
  --type    Filter by target type: API_ROUTE, GRAPHQL_RESOLVER
  --size    Total cap on results (0 = fetch every page; not a page size). Prefer 'scans coverage' for per-user coverage questions.`,
	Example: `  # List all targets for a scan
  escape-cli scans targets <scan-id>

  # List only REST endpoints
  escape-cli scans targets <scan-id> --type API_ROUTE

  # Export to JSON for coverage reporting
  escape-cli scans targets <scan-id> -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]v3.TargetDetailed{}) {
			return nil
		}

		if scanTargetsSize < 0 {
			return errors.New("--size must be greater than or equal to 0")
		}

		all, _, _, err := listScanTargets(cmd.Context(), args[0], scanTargetsType, scanTargetsSize)
		if err != nil {
			return fmt.Errorf("unable to list targets: %w", err)
		}

		out.Table(all, func() []string {
			res := []string{"ID\tTYPE\tMETHOD\tPATH/RESOLVER\tCOVERAGE\tREQUEST COUNT\tMEAN DURATION MS"}
			for _, t := range all {
				row := compactScanTarget(t)
				coverage := row.Coverage
				if coverage == "" {
					coverage = "-"
				}

				var meanDuration float32
				if route, ok := t.GetApiRouteOk(); ok && route != nil {
					meanDuration = route.GetMeanDuration()
				} else if resolver, ok := t.GetGraphqlResolverOk(); ok && resolver != nil {
					meanDuration = resolver.GetMeanDuration()
				}

				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%.0f", t.GetId(), row.Type, row.Method, row.Name, coverage, row.RequestCount, meanDuration))
			}

			return res
		})

		return nil
	},
}

var scansProblemsCmd = &cobra.Command{
	Use:     "problems",
	Aliases: []string{"pb"},
	Short:   "List scans with their problems",
	Long: `List Scans with Problems - Identify Failing Scans

Display scans whose execution surfaced validation problems (broken authentication,
unreachable target, schema invalid, etc.). Diagnostic counterpart of
'escape-cli profiles problems' but indexed by scan instead of profile.`,
	Example: `  # List all scans with problems
  escape-cli scans problems

  # Filter by profile
  escape-cli scans problems --profile-id <profile-id>

  # Filter by status
  escape-cli scans problems --status FAILED

  # Export to JSON
  escape-cli scans problems -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[v3.ScanSummarizedWithProblems2]{}) {
			return nil
		}

		kinds := resolveScanKinds(cmd)
		filters := &escape.ListScanProblemsFilters{
			After:         scanAfter,
			Before:        scanBefore,
			AssetIDs:      scanAssetIDs,
			ProfileIDs:    scanProfileIDs,
			ProjectIDs:    scanProjectIDs,
			Ignored:       scanIgnored,
			Initiator:     scanInitiator,
			Kinds:         kinds,
			Status:        scanStatus,
			SortType:      scanSortType,
			SortDirection: scanSortDirection,
		}

		return runPagedList(cmd, scanProblemsPage, func(ctx context.Context, cursor string, size int) ([]v3.ScanSummarizedWithProblems2, *string, int, error) {
			scans, next, total, err := escape.ListScanProblems(ctx, cursor, filters, size)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("unable to list scan problems: %w", err)
			}

			return scans, next, total, nil
		}, func(scans []v3.ScanSummarizedWithProblems2) []string {
			res := []string{"SCAN ID\tSTATUS\tKIND\tINITIATOR\tPROBLEMS\tCREATED AT\tLINK"}
			for _, scan := range scans {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%s\t%s", scan.GetId(), scan.GetStatus(), scan.GetKind(), scan.GetInitiator(), len(scan.GetProblems()), scan.GetCreatedAt(), scan.GetLinks().ScanIssues))
			}

			return res
		})
	},
}

func resolveScanKinds(cmd *cobra.Command) []string {
	if scanListAllKinds && cmd.Flags().Changed("kind") {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: --all-kinds is ignored when --kind is explicitly set")
	}

	if cmd.Flags().Changed("kind") {
		return scanKinds
	}

	if scanListAllKinds {
		return nil
	}

	return defaultScanKinds
}

func scanKindsFilter(kinds []string) *[]string {
	if kinds == nil {
		return nil
	}

	return &kinds
}

func fetchAllScans(
	ctx context.Context,
	filters *escape.ListScansFilters,
	limit int,
	progress bool,
) ([]v3.ScanSummarized2, error) {
	var allScans []v3.ScanSummarized2
	next := ""
	page := 0
	for {
		page++
		if progress {
			out.Log(fmt.Sprintf("Fetching scans (page %d, %d loaded)...", page, len(allScans)))
		}

		scans, cursor, err := escape.ListScans(ctx, next, filters)
		if err != nil {
			return nil, fmt.Errorf("unable to list scans: %w", err)
		}

		allScans = append(allScans, scans...)
		if limit > 0 && len(allScans) >= limit {
			return allScans[:limit], nil
		}

		if cursor == nil || *cursor == "" {
			break
		}

		next = *cursor
	}

	return allScans, nil
}

func init() {
	scansCmd.AddCommand(scansListCmd)
	scansListCmd.Flags().BoolVar(&scanListAllKinds, "all-kinds", false, "include ASM and all scan kinds (default: DAST and AI Pentesting kinds only)")
	scansListCmd.Flags().IntVar(&scanListLimit, "limit", 0, "maximum number of scans to return (0 = no limit)")
	// Local flags: these leaves have no subcommands, so nothing inherits them.
	// The MCP catalog reads LocalFlags, which covers both registrations.
	scansListCmd.Flags().StringSliceVarP(&scanProfileIDs, "profile-id", "p", []string{}, "filter by profile ID(s) - comma-separated for multiple")
	scansListCmd.Flags().StringSliceVar(&scanProjectIDs, "project-id", []string{}, "filter by project ID(s)")
	scansListCmd.Flags().StringSliceVarP(&scanAssetIDs, "asset-id", "a", []string{}, "filter by asset ID(s) - comma-separated for multiple")
	scansListCmd.Flags().StringVar(&scanAfter, "after", "", "show scans created after this date (RFC3339 format, e.g., 2025-01-01T00:00:00Z)")
	scansListCmd.Flags().StringVar(&scanBefore, "before", "", "show scans created before this date (RFC3339 format)")
	scansListCmd.Flags().StringVar(&scanIgnored, "ignored", "", "filter by ignored status (true/false)")
	scansListCmd.Flags().StringSliceVarP(&scanInitiator, "initiator", "i", []string{}, scanInitiatorUsage)
	scansListCmd.Flags().StringSliceVarP(&scanKinds, "kind", "k", []string{}, scanKindUsage)
	scansListCmd.Flags().StringSliceVarP(&scanStatus, "status", "s", []string{}, scanStatusUsage)
	scansListCmd.Flags().StringVar(&scanSortType, "sort-by", "", "sort field (e.g., createdAt)")
	scansListCmd.Flags().StringVar(&scanSortDirection, "sort-direction", "", "sort direction: asc, desc")
	scanStartCmd.Flags().BoolVarP(&scanStartCmdWatch, "watch", "w", false, "watch scan progress in real-time until completion")
	markMCPSkip(scanStartCmd.Flags(), "watch")
	scanStartCmd.Flags().StringVar(&scanFailOnSeverity, "fail-on-severity", "", failOnSeverityUsage)
	markMCPSkip(scanStartCmd.Flags(), "fail-on-severity")
	scanStartCmd.Flags().StringVar(&scanStartCmdCommitHash, "commit-hash", "", "git commit SHA for traceability (auto-detected in CI/CD)")
	scanStartCmd.Flags().StringVar(&scanStartCmdCommitLink, "commit-link", "", "URL to commit in your VCS")
	scanStartCmd.Flags().StringVar(&scanStartCmdCommitBranch, "commit-branch", "", "git branch name (auto-detected in CI/CD)")
	scanStartCmd.Flags().StringVar(&scanStartCmdCommitAuthor, "commit-author", "", "commit author name or email")
	scanStartCmd.Flags().StringVar(&scanStartCmdCommitAuthorProfilePictureLink, "profile-picture", "", "URL to author's profile picture")
	scanStartCmd.Flags().StringVarP(&scanStartCmdConfigurationOverride, "override", "c", "", "JSON configuration override for this scan")
	scanStartCmd.Flags().StringVar(&scanStartCmdAdditionalProperties, "additional-properties", "", "JSON additional properties for the scan request")
	scansCmd.AddCommand(scanStartCmd)
	scansCmd.AddCommand(scanGetCmd)
	scansCmd.AddCommand(scanIssuesCmd)
	scanIssuesPage.bind(scanIssuesCmd)
	scanWatchCmd.Flags().StringVar(&scanFailOnSeverity, "fail-on-severity", "", failOnSeverityUsage)
	markMCPSkip(scanWatchCmd.Flags(), "fail-on-severity")
	scansCmd.AddCommand(scanWatchCmd)
	scansCmd.AddCommand(scanCancelCmd)
	scansCmd.AddCommand(scanIgnoreCmd)
	scansCmd.AddCommand(scansReasoningCmd)
	scansCmd.AddCommand(scansAgentsCmd)
	scansReasoningCmd.Flags().StringVar(&reasoningAgentID, "agent-id", "", "limit logs to a single agent")
	scansReasoningCmd.Flags().StringVar(&reasoningSearch, "search", "", "search in log titles and descriptions")
	scansReasoningCmd.Flags().IntVar(
		&reasoningListLimit,
		"list-limit",
		defaultReasoningListLimit,
		fmt.Sprintf("maximum number of event summaries to fetch from the API (max %d)", maxReasoningListLimit),
	)
	scansReasoningCmd.Flags().IntVar(
		&reasoningHydrateLimit,
		"hydrate-limit",
		defaultReasoningHydrateLimit,
		fmt.Sprintf("maximum number of events to hydrate with full descriptions (0 = summaries only, max %d)", maxReasoningHydrateLimit),
	)
	scansAgentsCmd.Flags().StringVar(&scanAgentsSearch, "search", "", "filter agents by title")
	scansAgentsCmd.Flags().StringVar(&scanAgentsEventSearch, "event-search", "", "filter agents by text found in their reasoning logs")
	scansAgentsCmd.Flags().BoolVar(&scanAgentsRootsOnly, "roots-only", false, "return only root agents")
	scansCmd.AddCommand(scanTargetsCmd)
	scanTargetsCmd.Flags().StringVar(&scanTargetsType, "type", "", "filter by target type: API_ROUTE, GRAPHQL_RESOLVER")
	scanTargetsCmd.Flags().IntVar(
		&scanTargetsSize,
		"size",
		0,
		"total cap on returned targets, not a page size (0 = fetch every page)",
	)
	scansCmd.AddCommand(scansCoverageCmd)
	scansCoverageCmd.Flags().StringVar(&scanCoverageType, "type", "", "filter by target type: API_ROUTE, GRAPHQL_RESOLVER")
	scansCoverageCmd.Flags().StringVar(
		&scanCoverageStatus,
		"coverage",
		"",
		"filter the targets sample by coverage status (e.g. OK, SKIPPED). With --user, match that user's status, not overall. overall and byUser stay exhaustive.",
	)
	scansCoverageCmd.Flags().StringVar(
		&scanCoverageUser,
		"user",
		"",
		"filter the targets sample to routes that include this scanner user. overall and byUser still cover every route.",
	)
	scansCoverageCmd.Flags().IntVar(
		&scanCoverageSize,
		"size",
		defaultCoverageTargetListSize,
		fmt.Sprintf("max compact routes to return in targets (0 = all matching, max %d)", maxCoverageTargetListSize),
	)
	scansCmd.AddCommand(scansProblemsCmd)
	scanProblemsPage.bind(scansProblemsCmd)
	scansProblemsCmd.Flags().BoolVar(&scanListAllKinds, "all-kinds", false, "include ASM and all scan kinds (default: DAST and AI Pentesting kinds only)")
	scansProblemsCmd.PersistentFlags().StringSliceVarP(&scanProfileIDs, "profile-id", "p", []string{}, "filter by profile ID(s) - comma-separated for multiple")
	scansProblemsCmd.PersistentFlags().StringSliceVar(&scanProjectIDs, "project-id", []string{}, "filter by project ID(s)")
	scansProblemsCmd.PersistentFlags().StringSliceVarP(&scanAssetIDs, "asset-id", "a", []string{}, "filter by asset ID(s) - comma-separated for multiple")
	scansProblemsCmd.PersistentFlags().StringVar(&scanAfter, "after", "", "show scans created after this date (RFC3339 format, e.g., 2025-01-01T00:00:00Z)")
	scansProblemsCmd.PersistentFlags().StringVar(&scanBefore, "before", "", "show scans created before this date (RFC3339 format)")
	scansProblemsCmd.PersistentFlags().StringVar(&scanIgnored, "ignored", "", "filter by ignored status (true/false)")
	scansProblemsCmd.PersistentFlags().StringSliceVarP(&scanInitiator, "initiator", "i", []string{}, scanInitiatorUsage)
	scansProblemsCmd.PersistentFlags().StringSliceVarP(&scanKinds, "kind", "k", []string{}, scanKindUsage)
	scansProblemsCmd.PersistentFlags().StringSliceVarP(&scanStatus, "status", "s", []string{}, scanStatusUsage)
	scansProblemsCmd.PersistentFlags().StringVar(&scanSortType, "sort-by", "", "sort field (e.g., createdAt)")
	scansProblemsCmd.PersistentFlags().StringVar(&scanSortDirection, "sort-direction", "", "sort direction: asc, desc")
	rootCmd.AddCommand(scansCmd)
}
