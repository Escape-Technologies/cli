package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

// scanAgentsPageSize is the page the reasoning-log walk asks for.
// Agent listings pass 0 and use the API wrapper's own default.
const scanAgentsPageSize = 100

var scanAgentsSearch string
var scanAgentsEventSearch string
var scanAgentsRootsOnly bool
var scanAgentsPage pageFlags

// ScanAgents lists AI pentest agents for a scan.
type ScanAgents struct {
	ScanID string                   `json:"scanId"`
	Agents []escape.AgentSummarized `json:"agents"`
}

// scanAgentsMCPDocument is the MCP output for scans agents.
// MCP always requests one page, so items, nextCursor, and totalCount stay
// required. scanId and agents stay optional on the same object: the MCP
// boundary only rewrites a null into [] when that property's top-level schema
// type is array, and a page that omits them still matches.
type scanAgentsMCPDocument struct {
	Items      []escape.AgentSummarized `json:"items"`
	NextCursor string                   `json:"nextCursor"`
	TotalCount int                      `json:"totalCount"`
	ScanID     string                   `json:"scanId,omitempty"`
	Agents     []escape.AgentSummarized `json:"agents,omitempty"`
}

var scansAgentsCmd = &cobra.Command{
	Use:     "agents scan-id",
	Aliases: []string{"agent-list"},
	Short:   "List AI pentest agents for a scan (Beta)",
	Long: `List AI Pentest Agents

Returns the agents that ran during an AI pentest scan. Use --search to filter by
agent title and --event-search to find agents whose reasoning logs mention text.

--size or --cursor returns one page as {items, nextCursor, totalCount}.
Neither flag keeps the full listing.`,
	Example: `  escape-cli scans agents <scan-id> -o json
  escape-cli scans agents <scan-id> --search "Coverage" -o json
  escape-cli scans agents <scan-id> --event-search "XXE" -o json
  escape-cli scans agents <scan-id> --size 20 -o json
  escape-cli scans agents <scan-id> --cursor <nextCursor> -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Page mode prints {items, nextCursor, totalCount}. The default walk
		// still prints {scanId, agents}. MCP always pages; the registry type
		// is scanAgentsMCPDocument.
		if singlePageRequested(cmd) {
			if out.Schema(Page[escape.AgentSummarized]{}) {
				return nil
			}
		} else if out.Schema(ScanAgents{}) {
			return nil
		}

		// --size or --cursor is one page. The default is still every agent,
		// printed as {scanId, agents}.
		if singlePageRequested(cmd) {
			if err := runPagedList(cmd, scanAgentsPage, func(ctx context.Context, cursor string, size int) ([]escape.AgentSummarized, *string, int, error) {
				return escape.ListScanAgents(ctx, args[0], cursor, size, scanAgentFilters())
			}, scanAgentsTable); err != nil {
				return fmt.Errorf("unable to list scan agents: %w", err)
			}

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

func scanAgentFilters() *escape.ListScanAgentsFilters {
	return &escape.ListScanAgentsFilters{
		Search:      scanAgentsSearch,
		EventSearch: scanAgentsEventSearch,
		RootsOnly:   scanAgentsRootsOnly,
	}
}

func scanAgentsTable(agents []escape.AgentSummarized) []string {
	rows := []string{"ID\tTITLE\tSTATUS"}
	for _, agent := range agents {
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s", agent.ID, agent.Title, agent.Status))
	}

	return rows
}

func listScanAgents(ctx context.Context, scanID string) ([]escape.AgentSummarized, error) {
	var agents []escape.AgentSummarized
	next := ""
	for {
		page, cursor, _, err := escape.ListScanAgents(ctx, scanID, next, 0, scanAgentFilters())
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
