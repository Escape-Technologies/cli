package cmd

import (
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	eventLevels = []string{
		string(v3.ENUMPROPERTIESDATAITEMSPROPERTIESLEVEL_ERROR),
		//string(v3.ENUMPROPERTIESDATAITEMSPROPERTIESLEVEL_WARNING),
	}
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
		events, next, err := escape.ListEvents(cmd.Context(), "", eventLevels)
		if err != nil {
			return fmt.Errorf("unable to list events: %w", err)
		}

		result := []string{"ID\tLEVEL\tTITLE\tSTAGE\tCREATED AT"}
		for _, event := range events {
			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.GetId(), event.GetLevel(), event.GetTitle(), event.GetStage(), event.GetCreatedAt()))
		}

		for next != nil && *next != "" {
			events, next, err = escape.ListEvents(cmd.Context(), *next, eventLevels)
			if err != nil {
				return fmt.Errorf("unable to list profiles: %w", err)
			}
			for _, event := range events {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.GetId(), event.GetLevel(), event.GetTitle(), event.GetStage(), event.GetCreatedAt()))
			}
		}

		out.Table(result, func() []string {
			return result
		})

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
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eventID := args[0]
		event, err := escape.GetEvent(cmd.Context(), eventID)
		if err != nil {
			return fmt.Errorf("unable to get event: %w", err)
		}

		result := []string{"ID\tLEVEL\tTITLE\tSTAGE\tLINK"}
		result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", event.GetScanId(), event.GetLevel(), event.GetTitle(), event.GetStage(), strings.Replace(event.Scan.GetLinks().ScanIssues, "/issues", "/logs", 1)))

		out.Table(result, func() []string {
			return result
		})

		return nil
	},
}


func init() {
	eventsCmd.AddCommand(eventsListCmd)
	eventsListCmd.Flags().StringSliceVarP(&eventLevels, "levels", "l", eventLevels, fmt.Sprintf("levels of events: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESLEVELEnumValues))

	eventsCmd.AddCommand(eventGetCmd)

	rootCmd.AddCommand(eventsCmd)
}
