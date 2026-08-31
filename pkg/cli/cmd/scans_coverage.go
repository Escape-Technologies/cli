package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const (
	scanTargetsPageSize           = 100
	defaultCoverageTargetListSize = 200
	maxCoverageTargetListSize     = 500
	coverageOKExamplesPerUser     = 8
	coverageToolGuidance          = "byUser is exhaustive over every fetched target when complete=true, even if targetsTruncated is true. A user can have OK coverage on a route whose overall coverage is not OK; answer per-user questions from byUser.statuses, not overall coverage or scans_reasoning. The targets array is a compact sample."
)

var (
	scanCoverageType   string
	scanCoverageStatus string
	scanCoverageUser   string
	scanCoverageSize   int
)

// ScanCoverageTarget is a compact coverage row for MCP/CLI JSON.
type ScanCoverageTarget struct {
	ID             string                `json:"id"`
	Type           string                `json:"type"`
	Method         string                `json:"method,omitempty"`
	Name           string                `json:"name"`
	Coverage       string                `json:"coverage,omitempty"`
	RequestCount   int                   `json:"requestCount"`
	CoverageByUser []ScanCoverageUserRow `json:"coverageByUser,omitempty"`
}

// ScanCoverageUserRow is one scanner user's status on a target.
type ScanCoverageUserRow struct {
	Name     string `json:"name"`
	Coverage string `json:"coverage"`
}

// ScanCoverageUserSummary is per-user totals across the full scan.
type ScanCoverageUserSummary struct {
	Targets    int            `json:"targets"`
	Statuses   map[string]int `json:"statuses"`
	OKExamples []string       `json:"okExamples,omitempty"`
}

// ScanCoverage is the MCP-facing coverage payload.
type ScanCoverage struct {
	ScanID           string                             `json:"scanId"`
	Complete         bool                               `json:"complete"`
	TotalCount       int                                `json:"totalCount"`
	MatchedCount     int                                `json:"matchedCount"`
	ReturnedCount    int                                `json:"returnedCount"`
	TargetsTruncated bool                               `json:"targetsTruncated"`
	Guidance         string                             `json:"guidance"`
	ByUser           map[string]ScanCoverageUserSummary `json:"byUser"`
	Targets          []ScanCoverageTarget               `json:"targets"`
}

var scansCoverageCmd = &cobra.Command{
	Use:     "coverage scan-id",
	Aliases: []string{"cov"},
	Short:   "Per-user API coverage for a scan. Fetches all pages. Source of truth for successful requests by scanner user.",
	Long: `Scan Coverage By User

Walks every page of scan targets and aggregates coverageByUser.

OUTPUT SHAPE (JSON):
  {
    "complete": true,            // all target pages were fetched
    "totalCount": <int>,         // targets in the scan
    "matchedCount": <int>,       // targets after --coverage/--user filters
    "targetsTruncated": <bool>,  // targets array was capped by --size (false when --size 0)
    "byUser": { "<user>": { "targets": n, "statuses": {"OK": n, ...}, "okExamples": [...] } },
    "targets": [ compact routes ]
  }

byUser is computed over every fetched target, even when the targets array is capped.
Use this command for "did users A and B make successful requests?" — not scans reasoning.`,
	Example: `  escape-cli scans coverage <scan-id> -o json
  escape-cli scans coverage <scan-id> --coverage OK --user company_b_initiator -o json
  escape-cli scans coverage <scan-id> --type API_ROUTE -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("scan ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(ScanCoverage{}) {
			return nil
		}

		result, err := buildScanCoverage(
			cmd.Context(),
			args[0],
			scanCoverageType,
			scanCoverageStatus,
			scanCoverageUser,
			scanCoverageSize,
		)
		if err != nil {
			return err
		}

		out.Print(result, formatScanCoveragePretty(result))
		return nil
	},
}

func buildScanCoverage(
	ctx context.Context,
	scanID string,
	targetType string,
	coverage string,
	user string,
	listSize int,
) (ScanCoverage, error) {
	if listSize < 0 {
		return ScanCoverage{}, errors.New("size must be >= 0")
	}
	if listSize > maxCoverageTargetListSize {
		return ScanCoverage{}, fmt.Errorf("size must be <= %d (0 = all matching)", maxCoverageTargetListSize)
	}

	raw, totalCount, complete, err := listScanTargets(ctx, scanID, targetType, 0)
	if err != nil {
		return ScanCoverage{}, err
	}

	allRows := make([]ScanCoverageTarget, 0, len(raw))
	matched := make([]ScanCoverageTarget, 0, len(raw))
	for _, target := range raw {
		row := compactScanTarget(target)
		allRows = append(allRows, row)
		if matchCoverageFilter(row, coverage, user) {
			matched = append(matched, row)
		}
	}

	returned, truncated := capCoverageTargets(matched, listSize)

	return ScanCoverage{
		ScanID:           scanID,
		Complete:         complete,
		TotalCount:       totalCount,
		MatchedCount:     len(matched),
		ReturnedCount:    len(returned),
		TargetsTruncated: truncated,
		Guidance:         coverageToolGuidance,
		ByUser:           summarizeCoverageByUser(allRows),
		Targets:          returned,
	}, nil
}

func listScanTargets(
	ctx context.Context,
	scanID string,
	targetType string,
	limit int,
) ([]v3.TargetDetailed, int, bool, error) {
	var all []v3.TargetDetailed
	cursor := ""
	totalCount := 0
	for {
		if err := ctx.Err(); err != nil {
			return all, totalCount, false, fmt.Errorf("interrupted while listing targets: %w", err)
		}
		pageSize := scanTargetsPageSize
		if limit > 0 {
			remaining := limit - len(all)
			if remaining <= 0 {
				break
			}
			if remaining < pageSize {
				pageSize = remaining
			}
		}
		page, next, pageTotal, err := escape.ListScanTargets(ctx, scanID, cursor, targetType, pageSize)
		if err != nil {
			return nil, 0, false, err
		}
		if totalCount == 0 {
			totalCount = pageTotal
		}
		all = append(all, page...)
		if limit > 0 && len(all) >= limit {
			all = all[:limit]
			hasMore := next != nil && *next != ""
			return all, totalCount, !hasMore && len(all) >= totalCount && totalCount > 0, nil
		}
		if next == nil || *next == "" {
			complete := totalCount == 0 || len(all) >= totalCount
			return all, totalCount, complete, nil
		}
		cursor = *next
	}
	return all, totalCount, totalCount == 0 || len(all) >= totalCount, nil
}

func compactScanTarget(target v3.TargetDetailed) ScanCoverageTarget {
	row := ScanCoverageTarget{ID: target.GetId()}
	if route, ok := target.GetApiRouteOk(); ok && route != nil {
		row.Type = "API_ROUTE"
		row.Method = route.GetOperation()
		row.Name = route.GetDisplayName()
		row.RequestCount = int(route.GetRequestCount())
		if route.HasCoverage() {
			row.Coverage = string(route.GetCoverage())
		}
		row.CoverageByUser = compactCoverageByUser(route.GetCoverageByUser())
		return row
	}
	if resolver, ok := target.GetGraphqlResolverOk(); ok && resolver != nil {
		row.Type = "GRAPHQL_RESOLVER"
		row.Name = resolver.GetDisplayName()
		row.RequestCount = int(resolver.GetRequestCount())
		if resolver.HasCoverage() {
			row.Coverage = string(resolver.GetCoverage())
		}
		row.CoverageByUser = compactCoverageByUser(resolver.GetCoverageByUser())
		return row
	}
	if file, ok := target.GetCodeFileOk(); ok && file != nil {
		row.Type = "CODE_FILE"
		row.Method = file.GetLanguage()
		row.Name = file.GetPath()
		return row
	}
	if port, ok := target.GetPortOk(); ok && port != nil {
		row.Type = "PORT"
		row.Method = port.GetProtocol()
		row.Name = fmt.Sprintf("%.0f", port.GetPort())
		return row
	}
	row.Type = "UNKNOWN"
	return row
}

func compactCoverageByUser(entries []v3.CoverageByUserEntry) []ScanCoverageUserRow {
	if len(entries) == 0 {
		return nil
	}
	rows := make([]ScanCoverageUserRow, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, ScanCoverageUserRow{
			Name:     entry.GetName(),
			Coverage: string(entry.GetCoverage()),
		})
	}
	return rows
}

func capCoverageTargets(matched []ScanCoverageTarget, listSize int) ([]ScanCoverageTarget, bool) {
	if listSize == 0 || len(matched) <= listSize {
		return matched, false
	}
	return matched[:listSize], true
}

func matchCoverageFilter(row ScanCoverageTarget, coverage string, user string) bool {
	if user != "" {
		for _, entry := range row.CoverageByUser {
			if !strings.EqualFold(entry.Name, user) {
				continue
			}
			return coverage == "" || strings.EqualFold(entry.Coverage, coverage)
		}
		return false
	}
	if coverage != "" && !strings.EqualFold(row.Coverage, coverage) {
		return false
	}
	return true
}

func summarizeCoverageByUser(rows []ScanCoverageTarget) map[string]ScanCoverageUserSummary {
	summaries := map[string]ScanCoverageUserSummary{}
	for _, row := range rows {
		for _, entry := range row.CoverageByUser {
			current := summaries[entry.Name]
			if current.Statuses == nil {
				current.Statuses = map[string]int{}
			}
			current.Targets++
			status := entry.Coverage
			if status == "" {
				status = "UNKNOWN"
			}
			current.Statuses[status]++
			if strings.EqualFold(status, "OK") && len(current.OKExamples) < coverageOKExamplesPerUser {
				label := row.Name
				if row.Method != "" {
					label = row.Method + " " + row.Name
				}
				current.OKExamples = append(current.OKExamples, label)
			}
			summaries[entry.Name] = current
		}
	}
	return summaries
}

func formatScanCoveragePretty(result ScanCoverage) string {
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"scan %s  complete=%t  total=%d  matched=%d  returned=%d  truncated=%t\n",
		result.ScanID,
		result.Complete,
		result.TotalCount,
		result.MatchedCount,
		result.ReturnedCount,
		result.TargetsTruncated,
	)
	for name, summary := range result.ByUser {
		fmt.Fprintf(&b, "  %s  targets=%d  %v\n", name, summary.Targets, summary.Statuses)
	}
	return b.String()
}
