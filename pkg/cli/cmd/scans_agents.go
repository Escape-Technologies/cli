package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const scanAgentsPageSize = 100

var scanAgentsSearch string
var scanAgentsEventSearch string
var scanAgentsRootsOnly bool

// ScanAgents lists AI pentest agents for a scan.
type ScanAgents struct {
	ScanID string                `json:"scanId"`
	Agents []escape.AgentSummarized `json:"agents"`
}

var scansAgentsCmd = &cobra.Command{
	Use:     "agents scan-id",
	Aliases: []string{"agent-list"},
	Short:   "List AI pentest agents for a scan",
	Long: `List AI Pentest Agents

Returns the agents that ran during an AI pentest scan. Use --search to filter by
agent title and --event-search to find agents whose reasoning logs mention text.`,
	Example: `  escape-cli scans agents <scan-id> -o json
  escape-cli scans agents <scan-id> --search "Coverage" -o json
  escape-cli scans agents <scan-id> --event-search "XXE" -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(ScanAgents{}) {
			return nil
		}

		agents, err := listScanAgents(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		out.Print(ScanAgents{ScanID: args[0], Agents: agents}, "")
		return nil
	},
}

func listScanAgents(ctx context.Context, scanID string) ([]escape.AgentSummarized, error) {
	filters := &escape.ListScanAgentsFilters{
		Search:      scanAgentsSearch,
		EventSearch: scanAgentsEventSearch,
		RootsOnly:   scanAgentsRootsOnly,
	}

	var agents []escape.AgentSummarized
	next := ""
	for {
		page, cursor, err := escape.ListScanAgents(ctx, scanID, next, scanAgentsPageSize, filters)
		if err != nil {
			return nil, fmt.Errorf("unable to list scan agents: %w", err)
		}
		agents = append(agents, page...)
		if cursor == nil || *cursor == "" {
			break
		}
		next = *cursor
	}
	return agents, nil
}