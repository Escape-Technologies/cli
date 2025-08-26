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
	stages []string
	hasAttachments bool
	attachments []string
	eventLevels []string
)

var eventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event"},
	Short:   "Interact with events",
	Long:    "Interact with your escape events",
}

var eventsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List events",
	Long: `List events.

Example output:
ID                                      LEVEL    TITLE                      STAGE            CREATED AT
00000000-0000-0000-0000-000000000001    INFO     Scan started              	EXECUTION        2025-08-12T14:04:58.117Z`,
	Example: `escape-cli events list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		filters := &escape.ListEventsFilters{
			Search: search,
			ScanIDs: scanIDs,
			AssetIDs: assetIDs,
			IssueIDs: issueIDs,
			Stages: stages,
			HasAttachments: hasAttachments,
			Attachments: attachments,
			Levels: eventLevels,
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
