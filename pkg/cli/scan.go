package cli

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape/scans"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:     "scan",
	Aliases: []string{"sc"},
	Short:   "Scan commands",
}

var getScanCmd = &cobra.Command{
	Use:     "get [scanId]",
	Short:   "Get a scan by ID",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanId, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		scan, err := scans.GetScan(cmd.Context(), scanId)
		if err != nil {
			return fmt.Errorf("failed to get scan: %w", err)
		}
		return print(scan, func() {
			fmt.Println(scan.String())
		})
	},
}

var getScanIssuesCmd = &cobra.Command{
	Use:     "issues [scanId]",
	Short:   "Get issues for a scan by ID",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanId, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		issues, err := scans.GetScanIssues(cmd.Context(), scanId)
		if err != nil {
			return fmt.Errorf("failed to get scan issues: %w", err)
		}
		return print(issues, func() {
			fmt.Print(scans.FormatReportsTable(issues))
		})
	},
}

var getScanIssueCmd = &cobra.Command{
	Use:     "issue [scanId] [issueId]",
	Short:   "Get an issue for a scan by Issue ID",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanId, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		issueId, err := uuid.Parse(args[1])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		issues, err := scans.GetScanIssue(cmd.Context(), scanId, issueId)
		if err != nil {
			return fmt.Errorf("failed to get scan issue: %w", err)
		}
		return print(issues, func() {
			fmt.Print(scans.FormatIssuesTable(issues))
		})
	},
}

var getScanExchangeArchiveCmd = &cobra.Command{
	Use:     "exchange-archive [scanId]",
	Short:   "Get the exchange archive for a scan by ID",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanId, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		archive, err := scans.GetScanExchangeArchive(cmd.Context(), scanId)
		if err != nil {
			return fmt.Errorf("failed to get scan exchange archive: %w", err)
		}
		return print(archive, func() {
			fmt.Println(archive.String())
		})
	},
}

var getScanEventsCmd = &cobra.Command{
	Use:     "events [scanId]",
	Short:   "Get events for a scan by ID",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		size := 100
		scanId, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		
		events, nextCursor, err := scans.GetScanEvents(cmd.Context(), scanId, &size, nil)
		if err != nil {
			return fmt.Errorf("failed to get scan events: %w", err)
		}
		
		for _, event := range events {
			print(event, func() { fmt.Println(event.String()) })
		}

		for nextCursor != nil {
			events, nextCursor, err = scans.GetScanEvents(cmd.Context(), scanId, &size, nextCursor)
			if err != nil {
				return fmt.Errorf("failed to get scan events: %w", err)
			}
			
			for _, event := range events {
				print(event, func() { fmt.Println(event.String()) })
			}
		}
		return nil
	},
}