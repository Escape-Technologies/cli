package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/spf13/pflag"
)

func TestCommandSchemaRegistryIncludesRetestTools(t *testing.T) {
	t.Parallel()

	registry := CommandSchemaRegistry()

	list, ok := registry["escape-cli retests list"]
	if !ok {
		t.Fatal("expected retests list in CommandSchemaRegistry")
	}

	if _, ok := list.Output.(Page[v3.RetestDetailed]); !ok {
		t.Fatalf("retests list output type = %T", list.Output)
	}

	if list.Input != nil {
		t.Fatalf("retests list should not take a body, got %#v", list.Input)
	}

	get, ok := registry["escape-cli retests get"]
	if !ok {
		t.Fatal("expected retests get in CommandSchemaRegistry")
	}

	if _, ok := get.Output.(v3.GetRetest200Response); !ok {
		t.Fatalf("retests get output type = %T", get.Output)
	}

	if get.Input != nil {
		t.Fatalf("retests get should not take a body, got %#v", get.Input)
	}

	start, ok := registry["escape-cli retests start"]
	if !ok {
		t.Fatal("expected retests start in CommandSchemaRegistry")
	}

	if _, ok := start.Input.(v3.StartRetestRequest); !ok {
		t.Fatalf("retests start input type = %T", start.Input)
	}

	if _, ok := start.Output.(v3.StartRetest200Response); !ok {
		t.Fatalf("retests start output type = %T", start.Output)
	}
}

func TestBuildMCPToolSpecsIncludesRetestTools(t *testing.T) {
	t.Parallel()

	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	byName := make(map[string]int, len(specs))
	for i, spec := range specs {
		byName[spec.Name] = i
	}

	if _, ok := byName["retests"]; ok {
		t.Fatal("parent retests command should not be an MCP tool")
	}

	list := specs[mustSpec(t, byName, "retests_list")]
	if list.Path != "escape-cli retests list" || list.BodyProperty != "" {
		t.Fatalf("retests_list spec = path %q body %q", list.Path, list.BodyProperty)
	}

	assertFlag(t, list.FlagBindings, "profile_id", "profile-id", "stringSlice")
	assertFlag(t, list.FlagBindings, "sort_by", "sort-by", "string")
	assertFlag(t, list.FlagBindings, "size", "size", "int")
	assertFlag(t, list.FlagBindings, "cursor", "cursor", "string")
	if retestsListCmd.Flags().Lookup("limit") != nil {
		t.Fatal("retests list still exposes --limit; page size is --size")
	}

	if list.DefaultArgs["size"] != climcp.DefaultListPageSize {
		t.Fatalf("retests_list default size = %#v", list.DefaultArgs["size"])
	}

	if len(list.Tool.RawOutputSchema) == 0 {
		t.Fatal("retests_list should advertise the wrapped {items: [...]} output schema")
	}

	var listOutput map[string]any
	if err := json.Unmarshal(list.Tool.RawOutputSchema, &listOutput); err != nil {
		t.Fatalf("retests_list output schema: %v", err)
	}

	if listOutput["type"] != "object" {
		t.Fatalf("retests_list output schema type = %#v", listOutput["type"])
	}

	listProps, _ := listOutput["properties"].(map[string]any)
	listItems, _ := listProps["items"].(map[string]any)
	if listItems["type"] != "array" {
		t.Fatalf("retests_list properties.items = %#v", listProps["items"])
	}

	get := specs[mustSpec(t, byName, "retests_get")]
	if get.Path != "escape-cli retests get" {
		t.Fatalf("retests_get path = %q", get.Path)
	}

	if strings.Join(get.PositionalArgs, ",") != "retest_id" {
		t.Fatalf("retests_get positionals = %v", get.PositionalArgs)
	}

	assertFlag(t, get.FlagBindings, "cursor", "cursor", "string")
	assertFlag(t, get.FlagBindings, "size", "size", "int")
	if get.DefaultArgs != nil {
		t.Fatalf("retests_get should not inject a page size, got %#v", get.DefaultArgs)
	}

	if strings.Contains(get.Description, "{items") {
		t.Fatalf("retests_get description treats the command as a paged list: %q", get.Description)
	}

	if len(get.Tool.RawOutputSchema) == 0 {
		t.Fatal("retests_get should advertise an object output schema")
	}

	start := specs[mustSpec(t, byName, "retests_start")]
	if start.BodyProperty != "body" || start.Path != "escape-cli retests start" {
		t.Fatalf("retests_start spec = path %q body %q", start.Path, start.BodyProperty)
	}

	assertFlag(t, start.FlagBindings, "profile_id", "profile-id", "string")
	assertFlag(t, start.FlagBindings, "issue_id", "issue-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "filter_id", "filter-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "asset_id", "asset-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "severity", "severity", "stringSlice")
	assertFlag(t, start.FlagBindings, "tag_id", "tag-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "scanner_kind", "scanner-kind", "stringSlice")
	assertFlag(t, start.FlagBindings, "status", "status", "stringSlice")
	assertFlag(t, start.FlagBindings, "search", "search", "string")
	assertFlag(t, start.FlagBindings, "context", "context", "string")
	assertFlag(t, start.FlagBindings, "initiator", "initiator", "string")
	assertFlag(t, start.FlagBindings, "project_id", "project-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "target_id", "target-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "scan_id", "scan-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "security_test_uid", "security-test-uid", "stringSlice")
	assertFlag(t, start.FlagBindings, "domain", "domain", "stringSlice")
	assertFlag(t, start.FlagBindings, "name", "name", "stringSlice")
	assertFlag(t, start.FlagBindings, "blacklisted_id", "blacklisted-id", "stringSlice")
	assertFlag(t, start.FlagBindings, "blacklisted_name", "blacklisted-name", "stringSlice")
	assertFlag(t, start.FlagBindings, "risk", "risk", "stringSlice")
	assertFlag(t, start.FlagBindings, "asset_class", "asset-class", "stringSlice")
	assertFlag(t, start.FlagBindings, "category", "category", "stringSlice")
	assertFlag(t, start.FlagBindings, "asset_type", "asset-type", "stringSlice")
	assertFlag(t, start.FlagBindings, "asset_status", "asset-status", "stringSlice")
	assertFlag(t, start.FlagBindings, "jira_ticket", "jira-ticket", "bool")
	assertFlag(t, start.FlagBindings, "no_tags", "no-tags", "bool")
	assertFlag(t, start.FlagBindings, "ai_false_positive", "ai-false-positive", "bool")
	assertFlag(t, start.FlagBindings, "agentic", "agentic", "bool")
	assertFlag(t, start.FlagBindings, "dnf", "dnf", "string")

	bodyProps := mcpBodyProperties(t, start.Tool.RawInputSchema)
	for _, key := range []string{"profileId", "issueIds", "filter", "context", "initiator"} {
		if _, ok := bodyProps[key]; !ok {
			t.Fatalf("start body schema missing %q", key)
		}
	}

	// filter is a pointer, so its schema is a null union rather than a
	// nullable object. The fields stay on the object branch.
	filter, _ := bodyProps["filter"].(map[string]any)
	filterProps, _ := nullableObjectSchema(filter)["properties"].(map[string]any)
	for _, key := range []string{
		"ids", "assetIds", "severities", "tagIds", "scannerKinds", "status", "search",
		"projectIds", "targetIds", "scanIds", "securityTestUids", "domains", "names",
		"blacklistedIds", "blacklistedNames", "risks", "assetClasses", "categories",
		"assetTypes", "assetStatuses", "jiraTicket", "noTags", "aiFalsePositive",
		"agentic", "dnf",
	} {
		if _, ok := filterProps[key]; !ok {
			t.Fatalf("start filter schema missing %q", key)
		}
	}
}

func TestBuildStartRetestRequestFromFlags(t *testing.T) {
	resetRetestStartFlags(t)

	if err := retestsStartCmd.ParseFlags([]string{
		"--profile-id", "profile-1",
		"--issue-id", "issue-1",
		"--issue-id", "issue-2",
		"--context", "verify the fix",
		"--initiator", "MANUAL",
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	request, err := buildStartRetestRequest(startRetestInputFromCommand(retestsStartCmd), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.ProfileId != "profile-1" {
		t.Fatalf("profile = %q", request.ProfileId)
	}

	if strings.Join(request.IssueIds, ",") != "issue-1,issue-2" {
		t.Fatalf("issue ids = %v", request.IssueIds)
	}

	if request.Filter != nil {
		t.Fatalf("expected no filter, got %#v", request.Filter)
	}

	if request.GetContext() != "verify the fix" {
		t.Fatalf("context = %q", request.GetContext())
	}

	if request.GetInitiator() != v3.ENUMPROPERTIESINITIATOR_MANUAL {
		t.Fatalf("initiator = %q", request.GetInitiator())
	}
}

func TestBuildStartRetestRequestFromFilterFlags(t *testing.T) {
	resetRetestStartFlags(t)

	if err := retestsStartCmd.ParseFlags([]string{
		"--profile-id", "profile-1",
		"--severity", "HIGH",
		"--severity", "CRITICAL",
		"--status", "OPEN",
		"--asset-id", "asset-1",
		"--search", "checkout",
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	request, err := buildStartRetestRequest(startRetestInputFromCommand(retestsStartCmd), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if len(request.IssueIds) != 0 {
		t.Fatalf("expected no issue ids, got %v", request.IssueIds)
	}

	if request.Filter == nil {
		t.Fatal("expected filter")
	}

	if strings.Join(enumStrings(request.Filter.Severities), ",") != "HIGH,CRITICAL" {
		t.Fatalf("severities = %v", request.Filter.Severities)
	}

	if strings.Join(enumStrings(request.Filter.Status), ",") != "OPEN" {
		t.Fatalf("status = %v", request.Filter.Status)
	}

	if strings.Join(request.Filter.AssetIds, ",") != "asset-1" {
		t.Fatalf("asset ids = %v", request.Filter.AssetIds)
	}

	if request.Filter.GetSearch() != "checkout" {
		t.Fatalf("search = %q", request.Filter.GetSearch())
	}

	if request.HasInitiator() {
		t.Fatal("initiator should stay unset so the API can default it to MANUAL")
	}
}

func TestBuildStartRetestRequestFromExtendedFilterFlags(t *testing.T) {
	resetRetestStartFlags(t)

	if err := retestsStartCmd.ParseFlags([]string{
		"--profile-id", "profile-1",
		"--project-id", "project-1",
		"--target-id", "target-1",
		"--scan-id", "scan-1",
		"--security-test-uid", "ISSUE_SQL_INJECTION",
		"--domain", "example.com",
		"--name", "SQL injection found",
		"--blacklisted-id", "issue-2",
		"--blacklisted-name", "SQL injection",
		"--risk", "EXPOSED",
		"--asset-class", "FRONTEND",
		"--category", "INJECTION",
		"--asset-type", "WEBAPP",
		"--asset-status", "MONITORED",
		"--jira-ticket",
		"--no-tags",
		"--ai-false-positive",
		"--agentic",
		"--dnf", `{"and":[{"severity":"HIGH"}]}`,
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	request, err := buildStartRetestRequest(startRetestInputFromCommand(retestsStartCmd), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.Filter == nil {
		t.Fatal("expected filter")
	}

	filter := request.Filter
	if strings.Join(filter.ProjectIds, ",") != "project-1" {
		t.Errorf("project ids = %v", filter.ProjectIds)
	}

	if strings.Join(filter.TargetIds, ",") != "target-1" {
		t.Errorf("target ids = %v", filter.TargetIds)
	}

	if strings.Join(filter.ScanIds, ",") != "scan-1" {
		t.Errorf("scan ids = %v", filter.ScanIds)
	}

	if strings.Join(filter.SecurityTestUids, ",") != "ISSUE_SQL_INJECTION" {
		t.Errorf("security test uids = %v", filter.SecurityTestUids)
	}

	if strings.Join(filter.Domains, ",") != "example.com" {
		t.Errorf("domains = %v", filter.Domains)
	}

	if strings.Join(filter.Names, ",") != "SQL injection found" {
		t.Errorf("names = %v", filter.Names)
	}

	if strings.Join(filter.BlacklistedIds, ",") != "issue-2" {
		t.Errorf("blacklisted ids = %v", filter.BlacklistedIds)
	}

	if strings.Join(filter.BlacklistedNames, ",") != "SQL injection" {
		t.Errorf("blacklisted names = %v", filter.BlacklistedNames)
	}

	if strings.Join(enumStrings(filter.Risks), ",") != "EXPOSED" {
		t.Errorf("risks = %v", filter.Risks)
	}

	if strings.Join(enumStrings(filter.AssetClasses), ",") != "FRONTEND" {
		t.Errorf("asset classes = %v", filter.AssetClasses)
	}

	if strings.Join(enumStrings(filter.Categories), ",") != "INJECTION" {
		t.Errorf("categories = %v", filter.Categories)
	}

	if strings.Join(enumStrings(filter.AssetTypes), ",") != "WEBAPP" {
		t.Errorf("asset types = %v", filter.AssetTypes)
	}

	if strings.Join(enumStrings(filter.AssetStatuses), ",") != "MONITORED" {
		t.Errorf("asset statuses = %v", filter.AssetStatuses)
	}

	if !filter.GetJiraTicket() {
		t.Errorf("jira ticket = %v, want true", filter.JiraTicket)
	}

	if filter.NoTags == nil || !bool(*filter.NoTags) {
		t.Errorf("no tags = %v, want true", filter.NoTags)
	}

	if !filter.GetAiFalsePositive() {
		t.Errorf("ai false positive = %v, want true", filter.AiFalsePositive)
	}

	if !filter.GetAgentic() {
		t.Errorf("agentic = %v, want true", filter.Agentic)
	}

	if filter.Dnf["and"] == nil {
		t.Errorf("dnf = %v, want the parsed object", filter.Dnf)
	}
}

func TestBuildStartRetestRequestRejectsNoTagsFalse(t *testing.T) {
	resetRetestStartFlags(t)

	if err := retestsStartCmd.ParseFlags([]string{
		"--profile-id", "profile-1",
		"--no-tags=false",
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	_, err := buildStartRetestRequest(startRetestInputFromCommand(retestsStartCmd), nil)
	if err == nil || !strings.Contains(err.Error(), "--no-tags only supports true") {
		t.Fatalf("expected a --no-tags true-only error, got %v", err)
	}
}

func TestBuildStartRetestRequestRejectsInvalidDnf(t *testing.T) {
	resetRetestStartFlags(t)

	if err := retestsStartCmd.ParseFlags([]string{
		"--profile-id", "profile-1",
		"--dnf", "not-json",
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	_, err := buildStartRetestRequest(startRetestInputFromCommand(retestsStartCmd), nil)
	if err == nil || !strings.Contains(err.Error(), "invalid --dnf JSON") {
		t.Fatalf("expected an invalid --dnf error, got %v", err)
	}
}

func TestBuildStartRetestRequestRejectsBothSelectors(t *testing.T) {
	resetRetestStartFlags(t)

	if err := retestsStartCmd.ParseFlags([]string{
		"--profile-id", "profile-1",
		"--issue-id", "issue-1",
		"--severity", "HIGH",
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	_, err := buildStartRetestRequest(startRetestInputFromCommand(retestsStartCmd), nil)
	if err == nil || err.Error() != "provide either issueIds or filter, not both" {
		t.Fatalf("expected mutual exclusion error, got %v", err)
	}
}

func TestBuildStartRetestRequestMergesStdinAndFlags(t *testing.T) {
	resetRetestStartFlags(t)

	if err := retestsStartCmd.ParseFlags([]string{"--profile-id", "profile-flag"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	body := []byte(`{"issueIds":["issue-1"],"context":"from body"}`)
	request, err := buildStartRetestRequest(startRetestInputFromCommand(retestsStartCmd), body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.ProfileId != "profile-flag" {
		t.Fatalf("profile = %q", request.ProfileId)
	}

	if strings.Join(request.IssueIds, ",") != "issue-1" {
		t.Fatalf("issue ids = %v", request.IssueIds)
	}

	if request.GetContext() != "from body" {
		t.Fatalf("context = %q", request.GetContext())
	}
}

func TestBuildStartRetestRequestFlagReplacesBodyFilter(t *testing.T) {
	body := []byte(`{"profileId":"profile-1","filter":{"severities":["LOW"]}}`)
	input := startRetestInput{issueIDsSet: true, issueIDs: []string{"issue-9"}}

	request, err := buildStartRetestRequest(input, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.Filter != nil {
		t.Fatalf("expected flag to clear filter, got %#v", request.Filter)
	}

	if strings.Join(request.IssueIds, ",") != "issue-9" {
		t.Fatalf("issue ids = %v", request.IssueIds)
	}
}

func TestBuildStartRetestRequestExtendsBodyFilter(t *testing.T) {
	body := []byte(`{"profileId":"profile-1","filter":{"severities":["HIGH"]}}`)
	input := startRetestInput{statusesSet: true, statuses: []string{"OPEN"}}

	request, err := buildStartRetestRequest(input, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.Filter == nil {
		t.Fatal("expected filter")
	}

	if strings.Join(enumStrings(request.Filter.Severities), ",") != "HIGH" {
		t.Fatalf("severities = %v", request.Filter.Severities)
	}

	if strings.Join(enumStrings(request.Filter.Status), ",") != "OPEN" {
		t.Fatalf("status = %v", request.Filter.Status)
	}
}

func TestBuildStartRetestRequestAcceptsEmptyFilterBody(t *testing.T) {
	body := []byte(`{"profileId":"profile-1","filter":{}}`)
	request, err := buildStartRetestRequest(startRetestInput{}, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if request.Filter == nil || len(request.IssueIds) != 0 {
		t.Fatalf("request = %#v", request)
	}
}

func TestBuildStartRetestRequestRejectsBlankFilterFlagOverBodySelection(t *testing.T) {
	t.Parallel()

	// A blank --project-id used to drop out of the filter and clear the issue
	// selection from the body, which the API reads as "retest every issue".
	body := []byte(`{"profileId":"profile-1","issueIds":["issue-1"]}`)
	input := startRetestInput{projectIDsSet: true, projectIDs: []string{" "}}

	_, err := buildStartRetestRequest(input, body)
	if err == nil || !strings.Contains(err.Error(), "--project-id requires a non-empty value") {
		t.Fatalf("expected a blank --project-id error, got %v", err)
	}
}

func TestBuildStartRetestRequestValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input startRetestInput
		body  string
		want  string
	}{
		{name: "missing profile", want: "--profile-id is required"},
		{
			name:  "missing selector",
			input: startRetestInput{profileID: "profile-1", profileSet: true},
			want:  "provide exactly one of issueIds or filter",
		},
		{
			name: "both in body",
			body: `{"profileId":"profile-1","issueIds":["issue-1"],"filter":{"search":"checkout"}}`,
			want: "provide exactly one of issueIds or filter",
		},
		{
			name:  "bad severity",
			input: startRetestInput{profileID: "profile-1", profileSet: true, severitiesSet: true, severities: []string{"NOPE"}},
			want:  `invalid severity "NOPE"`,
		},
		{
			name:  "blank project id flag",
			input: startRetestInput{profileID: "profile-1", profileSet: true, projectIDsSet: true, projectIDs: []string{" "}},
			want:  "--project-id requires a non-empty value",
		},
		{
			name:  "blank severity flag",
			input: startRetestInput{profileID: "profile-1", profileSet: true, severitiesSet: true, severities: []string{""}},
			want:  "--severity requires a non-empty value",
		},
		{
			name:  "blank search flag",
			input: startRetestInput{profileID: "profile-1", profileSet: true, searchSet: true, search: "  "},
			want:  "--search requires a non-empty value",
		},
		{
			name:  "bad initiator",
			input: startRetestInput{profileID: "profile-1", profileSet: true, issueIDsSet: true, issueIDs: []string{"issue-1"}, initiatorSet: true, initiator: "API"},
			want:  `invalid initiator "API"`,
		},
		{
			name:  "context too long",
			input: startRetestInput{profileID: "profile-1", profileSet: true, issueIDsSet: true, issueIDs: []string{"issue-1"}, contextSet: true, context: strings.Repeat("a", retestContextMaxLen+1)},
			want:  "context must be at most",
		},
		{name: "invalid json", input: startRetestInput{profileSet: true, profileID: "profile-1"}, body: `{`, want: "invalid JSON"},
		{
			name: "bad body severity",
			body: `{"profileId":"profile-1","filter":{"severities":["NOPE"]}}`,
			want: "invalid JSON",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := buildStartRetestRequest(tc.input, []byte(tc.body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestRetestsListRejectsSortBeforeAPI(t *testing.T) {
	prevType := retestListSortType
	prevDirection := retestListSortDirection
	t.Cleanup(func() {
		retestListSortType = prevType
		retestListSortDirection = prevDirection
	})

	retestListSortType = "updatedAt"
	retestListSortDirection = ""
	err := retestsListCmd.RunE(retestsListCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--sort-by") {
		t.Fatalf("expected sort error, got %v", err)
	}

	retestListSortType = ""
	retestListSortDirection = "sideways"
	err = retestsListCmd.RunE(retestsListCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--sort-direction") {
		t.Fatalf("expected direction error, got %v", err)
	}
}

func TestRetestsGetRejectsPageSizeBeforeAPI(t *testing.T) {
	prev := retestGetSize
	t.Cleanup(func() { retestGetSize = prev })

	retestGetSize = retestReportPageSizeMax + 1
	err := retestsGetCmd.RunE(retestsGetCmd, []string{"retest-1"})
	if err == nil || !strings.Contains(err.Error(), "--size") {
		t.Fatalf("expected size error, got %v", err)
	}
}

func TestRetestListEmptyProfileIsErrorOnlyInPretty(t *testing.T) {
	t.Parallel()

	if !retestListEmptyIsError([]string{"profile-1"}, 0, true) {
		t.Fatal("pretty mode should still report that the profile has no retests")
	}

	if retestListEmptyIsError(nil, 0, true) || retestListEmptyIsError([]string{"profile-1"}, 2, true) {
		t.Fatal("empty error applies only to a profile filter with no rows")
	}

	if retestListEmptyIsError([]string{"profile-1"}, 0, false) {
		t.Fatal("json and yaml should print an empty list instead of failing")
	}
}

func TestRetestInitiatorHelpUsesEnum(t *testing.T) {
	labels := make([]string, len(v3.AllowedENUMPROPERTIESINITIATOREnumValues))
	for i, value := range v3.AllowedENUMPROPERTIESINITIATOREnumValues {
		labels[i] = string(value)
	}

	usage := retestsStartCmd.Flags().Lookup("initiator").Usage
	for _, label := range labels {
		if !strings.Contains(usage, label) {
			t.Fatalf("initiator help %q does not list enum value %s", usage, label)
		}
	}
}

func TestCollectRetestReportsFollowsPages(t *testing.T) {
	t.Parallel()

	calls := 0
	get := func(_ context.Context, id, cursor string, size int) (*v3.GetRetest200Response, error) {
		calls++
		if id != "retest-1" || size != retestReportPageSizeMax {
			t.Fatalf("id=%q size=%d", id, size)
		}

		switch cursor {
		case "":
			next := "page-2"

			return &v3.GetRetest200Response{
				Id:         "retest-1",
				ProfileId:  "profile-1",
				Status:     "RUNNING",
				Agents:     []v3.RetestAgent{{IssueId: "issue-1"}},
				NextCursor: &next,
			}, nil
		case "page-2":
			return &v3.GetRetest200Response{
				Id:     "retest-1",
				Agents: []v3.RetestAgent{{IssueId: "issue-2"}},
			}, nil
		default:
			t.Fatalf("unexpected cursor %q", cursor)
			return nil, nil
		}
	}

	got, err := collectRetestReports(context.Background(), "retest-1", "", retestReportPageSizeMax, true, get)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}

	if got.GetProfileId() != "profile-1" || got.HasNextCursor() {
		t.Fatalf("merged retest = id %q profile %q next %v", got.GetId(), got.GetProfileId(), got.NextCursor)
	}

	agents := got.GetAgents()
	if len(agents) != 2 || agents[0].IssueId != "issue-1" || agents[1].IssueId != "issue-2" {
		t.Fatalf("agents = %#v", agents)
	}
}

func TestCollectRetestReportsSinglePage(t *testing.T) {
	t.Parallel()

	calls := 0
	next := "page-3"
	get := func(_ context.Context, _, cursor string, _ int) (*v3.GetRetest200Response, error) {
		calls++
		if cursor != "page-2" {
			t.Fatalf("cursor = %q", cursor)
		}

		return &v3.GetRetest200Response{
			Id:         "retest-1",
			Agents:     []v3.RetestAgent{{IssueId: "issue-2"}},
			NextCursor: &next,
		}, nil
	}

	got, err := collectRetestReports(context.Background(), "retest-1", "page-2", 20, false, get)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	if calls != 1 || got.GetNextCursor() != "page-3" || len(got.GetAgents()) != 1 {
		t.Fatalf("calls=%d retest=%#v", calls, got)
	}
}

func mustSpec(t *testing.T, byName map[string]int, name string) int {
	t.Helper()
	index, ok := byName[name]
	if !ok {
		t.Fatalf("expected MCP tool %q", name)
	}

	return index
}

func assertFlag(t *testing.T, bindings []climcp.FlagBinding, property, flagName, kind string) {
	t.Helper()
	for _, binding := range bindings {
		if binding.Property != property {
			continue
		}

		if binding.FlagName != flagName || binding.Kind != kind {
			t.Fatalf("flag %s = name %q kind %q", property, binding.FlagName, binding.Kind)
		}

		return
	}

	t.Fatalf("missing flag property %q", property)
}

// A body supplied through cmd.SetIn must reach POST /retests. Reading
// os.Stdin directly would ignore it and fail with "--profile-id is required".
func TestRetestsStartReadsBodyFromCommandInput(t *testing.T) {
	resetRetestStartFlags(t)

	var sent map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &sent); err != nil {
			t.Errorf("request body %q: %v", raw, err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"message":"stop after capture"}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	retestsStartCmd.SetIn(bytes.NewBufferString(`{"profileId":"profile-body","issueIds":["issue-1"]}`))
	t.Cleanup(func() { retestsStartCmd.SetIn(nil) })
	retestsStartCmd.SetContext(context.Background())

	err := retestsStartCmd.RunE(retestsStartCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "unable to start retest") {
		t.Fatalf("expected the API error after the request was sent, got %v", err)
	}

	if sent["profileId"] != "profile-body" {
		t.Fatalf("request body = %#v, want the SetIn body", sent)
	}
}

func resetRetestStartFlags(t *testing.T) {
	t.Helper()
	resetFlags := func() {
		retestsStartCmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if slice, ok := flag.Value.(pflag.SliceValue); ok {
				_ = slice.Replace(nil)
			} else if err := flag.Value.Set(flag.DefValue); err != nil {
				t.Errorf("reset flag %s: %v", flag.Name, err)
			}

			flag.Changed = false
		})
	}
	resetFlags()
	t.Cleanup(resetFlags)
}

func enumStrings[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value)
	}

	return out
}

// nullableObjectSchema returns the object branch of a null union. Pointer
// fields are advertised as anyOf [object, null].
func nullableObjectSchema(schema map[string]any) map[string]any {
	branches, _ := schema["anyOf"].([]any)
	for _, branch := range branches {
		candidate, _ := branch.(map[string]any)
		if candidate["type"] == "object" || candidate["properties"] != nil {
			return candidate
		}
	}

	return schema
}

func mcpBodyProperties(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal input schema: %v", err)
	}

	properties, _ := schema["properties"].(map[string]any)
	body, _ := properties["body"].(map[string]any)
	bodyProperties, _ := body["properties"].(map[string]any)
	if bodyProperties == nil {
		t.Fatalf("start body properties missing: %#v", schema)
	}

	return bodyProperties
}
