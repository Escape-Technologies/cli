package cmd

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const (
	scanTargetsPageSize           = 100
	defaultCoverageTargetListSize = 200
	maxCoverageTargetListSize     = 500
	coverageOKExamples            = 8
	coverageToolGuidance          = "overall and byUser are exhaustive over every fetched target when complete=true, even if targetsTruncated is true. " +
		"byType counts every target in the scan by kind and sums to totalCount. " +
		"For how many routes/resolvers are OK, covered, or failing, answer from overall.statuses with overall.targets as the denominator (totalCount also counts non-API targets). " +
		"A route is OK overall when at least one user got OK on it. byUser.statuses counts per scanner user: a route OK for several users counts once per user, so never sum byUser to get route totals. " +
		"Do not count the targets array or use scans_reasoning; the targets array is a compact sample."
	coverageBlocklisted = "BLOCKLISTED"
	coverageSkipped     = "SKIPPED"
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

// ScanCoverageSummary is status totals across the full scan, overall or for one user.
type ScanCoverageSummary struct {
	Targets    int            `json:"targets"`
	Statuses   map[string]int `json:"statuses"`
	OKExamples []string       `json:"okExamples,omitempty"`
}

// ScanCoverage is the MCP-facing coverage payload.
type ScanCoverage struct {
	ScanID           string                         `json:"scanId"`
	Complete         bool                           `json:"complete"`
	TotalCount       int                            `json:"totalCount"`
	MatchedCount     int                            `json:"matchedCount"`
	ReturnedCount    int                            `json:"returnedCount"`
	TargetsTruncated bool                           `json:"targetsTruncated"`
	Guidance         string                         `json:"guidance"`
	ByType           map[string]int                 `json:"byType"`
	Overall          ScanCoverageSummary            `json:"overall"`
	ByUser           map[string]ScanCoverageSummary `json:"byUser"`
	Targets          []ScanCoverageTarget           `json:"targets"`
}

var scansCoverageCmd = &cobra.Command{
	Use:     "coverage scan-id",
	Aliases: []string{"cov"},
	Short:   "API coverage totals for a scan, overall and per scanner user. Fetches all pages. Source of truth for how many routes are OK and for successful requests by user.",
	Long: `Scan Coverage By User

Walks every page of scan targets and aggregates coverage overall and per user.

OUTPUT SHAPE (JSON):
  {
    "complete": true,            // all target pages were fetched
    "totalCount": <int>,         // targets in the scan, including non-API targets
    "matchedCount": <int>,       // targets after --coverage/--user filters
    "targetsTruncated": <bool>,  // targets array was capped by --size (false when --size 0)
    "byType": { "API_ROUTE": n, "WEB_PAGE": n, ... },  // sums to totalCount
    "overall": { "targets": n, "statuses": {"OK": n, "SKIPPED": n, ...}, "okExamples": [...] },
    "byUser": { "<user>": { "targets": n, "statuses": {"OK": n, ...}, "okExamples": [...] } },
    "targets": [ compact routes ]
  }

overall counts each API route and GraphQL resolver once, by the status shown in the
coverage table (BLOCKLISTED / SKIPPED when it has no coverage). byUser counts each
route once per scanner user, so its totals overlap. Both are computed over every
fetched target, even when the targets array is capped.
Use this command for "how many routes are OK?" or "did users A and B make successful
requests?" — not scans reasoning.`,
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
	byType := map[string]int{}
	for _, target := range raw {
		row := compactScanTarget(target)
		allRows = append(allRows, row)
		byType[row.Type]++
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
		ByType:           byType,
		Overall:          summarizeCoverage(allRows),
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
			return nil, 0, false, fmt.Errorf("unable to list targets: %w", err)
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
		row.Name = route.GetName()
		row.RequestCount = int(route.GetRequestCount())
		row.Coverage = coverageStatus(string(route.GetCoverage()), route.GetBlacklisted())
		row.CoverageByUser = compactCoverageByUser(route.GetCoverageByUser())

		return row
	}

	if resolver, ok := target.GetGraphqlResolverOk(); ok && resolver != nil {
		row.Type = "GRAPHQL_RESOLVER"
		row.Name = resolver.GetDisplayName()
		row.RequestCount = int(resolver.GetRequestCount())
		row.Coverage = coverageStatus(string(resolver.GetCoverage()), resolver.GetBlacklisted())
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

	if page, ok := target.GetWebPageOk(); ok && page != nil {
		row.Type = "WEB_PAGE"
		row.Name = page.GetUrl()

		return row
	}

	if crawled, ok := target.GetWebCrawledUrlOk(); ok && crawled != nil {
		row.Type = "WEB_CRAWLED_URL"
		row.Name = crawled.GetUrl()

		return row
	}

	if cve, ok := target.GetCveOk(); ok && cve != nil {
		row.Type = "CVE"
		row.Name = cve.GetCveId()

		return row
	}

	row.Type = "OTHER"

	return row
}

// coverageStatus mirrors the coverage table's Status column, which labels routes
// without any recorded coverage as blocklisted or skipped.
func coverageStatus(coverage string, blacklisted bool) string {
	switch {
	case coverage != "":
		return coverage
	case blacklisted:
		return coverageBlocklisted
	default:
		return coverageSkipped
	}
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

// summarizeCoverage counts each API route and GraphQL resolver once by its overall status.
func summarizeCoverage(rows []ScanCoverageTarget) ScanCoverageSummary {
	summary := ScanCoverageSummary{Statuses: map[string]int{}}
	for _, row := range rows {
		if row.Type != "API_ROUTE" && row.Type != "GRAPHQL_RESOLVER" {
			continue
		}

		summary.add(row, row.Coverage)
	}

	return summary
}

func summarizeCoverageByUser(rows []ScanCoverageTarget) map[string]ScanCoverageSummary {
	summaries := map[string]ScanCoverageSummary{}
	for _, row := range rows {
		for _, entry := range row.CoverageByUser {
			current := summaries[entry.Name]
			if current.Statuses == nil {
				current.Statuses = map[string]int{}
			}

			status := entry.Coverage
			if status == "" {
				status = "UNKNOWN"
			}

			current.add(row, status)
			summaries[entry.Name] = current
		}
	}

	return summaries
}

func (summary *ScanCoverageSummary) add(row ScanCoverageTarget, status string) {
	summary.Targets++
	summary.Statuses[status]++
	if !strings.EqualFold(status, "OK") || len(summary.OKExamples) >= coverageOKExamples {
		return
	}

	label := row.Name
	if row.Method != "" {
		label = row.Method + " " + row.Name
	}

	summary.OKExamples = append(summary.OKExamples, label)
}

func formatScanCoveragePretty(result ScanCoverage) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Scan %s\n", result.ScanID)
	if !result.Complete {
		b.WriteString("WARNING: not every target page was fetched, counts below are partial.\n")
	}

	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0) //nolint:mnd
	row := func(format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }
	row("\nTargets in scan\t%d\n", result.TotalCount)
	if result.MatchedCount != result.TotalCount {
		row("Targets matching --coverage/--user\t%d\n", result.MatchedCount)
	}

	_ = w.Flush()

	b.WriteString("\nTargets by type\n")
	types := slices.Sorted(maps.Keys(result.ByType))
	slices.SortStableFunc(types, func(left, right string) int { return result.ByType[right] - result.ByType[left] })
	for _, kind := range types {
		row("  %s\t%d\n", kind, result.ByType[kind])
	}

	_ = w.Flush()

	fmt.Fprintf(&b, "\nCoverage of %d API routes and GraphQL resolvers (each counted once)\n", result.Overall.Targets)
	for _, status := range coverageStatusOrder(result, []ScanCoverageSummary{result.Overall}) {
		count := result.Overall.Statuses[status]
		row("  %s\t%d\t%.1f%%\n", status, count, 100*float64(count)/float64(result.Overall.Targets)) //nolint:mnd
	}

	_ = w.Flush()

	if len(result.ByUser) == 0 {
		return b.String()
	}

	statuses := coverageStatusOrder(result, slices.Collect(maps.Values(result.ByUser)))
	b.WriteString("\nCoverage by scanner user (a route counts once per user, so rows overlap)\n")
	row("  USER\tROUTES\t%s\n", strings.Join(statuses, "\t"))
	users := slices.Sorted(maps.Keys(result.ByUser))
	slices.SortStableFunc(users, func(left, right string) int {
		return result.ByUser[right].Targets - result.ByUser[left].Targets
	})
	for _, user := range users {
		summary := result.ByUser[user]
		counts := make([]string, 0, len(statuses))
		for _, status := range statuses {
			counts = append(counts, strconv.Itoa(summary.Statuses[status]))
		}

		row("  %s\t%d\t%s\n", user, summary.Targets, strings.Join(counts, "\t"))
	}

	_ = w.Flush()

	return b.String()
}

// coverageStatusOrder lists every status seen in summaries: OK first, then by overall count.
func coverageStatusOrder(result ScanCoverage, summaries []ScanCoverageSummary) []string {
	seen := map[string]struct{}{}
	for _, summary := range summaries {
		for status := range summary.Statuses {
			seen[status] = struct{}{}
		}
	}

	statuses := slices.Sorted(maps.Keys(seen))
	slices.SortStableFunc(statuses, func(left, right string) int {
		switch {
		case left == "OK":
			return -1
		case right == "OK":
			return 1
		default:
			return result.Overall.Statuses[right] - result.Overall.Statuses[left]
		}
	})

	return statuses
}
