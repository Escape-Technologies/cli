package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const (
	// retestReportPageSizeMax is the public API cap for GET /retests/{id} report pages.
	retestReportPageSizeMax = 100
	// retestContextMaxLen matches the public API limit on operator context.
	retestContextMaxLen = 4000
	retestProgressScale = 100
)

var (
	retestListProfileIDs    []string
	retestListSortType      string
	retestListSortDirection string
	retestListPage          pageFlags

	retestGetCursor string
	retestGetSize   int

	retestStartProfileID    string
	retestStartIssueIDs     []string
	retestStartFilterIDs    []string
	retestStartAssetIDs     []string
	retestStartSeverities   []string
	retestStartTagIDs       []string
	retestStartScannerKinds []string
	retestStartStatuses     []string
	retestStartSearch       string
	retestStartContext      string
	retestStartInitiator    string
)

var retestsCmd = &cobra.Command{
	Use:     "retests",
	Aliases: []string{"retest"},
	Short:   "List, start, and inspect issue retests",
	Long: `Manage Issue Retests - Verify a Fix

Retests rerun selected issues on a DAST or AI pentest profile. The retest id
is the scan id returned by start. Poll get to read each issue outcome
(VERIFIED_FIXED, VULNERABLE, NOT_TESTED, or FAILED).

COMMON WORKFLOWS:
  • Retest specific issues:
    $ escape-cli retests start --profile-id <profile-id> --issue-id <issue-id>

  • Retest issues matching a filter:
    $ escape-cli retests start --profile-id <profile-id> --severity HIGH --status OPEN

  • Read outcomes:
    $ escape-cli retests get <retest-id>`,
}

var retestsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List issue retests",
	Long: `List Issue Retests - Review Verification Runs

List retests for the organization. Each id is the scan id of that retest.
Results are newest first unless --sort-by or --sort-direction is set.

FILTER OPTIONS:
  -p, --profile-id    Filter by one or more profile IDs
  --sort-by           createdAt or SEVERITY
  --sort-direction    asc or desc (default: desc)
  --size              Page size, from 1 to 100. With --size or --cursor, return one page
  --cursor            Page cursor (nextCursor from the previous page)`,
	Example: `  # List the newest retests
  escape-cli retests list

  # List retests for one profile
  escape-cli retests list --profile-id 00000000-0000-0000-0000-000000000000

  # Export the list
  escape-cli retests list -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[v3.RetestDetailed]{}) {
			return nil
		}

		if err := validateRetestSort(retestListSortType, retestListSortDirection); err != nil {
			return err
		}

		if err := validatePageFlags(cmd, retestListPage.size); err != nil {
			return err
		}

		filters := &escape.ListRetestsFilters{
			ProfileIDs:    retestListProfileIDs,
			SortType:      retestListSortType,
			SortDirection: retestListSortDirection,
		}
		// Same walk as runPagedList. A profile filter that matches nothing is
		// still an error in pretty mode, and that decision has to happen
		// before anything is printed. JSON and YAML print the empty page or
		// list so scripts and MCP do not treat "no rows" as a failure.
		single := singlePageRequested(cmd)
		retests, next, total, err := resolveList(cmd.Context(), single, retestListPage.cursor, retestListPage.size, func(ctx context.Context, cursor string, size int) ([]v3.RetestDetailed, *string, int, error) {
			rows, pageNext, pageTotal, err := escape.ListRetests(ctx, cursor, filters, size)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("unable to list retests: %w", err)
			}

			if rows == nil {
				rows = []v3.RetestDetailed{}
			}

			return rows, pageNext, pageTotal, nil
		})
		if err != nil {
			return err
		}

		if retests == nil {
			retests = []v3.RetestDetailed{}
		}

		if retestListEmptyIsError(retestListProfileIDs, len(retests), isPrettyOutput()) {
			return fmt.Errorf(
				"no retests found for profile ID(s) %s",
				strings.Join(retestListProfileIDs, ", "),
			)
		}

		emitList(single, retests, next, total, func() []string {
			return retestListTable(retests)
		})

		return nil
	},
}

var retestsGetCmd = &cobra.Command{
	Use:     "get retest-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get a retest and its issue outcomes",
	Long: `Get Retest - Read Status and Per-Issue Outcomes

Fetch a retest by scan id. By default every report page is loaded and the
agents are returned together. Pass --cursor to read a single page instead.

OUTCOMES:
  • VERIFIED_FIXED  - the issue no longer reproduces
  • VULNERABLE      - the issue still reproduces
  • NOT_TESTED      - the agent could not exercise the issue
  • FAILED          - the retest agent failed`,
	Example: `  # Follow every report page
  escape-cli retests get 00000000-0000-0000-0000-000000000000

  # Machine-readable outcomes
  escape-cli retests get <retest-id> -o json

  # One report page
  escape-cli retests get <retest-id> --cursor <cursor> --size 50`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("retest ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetRetest200Response{}) {
			return nil
		}

		size, err := retestReportPageSize(retestGetSize)
		if err != nil {
			return err
		}

		cursor := strings.TrimSpace(retestGetCursor)
		follow := cursor == ""
		retest, err := collectRetestReports(cmd.Context(), args[0], cursor, size, follow, escape.GetRetest)
		if err != nil {
			return fmt.Errorf("unable to get retest: %w", err)
		}

		printRetest(retest)

		return nil
	},
}

var retestsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an issue retest",
	Long: `Start Retest - Verify Issues After a Fix

Start a retest on a DAST or AI pentest profile. Provide exactly one of:
  • --issue-id (repeatable): explicit issue IDs
  • filter flags, or a "filter" object in the JSON body

Filter flags map to the public API filter: --filter-id, --asset-id,
--severity, --tag-id, --scanner-kind, --status, and --search.
--initiator defaults to MANUAL on the server when omitted.

The same request can be sent as JSON on stdin. Flags override the body.
escape-cli retests start --input-schema prints the body schema.`,
	Example: `  # Retest two issues
  escape-cli retests start --profile-id <profile-id> \
    --issue-id <issue-id> --issue-id <issue-id>

  # Retest open high issues on a profile
  escape-cli retests start --profile-id <profile-id> \
    --severity HIGH --status OPEN \
    --context "Verify the auth bypass fix"

  # Body on stdin; flags override matching fields
  echo '{"issueIds":["<issue-id>"]}' | \
    escape-cli retests start --profile-id <profile-id>

  # Every issue on the profile
  echo '{"profileId":"<profile-id>","filter":{}}' | escape-cli retests start`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.StartRetestRequest{}) {
			return nil
		}

		if out.Schema(v3.StartRetest200Response{}) {
			return nil
		}

		// cmd.InOrStdin keeps a reader supplied through cmd.SetIn. The body is
		// read even when flags name the profile and selection: it can still
		// carry context and initiator, and MCP pipes it next to those flags.
		body, err := readPipedStdin(cmd.InOrStdin())
		if err != nil {
			return err
		}

		request, err := buildStartRetestRequest(startRetestInputFromCommand(cmd), body)
		if err != nil {
			return err
		}

		retest, err := escape.StartRetest(cmd.Context(), *request)
		if err != nil {
			return fmt.Errorf("unable to start retest: %w", err)
		}

		printRetest(retest)

		return nil
	},
}

// startRetestBody is the stdin shape of POST /retests.
// It is decoded without the generated required-field check so --profile-id can
// fill profileId when the body only carries issueIds or filter.
type startRetestBody struct {
	ProfileID string                       `json:"profileId"`
	IssueIDs  []string                     `json:"issueIds"`
	Filter    *v3.StartRetestRequestFilter `json:"filter"`
	Context   *string                      `json:"context"`
	Initiator *v3.ENUMPROPERTIESINITIATOR  `json:"initiator"`
}

// startRetestInput is the flag side of POST /retests. A set bit means the user
// passed that flag, which overrides the same field from stdin.
type startRetestInput struct {
	profileID       string
	profileSet      bool
	issueIDs        []string
	issueIDsSet     bool
	filterIDs       []string
	filterIDsSet    bool
	assetIDs        []string
	assetIDsSet     bool
	severities      []string
	severitiesSet   bool
	tagIDs          []string
	tagIDsSet       bool
	scannerKinds    []string
	scannerKindsSet bool
	statuses        []string
	statusesSet     bool
	search          string
	searchSet       bool
	context         string
	contextSet      bool
	initiator       string
	initiatorSet    bool
}

func (input startRetestInput) filterFlagsSet() bool {
	return input.filterIDsSet ||
		input.assetIDsSet ||
		input.severitiesSet ||
		input.tagIDsSet ||
		input.scannerKindsSet ||
		input.statusesSet ||
		input.searchSet
}

// buildStartRetestRequest merges a StartRetestRequest JSON body with CLI flags.
// Flags override the same fields from the body. The result has a profile id and
// exactly one of issueIds or filter, matching POST /retests.
func buildStartRetestRequest(input startRetestInput, body []byte) (*v3.StartRetestRequest, error) {
	request := &v3.StartRetestRequest{}
	if len(strings.TrimSpace(string(body))) > 0 {
		var decoded startRetestBody
		if err := json.Unmarshal(body, &decoded); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}

		request.ProfileId = decoded.ProfileID
		request.IssueIds = decoded.IssueIDs
		request.Filter = decoded.Filter
		request.Context = decoded.Context
		request.Initiator = decoded.Initiator
	}

	if input.profileSet {
		request.ProfileId = input.profileID
	}

	request.ProfileId = strings.TrimSpace(request.ProfileId)
	if request.ProfileId == "" {
		return nil, errors.New("--profile-id is required")
	}

	if input.contextSet {
		request.Context = optionalString(input.context)
	}

	if request.HasContext() && utf8.RuneCountInString(request.GetContext()) > retestContextMaxLen {
		return nil, fmt.Errorf("context must be at most %d characters", retestContextMaxLen)
	}

	if input.initiatorSet {
		initiator, err := optionalInitiator(input.initiator)
		if err != nil {
			return nil, err
		}

		request.Initiator = initiator
	}

	if err := applyRetestSelection(request, input); err != nil {
		return nil, err
	}

	return request, nil
}

func applyRetestSelection(request *v3.StartRetestRequest, input startRetestInput) error {
	filterFlags := input.filterFlagsSet()
	if input.issueIDsSet && filterFlags {
		return errors.New("provide either issueIds or filter, not both")
	}

	if input.issueIDsSet {
		issueIDs := compactStrings(input.issueIDs)
		if len(issueIDs) == 0 {
			return errors.New("provide exactly one of issueIds or filter")
		}

		request.IssueIds = issueIDs
		request.Filter = nil
	}

	if filterFlags {
		filter, err := applyRetestFilter(request.Filter, input)
		if err != nil {
			return err
		}

		request.Filter = filter
		request.IssueIds = nil
	}

	request.IssueIds = compactStrings(request.IssueIds)

	hasIssueIDs := len(request.IssueIds) > 0
	hasFilter := request.Filter != nil
	if hasIssueIDs == hasFilter {
		return errors.New("provide exactly one of issueIds or filter")
	}

	return nil
}

func applyRetestFilter(base *v3.StartRetestRequestFilter, input startRetestInput) (*v3.StartRetestRequestFilter, error) {
	filter := v3.StartRetestRequestFilter{}
	if base != nil {
		filter = *base
	}

	if input.filterIDsSet {
		filter.Ids = compactStrings(input.filterIDs)
	}

	if input.assetIDsSet {
		filter.AssetIds = compactStrings(input.assetIDs)
	}

	if input.tagIDsSet {
		filter.TagIds = compactStrings(input.tagIDs)
	}

	if input.severitiesSet {
		severities, err := parseStringEnums(input.severities, "severity", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues)
		if err != nil {
			return nil, err
		}

		filter.Severities = severities
	}

	if input.scannerKindsSet {
		kinds, err := parseStringEnums(input.scannerKinds, "scanner kind", v3.AllowedENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMSEnumValues)
		if err != nil {
			return nil, err
		}

		filter.ScannerKinds = kinds
	}

	if input.statusesSet {
		statuses, err := parseStringEnums(input.statuses, "status", v3.AllowedENUMPROPERTIESFILTERPROPERTIESSTATUSITEMSEnumValues)
		if err != nil {
			return nil, err
		}

		filter.Status = statuses
	}

	if input.searchSet {
		filter.Search = optionalString(input.search)
	}

	return &filter, nil
}

func startRetestInputFromCommand(cmd *cobra.Command) startRetestInput {
	flags := cmd.Flags()

	return startRetestInput{
		profileID:       retestStartProfileID,
		profileSet:      flags.Changed("profile-id"),
		issueIDs:        retestStartIssueIDs,
		issueIDsSet:     flags.Changed("issue-id"),
		filterIDs:       retestStartFilterIDs,
		filterIDsSet:    flags.Changed("filter-id"),
		assetIDs:        retestStartAssetIDs,
		assetIDsSet:     flags.Changed("asset-id"),
		severities:      retestStartSeverities,
		severitiesSet:   flags.Changed("severity"),
		tagIDs:          retestStartTagIDs,
		tagIDsSet:       flags.Changed("tag-id"),
		scannerKinds:    retestStartScannerKinds,
		scannerKindsSet: flags.Changed("scanner-kind"),
		statuses:        retestStartStatuses,
		statusesSet:     flags.Changed("status"),
		search:          retestStartSearch,
		searchSet:       flags.Changed("search"),
		context:         retestStartContext,
		contextSet:      flags.Changed("context"),
		initiator:       retestStartInitiator,
		initiatorSet:    flags.Changed("initiator"),
	}
}

type getRetestPage func(context.Context, string, string, int) (*v3.GetRetest200Response, error)

// retestListEmptyIsError reports whether a profile filter with no rows should
// fail the command. Pretty mode keeps the human-readable error. Every other
// mode prints an empty list so -o json and MCP are not a failed tool call.
func retestListEmptyIsError(profileIDs []string, count int, pretty bool) bool {
	return len(profileIDs) > 0 && count == 0 && pretty
}

func retestListTable(retests []v3.RetestDetailed) []string {
	rows := []string{"ID\tPROFILE ID\tSTATUS\tPROGRESS\tCREATED AT\tFINISHED AT\tAGENTS"}
	for _, retest := range retests {
		rows = append(rows, formatRetestRow(&retest))
	}

	return rows
}

func retestInitiatorUsage() string {
	return fmt.Sprintf("initiator: %v (server default: MANUAL)", v3.AllowedENUMPROPERTIESINITIATOREnumValues)
}

// collectRetestReports loads a retest. When follow is false, it returns the
// requested reports page. When follow is true, it walks nextCursor and
// concatenates agents onto the first page.
func collectRetestReports(
	ctx context.Context,
	retestID string,
	cursor string,
	size int,
	follow bool,
	get getRetestPage,
) (*v3.GetRetest200Response, error) {
	if strings.TrimSpace(retestID) == "" {
		return nil, errors.New("retest ID is required")
	}

	page, err := get(ctx, retestID, cursor, size)
	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, errors.New("empty retest response")
	}

	if !follow {
		return page, nil
	}

	agents := append([]v3.RetestAgent{}, page.GetAgents()...)
	seen := map[string]struct{}{}
	if cursor != "" {
		seen[cursor] = struct{}{}
	}

	current := page
	for {
		next, ok := retestNextCursor(current)
		if !ok {
			break
		}

		if _, dup := seen[next]; dup {
			break
		}

		seen[next] = struct{}{}
		current, err = get(ctx, retestID, next, size)
		if err != nil {
			return nil, err
		}

		if current == nil {
			return nil, errors.New("empty retest response")
		}

		agents = append(agents, current.GetAgents()...)
	}

	page.SetAgents(agents)
	page.NextCursor = nil

	return page, nil
}

func validateRetestSort(sortType, sortDirection string) error {
	switch sortType {
	case "", "createdAt", "SEVERITY":
	default:
		return fmt.Errorf("invalid --sort-by %q; valid values: createdAt, SEVERITY", sortType)
	}

	switch sortDirection {
	case "", "asc", "desc":
	default:
		return fmt.Errorf("invalid --sort-direction %q; valid values: asc, desc", sortDirection)
	}

	return nil
}

func retestReportPageSize(size int) (int, error) {
	if size == 0 {
		return retestReportPageSizeMax, nil
	}

	if size < 1 || size > retestReportPageSizeMax {
		return 0, fmt.Errorf("--size must be between 1 and %d", retestReportPageSizeMax)
	}

	return size, nil
}

type retestRow interface {
	GetId() string
	GetProfileId() string
	GetStatus() string
	GetProgressRatio() float32
	GetCreatedAt() string
	GetFinishedAt() string
	GetAgents() []v3.RetestAgent
}

func printRetest(retest retestRow) {
	out.Table(retest, func() []string {
		return []string{
			"ID\tPROFILE ID\tSTATUS\tPROGRESS\tCREATED AT\tFINISHED AT\tAGENTS",
			formatRetestRow(retest),
		}
	})
	if !isPrettyOutput() {
		return
	}

	agents := retest.GetAgents()
	out.Table(agents, func() []string {
		rows := []string{"ISSUE ID\tAGENT ID\tSTATUS\tOUTCOME"}
		for _, agent := range agents {
			rows = append(rows, formatRetestAgent(agent))
		}

		return rows
	})
}

func formatRetestRow(retest retestRow) string {
	return fmt.Sprintf("%s\t%s\t%s\t%d%%\t%s\t%s\t%d",
		retest.GetId(),
		retest.GetProfileId(),
		retest.GetStatus(),
		int(retest.GetProgressRatio()*retestProgressScale),
		retest.GetCreatedAt(),
		formatOptional(retest.GetFinishedAt()),
		len(retest.GetAgents()),
	)
}

func formatRetestAgent(agent v3.RetestAgent) string {
	agentID := "-"
	if agent.HasId() {
		agentID = agent.GetId()
	}

	status := "-"
	if agent.HasStatus() {
		status = string(agent.GetStatus())
	}

	outcome := "-"
	if agent.HasOutcome() {
		outcome = string(agent.GetOutcome())
	}

	return fmt.Sprintf("%s\t%s\t%s\t%s", agent.GetIssueId(), agentID, status, outcome)
}

func formatOptional(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}

	return value
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func optionalInitiator(value string) (*v3.ENUMPROPERTIESINITIATOR, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	initiator, err := parseStringEnum(value, "initiator", v3.AllowedENUMPROPERTIESINITIATOREnumValues)
	if err != nil {
		return nil, err
	}

	return &initiator, nil
}

func compactStrings(values []string) []string {
	compact := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			compact = append(compact, value)
		}
	}

	if len(compact) == 0 {
		return nil
	}

	return compact
}

func parseStringEnums[T ~string](values []string, name string, allowed []T) ([]T, error) {
	parsed := make([]T, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		match, err := parseStringEnum(value, name, allowed)
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, match)
	}

	if len(parsed) == 0 {
		return nil, nil
	}

	return parsed, nil
}

func parseStringEnum[T ~string](value, name string, allowed []T) (T, error) {
	parsed := T(value)
	for _, candidate := range allowed {
		if candidate == parsed {
			return parsed, nil
		}
	}

	labels := make([]string, len(allowed))
	for i, candidate := range allowed {
		labels[i] = string(candidate)
	}

	var zero T

	return zero, fmt.Errorf("invalid %s %q; valid values: %s", name, value, strings.Join(labels, ", "))
}

func cursorValue(cursor *string) (string, bool) {
	if cursor == nil {
		return "", false
	}

	value := strings.TrimSpace(*cursor)
	if value == "" {
		return "", false
	}

	return value, true
}

func retestNextCursor(page *v3.GetRetest200Response) (string, bool) {
	if page == nil {
		return "", false
	}

	return cursorValue(page.NextCursor)
}

func init() {
	retestsListCmd.Flags().StringSliceVarP(&retestListProfileIDs, "profile-id", "p", nil, "filter by profile ID(s)")
	retestsListCmd.Flags().StringVar(&retestListSortType, "sort-by", "", "sort field (createdAt or SEVERITY)")
	retestsListCmd.Flags().StringVar(&retestListSortDirection, "sort-direction", "", "sort direction: asc, desc")
	retestListPage.bind(retestsListCmd)

	retestsGetCmd.Flags().StringVar(&retestGetCursor, "cursor", "", "return this reports page only, instead of following every page")
	retestsGetCmd.Flags().IntVar(&retestGetSize, "size", 0, "reports per page (1-100, default 100)")

	retestsStartCmd.Flags().StringVarP(&retestStartProfileID, "profile-id", "p", "", "profile ID to retest")
	retestsStartCmd.Flags().StringSliceVarP(&retestStartIssueIDs, "issue-id", "i", nil, "explicit issue IDs to retest; mutually exclusive with filter flags")
	retestsStartCmd.Flags().StringSliceVar(&retestStartFilterIDs, "filter-id", nil, "filter.ids: issue IDs combined with other filter flags")
	retestsStartCmd.Flags().StringSliceVar(&retestStartAssetIDs, "asset-id", nil, "filter by asset ID(s)")
	retestsStartCmd.Flags().StringSliceVar(&retestStartSeverities, "severity", nil, fmt.Sprintf("filter by severity: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues))
	retestsStartCmd.Flags().StringSliceVar(&retestStartTagIDs, "tag-id", nil, "filter by tag ID(s)")
	retestsStartCmd.Flags().StringSliceVar(&retestStartScannerKinds, "scanner-kind", nil, fmt.Sprintf("filter by scanner kind: %v", v3.AllowedENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMSEnumValues))
	retestsStartCmd.Flags().StringSliceVar(&retestStartStatuses, "status", nil, fmt.Sprintf("filter by issue status: %v", v3.AllowedENUMPROPERTIESFILTERPROPERTIESSTATUSITEMSEnumValues))
	retestsStartCmd.Flags().StringVar(&retestStartSearch, "search", "", "filter issues by name or description")
	retestsStartCmd.Flags().StringVar(&retestStartContext, "context", "", "operator context passed to the retest agents")
	retestsStartCmd.Flags().StringVar(&retestStartInitiator, "initiator", "", retestInitiatorUsage())

	retestsCmd.AddCommand(retestsListCmd, retestsGetCmd, retestsStartCmd)
	rootCmd.AddCommand(retestsCmd)
}
