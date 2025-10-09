package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	stages         []string
	hasAttachments bool
	attachments    []string
	eventLevels    []string
)

var eventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event"},
	Short:   "View scan events and execution logs",
	Long: `View Scan Events - Monitor Security Testing Activity

Events track detailed activities during security scans including test execution,
discoveries, errors, and progress updates. Use events for troubleshooting scans
and understanding test behavior.

EVENT LEVELS:
  • ERROR   - Scan errors and failures
  • WARN    - Warnings and potential issues
  • INFO    - General informational messages
  • DEBUG   - Detailed debugging information

EVENT STAGES:
  • DISCOVERY  - API endpoint discovery
  • EXECUTION  - Active security testing
  • ANALYSIS   - Results processing
  • REPORTING  - Report generation`,
}

var eventsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List scan events",
	Long: `List Scan Events - View Testing Activity Logs

Display events from security scans with filtering by scan, asset, severity, and type.

FILTER OPTIONS:
  -s, --search         Free-text search
  --scan-id            Filter by scan ID
  -a, --asset-id       Filter by asset ID
  -i, --issue-id       Filter by issue ID
  --stage              Filter by execution stage
  -l, --levels         Filter by level (ERROR, WARN, INFO, DEBUG)
  --has-attachments    Show only events with attachments`,
	Example: `  # List recent events
  escape-cli events list

  # List events for a specific scan
  escape-cli events list --scan-id <scan-id>

  # List errors only
  escape-cli events list --levels ERROR

  # Search event logs
  escape-cli events list --search "timeout"`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		filters := &escape.ListEventsFilters{
			Search:         search,
			ScanIDs:        scanIDs,
			AssetIDs:       assetIDs,
			IssueIDs:       issueIDs,
			Stages:         stages,
			HasAttachments: hasAttachments,
			Attachments:    attachments,
			Levels:         eventLevels,
		}
		events, next, err := escape.ListEvents(cmd.Context(), "", filters)
		if err != nil {
			return fmt.Errorf("unable to list events: %w", err)
		}
		out.Table(events, func() []string {
			res := []string{"ID\tCREATED AT\tLEVEL\tSTAGE\tTITLE"}
			for _, event := range events {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.GetId(), event.GetCreatedAt(), event.GetLevel(), event.GetStage(), event.GetTitle()))
			}
			return res
		})

		for next != nil && *next != "" {
			events, next, err = escape.ListEvents(cmd.Context(), *next, filters)

			if err != nil {
				return fmt.Errorf("unable to list profiles: %w", err)
			}
			out.Table(events, func() []string {
				res := []string{"ID\tCREATED AT\tLEVEL\tSTAGE\tTITLE"}
				for _, event := range events {
					res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.GetId(), event.GetCreatedAt(), event.GetLevel(), event.GetStage(), event.GetTitle()))
				}
				return res
			})
		}

		return nil
	},
}

var eventGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get an event",
	Long: `Get an event.

Example output:
ID                                      LEVEL    TITLE                          STAGE           LINK
00000000-0000-0000-0000-000000000001    INFO     Scan started              	    EXECUTION       https://app.escape.tech/events/00000000-0000-0000-0000-000000000001/logs`,
	Example: `escape-cli events get event-id`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("event ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		eventID := args[0]
		event, err := escape.GetEvent(cmd.Context(), eventID)
		if err != nil {
			return fmt.Errorf("unable to get event: %w", err)
		}

		out.Table(event, func() []string {
			res := []string{"ID\tCREATED AT\tLEVEL\tSTAGE\tTITLE\tLINK"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", event.GetId(), event.GetCreatedAt(), event.GetLevel(), event.GetStage(), event.GetTitle(), strings.Replace(event.Scan.GetLinks().ScanIssues, "/issues", "/logs", 1)))
			return res
		})

		return nil
	},
}

func init() {
	eventsCmd.AddCommand(eventsListCmd)
	eventsListCmd.Flags().StringVarP(&search, "search", "s", "", "Search term to filter events by")
	eventsListCmd.Flags().StringSliceVarP(&scanIDs, "scan-id", "", []string{}, "Scan ID to filter events by")
	eventsListCmd.Flags().StringSliceVarP(&assetIDs, "asset-id", "a", assetIDs, "Asset ID to filter events by")
	eventsListCmd.Flags().StringSliceVarP(&issueIDs, "issue-id", "i", []string{}, "Issue ID to filter events by")
	eventsListCmd.Flags().StringSliceVarP(&stages, "stage", "", stages, fmt.Sprintf("Stages of events: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESSTAGEEnumValues))
	eventsListCmd.Flags().BoolVarP(&hasAttachments, "has-attachments", "", hasAttachments, "Has attachments")
	eventsListCmd.Flags().StringSliceVarP(&attachments, "attachments", "t", attachments, "Attachments to filter events by")
	eventsListCmd.Flags().StringSliceVarP(&eventLevels, "levels", "l", eventLevels, fmt.Sprintf("levels of events: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESLEVELEnumValues))

	eventsCmd.AddCommand(eventGetCmd)

	rootCmd.AddCommand(eventsCmd)
}
