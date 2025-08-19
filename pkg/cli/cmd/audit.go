package cmd

import (
	"fmt"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	auditCmdDateFrom  = time.Now().Add(-12 * time.Hour).Format(time.RFC3339)
	auditCmdDateTo    = time.Now().Format(time.RFC3339)
	auditCmdEventType = ""
)

var auditCmd = &cobra.Command{
	Use:     "audit",
	Aliases: []string{"audits"},
	Short:   "Interact with audits",
	Long:    `List audit logs of the organization.`,
	Example: `escape-cli audit`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		logs, next, err := escape.ListAuditLogs(
			cmd.Context(),
			auditCmdDateFrom,
			auditCmdDateTo,
			auditCmdEventType,
			"",
		)
		if err != nil {
			return fmt.Errorf("unable to list audits: %w", err)
		}

		result := make([]*v3.AuditLogSummarized, 0, len(logs))
		for _, log := range logs {
			result = append(result, &log)
		}

		for next != nil && *next != "" {
			logs, next, err = escape.ListAuditLogs(
				cmd.Context(),
				auditCmdDateFrom,
				auditCmdDateTo,
				auditCmdEventType,
				*next,
			)
			if err != nil {
				return fmt.Errorf("unable to list audits: %w", err)
			}
			for _, log := range logs {
				result = append(result, &log)
			}
		}

		out.Table(result, func() []string {
			fields := []string{"ACTOR ID\t ACTION\t NAME\t DATE\t"}
			for _, log := range result {
				fields = append(fields, fmt.Sprintf(
					"%s\t%s\t%s\t%s\t%s",
					log.GetActor(),
					log.GetAction(),
					log.GetTitle(),
					log.GetDate(),
					log.GetTarget(),
				))
			}
			return fields
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().StringVarP(&auditCmdDateFrom, "date-from", "f", auditCmdDateFrom, "Filter by date from (ISO 8601)")
	auditCmd.Flags().StringVarP(&auditCmdDateTo, "date-to", "t", auditCmdDateTo, "Filter by date to (ISO 8601)")
	auditCmd.Flags().StringVarP(&auditCmdEventType, "event-type", "e", "", "Filter by event type: (scan.scheduled, scan.started, scan.finished)")
}
