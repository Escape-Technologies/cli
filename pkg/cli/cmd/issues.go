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
	issueStatus          = []string{
		string(v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS_OPEN),
		string(v3.ENUMPROPERTIESDATAITEMSPROPERTIESSTATUS_MANUAL_REVIEW),
	}
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
ID                                      TYPE         NAME                                                           CREATED AT              HAS CI    CRON
00000000-0000-0000-0000-000000000001    REST         Example-Application-1                                          2025-02-21T11:15:07Z    false     0 11 * * 5`,
	Example: `escape-cli issues list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		issues, next, err := escape.ListIssues(cmd.Context(), "", issueStatus, issueSeverity)
		if err != nil {
			return fmt.Errorf("unable to list issues: %w", err)
		}

		result := []string{"ID\tSEVERITY\tCATEGORY\tNAME\tSTATUS\tASSET\tCREATED AT\tLINK"}
		for _, issue := range issues {
			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), issue.GetSeverity(), issue.GetCategory(), issue.GetName(), issue.GetStatus(), issue.GetAsset().Name, issue.GetCreatedAt(), issue.GetLinks().IssueOverview))
		}

		for next != nil && *next != "" {
			issues, next, err = escape.ListIssues(cmd.Context(), *next, issueStatus, issueSeverity)
			if err != nil {
				return fmt.Errorf("unable to list profiles: %w", err)
			}
			for _, issue := range issues {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), issue.GetSeverity(), issue.GetCategory(), issue.GetName(), issue.GetAsset().Name, issue.GetCreatedAt(), issue.GetLinks().IssueOverview))
			}
		}

		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}

var issueGetCmd = &cobra.Command{
	Use:     "get issue-id",
	Aliases: []string{"describe"},
	Short:   "Get an issue",
	Long: `Get a profile by ID.

Example output:
ID                                      TYPE         NAME                                                           CREATED AT              HAS CI    CRON
00000000-0000-0000-0000-000000000001    REST         Example-Application-1                                          2025-02-21T11:15:07Z    false     0 11 * * 5`,
	Example: `escape-cli profiles get 00000000-0000-0000-0000-000000000001`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		result := []string{"SEVERITY\tCATEGORY\tNAME\tSTATUS\tASSET\tCREATED AT\tLINK"}
		result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetSeverity(), issue.GetCategory(), issue.GetName(), issue.GetStatus(), issue.GetAsset().Name, issue.GetCreatedAt(), issue.GetLinks().IssueOverview))

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
ID                                      TYPE         NAME                                                           CREATED AT              HAS CI    CRON
00000000-0000-0000-0000-000000000001    REST         Example-Application-1                                          2025-02-21T11:15:07Z    false     0 11 * * 5`,
	Example: `escape-cli issues update 00000000-0000-0000-0000-000000000001`,
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
ID                                      TYPE         NAME                                                           CREATED AT              HAS CI    CRON
00000000-0000-0000-0000-000000000001    REST         Example-Application-1                                          2025-02-21T11:15:07Z    false     0 11 * * 5`,
	Example: `escape-cli issues list-activities 00000000-0000-0000-0000-000000000001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		activities, err := escape.ListIssueActivities(cmd.Context(), issueID)
		if err != nil {
			return fmt.Errorf("unable to list activities: %w", err)
		}

		result := []string{"ID\tKIND\tCREATED AT"}
		result = append(result, fmt.Sprintf("%s\t%s\t%s", activities.GetId(), activities.GetKind(), activities.GetCreatedAt()))

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

	rootCmd.AddCommand(issuesCmd)
}
