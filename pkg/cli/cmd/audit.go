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
	Aliases: []string{"audits"},
	Short:   "Interact with audits",
	Long:    `Interact with your escape audit logs`,
}

var auditListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List audit logs",
	Long:    `List audit logs of the organization.`,
	Example: `escape-cli audit list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		logs, next, err := escape.ListAuditLogs(
			cmd.Context(),
			"",
			&escape.ListAuditLogsFilters{
				DateFrom:   auditCmdDateFrom,
				DateTo:     auditCmdDateTo,
				ActionType: auditCmdEventType,
				Actor:      auditCmdActor,
				Search:     auditCmdSearch,
			},
		)
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
			logs, next, err = escape.ListAuditLogs(
				cmd.Context(),
				*next,
				&escape.ListAuditLogsFilters{
					DateFrom: auditCmdDateFrom,
					DateTo:   auditCmdDateTo,
					ActionType: auditCmdEventType,
				},
			)
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
