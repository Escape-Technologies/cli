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
	Short:   "Interact with issues",
	Long:    "Interact with your escape issues",
}

var issueListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List issues",
	Long: `List issues.

Example output:
ID                                      CREATED AT  SEVERITY  STATUS  NAME                                    ASSET                          LINK
00000000-0000-0000-0000-000000000001    2025-07-24  INFO      OPEN    Misconfigured CSP Header                https://my.app                 Link`,
	Example: `escape-cli issues list`, RunE: func(cmd *cobra.Command, _ []string) error {
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
				return fmt.Errorf("unable to list profiles: %w", err)
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
	Aliases: []string{"describe"},
	Short:   "Get an issue",
	Long: `Get a profile by ID.

Example output:
ID                                      CREATED AT                  \tLINK       \t\t\t\t\t\t\tSEVERITY    CATEGORY                  STATUS        NAME                     ASSET
00000000-0000-0000-0000-000000000001    2025-06-26T06:03:26.128Z        https://app.escape.tech/all-risks/issues/00000000-0000-0000-0000-000000000001/overview/       HIGH          XXE Injection            https://gontoz.escape.tech/`,
	Example: `escape-cli profiles get 00000000-0000-0000-0000-000000000001`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		result := []string{"ID\tCREATED AT\tSEVERITY\tCATEGORY\tSTATUS\tNAME\tASSET\tLINK"}
		result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), issue.GetCreatedAt(), issue.GetSeverity(), issue.GetCategory(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetLinks().IssueOverview))

		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}

var issueUpdateStatusCmd = &cobra.Command{
	Use:     "update issue-id --status MANUAL_REVIEW",
	Aliases: []string{"update"},
	Short:   "Update an issue",
	Args:    cobra.ExactArgs(1),
	Long: `Update the status of an issue.

Example output:
Issue dc8a5509-348c-4319-ab68-3d8382c6f084 updated to status MANUAL_REVIEW`,
	Example: `escape-cli issues update 00000000-0000-0000-0000-000000000001 -s MANUAL_REVIEW`,
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		if issueUpdateStatusStr == "" {
			return errors.New("no new issue status passed, please use --status to update the issue status")
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
	Aliases: []string{"ls-activities"},
	Short:   "List the activities of an issue",
	Long: `List the activities of an issue.

Example output:
ID                                      KIND    CREATED AT
00000000-0000-0000-0000-000000000001    CREATED 2025-06-27T06:02:18.874Z`,
	Example: `escape-cli issues list-activities 00000000-0000-0000-0000-000000000001`,
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
	issueUpdateStatusCmd.Flags().StringVarP(&issueUpdateStatusStr, "status", "s", issueUpdateStatusStr, "Status to update the issue to")

	issuesCmd.AddCommand(issueListCmd)

	issueListCmd.Flags().StringSliceVarP(&issueStatus, "status", "s", issueStatus, fmt.Sprintf("Status of issues: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSTATUSEnumValues))
	issueListCmd.Flags().StringSliceVarP(&issueSeverity, "severity", "l", issueSeverity, fmt.Sprintf("Severity of issues: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSEVERITYEnumValues))
	issueListCmd.Flags().StringSliceVarP(&profileIDs, "profile-id", "p", profileIDs, "Profile ID to filter issues by")
	issueListCmd.Flags().StringSliceVarP(&assetIDs, "asset-id", "a", assetIDs, "Asset ID to filter issues by")
	issueListCmd.Flags().StringSliceVarP(&domains, "domain", "d", domains, "Domain to filter issues by")
	issueListCmd.Flags().StringSliceVarP(&issueIDs, "issue-id", "i", issueIDs, "Issue ID to filter issues by")
	issueListCmd.Flags().StringSliceVarP(&scanIDs, "scan-id", "", []string{}, "Scan ID to filter issues by")
	issueListCmd.Flags().StringSliceVarP(&tagsIDs, "tag-id", "t", []string{}, "Tag ID to filter issues by")
	issueListCmd.Flags().StringVarP(&search, "search", "", "", "Search term to filter issues by")
	issueListCmd.Flags().StringVarP(&jiraTicket, "jira-ticket", "j", "", "Jira ticket to filter issues by")
	issueListCmd.Flags().StringSliceVarP(&risks, "risk", "r", []string{}, fmt.Sprintf("Risk of issues: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESRISKSITEMSEnumValues))
	issueListCmd.Flags().StringSliceVarP(&assetClasses, "asset-class", "", []string{}, fmt.Sprintf("Asset class of issues: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESCLASSEnumValues))

	rootCmd.AddCommand(issuesCmd)
}
