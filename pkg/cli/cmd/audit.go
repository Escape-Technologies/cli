package cmd

import (
	"fmt"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	auditCmdDateFrom  = time.Now().Add(-12 * time.Hour).Format(time.RFC3339)
	auditCmdDateTo    = time.Now().Format(time.RFC3339)
	auditCmdEventType = ""
	auditCmdActor     = ""
	auditCmdSearch    = ""
)

var auditCmd = &cobra.Command{
	Use:     "audit",
	Aliases: []string{"audits", "logs"},
	Short:   "View audit logs and activity history",
	Long: `View Audit Logs - Organization Activity Timeline

Audit logs track all actions performed in your organization including scans,
configuration changes, user actions, and security events. Essential for compliance
and security monitoring.

LOGGED ACTIVITIES:
  • User authentication and access
  • Scan lifecycle events (started, finished, failed)
  • Configuration changes
  • Issue status updates
  • Profile and asset modifications

COMPLIANCE:
  Use audit logs for compliance reporting (SOC 2, ISO 27001, PCI DSS, etc.)`,
}

var auditListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List organization audit logs",
	Long: `List Audit Logs - View Activity History

Display audit logs with filtering by date range, event type, and actor.
By default shows logs from the last 12 hours.

FILTER OPTIONS:
  -f, --date-from    Start date (RFC3339 format)
  -t, --date-to      End date (RFC3339 format)
  -e, --event-type   Event type (scan.started, scan.finished, user.authenticated, etc.)
  -a, --actor        Filter by actor (user ID or email)
  -s, --search       Free-text search`,
	Example: `  # List recent audit logs
  escape-cli audit list

  # List logs for specific date range
  escape-cli audit list --date-from 2025-01-01T00:00:00Z --date-to 2025-01-31T23:59:59Z

  # List scan events
  escape-cli audit list --event-type scan.started

  # List actions by specific user
  escape-cli audit list --actor user@example.com

  # Export for compliance reporting
  escape-cli audit list --date-from 2025-01-01T00:00:00Z -o json > audit-report-jan2025.json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		filters := &escape.ListAuditLogsFilters{
			DateFrom:   auditCmdDateFrom,
			DateTo:     auditCmdDateTo,
			ActionType: auditCmdEventType,
			Actor:      auditCmdActor,
			Search:     auditCmdSearch,
		}

		logs, next, err := escape.ListAuditLogs(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("unable to list audits: %w", err)
		}

		out.Table(logs, func() []string {
			fields := []string{"DATE\tACTION\tACTOR\tACTOR EMAIL\tTITLE"}
			for _, log := range logs {
				fields = append(fields, fmt.Sprintf(
					"%s\t%s\t%s\t%s\t%s",
					log.GetDate(),
					log.GetAction(),
					log.GetActor(),
					log.GetActorEmail(),
					log.GetTitle(),
				))
			}
			return fields
		})

		for next != nil && *next != "" {
			logs, next, err = escape.ListAuditLogs(cmd.Context(), *next, filters)
			if err != nil {
				return fmt.Errorf("unable to list audits: %w", err)
			}
			out.Table(logs, func() []string {
				fields := []string{"DATE\tACTION\tACTOR\tACTOR EMAIL\tTITLE"}
				for _, log := range logs {
					fields = append(fields, fmt.Sprintf(
						"%s\t%s\t%s\t%s\t%s",
						log.GetDate(),
						log.GetAction(),
						log.GetActor(),
						log.GetActorEmail(),
						log.GetTitle(),
					))
				}
				return fields
			})
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)

	auditCmd.AddCommand(auditListCmd)
	// theses are optional flags for the list command
	auditListCmd.Flags().StringVarP(&auditCmdDateFrom, "date-from", "f", auditCmdDateFrom, "Filter by date from (ISO 8601)")
	auditListCmd.Flags().StringVarP(&auditCmdDateTo, "date-to", "t", auditCmdDateTo, "Filter by date to (ISO 8601)")
	auditListCmd.Flags().StringVarP(&auditCmdEventType, "event-type", "e", "", "Filter by event type: (scan.scheduled, scan.started, scan.finished, user.authenticated)")
	auditListCmd.Flags().StringVarP(&auditCmdActor, "actor", "a", "", "Filter by actor")
	auditListCmd.Flags().StringVarP(&auditCmdSearch, "search", "s", "", "Search term to filter audit logs by")
}
