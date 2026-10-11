package escape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	prd2312EmptyPage  = `{"data":[],"nextCursor":null,"totalCount":0}`
	prd2312EmptyArray = `[]`
	prd2312APIKey     = "00000000-0000-0000-0000-000000000000"
)

func capturePRD2312(t *testing.T, body string) (*[]string, *[]url.Values) {
	t.Helper()

	paths := []string{}
	queries := []url.Values{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	return &paths, &queries
}

func lastQuery(t *testing.T, queries *[]url.Values) url.Values {
	t.Helper()

	if queries == nil || len(*queries) == 0 {
		t.Fatal("no request was sent")
	}

	return (*queries)[len(*queries)-1]
}

func assertQueryValues(t *testing.T, query url.Values, key string, want []string) {
	t.Helper()

	got := query[key]
	if len(got) != len(want) {
		t.Fatalf("%s = %#v, want %#v", key, got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %#v, want %#v", key, got, want)
		}

		if strings.Contains(got[i], "Parameter value") {
			t.Fatalf("%s serialized as a generated union (%q)", key, got[i])
		}
	}
}

func TestListJiraProjectsSendsSearch(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyArray)

	projects, err := ListJiraProjects(context.Background(), "jira-1", "core")
	if err != nil {
		t.Fatalf("list jira projects: %v", err)
	}

	if projects == nil {
		t.Fatal("nil slice, want an empty list")
	}

	if (*paths)[0] != "/v3/integrations/jira/jira-1/projects" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	assertQueryValues(t, lastQuery(t, queries), "search", []string{"core"})
}

func TestListJiraProjectsDrainsCursorHeader(t *testing.T) {
	var queries []url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "" {
			w.Header().Set("X-Next-Cursor", "page-2")
			_, _ = w.Write([]byte(`[{"id":"10001","name":"Core","projectTypeKey":"software"}]`))

			return
		}

		_, _ = w.Write([]byte(`[{"id":"10002","name":"Platform","projectTypeKey":"software"}]`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	projects, err := ListJiraProjects(context.Background(), "jira-1", "core")
	if err != nil {
		t.Fatalf("list jira projects: %v", err)
	}

	if len(queries) != 2 {
		t.Fatalf("requests = %d, want the cursor page and the page after it", len(queries))
	}

	if queries[0].Get("cursor") != "" {
		t.Fatalf("first cursor = %q, want it omitted", queries[0].Get("cursor"))
	}

	assertQueryValues(t, queries[1], "cursor", []string{"page-2"})
	for _, query := range queries {
		assertQueryValues(t, query, "search", []string{"core"})
		if query.Get("size") != "" {
			t.Fatalf("size = %q, want it omitted so the API walks its page budget", query.Get("size"))
		}
	}

	if len(projects) != 2 || projects[0].ID != "10001" || projects[1].ID != "10002" {
		t.Fatalf("projects = %#v, want both pages", projects)
	}
}

func TestListJiraProjectsStopsWhenCursorHeaderIsAbsent(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"10001","name":"Core","projectTypeKey":"software"}]`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	projects, err := ListJiraProjects(context.Background(), "jira-1", "")
	if err != nil {
		t.Fatalf("list jira projects: %v", err)
	}

	if requests != 1 {
		t.Fatalf("requests = %d, want one request when X-Next-Cursor is absent", requests)
	}

	if len(projects) != 1 || projects[0].ID != "10001" {
		t.Fatalf("projects = %#v", projects)
	}
}

func TestListJiraProjectsStopsOnRepeatedCursor(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Next-Cursor", "page-2")
		_, _ = w.Write([]byte(`[{"id":"10001","name":"Core","projectTypeKey":"software"}]`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	_, err := ListJiraProjects(context.Background(), "jira-1", "core")
	if err == nil || !strings.Contains(err.Error(), "did not advance") {
		t.Fatalf("err = %v, want a pagination error", err)
	}

	if requests != 2 {
		t.Fatalf("requests = %d, want the repeated cursor to stop the loop", requests)
	}
}

func TestListJiraIssueTypesSendsProjectID(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyArray)

	if _, err := ListJiraIssueTypes(context.Background(), "jira-1", "CT"); err != nil {
		t.Fatalf("list jira issue types: %v", err)
	}

	if (*paths)[0] != "/v3/integrations/jira/jira-1/issue-types" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	assertQueryValues(t, lastQuery(t, queries), "projectId", []string{"CT"})
}

func TestListAzureBoardsAreasSendsProjectAndWorkItemType(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyArray)

	areas, err := ListAzureBoardsAreas(context.Background(), "azure-1", "My Project", "Bug")
	if err != nil {
		t.Fatalf("list azure areas: %v", err)
	}

	if areas == nil {
		t.Fatal("nil slice, want an empty list")
	}

	if (*paths)[0] != "/v3/integrations/azure-boards/azure-1/areas" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "project", []string{"My Project"})
	assertQueryValues(t, query, "workItemType", []string{"Bug"})
}

func TestListScanScenariosSendsFilters(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyPage)

	if _, _, _, err := ListScanScenarios(context.Background(), "scan-1", "cursor-1", "INJECTION", "login", 20); err != nil {
		t.Fatalf("list scenarios: %v", err)
	}

	if (*paths)[0] != "/v3/scans/scan-1/scenarios" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "cursor", []string{"cursor-1"})
	assertQueryValues(t, query, "size", []string{"20"})
	assertQueryValues(t, query, "vector", []string{"INJECTION"})
	assertQueryValues(t, query, "search", []string{"login"})
}

func TestListScanTargetsSendsFilters(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyPage)

	_, _, _, err := ListScanTargets(context.Background(), "scan-1", "cursor-1", &ListScanTargetsFilters{
		Types:         []string{"API_ROUTE", "GRAPHQL_RESOLVER"},
		Sources:       []string{"SOURCE_INFERRED", "SOURCE_SPECIFICATION"},
		Coverage:      []string{"OK", "TIMEOUT"},
		Severities:    []string{"HIGH", "CRITICAL"},
		Issues:        []string{"issue-1", "issue-2"},
		Methods:       []string{"GET", "POST"},
		Search:        "checkout",
		SortType:      "NAME",
		SortDirection: "desc",
	}, 25)
	if err != nil {
		t.Fatalf("list targets: %v", err)
	}

	if (*paths)[0] != "/v3/scans/scan-1/targets" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "cursor", []string{"cursor-1"})
	assertQueryValues(t, query, "size", []string{"25"})
	assertQueryValues(t, query, "types", []string{"API_ROUTE", "GRAPHQL_RESOLVER"})
	assertQueryValues(t, query, "sources", []string{"SOURCE_INFERRED", "SOURCE_SPECIFICATION"})
	assertQueryValues(t, query, "coverage", []string{"OK", "TIMEOUT"})
	assertQueryValues(t, query, "severities", []string{"HIGH", "CRITICAL"})
	assertQueryValues(t, query, "issues", []string{"issue-1", "issue-2"})
	assertQueryValues(t, query, "methods", []string{"GET", "POST"})
	assertQueryValues(t, query, "search", []string{"checkout"})
	assertQueryValues(t, query, "sortType", []string{"NAME"})
	assertQueryValues(t, query, "sortDirection", []string{"desc"})
}

func TestListScansSendsBranchFilters(t *testing.T) {
	_, queries := capturePRD2312(t, prd2312EmptyPage)

	if _, _, _, err := ListScans(context.Background(), "", &ListScansFilters{
		CommitBranch:      "feature/prd",
		LatestBranchScans: "true",
	}, 0); err != nil {
		t.Fatalf("list scans: %v", err)
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "commitBranch", []string{"feature/prd"})
	assertQueryValues(t, query, "latestBranchScans", []string{"true"})
}

func TestListAuditLogsSendsTarget(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyPage)

	if _, _, err := ListAuditLogs(context.Background(), "", &ListAuditLogsFilters{Target: "scan-1"}); err != nil {
		t.Fatalf("list audit logs: %v", err)
	}

	if (*paths)[0] != "/v3/audit/logs" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	assertQueryValues(t, lastQuery(t, queries), "target", []string{"scan-1"})
}

func TestListProjectsSendsIDsAndPermissions(t *testing.T) {
	_, queries := capturePRD2312(t, prd2312EmptyPage)

	if _, _, _, err := ListProjects(context.Background(), "", &ListProjectsFilters{
		IDs:         []string{"proj-1", "proj-2"},
		Permissions: []string{"ADMIN", "EDIT_ASSETS"},
	}, 0); err != nil {
		t.Fatalf("list projects: %v", err)
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "ids", []string{"proj-1,proj-2"})
	assertQueryValues(t, query, "permissions", []string{"ADMIN", "EDIT_ASSETS"})
}

func TestListWorkflowsSendsDnf(t *testing.T) {
	_, queries := capturePRD2312(t, prd2312EmptyPage)

	dnf := `{"op":"and"}`
	if _, _, _, err := ListWorkflows(context.Background(), "", &ListWorkflowsFilters{Dnf: dnf}, 0); err != nil {
		t.Fatalf("list workflows: %v", err)
	}

	assertQueryValues(t, lastQuery(t, queries), "dnf", []string{dnf})
}

func TestListUsersSendsDeactivated(t *testing.T) {
	_, queries := capturePRD2312(t, prd2312EmptyArray)

	if _, err := ListUsers(context.Background(), "", "false"); err != nil {
		t.Fatalf("list users: %v", err)
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "deactivated", []string{"false"})
	if query.Get("search") != "" {
		t.Fatalf("search = %q, want it omitted", query.Get("search"))
	}
}

func TestListRolesSendsSearchOnEveryPage(t *testing.T) {
	var queries []url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "" {
			w.Header().Set("X-Next-Cursor", "page-2")
		}

		_, _ = w.Write([]byte(prd2312EmptyArray))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	if _, err := ListRoles(context.Background(), "admin"); err != nil {
		t.Fatalf("list roles: %v", err)
	}

	if len(queries) != 2 {
		t.Fatalf("requests = %d, want search on both pages", len(queries))
	}

	for _, query := range queries {
		assertQueryValues(t, query, "search", []string{"admin"})
	}
}

func TestListCustomRulesSendsFiltersOnEveryPage(t *testing.T) {
	var queries []url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("cursor") == "" {
			w.Header().Set("X-Next-Cursor", "page-2")
		}

		_, _ = w.Write([]byte(prd2312EmptyArray))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	if _, err := ListCustomRules(context.Background(), &ListCustomRulesFilters{Search: "auth", Context: "DAST"}); err != nil {
		t.Fatalf("list custom rules: %v", err)
	}

	if len(queries) != 2 {
		t.Fatalf("requests = %d, want filters on both pages", len(queries))
	}

	for _, query := range queries {
		assertQueryValues(t, query, "search", []string{"auth"})
		assertQueryValues(t, query, "context", []string{"DAST"})
	}
}

func TestListProblemsSendsProblemFilters(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyPage)

	_, _, _, err := ListProblems(context.Background(), "", &ListProblemsFilters{
		ProblemCodes:      []string{"AUTH_FAILURE", "TIMEOUT"},
		ProblemSeverities: []string{"ERROR", "WARNING"},
	}, 0)
	if err != nil {
		t.Fatalf("list problems: %v", err)
	}

	if (*paths)[0] != "/v3/profiles/problems" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "problemCodes", []string{"AUTH_FAILURE", "TIMEOUT"})
	assertQueryValues(t, query, "problemSeverities", []string{"ERROR", "WARNING"})
}

func TestListScanProblemsSendsProblemFilters(t *testing.T) {
	paths, queries := capturePRD2312(t, prd2312EmptyPage)

	_, _, _, err := ListScanProblems(context.Background(), "", &ListScanProblemsFilters{
		ProblemCodes:      []string{"AUTH_FAILURE", "TIMEOUT"},
		ProblemSeverities: []string{"ERROR", "WARNING"},
	}, 0)
	if err != nil {
		t.Fatalf("list scan problems: %v", err)
	}

	if (*paths)[0] != "/v3/scans/problems" {
		t.Fatalf("path = %s", (*paths)[0])
	}

	query := lastQuery(t, queries)
	assertQueryValues(t, query, "problemCodes", []string{"AUTH_FAILURE", "TIMEOUT"})
	assertQueryValues(t, query, "problemSeverities", []string{"ERROR", "WARNING"})
}

func TestPRD2312FiltersAreOmittedWhenEmpty(t *testing.T) {
	queries := []url.Values{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		body := prd2312EmptyPage
		if strings.Contains(r.URL.Path, "/jira/") || strings.HasPrefix(r.URL.Path, "/v3/users") {
			body = prd2312EmptyArray
		}

		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", prd2312APIKey)

	if _, err := ListJiraProjects(context.Background(), "jira-1", ""); err != nil {
		t.Fatalf("jira projects: %v", err)
	}

	if lastQuery(t, &queries).Get("search") != "" {
		t.Fatal("empty jira search was sent")
	}

	if _, _, _, err := ListScanScenarios(context.Background(), "scan-1", "", "", "", 0); err != nil {
		t.Fatalf("scenarios: %v", err)
	}

	scenarioQuery := lastQuery(t, &queries)
	for _, key := range []string{"cursor", "size", "vector", "search"} {
		if scenarioQuery.Get(key) != "" {
			t.Fatalf("scenarios %s = %q, want it omitted", key, scenarioQuery.Get(key))
		}
	}

	if _, _, _, err := ListScanTargets(context.Background(), "scan-1", "", nil, 0); err != nil {
		t.Fatalf("targets: %v", err)
	}

	targetQuery := lastQuery(t, &queries)
	for _, key := range []string{"types", "sources", "coverage", "severities", "issues", "methods", "search", "sortType", "sortDirection", "cursor"} {
		if targetQuery.Get(key) != "" {
			t.Fatalf("targets %s = %q, want it omitted", key, targetQuery.Get(key))
		}
	}

	if _, _, _, err := ListScans(context.Background(), "", &ListScansFilters{}, 0); err != nil {
		t.Fatalf("scans: %v", err)
	}

	scanQuery := lastQuery(t, &queries)
	if scanQuery.Get("commitBranch") != "" || scanQuery.Get("latestBranchScans") != "" {
		t.Fatalf("empty branch filters were sent: %v", scanQuery)
	}

	if _, _, err := ListAuditLogs(context.Background(), "", &ListAuditLogsFilters{}); err != nil {
		t.Fatalf("audit: %v", err)
	}

	if lastQuery(t, &queries).Get("target") != "" {
		t.Fatal("empty audit target was sent")
	}

	if _, _, _, err := ListProjects(context.Background(), "", &ListProjectsFilters{}, 0); err != nil {
		t.Fatalf("projects: %v", err)
	}

	projectQuery := lastQuery(t, &queries)
	if projectQuery.Get("ids") != "" || len(projectQuery["permissions"]) != 0 {
		t.Fatalf("empty project filters were sent: %v", projectQuery)
	}

	if _, _, _, err := ListWorkflows(context.Background(), "", &ListWorkflowsFilters{}, 0); err != nil {
		t.Fatalf("workflows: %v", err)
	}

	if lastQuery(t, &queries).Get("dnf") != "" {
		t.Fatal("empty workflow dnf was sent")
	}

	if _, err := ListUsers(context.Background(), "", ""); err != nil {
		t.Fatalf("users: %v", err)
	}

	if lastQuery(t, &queries).Get("deactivated") != "" {
		t.Fatal("empty deactivated was sent")
	}

	if _, _, _, err := ListProblems(context.Background(), "", &ListProblemsFilters{}, 0); err != nil {
		t.Fatalf("problems: %v", err)
	}

	problemQuery := lastQuery(t, &queries)
	if len(problemQuery["problemCodes"]) != 0 || len(problemQuery["problemSeverities"]) != 0 {
		t.Fatalf("empty problem filters were sent: %v", problemQuery)
	}

	if _, _, _, err := ListScanProblems(context.Background(), "", &ListScanProblemsFilters{}, 0); err != nil {
		t.Fatalf("scan problems: %v", err)
	}

	scanProblemQuery := lastQuery(t, &queries)
	if len(scanProblemQuery["problemCodes"]) != 0 || len(scanProblemQuery["problemSeverities"]) != 0 {
		t.Fatalf("empty scan problem filters were sent: %v", scanProblemQuery)
	}
}
