package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// maxLatestEventsHydrationConcurrency caps how many events we hydrate in
// parallel from `issues get-with-events`. Matches the public API's hard cap of
// 5 latest event IDs per issue, so this is also the worst-case fan-out.
const maxLatestEventsHydrationConcurrency = 5

// formatLatestEventIDs renders the LATEST EVENTS column for `issues get`.
// Uses presence-aware inputs (pointers) so the three states stay distinct:
//   - ids == nil      → enrichment unavailable / field omitted by API → "(n/a)"
//   - len(*ids) == 0  → enrichment succeeded but no events             → "-"
//   - otherwise       → comma-joined IDs, with a trailing "…" when truncated
func formatLatestEventIDs(ids *[]string, truncated bool) string {
	if ids == nil {
		return "(n/a)"
	}

	if len(*ids) == 0 {
		return "-"
	}

	joined := strings.Join(*ids, ", ")
	if truncated {
		joined += ", …"
	}

	return joined
}

// IssueWithEvents bundles an issue with its hydrated latest-scan events so the
// MCP/CLI can hand an AI agent the full context (including each event's
// Exchange) in a single tool call.
//
// LatestEvents and LatestEventsTruncated are pointers so the JSON output can
// distinguish "field omitted by the API" (nil, key absent) from "no events"
// (empty slice) or "not truncated" (false). This mirrors the public API's
// optional wire contract — see `IssueDetailedSchema` in services/public-api.
type IssueWithEvents struct {
	Issue                 v3.GetIssue200Response    `json:"issue"`
	LatestEvents          *[]v3.GetEvent200Response `json:"latestEvents,omitempty"`
	LatestEventsTruncated *bool                     `json:"latestEventsTruncated,omitempty"`
	EventErrors           []IssueEventHydrateError  `json:"eventErrors,omitempty"`
}

// IssueEventHydrateError captures a per-event hydration failure so partial
// success is observable rather than swallowed.
type IssueEventHydrateError struct {
	EventID string `json:"eventId"`
	Error   string `json:"error"`
}

// optionalBool returns a pointer to the flag value, or nil when it is the zero
// value, so an unset boolean filter is not forwarded to the API at all.
func optionalBool(value bool) *bool {
	if !value {
		return nil
	}

	return &value
}

// issueListFilters assembles the GET /issues filter set from the `issues list`
// flags. The boolean API filters are string flags ("true"/"false") so an
// explicit false is forwarded instead of being dropped. noTags only honours
// true, so its flag stays nil unless set.
func issueListFilters() *escape.ListIssuesFilters {
	return &escape.ListIssuesFilters{
		Status:           issueStatus,
		Severities:       issueSeverity,
		ProfileIDs:       profileIDs,
		AssetIDs:         assetIDs,
		Domains:          domains,
		IssueIDs:         issueIDs,
		ScanIDs:          scanIDs,
		TagsIDs:          tagsIDs,
		Risks:            risks,
		Search:           search,
		AssetClasses:     assetClasses,
		JiraTicket:       jiraTicket,
		ScannerKinds:     issueScannerKinds,
		Names:            issueNames,
		ProjectIDs:       issueProjectIDs,
		TargetIDs:        issueTargetIDs,
		Categories:       issueCategories,
		AssetTypes:       issueAssetTypes,
		AssetStatuses:    issueAssetStatuses,
		SecurityTestUids: issueSecurityTestUids,
		BlacklistedIDs:   issueBlacklistedIDs,
		BlacklistedNames: issueBlacklistedNames,
		AiFalsePositive:  strings.TrimSpace(issueAiFalsePositive),
		Agentic:          strings.TrimSpace(issueAgentic),
		NoTags:           optionalBool(issueNoTags),
		Dnf:              strings.TrimSpace(issueDnf),
	}
}

func formatIssueCompliances(items []v3.GetIssue200ResponseCompliancesInner) string {
	if len(items) == 0 {
		return "-"
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		if item.GetFramework() == "" && item.GetItem() == "" {
			continue
		}

		values = append(values, strings.Trim(strings.Join([]string{item.GetFramework(), item.GetItem()}, ":"), ":"))
	}

	if len(values) == 0 {
		return "-"
	}

	return strings.Join(values, ", ")
}

var validIssueSortFields = map[string]struct{}{
	"LAST_SEEN":  {},
	"FIRST_SEEN": {},
	"SEVERITY":   {},
	"STATUS":     {},
}

var (
	issueUpdateStatusStr string
	issueUpdateReason    string
	issueUpdateSeverity  string
	issueResetSeverity   bool
	issueSortType        string
	issueSortDirection   string
	issueSeverity        []string
	issueStatus          []string
	profileIDs           []string
	assetIDs             []string
	domains              []string
	issueIDs             []string
	scanIDs              []string
	tagsIDs              []string
	search               string
	jiraTicket           string
	risks                []string
	assetClasses         []string
	issueScannerKinds    []string
	issueNames           []string
	issueListPage        pageFlags

	issueProjectIDs       []string
	issueTargetIDs        []string
	issueCategories       []string
	issueAssetTypes       []string
	issueAssetStatuses    []string
	issueSecurityTestUids []string
	issueBlacklistedIDs   []string
	issueBlacklistedNames []string
	issueAiFalsePositive  string
	issueAgentic          string
	issueNoTags           bool
	issueDnf              string
)

var issuesCmd = &cobra.Command{
	Use:     "issues",
	Aliases: []string{"issue"},
	Short:   "Manage and track security vulnerabilities",
	Long: `Manage Security Issues - Track and Remediate Vulnerabilities

Issues are security vulnerabilities, misconfigurations, and compliance violations
discovered during security scans. Each issue represents a specific security concern
that should be reviewed and remediated.

ISSUE LIFECYCLE:
  1. OPEN              - Newly discovered, needs review
  2. MANUAL_REVIEW     - Under investigation
  3. IN_PROGRESS       - Actively being fixed
  4. RESOLVED          - Fixed and verified
  5. FALSE_POSITIVE    - Not a real issue
  6. ACCEPTED_RISK     - Acknowledged but not fixing

COMMON WORKFLOWS:
  • List high-priority issues:
    $ escape-cli issues list --severity HIGH,CRITICAL --status OPEN

  • Review issues for a specific asset:
    $ escape-cli issues list --asset-id <asset-id>

  • Update issue status as you fix them:
    $ escape-cli issues update <issue-id> --status IN_PROGRESS

  • Track issue history:
    $ escape-cli issues list-activities <issue-id>`,
}

var issueListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List security issues with powerful filtering",
	Long: `List Security Issues - Query Your Vulnerability Database

List and filter security issues across your entire organization. Use powerful
filtering options to find exactly the issues you need to review or remediate.

FILTER OPTIONS:
  --severity         Filter by severity: CRITICAL, HIGH, MEDIUM, LOW, INFO
  --status           Filter by status: OPEN, MANUAL_REVIEW, IN_PROGRESS, RESOLVED
  -p, --profile-id   Filter by profile ID
  -a, --asset-id     Filter by asset ID
  -d, --domain       Filter by domain name
  -i, --issue-id     Filter by specific issue IDs
  --scan-id          Filter by scan ID
  -t, --tag-id       Filter by tag ID
  -r, --risk         Filter by risk level
  --asset-class      Filter by asset classification
  --project-id       Filter by project ID
  --target-id        Filter by target ID
  --category         Filter by issue category
  --asset-type       Filter by asset type
  --asset-status     Filter by asset status
  --security-test-uid Filter by security test UID
  --blacklisted-id   Exclude these issue IDs
  --blacklisted-name Exclude issues by their raw name
  --ai-false-positive Filter by AI false positive classification (true/false)
  --agentic          Filter by agentic (AI pentest) issues (true/false)
  --no-tags           Filter by issues whose assets have no tags (true only)
  --dnf              Advanced filter as a DNF expression (URL-encoded JSON object)
  -s, --search       Free-text search across issue names

SEVERITY PRIORITY:
  🔴 CRITICAL   - Critical security flaws, immediate action required
  🟠 HIGH       - Serious vulnerabilities, fix ASAP
  🟡 MEDIUM     - Moderate risk, schedule fixes
  🔵 LOW        - Minor issues, address when possible
  ⚪ INFO       - Informational findings

Example output:
ID                                      CREATED AT  SEVERITY  STATUS  NAME                        ASSET           LINK
00000000-0000-0000-0000-000000000001    2025-07-24  HIGH      OPEN    SQL Injection               api.example.com https://...
00000000-0000-0000-0000-000000000002    2025-07-23  MEDIUM    OPEN    Misconfigured CSP Header    my.app          https://...`,
	Example: `  # List all open critical and high severity issues
  escape-cli issues list --severity CRITICAL,HIGH --status OPEN

  # List issues for a specific asset
  escape-cli issues list --asset-id 00000000-0000-0000-0000-000000000000

  # Search for specific vulnerability types
  escape-cli issues list --search "SQL injection"

  # List issues from a specific scan
  escape-cli issues list --scan-id <scan-id>

  # Export to JSON for custom processing
  escape-cli issues list -o json | jq '.[] | select(.severity == "CRITICAL")'

  # List unresolved issues across all assets
  escape-cli issues list --status OPEN,MANUAL_REVIEW,IN_PROGRESS

  # List AI-classified false positives in a project
  escape-cli issues list --project-id <project-id> --ai-false-positive true

  # List agentic (AI pentest) findings of a given category
  escape-cli issues list --agentic true --category INJECTION`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Output JSON Schema if requested
		if out.Schema([]v3.IssueSummarized{}) {
			return nil
		}

		if issueSortDirection != "" && issueSortType == "" {
			return errors.New("--sort-direction requires --sort-by")
		}

		if issueSortType != "" {
			issueSortType = strings.ToUpper(issueSortType)
			if _, ok := validIssueSortFields[issueSortType]; !ok {
				return fmt.Errorf("invalid --sort-by %q; valid values: LAST_SEEN, FIRST_SEEN, SEVERITY, STATUS", issueSortType)
			}
		}

		switch normalizedDirection := strings.ToLower(issueSortDirection); normalizedDirection {
		case "":
		case "asc", "desc":
			issueSortDirection = normalizedDirection
		default:
			return fmt.Errorf("invalid --sort-direction %q; valid values: asc, desc", issueSortDirection)
		}

		filters := issueListFilters()
		if err := runPagedList(cmd, issueListPage, func(ctx context.Context, cursor string, size int) ([]v3.IssueSummarized, *string, int, error) {
			return escape.ListIssues(ctx, cursor, filters, issueSortType, issueSortDirection, size)
		}, func(issues []v3.IssueSummarized) []string {
			res := []string{"ID\tCREATED AT\tSEVERITY\tSTATUS\tNAME\tASSET\tLINK"}
			for _, issue := range issues {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), out.GetShortDate(issue.GetCreatedAt()), issue.GetSeverity(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetLinks().IssueOverview))
			}

			return res
		}); err != nil {
			return fmt.Errorf("unable to list issues: %w", err)
		}

		return nil
	},
}

var issueGetCmd = &cobra.Command{
	Use:     "get issue-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get detailed information about a security issue",
	Long: `Get Issue Details - View Complete Vulnerability Information

Retrieve comprehensive details about a specific security issue including severity,
category, status, affected asset, and a direct link to full remediation guidance.

DISPLAYED INFORMATION:
  • ID          - Unique issue identifier
  • CREATED AT  - When the issue was first discovered
  • SEVERITY    - Risk level (CRITICAL, HIGH, MEDIUM, LOW, INFO)
  • CATEGORY    - Vulnerability classification (e.g., INJECTION, AUTH, CRYPTO)
  • STATUS      - Current remediation status
  • NAME        - Human-readable vulnerability description
  • ASSET       - Affected API or application
  • LINK        - URL to detailed analysis and remediation steps

AI PENTEST DETAILS (pretty output):
  • Reproduction steps, exploits and attack chain
  • Impact and AI false-positive verdict
  • Linked ticket (external ID and URL)
  • Ready-to-use fix prompt for a coding agent
  All of these fields are also present in -o json output.

USE CASES:
  • Review vulnerability details before fixing
  • Share issue information with team members
  • Verify issue details in incident response
  • Get remediation guidance link
  • Hand the fix prompt to a coding agent

Example output:
ID                                      CREATED AT                SEVERITY  CATEGORY         STATUS  NAME              ASSET                  LINK
00000000-0000-0000-0000-000000000001    2025-06-26T06:03:26.128Z  HIGH      XXE Injection    OPEN    XML External...   api.example.com        https://...`,
	Example: `  # Get issue details
  escape-cli issues get 00000000-0000-0000-0000-000000000001

  # Get issue in JSON format
  escape-cli issues get <issue-id> -o json`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output JSON Schema if requested
		if out.Schema(v3.GetIssue200Response{}) {
			return nil
		}

		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		out.Table(issue, func() []string {
			cvssData := issue.GetCvss()
			cvss := "-"
			if cvssData.GetScore() != 0 || cvssData.GetVector() != "" {
				cvss = fmt.Sprintf("%.1f %s", cvssData.GetScore(), cvssData.GetVector())
			}

			var latestEventIDs *[]string
			if ids, ok := issue.GetLatestEventIdsOk(); ok {
				latestEventIDs = &ids
			}

			latestEvents := formatLatestEventIDs(latestEventIDs, issue.GetLatestEventsTruncated())
			res := []string{"ID\tCREATED AT\tSEVERITY\tCATEGORY\tSTATUS\tNAME\tASSET\tFIRST SEEN SCAN\tCVSS\tFRAMEWORK\tCOMPLIANCES\tREMEDIATION\tCONTEXT\tLATEST EVENTS\tLINK"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", issue.GetId(), issue.GetCreatedAt(), issue.GetSeverity(), issue.GetCategory(), issue.GetStatus(), issue.GetName(), issue.GetAsset().Name, issue.GetFirstSeenScanId(), cvss, issue.GetAiRemediationFramework(), formatIssueCompliances(issue.GetCompliances()), issue.GetRemediation(), issue.GetContext(), latestEvents, issue.GetLinks().IssueOverview))

			return res
		})

		printIssueDetails(issue)

		return nil
	},
}

// printIssueDetails renders the AI pentest detail fields of `issues get`:
// reproduction steps, exploits, attack chain, impact, the AI false-positive
// verdict, the linked ticket and the fix prompt. JSON and YAML output already
// carry every one of these fields on the issue document, so this only runs
// when stdout is the human table.
func printIssueDetails(issue *v3.GetIssue200Response) {
	if !isPrettyOutput() {
		return
	}

	if steps := issue.GetReproductionSteps(); len(steps) > 0 {
		out.Table(steps, func() []string {
			res := []string{"REPRODUCTION STEPS"}
			res = append(res, "ORDER\tTITLE\tCONTENT")
			for _, step := range steps {
				res = append(res, fmt.Sprintf("%d\t%s\t%s", step.GetOrder(), step.GetTitle(), step.GetContent()))
			}

			return res
		})
	}

	if exploits := issue.GetExploits(); len(exploits) > 0 {
		out.Table(exploits, func() []string {
			res := []string{"EXPLOITS"}
			res = append(res, exploits...)

			return res
		})
	}

	if chain := issue.GetAttackChain(); len(chain) > 0 {
		out.Table(chain, func() []string {
			res := []string{"ATTACK CHAIN"}
			res = append(res, "ORDER\tSTAGE\tTITLE\tDESCRIPTION\tCONTENT")
			for _, step := range chain {
				res = append(res, fmt.Sprintf("%d\t%s\t%s\t%s\t%s", step.GetOrder(), step.GetStage(), step.GetTitle(), step.GetDescription(), step.GetContent()))
			}

			return res
		})
	}

	verdict := issue.GetAiFalsePositive()
	out.Table(verdict, func() []string {
		res := []string{"AI FALSE POSITIVE"}
		res = append(res, "IS FALSE POSITIVE\tREASONING\tREASONING SUMMARY")
		res = append(res, fmt.Sprintf("%t\t%s\t%s", verdict.GetIsFalsePositive(), formatOptional(verdict.GetReasoning()), formatOptional(verdict.GetReasoningSummary())))

		return res
	})

	if ticket, ok := issue.GetTicketOk(); ok {
		out.Table(ticket, func() []string {
			res := []string{"TICKET"}
			res = append(res, "EXTERNAL ID\tEXTERNAL URL\tCREATED AT")
			res = append(res, fmt.Sprintf("%s\t%s\t%s", ticket.GetExternalId(), ticket.GetExternalUrl(), ticket.GetCreatedAt()))

			return res
		})
	}

	if issue.HasImpact() {
		out.Table(issue, func() []string {
			res := []string{"IMPACT"}
			res = append(res, issue.GetImpact())

			return res
		})
	}

	out.Table(issue, func() []string {
		res := []string{"FIX PROMPT"}
		res = append(res, issue.GetFixPrompt())

		return res
	})
}

var issueGetWithEventsCmd = &cobra.Command{
	Use:     "get-with-events issue-id",
	Aliases: []string{"with-events", "events"},
	Short:   "Get an issue plus every event from its last-seen scan in a single call",
	Long: `Get Issue With Hydrated Events - Single-Call Context for AI Agents

Fetches an issue and concurrently hydrates every event ID listed in
` + "`latestEventIds`" + ` (capped at 5 by the API). Each hydrated event includes the
Exchange (request/response payload) on its attachments when available, so an
agent receives the full investigation context in one tool call.

OUTPUT SHAPE (JSON):
  {
    "issue": <IssueDetailed>,
    "latestEvents": [<EventDetailed>, ...],
    "latestEventsTruncated": <bool>,
    "eventErrors": [{"eventId": "...", "error": "..."}, ...]   // omitted on full success
  }

Per-event hydration failures do NOT cancel siblings and do NOT fail the
command — they surface in 'eventErrors'. The command only fails if the issue
itself cannot be fetched.`,
	Example: `  # Hydrate full context for an AI agent
  escape-cli issues get-with-events <issue-id> -o json

  # Aliases
  escape-cli issues with-events <issue-id>
  escape-cli issues events <issue-id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(IssueWithEvents{}) {
			return nil
		}

		issueID := args[0]
		issue, err := escape.GetIssue(cmd.Context(), issueID)
		if err != nil || issue == nil {
			return fmt.Errorf("unable to get issue %s: %w", issueID, err)
		}

		result := IssueWithEvents{Issue: *issue}

		// `latestEventIds` and `latestEventsTruncated` are optional on the wire:
		// the public API omits them when `lastSeenScanId` is null or the extra
		// GraphQL enrichment fails. Propagate that absence to the output so
		// clients can tell "unavailable" from "empty" / "not truncated".
		eventIDs, idsOK := issue.GetLatestEventIdsOk()
		if truncated, ok := issue.GetLatestEventsTruncatedOk(); ok {
			result.LatestEventsTruncated = truncated
		}

		if !idsOK {
			if err := renderIssueWithEvents(result); err != nil {
				return err
			}

			return nil
		}

		latestEvents := make([]v3.GetEvent200Response, len(eventIDs))
		eventErrors := make([]IssueEventHydrateError, len(eventIDs))
		eventOK := make([]bool, len(eventIDs))

		g, gctx := errgroup.WithContext(cmd.Context())
		g.SetLimit(maxLatestEventsHydrationConcurrency)
		var mu sync.Mutex
		for i, eid := range eventIDs {
			g.Go(func() error {
				ev, err := escape.GetEvent(gctx, eid)
				mu.Lock()
				defer mu.Unlock()
				if err != nil || ev == nil {
					msg := "nil response"
					if err != nil {
						msg = err.Error()
					}

					eventErrors[i] = IssueEventHydrateError{EventID: eid, Error: msg}

					return nil
				}

				latestEvents[i] = *ev
				eventOK[i] = true

				return nil
			})
		}

		_ = g.Wait()

		hydratedEvents := make([]v3.GetEvent200Response, 0, len(latestEvents))
		hydratedErrors := make([]IssueEventHydrateError, 0)
		for i := range eventIDs {
			if eventOK[i] {
				hydratedEvents = append(hydratedEvents, latestEvents[i])
			} else {
				hydratedErrors = append(hydratedErrors, eventErrors[i])
			}
		}

		result.LatestEvents = &hydratedEvents
		if len(hydratedErrors) > 0 {
			result.EventErrors = hydratedErrors
		}

		return renderIssueWithEvents(result)
	},
}

func renderIssueWithEvents(result IssueWithEvents) error {
	out.Table(result, func() []string {
		hydrated := "(n/a)"
		if result.LatestEvents != nil {
			hydrated = strconv.Itoa(len(*result.LatestEvents))
		}

		truncated := "(n/a)"
		if result.LatestEventsTruncated != nil {
			truncated = strconv.FormatBool(*result.LatestEventsTruncated)
		}

		res := []string{"ISSUE ID\tEVENTS HYDRATED\tEVENTS FAILED\tTRUNCATED"}
		res = append(res, fmt.Sprintf("%s\t%s\t%d\t%s",
			result.Issue.GetId(),
			hydrated,
			len(result.EventErrors),
			truncated,
		))

		return res
	})

	return nil
}

// buildUpdateIssueRequest assembles the PUT payload for `issues update` from the command flags.
// A --reason is forwarded with the status change, the severity change, or a severity reset.
// Severity changes use the { value, reason } object form only when a reason is given.
// A reset with a reason uses the same object with value null, so organizations enforcing
// REQUIRE_CHANGE_REASON can clear a manual severity. A reset without a reason stays a bare null.
func buildUpdateIssueRequest() (*v3.UpdateIssueRequest, error) {
	if issueResetSeverity && issueUpdateSeverity != "" {
		return nil, errors.New("--reset-severity and --severity are mutually exclusive")
	}

	if issueUpdateReason != "" && issueUpdateStatusStr == "" && issueUpdateSeverity == "" && !issueResetSeverity {
		return nil, errors.New("--reason requires --status, --severity, or --reset-severity")
	}

	body := v3.NewUpdateIssueRequestWithDefaults()
	if issueUpdateStatusStr != "" {
		newStatus := v3.ENUMPROPERTIESFILTERPROPERTIESSTATUSITEMS(issueUpdateStatusStr)
		if !newStatus.IsValid() {
			return nil, fmt.Errorf("invalid status %q; valid values: %v", issueUpdateStatusStr, v3.AllowedENUMPROPERTIESFILTERPROPERTIESSTATUSITEMSEnumValues)
		}

		statusPayload := v3.NewBulkUpdateIssuesRequestStatusAnyOf(newStatus)
		if issueUpdateReason != "" {
			statusPayload.SetReason(issueUpdateReason)
		}

		body.SetStatus(v3.UpdateIssueRequestStatus{BulkUpdateIssuesRequestStatusAnyOf: statusPayload})
	}

	if issueResetSeverity {
		if issueUpdateReason != "" {
			body.SetSeverity(v3.UpdateIssueRequestSeverity{
				BulkUpdateIssuesRequestSeverityAnyOf: severityResetWithReason(issueUpdateReason),
			})
		} else {
			body.SetSeverityNil()
		}
	} else if issueUpdateSeverity != "" {
		severity := v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY(issueUpdateSeverity)
		if !severity.IsValid() {
			return nil, fmt.Errorf("invalid severity %q; valid values: %v", issueUpdateSeverity, v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues)
		}

		if issueUpdateReason != "" {
			severityPayload := v3.NewBulkUpdateIssuesRequestSeverityAnyOf()
			severityPayload.SetValue(severity)
			severityPayload.SetReason(issueUpdateReason)
			body.SetSeverity(v3.UpdateIssueRequestSeverity{BulkUpdateIssuesRequestSeverityAnyOf: severityPayload})
		} else {
			body.SetSeverity(v3.UpdateIssueRequestSeverity{
				ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY: &severity,
			})
		}
	}

	return body, nil
}

// severityResetWithReason is the object form {"value": null, "reason": reason}.
// The generated Value field is omitempty, so assigning a nil Value drops the key.
// ToMap copies AdditionalProperties afterwards, which is what keeps the explicit null.
func severityResetWithReason(reason string) *v3.BulkUpdateIssuesRequestSeverityAnyOf {
	payload := v3.NewBulkUpdateIssuesRequestSeverityAnyOf()
	payload.SetReason(reason)
	payload.AdditionalProperties = map[string]any{"value": nil}

	return payload
}

var issueUpdateStatusCmd = &cobra.Command{
	Use:     "update issue-id",
	Aliases: []string{"set-status"},
	Short:   "Update issue status or severity to track remediation progress",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	Long: `Update Issue Status - Track Vulnerability Remediation

Change the status of a security issue as you progress through remediation.
Status updates create an audit trail and help teams track security work.

AVAILABLE STATUSES:
  OPEN              - Newly discovered, awaiting review
  MANUAL_REVIEW     - Under investigation by security team
  IN_PROGRESS       - Actively being fixed by developers
  RESOLVED          - Fixed and verified
  FALSE_POSITIVE    - Determined not to be a real issue
  ACCEPTED_RISK     - Acknowledged but not fixing (with justification)
  REOPENED          - Previously resolved but found again

WORKFLOW EXAMPLE:
  1. New issue discovered:        OPEN
  2. Security team reviews:       MANUAL_REVIEW
  3. Assigned to developers:      IN_PROGRESS
  4. Fix deployed and tested:     RESOLVED

TRACKING:
  All status changes are logged in the issue's activity history.
  Use 'escape-cli issues list-activities <issue-id>' to view the full timeline.`,
	Example: `  # Mark issue under review
  escape-cli issues update <issue-id> --status MANUAL_REVIEW

  # Mark as in progress when fixing
  escape-cli issues update <issue-id> --status IN_PROGRESS

  # Mark as resolved after fixing
  escape-cli issues update <issue-id> --status RESOLVED

  # Mark as false positive
  escape-cli issues update <issue-id> --status FALSE_POSITIVE

  # Change severity, with the reason some organizations require
  escape-cli issues update <issue-id> --severity CRITICAL --reason "Weaponized in the wild"

  # Bulk update issues from a list
  cat issue_ids.txt | xargs -I {} escape-cli issues update {} --status IN_PROGRESS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.UpdateIssue200Response{}) {
			return nil
		}

		issueID := args[0]
		if issueUpdateStatusStr == "" && issueUpdateSeverity == "" && !issueResetSeverity {
			_ = cmd.Help()
			return errors.New("at least one of --status, --severity, or --reset-severity is required")
		}

		body, err := buildUpdateIssueRequest()
		if err != nil {
			return err
		}

		result, err := escape.UpdateIssue(cmd.Context(), issueID, *body)
		if err != nil {
			return fmt.Errorf("unable to update issue %s: %w", issueID, err)
		}

		out.Table(result, func() []string {
			res := []string{"UPDATED IDS"}
			res = append(res, result.GetIds()...)

			return res
		})
		out.Log("Updated issue " + issueID)

		return nil
	},
}

var issueListActivitiesCmd = &cobra.Command{
	Use:     "list-activities issue-id",
	Aliases: []string{"ls-activities", "activities", "history", "timeline"},
	Short:   "View complete activity history and timeline for an issue",
	Long: `View Issue Activity Timeline - Audit Trail and History

Display the complete activity history for a security issue, including all status
changes, comments, and modifications. This provides a full audit trail of who
did what and when.

ACTIVITY TYPES:
  • CREATED          - Issue first discovered
  • STATUS_CHANGED   - Status updated (e.g., OPEN → IN_PROGRESS)
  • COMMENT_ADDED    - Team member added a comment
  • ASSIGNED         - Issue assigned to a team member
  • SEVERITY_CHANGED - Severity level adjusted
  • REOPENED         - Previously resolved issue found again

DISPLAYED INFORMATION:
  • ID           - Activity identifier
  • CREATED AT   - When the activity occurred
  • KIND         - Type of activity
  • AUTHOR ID    - Who performed the action
  • AUTHOR EMAIL - User's email address

USE CASES:
  • Review remediation progress
  • Create compliance audit trails
  • Investigate who changed issue status
  • Track time-to-resolution metrics
  • Generate reports for management

Example output:
ID                                      CREATED AT                KIND              AUTHOR ID    AUTHOR EMAIL
00000000-0000-0000-0000-000000000001    2025-06-27T06:02:18.874Z  CREATED           sys-001      system@escape.tech
00000000-0000-0000-0000-000000000002    2025-06-27T08:15:32.120Z  STATUS_CHANGED    usr-123      john@example.com
00000000-0000-0000-0000-000000000003    2025-06-28T14:22:01.543Z  COMMENT_ADDED     usr-456      jane@example.com`,
	Example: `  # View issue activity timeline
  escape-cli issues list-activities 00000000-0000-0000-0000-000000000001

  # Export timeline to JSON
  escape-cli issues list-activities <issue-id> -o json

  # View activities in YAML format
  escape-cli issues list-activities <issue-id> -o yaml`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Output JSON Schema if requested
		if out.Schema([]v3.ActivitySummarized{}) {
			return nil
		}

		issueID := args[0]
		activities, err := escape.ListIssueActivities(cmd.Context(), issueID)
		if err != nil {
			return fmt.Errorf("unable to list activities: %w", err)
		}

		result := []string{"ID\tCREATED AT\tKIND\tAUTHOR ID\tAUTHOR EMAIL"}
		for _, activity := range activities {
			author := activity.GetAuthor()
			authorID := "null"
			authorEmail := "null"
			if author.GetId() != "" || author.GetEmail() != "" {
				authorID = author.GetId()
				authorEmail = author.GetEmail()
			}

			result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", activity.GetId(), activity.GetCreatedAt(), activity.GetKind(), authorID, authorEmail))
		}

		out.Table(activities, func() []string {
			return result
		})

		return nil
	},
}

var issueCommentCmd = &cobra.Command{
	Use:     "comment issue-id",
	Aliases: []string{"add-comment"},
	Short:   "Add a comment to an issue. Pass message or body.",
	Long: `Add a comment to an issue. Pass message or a JSON body {"comment":"..."}.
When both are set, --message wins.`,
	Example: `  escape-cli issues comment <issue-id> --message "Looks like a false positive"

  echo '{"comment":"Looks like a false positive"}' | escape-cli issues comment <issue-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.InputSchema(v3.CreateAssetCommentRequest{}) {
			return nil
		}

		if out.Schema(v3.CreateAssetComment200Response{}) {
			return nil
		}

		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		msg, _ := cmd.Flags().GetString("message")
		var stdin []byte
		// A loop that passes --message must not consume the caller's stdin.
		if strings.TrimSpace(msg) == "" {
			var err error
			stdin, err = readPipedStdin(cmd.InOrStdin())
			if err != nil {
				return err
			}
		}

		text, err := issueCommentText(msg, stdin)
		if err != nil {
			return err
		}

		issueID := args[0]
		comment, err := escape.CommentIssue(cmd.Context(), issueID, text)
		if err != nil {
			return fmt.Errorf("unable to add comment: %w", err)
		}

		out.Print(comment, "Comment added to issue "+issueID)

		return nil
	},
}

// issueCommentText resolves the comment the MCP schema advertises.
// --message wins. Otherwise the stdin body is a CreateAssetCommentRequest,
// whose required field is comment.
func issueCommentText(flagMessage string, stdin []byte) (string, error) {
	if msg := strings.TrimSpace(flagMessage); msg != "" {
		return msg, nil
	}

	if len(bytes.TrimSpace(stdin)) == 0 {
		return "", errors.New("--message is required")
	}

	var body v3.CreateAssetCommentRequest
	if err := json.Unmarshal(stdin, &body); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	if strings.TrimSpace(body.Comment) == "" {
		return "", errors.New("--message is required")
	}

	return body.Comment, nil
}

var (
	funnelProjectIDs []string

	trendAfter          string
	trendBefore         string
	trendInterval       string
	trendApplicationIDs []string
	trendProjectIDs     []string

	bulkIssueStatus       string
	bulkIssueReason       string
	setSeverity           string
	resetSeverity         bool
	bulkIssueIDs          []string
	bulkIssueAssetIDs     []string
	bulkIssueSeverities   []string
	bulkIssueProfileIDs   []string
	bulkIssueTagIDs       []string
	bulkIssueScannerKinds []string
	bulkIssueAll          bool
	bulkIssueDryRun       bool

	notifyScanID string

	triggerWorkflowID string
)

var issueFunnelCmd = &cobra.Command{
	Use:   "funnel",
	Short: "Show issue funnel breakdown",
	Long:  `Display the issue funnel: ALL → OPEN_ISSUES → EXPOSED → UNAUTHENTICATED → HIGH_BUSINESS_IMPACT → CRITICAL with counts per category and step.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]escape.IssueFunnelStep{}) {
			return nil
		}

		steps, err := escape.GetIssueFunnel(cmd.Context(), funnelProjectIDs)
		if err != nil {
			return fmt.Errorf("unable to get issue funnel: %w", err)
		}

		out.Table(steps, func() []string {
			res := []string{"CATEGORY\tSTEP\tCOUNT"}
			for _, s := range steps {
				res = append(res, fmt.Sprintf("%s\t%s\t%.0f", s.Category, s.Step, s.Count))
			}

			return res
		})

		return nil
	},
}

var issueTrendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Show issue severity trends over time",
	Long:  `Display time-bucketed issue counts by severity. Track whether your security posture is improving or degrading.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]escape.IssueTrendPoint{}) {
			return nil
		}

		if trendAfter == "" || trendBefore == "" {
			return errors.New("--after and --before are required")
		}

		points, err := escape.GetIssueTrends(cmd.Context(), trendAfter, trendBefore, trendInterval, trendApplicationIDs, trendProjectIDs)
		if err != nil {
			return fmt.Errorf("unable to get issue trends: %w", err)
		}

		out.Table(points, func() []string {
			res := []string{"DATE\tHIGH\tMEDIUM\tLOW\tINFO"}
			for _, p := range points {
				res = append(res, fmt.Sprintf("%s\t%.0f\t%.0f\t%.0f\t%.0f", p.Date, p.HIGH, p.MEDIUM, p.LOW, p.INFO))
			}

			return res
		})

		return nil
	},
}

// buildBulkUpdateIssuesRequest assembles the bulk update payload from the bulk-update command flags.
// --reason is forwarded with the status change, the severity change, or a severity reset so
// organizations enforcing REQUIRE_CHANGE_REASON can change or clear severities in bulk.
func buildBulkUpdateIssuesRequest() (*v3.BulkUpdateIssuesRequest, error) {
	if resetSeverity && setSeverity != "" {
		return nil, errors.New("--reset-severity and --set-severity are mutually exclusive")
	}

	if bulkIssueReason != "" && bulkIssueStatus == "" && setSeverity == "" && !resetSeverity {
		return nil, errors.New("--reason requires --status, --set-severity, or --reset-severity")
	}

	body := v3.NewBulkUpdateIssuesRequestWithDefaults()
	if bulkIssueStatus != "" {
		status := v3.ENUMPROPERTIESFILTERPROPERTIESSTATUSITEMS(bulkIssueStatus)
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status %q; valid values: %v", bulkIssueStatus, v3.AllowedENUMPROPERTIESFILTERPROPERTIESSTATUSITEMSEnumValues)
		}

		statusPayload := v3.NewBulkUpdateIssuesRequestStatusAnyOf(status)
		if bulkIssueReason != "" {
			statusPayload.SetReason(bulkIssueReason)
		}

		body.SetStatus(v3.BulkUpdateIssuesRequestStatus{BulkUpdateIssuesRequestStatusAnyOf: statusPayload})
	}

	if resetSeverity {
		if bulkIssueReason != "" {
			body.SetSeverity(v3.BulkUpdateIssuesRequestSeverity{
				BulkUpdateIssuesRequestSeverityAnyOf: severityResetWithReason(bulkIssueReason),
			})
		} else {
			body.SetSeverityNil()
		}
	} else if setSeverity != "" {
		severity := v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY(setSeverity)
		if !severity.IsValid() {
			return nil, fmt.Errorf("invalid severity %q; valid values: %v", setSeverity, v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues)
		}

		severityPayload := v3.NewBulkUpdateIssuesRequestSeverityAnyOf()
		severityPayload.SetValue(severity)
		if bulkIssueReason != "" {
			severityPayload.SetReason(bulkIssueReason)
		}

		body.SetSeverity(v3.BulkUpdateIssuesRequestSeverity{BulkUpdateIssuesRequestSeverityAnyOf: severityPayload})
	}

	if !body.HasStatus() && !body.HasSeverity() {
		return nil, errors.New("at least one of --status, --set-severity, or --reset-severity is required")
	}

	where := v3.BulkUpdateIssuesRequestWhere{}
	if len(bulkIssueIDs) > 0 {
		where.Ids = bulkIssueIDs
	}

	if len(bulkIssueAssetIDs) > 0 {
		where.AssetIds = bulkIssueAssetIDs
	}

	if len(bulkIssueSeverities) > 0 {
		severities := make([]v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY, len(bulkIssueSeverities))
		for i, s := range bulkIssueSeverities {
			severities[i] = v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY(s)
		}

		where.Severities = severities
	}

	if len(bulkIssueProfileIDs) > 0 {
		where.ProfileIds = bulkIssueProfileIDs
	}

	if len(bulkIssueTagIDs) > 0 {
		where.TagIds = bulkIssueTagIDs
	}

	if len(bulkIssueScannerKinds) > 0 {
		kinds := make([]v3.ENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMS, len(bulkIssueScannerKinds))
		for i, k := range bulkIssueScannerKinds {
			kinds[i] = v3.ENUMPROPERTIESFILTERPROPERTIESSCANNERKINDSITEMS(k)
		}

		where.ScannerKinds = kinds
	}

	// An empty where is only sent with a filter: with --all the API rejects a
	// non-empty where, and an empty one is noise.
	if bulkIssueFilterFlagsSet() {
		body.SetWhere(where)
	}

	// All and DryRun are pointers with omitempty, so they are only forwarded
	// when set: an explicit false would be accepted by the API but adds noise.
	if bulkIssueAll {
		body.SetAll(true)
	}

	if bulkIssueDryRun {
		body.SetDryRun(true)
	}

	return body, nil
}

// bulkIssueFilterFlagsSet reports whether any where-filter flag carries a
// value. It mirrors the fields requireIssueBulkSelection checks.
func bulkIssueFilterFlagsSet() bool {
	return len(bulkIssueIDs) > 0 ||
		len(bulkIssueAssetIDs) > 0 ||
		len(bulkIssueSeverities) > 0 ||
		len(bulkIssueProfileIDs) > 0 ||
		len(bulkIssueTagIDs) > 0 ||
		len(bulkIssueScannerKinds) > 0
}

// requireIssueBulkSelection rejects a bulk issue update that would match
// every issue in the organization. Since the breaking API change
// (POST /issues/bulk-update), the server requires a non-empty where filter
// or all=true, and rejects all=true combined with a filter.
func requireIssueBulkSelection() error {
	if bulkIssueAll {
		if bulkIssueFilterFlagsSet() {
			return errors.New("--all cannot be combined with a filter; pass --all alone to target every issue")
		}

		return nil
	}

	if bulkIssueFilterFlagsSet() {
		return nil
	}

	return errors.New("at least one of --issue-id, --asset-id, --severity, --profile-id, --tag-id, --scanner-kind, or --all is required")
}

var issueBulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Update status and/or severity of multiple issues matching a filter",
	Long: `Bulk update issues. For example, mark all LOW severity issues on a given asset as IGNORED, or reset severities to scanner values. --reason is forwarded with both --status and --set-severity changes.

A non-empty filter or --all is required: the API rejects an unscoped update
that would match every issue in the organization. --all cannot be combined
with filter flags. --dry-run prints the matching issue IDs without updating
anything.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.BulkUpdateIssues200Response{}) {
			return nil
		}

		if err := requireIssueBulkSelection(); err != nil {
			return err
		}

		body, err := buildBulkUpdateIssuesRequest()
		if err != nil {
			return err
		}

		result, err := escape.BulkUpdateIssues(cmd.Context(), *body)
		if err != nil {
			return fmt.Errorf("unable to bulk update issues: %w", err)
		}

		changed := "Updated IDs"
		summary := fmt.Sprintf("Updated %d issues", len(result.GetIds()))
		if result.GetDryRun() {
			changed = "MATCHING IDS (DRY RUN)"
			summary = fmt.Sprintf("Dry run: %d issues would be updated", len(result.GetIds()))
		}

		out.Table(result, func() []string {
			res := []string{changed}
			res = append(res, result.GetIds()...)

			return res
		})
		out.Log(summary)

		return nil
	},
}

var issueNotifyCmd = &cobra.Command{
	Use:   "notify issue-id",
	Short: "Send notification to asset owners about an issue",
	Long:  `Send an email notification to the owners of the asset associated with this issue.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.NotifyIssueOwners200Response{}) {
			return nil
		}

		issueID := args[0]
		if notifyScanID == "" {
			return errors.New("--scan-id is required")
		}

		result, err := escape.NotifyIssueOwners(cmd.Context(), issueID, notifyScanID)
		if err != nil {
			return fmt.Errorf("unable to notify owners: %w", err)
		}

		pretty := "No owners found to notify for issue " + issueID
		if result.GetNotified() {
			pretty = "Notification sent to asset owners for issue " + issueID
		}

		out.Print(result, pretty)

		return nil
	},
}

var issueTriggerWorkflowCmd = &cobra.Command{
	Use:   "trigger-workflow issue-id",
	Short: "Run a manual workflow on an issue (e.g. create a Jira ticket). Use workflows_list to find manual workflow IDs.",
	Long: `Run a Manual Workflow on an Issue - Automate Issue Actions

Manual workflows are automations configured in Escape with a MANUAL trigger
(for example, exporting the issue to Jira). This command runs one against a
single issue and returns the workflow that was triggered.

Discover the available workflow IDs with:
  $ escape-cli workflows list --trigger MANUAL`,
	Example: `  # Find manual workflow IDs
  escape-cli workflows list --trigger MANUAL

  # Create a Jira ticket for an issue via a manual workflow
  escape-cli issues trigger-workflow 00000000-0000-0000-0000-000000000001 --workflow-id 00000000-0000-0000-0000-000000000002`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("issue ID is required")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.TriggerIssueManualWorkflow200Response{}) {
			return nil
		}

		issueID := args[0]
		if triggerWorkflowID == "" {
			return errors.New("--workflow-id is required")
		}

		result, err := escape.TriggerIssueManualWorkflow(cmd.Context(), issueID, triggerWorkflowID)
		if err != nil {
			return fmt.Errorf("unable to trigger workflow on issue %s: %w", issueID, err)
		}

		workflow := result.GetWorkflow()
		out.Table(result, func() []string {
			return []string{
				"WORKFLOW ID\tWORKFLOW NAME",
				fmt.Sprintf("%s\t%s", workflow.GetId(), workflow.GetName()),
			}
		})
		out.Log(fmt.Sprintf("Triggered workflow %q on issue %s", workflow.GetName(), issueID))

		return nil
	},
}

func init() {
	issuesCmd.AddCommand(issueGetCmd)
	issuesCmd.AddCommand(issueGetWithEventsCmd)
	issuesCmd.AddCommand(issueListActivitiesCmd)
	issuesCmd.AddCommand(issueCommentCmd)
	issueCommentCmd.Flags().String("message", "", "comment text; alternatively pipe {\"comment\":\"...\"}. --message wins when both are set")

	issuesCmd.AddCommand(issueUpdateStatusCmd)
	issueUpdateStatusCmd.Flags().StringVarP(&issueUpdateStatusStr, "status", "s", issueUpdateStatusStr, fmt.Sprintf("new status for the issue: %v", v3.AllowedENUMPROPERTIESFILTERPROPERTIESSTATUSITEMSEnumValues))
	issueUpdateStatusCmd.Flags().StringVar(&issueUpdateReason, "reason", "", "reason for the status and/or severity change (required if your organization enforces it)")
	issueUpdateStatusCmd.Flags().StringVar(&issueUpdateReason, "comment", "", "deprecated: use --reason")
	issueUpdateStatusCmd.Flags().StringVar(&issueUpdateSeverity, "severity", "", fmt.Sprintf("new severity for the issue: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues))
	issueUpdateStatusCmd.Flags().BoolVar(&issueResetSeverity, "reset-severity", false, "reset severity to the scanner value")
	_ = issueUpdateStatusCmd.Flags().MarkDeprecated("comment", "use --reason instead")

	issuesCmd.AddCommand(issueFunnelCmd)
	issueFunnelCmd.Flags().StringSliceVar(&funnelProjectIDs, "project-id", nil, "filter by project ID(s)")

	issuesCmd.AddCommand(issueTrendsCmd)
	issueTrendsCmd.Flags().StringVar(&trendAfter, "after", "", "start date (ISO 8601, required)")
	issueTrendsCmd.Flags().StringVar(&trendBefore, "before", "", "end date (ISO 8601, required)")
	issueTrendsCmd.Flags().StringVar(&trendInterval, "interval", "1 day", "time bucket interval (e.g. '1 day', '1 week')")
	issueTrendsCmd.Flags().StringSliceVar(&trendApplicationIDs, "application-id", nil, "filter by application ID(s)")
	issueTrendsCmd.Flags().StringSliceVar(&trendProjectIDs, "project-id", nil, "filter by project ID(s)")

	issuesCmd.AddCommand(issueBulkUpdateCmd)
	issueBulkUpdateCmd.Flags().StringVar(&bulkIssueStatus, "status", "", "new status to apply")
	issueBulkUpdateCmd.Flags().StringVar(&bulkIssueReason, "reason", "", "reason for the status and/or severity change (required if your organization enforces it)")
	issueBulkUpdateCmd.Flags().StringVar(&setSeverity, "set-severity", "", "new severity to apply")
	issueBulkUpdateCmd.Flags().BoolVar(&resetSeverity, "reset-severity", false, "reset severity to the scanner value")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueIDs, "issue-id", nil, "filter by issue ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueAssetIDs, "asset-id", nil, "filter by asset ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueSeverities, "severity", nil, "filter by severity")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueProfileIDs, "profile-id", nil, "filter by profile ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueTagIDs, "tag-id", nil, "filter by tag ID(s)")
	issueBulkUpdateCmd.Flags().StringSliceVar(&bulkIssueScannerKinds, "scanner-kind", nil, "filter by scanner kind")
	issueBulkUpdateCmd.Flags().BoolVar(&bulkIssueAll, "all", false, "target every issue in the organization; cannot be combined with filter flags")
	issueBulkUpdateCmd.Flags().BoolVar(&bulkIssueDryRun, "dry-run", false, "print the matching issue IDs without updating anything")

	issuesCmd.AddCommand(issueNotifyCmd)
	issueNotifyCmd.Flags().StringVar(&notifyScanID, "scan-id", "", "scan ID to reference in the notification (required)")

	issuesCmd.AddCommand(issueTriggerWorkflowCmd)
	issueTriggerWorkflowCmd.Flags().StringVar(&triggerWorkflowID, "workflow-id", "", "manual workflow ID to trigger (required; use 'workflows list --trigger MANUAL' to find IDs)")
	_ = issueTriggerWorkflowCmd.MarkFlagRequired("workflow-id")

	issuesCmd.AddCommand(issueListCmd)

	issueListCmd.Flags().StringVarP(&search, "search", "s", "", "free-text search across issue names and descriptions")
	issueListCmd.Flags().StringSliceVarP(&issueStatus, "status", "", issueStatus, fmt.Sprintf("filter by status: %v", v3.AllowedENUMPROPERTIESFILTERPROPERTIESSTATUSITEMSEnumValues))
	issueListCmd.Flags().StringSliceVarP(&issueSeverity, "severity", "l", issueSeverity, fmt.Sprintf("filter by severity level: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITYEnumValues))
	issueListCmd.Flags().StringSliceVarP(&profileIDs, "profile-id", "p", profileIDs, "filter by profile ID(s) - comma-separated for multiple")
	issueListCmd.Flags().StringSliceVarP(&assetIDs, "asset-id", "a", assetIDs, "filter by asset ID(s) - comma-separated for multiple")
	issueListCmd.Flags().StringSliceVarP(&domains, "domain", "d", domains, "filter by domain name(s)")
	issueListCmd.Flags().StringSliceVarP(&issueIDs, "issue-id", "i", issueIDs, "filter by specific issue ID(s)")
	issueListCmd.Flags().StringSliceVarP(&scanIDs, "scan-id", "", []string{}, "filter by scan ID(s) that discovered the issues")
	issueListCmd.Flags().StringSliceVarP(&tagsIDs, "tag-id", "t", []string{}, "filter by tag ID(s)")
	issueListCmd.Flags().StringVarP(&jiraTicket, "jira-ticket", "j", "", "filter by associated Jira ticket ID")
	issueListCmd.Flags().StringSliceVarP(&risks, "risk", "r", []string{}, fmt.Sprintf("filter by asset risk level: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESRISKSITEMSEnumValues))
	issueListCmd.Flags().StringSliceVarP(&assetClasses, "asset-class", "", []string{}, fmt.Sprintf("filter by asset classification: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESCLASSEnumValues))
	issueListCmd.Flags().StringSliceVar(&issueScannerKinds, "scanner-kind", []string{}, "filter by scanner kind (e.g., DAST, BLST_REST)")
	issueListCmd.Flags().StringSliceVar(&issueNames, "name", []string{}, "filter by issue name(s)")
	issueListCmd.Flags().StringSliceVar(&issueProjectIDs, "project-id", []string{}, "filter by project ID(s)")
	issueListCmd.Flags().StringSliceVar(&issueTargetIDs, "target-id", []string{}, "filter by target ID(s)")
	issueListCmd.Flags().StringSliceVar(&issueCategories, "category", []string{}, fmt.Sprintf("filter by issue category: %v", v3.AllowedENUMPROPERTIESISSUEPROPERTIESCATEGORIESITEMSPROPERTIESCATEGORYEnumValues))
	issueListCmd.Flags().StringSliceVar(&issueAssetTypes, "asset-type", []string{}, fmt.Sprintf("filter by asset type: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESTYPEEnumValues))
	issueListCmd.Flags().StringSliceVar(&issueAssetStatuses, "asset-status", []string{}, fmt.Sprintf("filter by asset status: %v", v3.AllowedENUMPROPERTIESDATAITEMSPROPERTIESEXTRAASSETSITEMSPROPERTIESSTATUSEnumValues))
	issueListCmd.Flags().StringSliceVar(&issueSecurityTestUids, "security-test-uid", []string{}, "filter by security test UID(s)")
	issueListCmd.Flags().StringSliceVar(&issueBlacklistedIDs, "blacklisted-id", []string{}, "exclude these issue ID(s)")
	issueListCmd.Flags().StringSliceVar(&issueBlacklistedNames, "blacklisted-name", []string{}, "exclude issues by their raw name(s)")
	issueListCmd.Flags().StringVar(&issueAiFalsePositive, "ai-false-positive", "", "filter by AI false positive classification (true/false)")
	issueListCmd.Flags().StringVar(&issueAgentic, "agentic", "", "filter by agentic (AI pentest) issues (true/false)")
	issueListCmd.Flags().BoolVar(&issueNoTags, "no-tags", false, "filter by issues whose assets have no tags (only true is supported)")
	issueListCmd.Flags().StringVar(&issueDnf, "dnf", "", "advanced filter as a DNF expression (URL-encoded JSON object)")
	issueListCmd.Flags().StringVar(&issueSortType, "sort-by", "", "sort field: LAST_SEEN, FIRST_SEEN, SEVERITY, STATUS")
	issueListCmd.Flags().StringVar(&issueSortDirection, "sort-direction", "", "sort direction: asc, desc")
	issueListPage.bind(issueListCmd)

	rootCmd.AddCommand(issuesCmd)
}
