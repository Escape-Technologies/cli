package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var auditLogsCmd = &cobra.Command{
	Use:     "audit-logs",
	Aliases: []string{"audit", "audit-log"},
	Short:   "Interact with your escape audit logs",
}

var auditLogsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all audit logs",
	RunE: func(cmd *cobra.Command, _ []string) error {
		auditLogs, _, err := escape.ListAuditLogs(cmd.Context(), nil, nil)
		if err != nil {
			return fmt.Errorf("failed to list audit logs: %w", err)
		}
		out.Table(auditLogs, func() []string {
			res := []string{"DATE\tACTION\tNAME\tACTOR\tACTOR EMAIL"}
			for _, auditLog := range auditLogs {
				res = append(
					res,
					fmt.Sprintf(
						"%s\t%s\t%s\t%s\t%s",
						auditLog.GetDate(),
						auditLog.GetAction(),
						auditLog.GetName(),
						auditLog.GetActor(),
						auditLog.GetActorEmail(),
					),
				)
			}
			return res
		})
		return nil
	},
}

func init() {
	auditLogsCmd.AddCommand(auditLogsListCmd)
	rootCmd.AddCommand(auditLogsCmd)
}
