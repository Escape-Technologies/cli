package cmd

import (
	"errors"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	issueUpdateStatusStr string
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
  1. OPEN           - Newly discovered, needs review
  2. MANUAL_REVIEW  - Under investigation
  3. RESOLVED       - Fixed and verified
  4. FALSE_POSITIVE - Not a real issue
  5. IGNORED        - Ignored / excluded from tracking

COMMON WORKFLOWS:
  • List high-priority issues:
    $ escape-cli issues list --severity HIGH,CRITICAL --status OPEN

  • Review issues for a specific asset:
    $ escape-cli issues list --asset-id <asset-id>

  • Update issue status as you fix them:
    $ escape-cli issues update <issue-id> --status MANUAL_REVIEW

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
  --status           Filter by status: FALSE_POSITIVE, IGNORED, MANUAL_REVIEW, OPEN, RESOLVED
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
		}
		issues, next, err := escape.ListIssues(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("unable to list issues: %w", err)
		}

		out.Table(issues, func() []string {
			res := []string{"ID\tCREATED AT\tSEVERITY\tSTATUS\tNAME\tASSET\tLINK"}
			for _, issue := range issues {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), out.GetShortDate(issue.GetCreatedAt()), issue.GetSeverity(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetLinks().IssueOverview))
			}
			return res
		})

		for next != nil && *next != "" {
			issues, next, err = escape.ListIssues(cmd.Context(), *next, filters)
			if err != nil {
				return fmt.Errorf("unable to list issues: %w", err)
			}
			out.Table(issues, func() []string {
				res := []string{"ID\tCREATED AT\tSEVERITY\tSTATUS\tNAME\tASSET\tLINK"}
				for _, issue := range issues {
					res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), out.GetShortDate(issue.GetCreatedAt()), issue.GetSeverity(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetLinks().IssueOverview))
				}
				return res
			})
		}

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
		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		out.Table(issue, func() []string {
			res := []string{"ID\tCREATED AT\tSEVERITY\tCATEGORY\tSTATUS\tNAME\tASSET\tLINK"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), issue.GetCreatedAt(), issue.GetSeverity(), issue.GetCategory(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetLinks().IssueOverview))
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
  OPEN           - Newly discovered, awaiting review
  MANUAL_REVIEW  - Under investigation by security team
  RESOLVED       - Fixed and verified
  FALSE_POSITIVE - Determined not to be a real issue
  IGNORED        - Ignored / excluded from tracking

WORKFLOW EXAMPLE:
  1. New issue discovered:        OPEN
  2. Security team reviews:       MANUAL_REVIEW
  3. Fix deployed and tested:     RESOLVED

TRACKING:
  All status changes are logged in the issue's activity history.
  Use 'escape-cli issues list-activities <issue-id>' to view the full timeline.`,
	Example: `  # Mark issue under review
  escape-cli issues update <issue-id> --status MANUAL_REVIEW

  # Mark as resolved after fixing
  escape-cli issues update <issue-id> --status RESOLVED

  # Mark as false positive
  escape-cli issues update <issue-id> --status FALSE_POSITIVE
  
  # Ignore an issue
  escape-cli issues update <issue-id> --status IGNORED

  # Bulk update issues from a list
  cat issue_ids.txt | xargs -I {} escape-cli issues update {} --status MANUAL_REVIEW`,
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

		isUpdated, err := escape.UpdateIssue(cmd.Context(), issueID, newStatus)
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

func init() {
	issuesCmd.AddCommand(issueGetCmd)
	issuesCmd.AddCommand(issueListActivitiesCmd)

	issuesCmd.AddCommand(issueUpdateStatusCmd)
	issueUpdateStatusCmd.Flags().StringVarP(&issueUpdateStatusStr, "status", "s", issueUpdateStatusStr, fmt.Sprintf("new status for the issue: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSTATUSEnumValues))

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
	issueListCmd.Flags().StringSliceVarP(&assetClasses, "asset-class", "", []string{}, fmt.Sprintf("filter by asset classification: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESCLASSEnumValues))

	rootCmd.AddCommand(issuesCmd)
}
