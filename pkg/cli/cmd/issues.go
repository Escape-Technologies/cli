package cmd

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// maxLatestEventsHydrationConcurrency caps how many events we hydrate in
// parallel from `issues get-with-events`. Matches the public API's hard cap of
// 5 latest event IDs per issue, so this is also the worst-case fan-out.
const maxLatestEventsHydrationConcurrency = 5

func formatLatestEventIDs(ids []string, truncated bool) string {
	if len(ids) == 0 {
		return "-"
	}
	joined := strings.Join(ids, ", ")
	if truncated {
		joined += ", …"
	}
	return joined
}

// IssueWithEvents bundles an issue with its hydrated latest-scan events so the
// MCP/CLI can hand an AI agent the full context (including each event's
// Exchange) in a single tool call.
type IssueWithEvents struct {
	Issue                 v3.GetIssue200Response   `json:"issue"`
	LatestEvents          []v3.GetEvent200Response `json:"latestEvents"`
	LatestEventsTruncated bool                     `json:"latestEventsTruncated"`
	EventErrors           []IssueEventHydrateError `json:"eventErrors,omitempty"`
}

// IssueEventHydrateError captures a per-event hydration failure so partial
// success is observable rather than swallowed.
type IssueEventHydrateError struct {
	EventID string `json:"eventId"`
	Error   string `json:"error"`
}

func formatIssueCompliances(items []v3.GetIssue200ResponseCompliancesInner) string {
	if len(items) == 0 {
		return "-"
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		if item.GetFramework() == "" && item.GetItem() == "" {
			continue
		}
		values = append(values, strings.Trim(strings.Join([]string{item.GetFramework(), item.GetItem()}, ":"), ":"))
	}
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

var validIssueSortFields = map[string]struct{}{
	"LAST_SEEN":  {},
	"FIRST_SEEN": {},
	"SEVERITY":   {},
	"STATUS":     {},
}

var (
	issueUpdateStatusStr string
	issueUpdateComment   string
	issueSortType        string
	issueSortDirection   string
	issueSeverity        []string
	issueStatus          []string
	profileIDs           []string
	assetIDs             []string
	domains              []string
	issueIDs             []string
	scanIDs              []string
	tagsIDs              []string
	search               string
	jiraTicket           string
	risks                []string
	assetClasses         []string
	issueScannerKinds    []string
	issueNames           []string
)

var issuesCmd = &cobra.Command{
	Use:     "issues",
	Aliases: []string{"issue"},
	Short:   "Manage and track security vulnerabilities",
	Long: `Manage Security Issues - Track and Remediate Vulnerabilities

Issues are security vulnerabilities, misconfigurations, and compliance violations
discovered during security scans. Each issue represents a specific security concern
that should be reviewed and remediated.

ISSUE LIFECYCLE:
  1. OPEN              - Newly discovered, needs review
  2. MANUAL_REVIEW     - Under investigation
  3. IN_PROGRESS       - Actively being fixed
  4. RESOLVED          - Fixed and verified
  5. FALSE_POSITIVE    - Not a real issue
  6. ACCEPTED_RISK     - Acknowledged but not fixing

COMMON WORKFLOWS:
  • List high-priority issues:
    $ escape-cli issues list --severity HIGH,CRITICAL --status OPEN

  • Review issues for a specific asset:
    $ escape-cli issues list --asset-id <asset-id>

  • Update issue status as you fix them:
    $ escape-cli issues update <issue-id> --status IN_PROGRESS

  • Track issue history:
    $ escape-cli issues list-activities <issue-id>`,
}

var issueListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List security issues with powerful filtering",
	Long: `List Security Issues - Query Your Vulnerability Database

List and filter security issues across your entire organization. Use powerful
filtering options to find exactly the issues you need to review or remediate.

FILTER OPTIONS:
  --severity         Filter by severity: CRITICAL, HIGH, MEDIUM, LOW, INFO
  --status           Filter by status: OPEN, MANUAL_REVIEW, IN_PROGRESS, RESOLVED
  -p, --profile-id   Filter by profile ID
  -a, --asset-id     Filter by asset ID  
  -d, --domain       Filter by domain name
  -i, --issue-id     Filter by specific issue IDs
  --scan-id          Filter by scan ID
  -t, --tag-id       Filter by tag ID
  -r, --risk         Filter by risk level
  --asset-class      Filter by asset classification
  -s, --search       Free-text search across issue names

SEVERITY PRIORITY:
  🔴 CRITICAL   - Critical security flaws, immediate action required
  🟠 HIGH       - Serious vulnerabilities, fix ASAP
  🟡 MEDIUM     - Moderate risk, schedule fixes
  🔵 LOW        - Minor issues, address when possible
  ⚪ INFO       - Informational findings

Example output:
ID                                      CREATED AT  SEVERITY  STATUS  NAME                        ASSET           LINK
00000000-0000-0000-0000-000000000001    2025-07-24  HIGH      OPEN    SQL Injection               api.example.com https://...
00000000-0000-0000-0000-000000000002    2025-07-23  MEDIUM    OPEN    Misconfigured CSP Header    my.app          https://...`,
	Example: `  # List all open critical and high severity issues
  escape-cli issues list --severity CRITICAL,HIGH --status OPEN

  # List issues for a specific asset
  escape-cli issues list --asset-id 00000000-0000-0000-0000-000000000000

  # Search for specific vulnerability types
  escape-cli issues list --search "SQL injection"

  # List issues from a specific scan
  escape-cli issues list --scan-id <scan-id>

  # Export to JSON for custom processing
  escape-cli issues list -o json | jq '.[] | select(.severity == "CRITICAL")'

  # List unresolved issues across all assets
  escape-cli issues list --status OPEN,MANUAL_REVIEW,IN_PROGRESS`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Output JSON Schema if requested
		if out.Schema([]v3.IssueSummarized{}) {
			return nil
		}

		if issueSortDirection != "" && issueSortType == "" {
			return errors.New("--sort-direction requires --sort-by")
		}
		if issueSortType != "" {
			issueSortType = strings.ToUpper(issueSortType)
			if _, ok := validIssueSortFields[issueSortType]; !ok {
				return fmt.Errorf("invalid --sort-by %q; valid values: LAST_SEEN, FIRST_SEEN, SEVERITY, STATUS", issueSortType)
			}
		}
		switch normalizedDirection := strings.ToLower(issueSortDirection); normalizedDirection {
		case "":
		case "asc", "desc":
			issueSortDirection = normalizedDirection
		default:
			return fmt.Errorf("invalid --sort-direction %q; valid values: asc, desc", issueSortDirection)
		}

		filters := &escape.ListIssuesFilters{
			Status:       issueStatus,
			Severities:   issueSeverity,
			ProfileIDs:   profileIDs,
			AssetIDs:     assetIDs,
			Domains:      domains,
			IssueIDs:     issueIDs,
			ScanIDs:      scanIDs,
			TagsIDs:      tagsIDs,
			Risks:        risks,
			Search:       search,
			AssetClasses: assetClasses,
			JiraTicket:   jiraTicket,
			ScannerKinds: issueScannerKinds,
			Names:        issueNames,
		}
		issues, next, err := escape.ListIssues(cmd.Context(), "", filters, issueSortType, issueSortDirection)
		if err != nil {
			return fmt.Errorf("unable to list issues: %w", err)
		}
		allIssues := issues
		for next != nil && *next != "" {
			issues, next, err = escape.ListIssues(cmd.Context(), *next, filters, issueSortType, issueSortDirection)
			if err != nil {
				return fmt.Errorf("unable to list issues: %w", err)
			}
			allIssues = append(allIssues, issues...)
		}
		out.Table(allIssues, func() []string {
			res := []string{"ID\tCREATED AT\tSEVERITY\tSTATUS\tNAME\tASSET\tLINK"}
			for _, issue := range allIssues {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), out.GetShortDate(issue.GetCreatedAt()), issue.GetSeverity(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetLinks().IssueOverview))
			}
			return res
		})

		return nil
	},
}

var issueGetCmd = &cobra.Command{
	Use:     "get issue-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get detailed information about a security issue",
	Long: `Get Issue Details - View Complete Vulnerability Information

Retrieve comprehensive details about a specific security issue including severity,
category, status, affected asset, and a direct link to full remediation guidance.

DISPLAYED INFORMATION:
  • ID          - Unique issue identifier
  • CREATED AT  - When the issue was first discovered
  • SEVERITY    - Risk level (CRITICAL, HIGH, MEDIUM, LOW, INFO)
  • CATEGORY    - Vulnerability classification (e.g., INJECTION, AUTH, CRYPTO)
  • STATUS      - Current remediation status
  • NAME        - Human-readable vulnerability description
  • ASSET       - Affected API or application
  • LINK        - URL to detailed analysis and remediation steps

USE CASES:
  • Review vulnerability details before fixing
  • Share issue information with team members
  • Verify issue details in incident response
  • Get remediation guidance link

Example output:
ID                                      CREATED AT                SEVERITY  CATEGORY         STATUS  NAME              ASSET                  LINK
00000000-0000-0000-0000-000000000001    2025-06-26T06:03:26.128Z  HIGH      XXE Injection    OPEN    XML External...   api.example.com        https://...`,
	Example: `  # Get issue details
  escape-cli issues get 00000000-0000-0000-0000-000000000001

  # Get issue in JSON format
  escape-cli issues get <issue-id> -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output JSON Schema if requested
		if out.Schema(v3.GetIssue200Response{}) {
			return nil
		}

		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		out.Table(issue, func() []string {
			cvssData := issue.GetCvss()
			cvss := "-"
			if cvssData.GetScore() != 0 || cvssData.GetVector() != "" {
				cvss = fmt.Sprintf("%.1f %s", cvssData.GetScore(), cvssData.GetVector())
			}
			latestEvents := formatLatestEventIDs(issue.GetLatestEventIds(), issue.GetLatestEventsTruncated())
			res := []string{"ID\tCREATED AT\tSEVERITY\tCATEGORY\tSTATUS\tNAME\tASSET\tFIRST SEEN SCAN\tCVSS\tFRAMEWORK\tCOMPLIANCES\tREMEDIATION\tCONTEXT\tLATEST EVENTS\tLINK"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), issue.GetCreatedAt(), issue.GetSeverity(), issue.GetCategory(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetFirstSeenScanId(), cvss, issue.GetAiRemediationFramework(), formatIssueCompliances(issue.GetCompliances()), issue.GetRemediation(), issue.GetContext(), latestEvents, issue.GetLinks().IssueOverview))
			return res
		})

		return nil
	},
}

var issueGetWithEventsCmd = &cobra.Command{
	Use:     "get-with-events issue-id",
	Aliases: []string{"with-events", "events"},
	Short:   "Get an issue plus every event from its last-seen scan in a single call",
	Long: `Get Issue With Hydrated Events - Single-Call Context for AI Agents

Fetches an issue and concurrently hydrates every event ID listed in
` + "`latestEventIds`" + ` (capped at 5 by the API). Each hydrated event includes the
Exchange (request/response payload) on its attachments when available, so an
agent receives the full investigation context in one tool call.

OUTPUT SHAPE (JSON):
  {
    "issue": <IssueDetailed>,
    "latestEvents": [<EventDetailed>, ...],
    "latestEventsTruncated": <bool>,
    "eventErrors": [{"eventId": "...", "error": "..."}, ...]   // omitted on full success
  }

Per-event hydration failures do NOT cancel siblings and do NOT fail the
command — they surface in 'eventErrors'. The command only fails if the issue
itself cannot be fetched.`,
	Example: `  # Hydrate full context for an AI agent
  escape-cli issues get-with-events <issue-id> -o json

  # Aliases
  escape-cli issues with-events <issue-id>
  escape-cli issues events <issue-id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(IssueWithEvents{}) {
			return nil
		}

		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		eventIDs := issue.GetLatestEventIds()
		latestEvents := make([]v3.GetEvent200Response, len(eventIDs))
		eventErrors := make([]IssueEventHydrateError, len(eventIDs))
		eventOK := make([]bool, len(eventIDs))

		g, gctx := errgroup.WithContext(cmd.Context())
		g.SetLimit(maxLatestEventsHydrationConcurrency)
		var mu sync.Mutex
		for i, eid := range eventIDs {
			i, eid := i, eid
			g.Go(func() error {
				ev, err := escape.GetEvent(gctx, eid)
				mu.Lock()
				defer mu.Unlock()
				if err != nil || ev == nil {
					msg := "nil response"
					if err != nil {
						msg = err.Error()
					}
					eventErrors[i] = IssueEventHydrateError{EventID: eid, Error: msg}
					return nil
				}
				latestEvents[i] = *ev
				eventOK[i] = true
				return nil
			})
		}
		_ = g.Wait()

		hydratedEvents := make([]v3.GetEvent200Response, 0, len(latestEvents))
		hydratedErrors := make([]IssueEventHydrateError, 0)
		for i := range eventIDs {
			if eventOK[i] {
				hydratedEvents = append(hydratedEvents, latestEvents[i])
			} else {
				hydratedErrors = append(hydratedErrors, eventErrors[i])
			}
		}

		result := IssueWithEvents{
			Issue:                 *issue,
			LatestEvents:          hydratedEvents,
			LatestEventsTruncated: issue.GetLatestEventsTruncated(),
		}
		if len(hydratedErrors) > 0 {
			result.EventErrors = hydratedErrors
		}

		out.Table(result, func() []string {
			res := []string{"ISSUE ID\tEVENTS HYDRATED\tEVENTS FAILED\tTRUNCATED"}
			res = append(res, fmt.Sprintf("%s\t%d\t%d\t%t",
				result.Issue.GetId(),
				len(result.LatestEvents),
				len(result.EventErrors),
				result.LatestEventsTruncated,
			))
			return res
		})
		return nil
	},
}

var issueUpdateStatusCmd = &cobra.Command{
	Use:     "update issue-id --status STATUS",
	Aliases: []string{"set-status"},
	Short:   "Update issue status to track remediation progress",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}
		return nil
	},
	Long: `Update Issue Status - Track Vulnerability Remediation

Change the status of a security issue as you progress through remediation.
Status updates create an audit trail and help teams track security work.

AVAILABLE STATUSES:
  OPEN              - Newly discovered, awaiting review
  MANUAL_REVIEW     - Under investigation by security team
  IN_PROGRESS       - Actively being fixed by developers
  RESOLVED          - Fixed and verified
  FALSE_POSITIVE    - Determined not to be a real issue
  ACCEPTED_RISK     - Acknowledged but not fixing (with justification)
  REOPENED          - Previously resolved but found again

WORKFLOW EXAMPLE:
  1. New issue discovered:        OPEN
  2. Security team reviews:       MANUAL_REVIEW
  3. Assigned to developers:      IN_PROGRESS
  4. Fix deployed and tested:     RESOLVED

TRACKING:
  All status changes are logged in the issue's activity history.
  Use 'escape-cli issues list-activities <issue-id>' to view the full timeline.`,
	Example: `  # Mark issue under review
  escape-cli issues update <issue-id> --status MANUAL_REVIEW

  # Mark as in progress when fixing
  escape-cli issues update <issue-id> --status IN_PROGRESS

  # Mark as resolved after fixing
  escape-cli issues update <issue-id> --status RESOLVED

  # Mark as false positive
  escape-cli issues update <issue-id> --status FALSE_POSITIVE

  # Bulk update issues from a list
  cat issue_ids.txt | xargs -I {} escape-cli issues update {} --status IN_PROGRESS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		if err := cmd.MarkFlagRequired("status"); err != nil {
			return fmt.Errorf("failed to mark status flag as required: %w", err)
		}
		if issueUpdateStatusStr == "" {
			_ = cmd.Help()
			return errors.New("--status flag is required")
		}

		// Validate provided status against generated enum
		newStatus := v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS(issueUpdateStatusStr)
		if !newStatus.IsValid() {
			return fmt.Errorf("invalid status %q; valid values: %v", issueUpdateStatusStr, v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSTATUSEnumValues)
		}

		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		isUpdated, err := escape.UpdateIssue(cmd.Context(), issueID, newStatus, issueUpdateComment)
		if err != nil || !isUpdated {
			return fmt.Errorf("unable to update issue %s: %w", issueID, err)
		}

		fmt.Printf("Issue %s updated to status %s\n", issueID, newStatus)

		return nil
	},
}

var issueListActivitiesCmd = &cobra.Command{
	Use:     "list-activities issue-id",
	Aliases: []string{"ls-activities", "activities", "history", "timeline"},
	Short:   "View complete activity history and timeline for an issue",
	Long: `View Issue Activity Timeline - Audit Trail and History

Display the complete activity history for a security issue, including all status
changes, comments, and modifications. This provides a full audit trail of who
did what and when.

ACTIVITY TYPES:
  • CREATED          - Issue first discovered
  • STATUS_CHANGED   - Status updated (e.g., OPEN → IN_PROGRESS)
  • COMMENT_ADDED    - Team member added a comment
  • ASSIGNED         - Issue assigned to a team member
  • SEVERITY_CHANGED - Severity level adjusted
  • REOPENED         - Previously resolved issue found again

DISPLAYED INFORMATION:
  • ID           - Activity identifier
  • CREATED AT   - When the activity occurred
  • KIND         - Type of activity
  • AUTHOR ID    - Who performed the action
  • AUTHOR EMAIL - User's email address

USE CASES:
  • Review remediation progress
  • Create compliance audit trails
  • Investigate who changed issue status
  • Track time-to-resolution metrics
  • Generate reports for management

Example output:
ID                                      CREATED AT                KIND              AUTHOR ID    AUTHOR EMAIL
00000000-0000-0000-0000-000000000001    2025-06-27T06:02:18.874Z  CREATED           sys-001      system@escape.tech
00000000-0000-0000-0000-000000000002    2025-06-27T08:15:32.120Z  STATUS_CHANGED    usr-123      john@example.com
00000000-0000-0000-0000-000000000003    2025-06-28T14:22:01.543Z  COMMENT_ADDED     usr-456      jane@example.com`,
	Example: `  # View issue activity timeline
  escape-cli issues list-activities 00000000-0000-0000-0000-000000000001

  # Export timeline to JSON
  escape-cli issues list-activities <issue-id> -o json

  # View activities in YAML format
  escape-cli issues list-activities <issue-id> -o yaml`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output JSON Schema if requested
		if out.Schema([]v3.ActivitySummarized{}) {
			return nil
		}

		issueID := args[0]
		activities, err := escape.ListIssueActivities(cmd.Context(), issueID)
		if err != nil {
			return fmt.Errorf("unable to list activities: %w", err)
		}

		result := []string{"ID\tCREATED AT\tKIND\tAUTHOR ID\tAUTHOR EMAIL"}
		for _, activity := range activities {
			author := activity.GetAuthor()
			authorID := "null"
			authorEmail := "null"
			if author.GetId() != "" || author.GetEmail() != "" {
				authorID = author.GetId()
				authorEmail = author.GetEmail()
			}
			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", activity.GetId(), activity.GetCreatedAt(), activity.GetKind(), authorID, authorEmail))
		}

		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}

var issueCommentCmd = &cobra.Command{
	Use:     "comment issue-id",
	Aliases: []string{"add-comment"},
	Short:   "Add a comment to an issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.CreateAssetCommentRequest{}) {
			return nil
		}
		if out.Schema(v3.CreateAssetComment200Response{}) {
			return nil
		}
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		issueID := args[0]
		msg, _ := cmd.Flags().GetString("message")
		if strings.TrimSpace(msg) == "" {
			return errors.New("--message is required")
		}
		if err := escape.CommentIssue(cmd.Context(), issueID, msg); err != nil {
			return fmt.Errorf("unable to add comment: %w", err)
		}
		out.Log("Comment added to issue " + issueID)
		return nil
	},
}

var (
	funnelProjectIDs []string

	trendAfter          string
	trendBefore         string
	trendInterval       string
	trendApplicationIDs []string
	trendProjectIDs     []string

	bulkIssueStatus       string
	bulkIssueIDs          []string
	bulkIssueAssetIDs     []string
	bulkIssueSeverities   []string
	bulkIssueProfileIDs   []string
	bulkIssueTagIDs       []string
	bulkIssueScannerKinds []string

	notifyScanID string
)

var issueFunnelCmd = &cobra.Command{
	Use:   "funnel",
	Short: "Show issue funnel breakdown",
	Long:  `Display the issue funnel: ALL → OPEN_ISSUES → EXPOSED → UNAUTHENTICATED → HIGH_BUSINESS_IMPACT → CRITICAL with counts per category and step.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]escape.IssueFunnelStep{}) {
			return nil
		}
		steps, err := escape.GetIssueFunnel(cmd.Context(), funnelProjectIDs)
		if err != nil {
			return fmt.Errorf("unable to get issue funnel: %w", err)
		}
		out.Table(steps, func() []string {
			res := []string{"CATEGORY\tSTEP\tCOUNT"}
			for _, s := range steps {
				res = append(res, fmt.Sprintf("%s\t%s\t%.0f", s.Category, s.Step, s.Count))
			}
			return res
		})
		return nil
	},
}

var issueTrendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Show issue severity trends over time",
	Long:  `Display time-bucketed issue counts by severity. Track whether your security posture is improving or degrading.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]escape.IssueTrendPoint{}) {
			return nil
		}
		if trendAfter == "" || trendBefore == "" {
			return errors.New("--after and --before are required")
		}
		points, err := escape.GetIssueTrends(cmd.Context(), trendAfter, trendBefore, trendInterval, trendApplicationIDs, trendProjectIDs)
		if err != nil {
			return fmt.Errorf("unable to get issue trends: %w", err)
		}
		out.Table(points, func() []string {
			res := []string{"DATE\tHIGH\tMEDIUM\tLOW\tINFO"}
			for _, p := range points {
				res = append(res, fmt.Sprintf("%s\t%.0f\t%.0f\t%.0f\t%.0f", p.Date, p.HIGH, p.MEDIUM, p.LOW, p.INFO))
			}
			return res
		})
		return nil
	},
}

var issueBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Update status of multiple issues matching a filter",
	Long:  `Bulk update the status of issues. For example, mark all LOW severity issues on a given asset as IGNORED.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if bulkIssueStatus == "" {
			return errors.New("--status is required")
		}
		status := v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS(bulkIssueStatus)
		where := v3.BulkUpdateIssuesRequestWhere{}
		if len(bulkIssueIDs) > 0 {
			where.Ids = bulkIssueIDs
		}
		if len(bulkIssueAssetIDs) > 0 {
			where.AssetIds = bulkIssueAssetIDs
		}
		if len(bulkIssueSeverities) > 0 {
			severities := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESSEVERITY, len(bulkIssueSeverities))
			for i, s := range bulkIssueSeverities {
				severities[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESSEVERITY(s)
			}
			where.Severities = severities
		}
		if len(bulkIssueProfileIDs) > 0 {
			where.ProfileIds = bulkIssueProfileIDs
		}
		if len(bulkIssueTagIDs) > 0 {
			where.TagIds = bulkIssueTagIDs
		}
		if len(bulkIssueScannerKinds) > 0 {
			kinds := make([]v3.ENUMPROPERTIESWHEREPROPERTIESSCANNERKINDSITEMS, len(bulkIssueScannerKinds))
			for i, k := range bulkIssueScannerKinds {
				kinds[i] = v3.ENUMPROPERTIESWHEREPROPERTIESSCANNERKINDSITEMS(k)
			}
			where.ScannerKinds = kinds
		}
		result, err := escape.BulkUpdateIssues(cmd.Context(), status, &where)
		if err != nil {
			return fmt.Errorf("unable to bulk update issues: %w", err)
		}
		out.Log(fmt.Sprintf("Updated %d issues", len(result.GetIds())))
		out.Table(result, func() []string {
			res := []string{"UPDATED IDS"}
			res = append(res, result.GetIds()...)
			return res
		})
		return nil
	},
}

var issueNotifyCmd = &cobra.Command{
	Use:   "notify issue-id",
	Short: "Send notification to asset owners about an issue",
	Long:  `Send an email notification to the owners of the asset associated with this issue.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		if notifyScanID == "" {
			return errors.New("--scan-id is required")
		}
		notified, err := escape.NotifyIssueOwners(cmd.Context(), issueID, notifyScanID)
		if err != nil {
			return fmt.Errorf("unable to notify owners: %w", err)
		}
		if notified {
			out.Log("Notification sent to asset owners for issue " + issueID)
		} else {
			out.Log("No owners found to notify for issue " + issueID)
		}
		return nil
	},
}

func init() {
	issuesCmd.AddCommand(issueGetCmd)
	issuesCmd.AddCommand(issueGetWithEventsCmd)
	issuesCmd.AddCommand(issueListActivitiesCmd)
	issuesCmd.AddCommand(issueCommentCmd)
	issueCommentCmd.Flags().String("message", "", "comment message to add to the issue")

	issuesCmd.AddCommand(issueUpdateStatusCmd)
	issueUpdateStatusCmd.Flags().StringVarP(&issueUpdateStatusStr, "status", "s", issueUpdateStatusStr, fmt.Sprintf("new status for the issue: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSTATUSEnumValues))
	issueUpdateStatusCmd.Flags().StringVar(&issueUpdateComment, "comment", "", "optional comment explaining the status change")

	issuesCmd.AddCommand(issueFunnelCmd)
	issueFunnelCmd.Flags().StringSliceVar(&funnelProjectIDs, "project-id", nil, "filter by project ID(s)")

	issuesCmd.AddCommand(issueTrendsCmd)
	issueTrendsCmd.Flags().StringVar(&trendAfter, "after", "", "start date (ISO 8601, required)")
	issueTrendsCmd.Flags().StringVar(&trendBefore, "before", "", "end date (ISO 8601, required)")
	issueTrendsCmd.Flags().StringVar(&trendInterval, "interval", "1 day", "time bucket interval (e.g. '1 day', '1 week')")
	issueTrendsCmd.Flags().StringSliceVar(&trendApplicationIDs, "application-id", nil, "filter by application ID(s)")
	issueTrendsCmd.Flags().StringSliceVar(&trendProjectIDs, "project-id", nil, "filter by project ID(s)")

	issuesCmd.AddCommand(issueBulkUpdateCmd)
	issueBulkUpdateCmd.Flags().StringVar(&bulkIssueStatus, "status", "", "new status to apply (required)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueIDs, "issue-id", nil, "filter by issue ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueAssetIDs, "asset-id", nil, "filter by asset ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueSeverities, "severity", nil, "filter by severity")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueProfileIDs, "profile-id", nil, "filter by profile ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueTagIDs, "tag-id", nil, "filter by tag ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueScannerKinds, "scanner-kind", nil, "filter by scanner kind")

	issuesCmd.AddCommand(issueNotifyCmd)
	issueNotifyCmd.Flags().StringVar(&notifyScanID, "scan-id", "", "scan ID to reference in the notification (required)")

	issuesCmd.AddCommand(issueListCmd)

	issueListCmd.Flags().StringVarP(&search, "search", "s", "", "free-text search across issue names and descriptions")
	issueListCmd.Flags().StringSliceVarP(&issueStatus, "status", "", issueStatus, fmt.Sprintf("filter by status: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSTATUSEnumValues))
	issueListCmd.Flags().StringSliceVarP(&issueSeverity, "severity", "l", issueSeverity, fmt.Sprintf("filter by severity level: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSEVERITYEnumValues))
	issueListCmd.Flags().StringSliceVarP(&profileIDs, "profile-id", "p", profileIDs, "filter by profile ID(s) - comma-separated for multiple")
	issueListCmd.Flags().StringSliceVarP(&assetIDs, "asset-id", "a", assetIDs, "filter by asset ID(s) - comma-separated for multiple")
	issueListCmd.Flags().StringSliceVarP(&domains, "domain", "d", domains, "filter by domain name(s)")
	issueListCmd.Flags().StringSliceVarP(&issueIDs, "issue-id", "i", issueIDs, "filter by specific issue ID(s)")
	issueListCmd.Flags().StringSliceVarP(&scanIDs, "scan-id", "", []string{}, "filter by scan ID(s) that discovered the issues")
	issueListCmd.Flags().StringSliceVarP(&tagsIDs, "tag-id", "t", []string{}, "filter by tag ID(s)")
	issueListCmd.Flags().StringVarP(&jiraTicket, "jira-ticket", "j", "", "filter by associated Jira ticket ID")
	issueListCmd.Flags().StringSliceVarP(&risks, "risk", "r", []string{}, fmt.Sprintf("filter by asset risk level: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESRISKSITEMSEnumValues))
	issueListCmd.Flags().StringSliceVarP(&assetClasses, "asset-class", "", []string{}, fmt.Sprintf("filter by asset classification: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESCLASSEnumValues))
	issueListCmd.Flags().StringSliceVar(&issueScannerKinds, "scanner-kind", []string{}, "filter by scanner kind (e.g., DAST, BLST_REST)")
	issueListCmd.Flags().StringSliceVar(&issueNames, "name", []string{}, "filter by issue name(s)")
	issueListCmd.Flags().StringVar(&issueSortType, "sort-by", "", "sort field: LAST_SEEN, FIRST_SEEN, SEVERITY, STATUS")
	issueListCmd.Flags().StringVar(&issueSortDirection, "sort-direction", "", "sort direction: asc, desc")

	rootCmd.AddCommand(issuesCmd)
}
