package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

func captureCommandQueries(t *testing.T) *[]struct {
	path  string
	query url.Values
} {
	t.Helper()

	seen := []struct {
		path  string
		query url.Values
	}{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, struct {
			path  string
			query url.Values
		}{path: r.URL.Path, query: r.URL.Query()})
		w.Header().Set("Content-Type", "application/json")
		body := `{"data":[],"nextCursor":null,"totalCount":0}`
		switch {
		case strings.HasPrefix(r.URL.Path, "/v3/users"),
			strings.HasPrefix(r.URL.Path, "/v3/roles"),
			strings.HasPrefix(r.URL.Path, "/v3/custom-rules"),
			strings.Contains(r.URL.Path, "/jira/"),
			strings.Contains(r.URL.Path, "/azure-boards/"):
			body = `[]`
		}

		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
	if err := out.SetOutput("json"); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })

	return &seen
}

func runCommand(t *testing.T, command *cobra.Command, args []string) {
	t.Helper()

	command.SetContext(context.Background())
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	if err := command.RunE(command, args); err != nil {
		t.Fatalf("%s: %v", command.CommandPath(), err)
	}
}

func setStringFlag(t *testing.T, target *string, value string) {
	t.Helper()

	previous := *target
	*target = value
	t.Cleanup(func() { *target = previous })
}

func setStringSliceFlag(t *testing.T, target *[]string, value []string) {
	t.Helper()

	previous := *target
	*target = value
	t.Cleanup(func() { *target = previous })
}

func lastCommandQuery(t *testing.T, seen *[]struct {
	path  string
	query url.Values
}, path string) url.Values {
	t.Helper()

	for i := len(*seen) - 1; i >= 0; i-- {
		if (*seen)[i].path == path {
			return (*seen)[i].query
		}
	}

	t.Fatalf("no request for %s in %#v", path, *seen)

	return nil
}

func TestJiraAndAzureLookupsSendTheirFlags(t *testing.T) {
	seen := captureCommandQueries(t)
	setStringFlag(t, &integrationsJiraProjectsSearch, "core")
	setStringFlag(t, &integrationsJiraIssueTypesProjectID, "CT")
	setStringFlag(t, &integrationsAzureAreasProject, "My Project")
	setStringFlag(t, &integrationsAzureAreasWorkItemType, "Bug")

	runCommand(t, integrationsJiraProjectsCmd, []string{"jira-1"})
	runCommand(t, integrationsJiraIssueTypesCmd, []string{"jira-1"})
	runCommand(t, integrationsAzureAreasCmd, []string{"azure-1"})

	if got := lastCommandQuery(t, seen, "/v3/integrations/jira/jira-1/projects").Get("search"); got != "core" {
		t.Fatalf("jira projects search = %q", got)
	}

	if got := lastCommandQuery(t, seen, "/v3/integrations/jira/jira-1/issue-types").Get("projectId"); got != "CT" {
		t.Fatalf("jira issue types projectId = %q", got)
	}

	areas := lastCommandQuery(t, seen, "/v3/integrations/azure-boards/azure-1/areas")
	if areas.Get("project") != "My Project" || areas.Get("workItemType") != "Bug" {
		t.Fatalf("azure areas query = %v", areas)
	}
}

func TestJiraIssueTypesAndAzureAreasRequireTheirFlags(t *testing.T) {
	if err := out.SetOutput("pretty"); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })
	setStringFlag(t, &integrationsJiraIssueTypesProjectID, "")
	err := integrationsJiraIssueTypesCmd.RunE(integrationsJiraIssueTypesCmd, []string{"jira-1"})
	if err == nil || !strings.Contains(err.Error(), "--project-id is required") {
		t.Fatalf("jira issue types err = %v", err)
	}

	setStringFlag(t, &integrationsAzureAreasProject, "")
	setStringFlag(t, &integrationsAzureAreasWorkItemType, "Bug")
	err = integrationsAzureAreasCmd.RunE(integrationsAzureAreasCmd, []string{"azure-1"})
	if err == nil || !strings.Contains(err.Error(), "--project is required") {
		t.Fatalf("azure areas project err = %v", err)
	}

	setStringFlag(t, &integrationsAzureAreasProject, "My Project")
	setStringFlag(t, &integrationsAzureAreasWorkItemType, "")
	err = integrationsAzureAreasCmd.RunE(integrationsAzureAreasCmd, []string{"azure-1"})
	if err == nil || !strings.Contains(err.Error(), "--work-item-type is required") {
		t.Fatalf("azure areas work item err = %v", err)
	}
}

func TestScansScenariosSendsVectorAndSearch(t *testing.T) {
	seen := captureCommandQueries(t)
	resetFlag(t, scansScenariosCmd, "size")
	resetFlag(t, scansScenariosCmd, "cursor")
	setStringFlag(t, &scanScenariosVector, "INJECTION")
	setStringFlag(t, &scanScenariosSearch, "login")

	runCommand(t, scansScenariosCmd, []string{"scan-1"})

	query := lastCommandQuery(t, seen, "/v3/scans/scan-1/scenarios")
	if query.Get("vector") != "INJECTION" || query.Get("search") != "login" {
		t.Fatalf("scenarios query = %v", query)
	}
}

func TestScansTargetsSendsFilters(t *testing.T) {
	seen := captureCommandQueries(t)
	setStringFlag(t, &scanTargetsType, "API_ROUTE,GRAPHQL_RESOLVER")
	setStringSliceFlag(t, &scanTargetsSources, []string{"SOURCE_INFERRED", "SOURCE_SPECIFICATION"})
	setStringSliceFlag(t, &scanTargetsCoverage, []string{"OK", "TIMEOUT"})
	setStringSliceFlag(t, &scanTargetsSeverities, []string{"HIGH"})
	setStringSliceFlag(t, &scanTargetsIssues, []string{"issue-1", "issue-2"})
	setStringSliceFlag(t, &scanTargetsMethods, []string{"GET", "POST"})
	setStringFlag(t, &scanTargetsSearch, "checkout")
	setStringFlag(t, &scanTargetsSortType, "NAME")
	setStringFlag(t, &scanTargetsSortDirection, "desc")
	// PRD-2313: --size is a page size when the flag is set, and --size 0 is
	// rejected. Reset both page flags so this stays the full listing.
	resetFlag(t, scanTargetsCmd, "size")
	resetFlag(t, scanTargetsCmd, "cursor")

	runCommand(t, scanTargetsCmd, []string{"scan-1"})

	query := lastCommandQuery(t, seen, "/v3/scans/scan-1/targets")
	if strings.Join(query["types"], ",") != "API_ROUTE,GRAPHQL_RESOLVER" {
		t.Fatalf("types = %#v", query["types"])
	}

	if strings.Join(query["sources"], ",") != "SOURCE_INFERRED,SOURCE_SPECIFICATION" {
		t.Fatalf("sources = %#v", query["sources"])
	}

	if strings.Join(query["coverage"], ",") != "OK,TIMEOUT" {
		t.Fatalf("coverage = %#v", query["coverage"])
	}

	if strings.Join(query["severities"], ",") != "HIGH" {
		t.Fatalf("severities = %#v", query["severities"])
	}

	if strings.Join(query["issues"], ",") != "issue-1,issue-2" {
		t.Fatalf("issues = %#v", query["issues"])
	}

	if strings.Join(query["methods"], ",") != "GET,POST" {
		t.Fatalf("methods = %#v", query["methods"])
	}

	if query.Get("search") != "checkout" || query.Get("sortType") != "NAME" || query.Get("sortDirection") != "desc" {
		t.Fatalf("targets query = %v", query)
	}
}

func TestScansListSendsBranchFilters(t *testing.T) {
	seen := captureCommandQueries(t)
	resetFlag(t, scansListCmd, "size")
	resetFlag(t, scansListCmd, "cursor")
	setStringSliceFlag(t, &scanProfileIDs, nil)
	setStringFlag(t, &scanCommitBranch, "feature/prd")
	setStringFlag(t, &scanLatestBranchScans, "true")

	runCommand(t, scansListCmd, nil)

	query := lastCommandQuery(t, seen, "/v3/scans")
	if query.Get("commitBranch") != "feature/prd" || query.Get("latestBranchScans") != "true" {
		t.Fatalf("scans query = %v", query)
	}
}

func TestAuditListSendsTarget(t *testing.T) {
	seen := captureCommandQueries(t)
	setStringFlag(t, &auditCmdTarget, "scan-1")

	runCommand(t, auditListCmd, nil)

	if got := lastCommandQuery(t, seen, "/v3/audit/logs").Get("target"); got != "scan-1" {
		t.Fatalf("target = %q", got)
	}
}

func TestProjectsListSendsIDAndPermission(t *testing.T) {
	seen := captureCommandQueries(t)
	resetFlag(t, projectsListCmd, "size")
	resetFlag(t, projectsListCmd, "cursor")
	setStringSliceFlag(t, &projectsIDs, []string{"proj-1", "proj-2"})
	setStringSliceFlag(t, &projectsPermissions, []string{"ADMIN", "EDIT_ASSETS"})

	runCommand(t, projectsListCmd, nil)

	query := lastCommandQuery(t, seen, "/v3/projects")
	if query.Get("ids") != "proj-1,proj-2" {
		t.Fatalf("ids = %q", query.Get("ids"))
	}

	if strings.Join(query["permissions"], ",") != "ADMIN,EDIT_ASSETS" {
		t.Fatalf("permissions = %#v", query["permissions"])
	}
}

func TestWorkflowsListSendsDnf(t *testing.T) {
	seen := captureCommandQueries(t)
	resetFlag(t, workflowsListCmd, "size")
	resetFlag(t, workflowsListCmd, "cursor")
	setStringFlag(t, &workflowsDnf, `{"op":"and"}`)

	runCommand(t, workflowsListCmd, nil)

	if got := lastCommandQuery(t, seen, "/v3/workflows").Get("dnf"); got != `{"op":"and"}` {
		t.Fatalf("dnf = %q", got)
	}
}

func TestUsersListSendsDeactivated(t *testing.T) {
	seen := captureCommandQueries(t)
	setStringFlag(t, &usersSearch, "")
	setStringFlag(t, &usersDeactivated, "false")

	runCommand(t, usersListCmd, nil)

	if got := lastCommandQuery(t, seen, "/v3/users").Get("deactivated"); got != "false" {
		t.Fatalf("deactivated = %q", got)
	}
}

func TestRolesListSendsSearch(t *testing.T) {
	seen := captureCommandQueries(t)
	setStringFlag(t, &rolesSearch, "admin")

	runCommand(t, rolesListCmd, nil)

	if got := lastCommandQuery(t, seen, "/v3/roles").Get("search"); got != "admin" {
		t.Fatalf("search = %q", got)
	}
}

func TestCustomRulesListSendsSearchAndContext(t *testing.T) {
	seen := captureCommandQueries(t)
	setStringFlag(t, &customRulesListSearch, "auth")
	setStringFlag(t, &customRulesListContext, "DAST")

	runCommand(t, customRulesListCmd, nil)

	query := lastCommandQuery(t, seen, "/v3/custom-rules")
	if query.Get("search") != "auth" || query.Get("context") != "DAST" {
		t.Fatalf("custom rules query = %v", query)
	}
}

func TestProblemsCommandsSendProblemFilters(t *testing.T) {
	seen := captureCommandQueries(t)
	resetFlag(t, problemsCmd, "size")
	resetFlag(t, problemsCmd, "cursor")
	resetFlag(t, profileProblemsCmd, "size")
	resetFlag(t, profileProblemsCmd, "cursor")
	resetFlag(t, scansProblemsCmd, "size")
	resetFlag(t, scansProblemsCmd, "cursor")
	setStringSliceFlag(t, &problemsProblemCodes, []string{"AUTH_FAILURE"})
	setStringSliceFlag(t, &problemsProblemSeverities, []string{"ERROR"})
	setStringSliceFlag(t, &profileProblemsProblemCodes, []string{"TIMEOUT"})
	setStringSliceFlag(t, &profileProblemsProblemSeverities, []string{"WARNING"})
	setStringSliceFlag(t, &scanProblemsProblemCodes, []string{"UNREACHABLE_ASSET"})
	setStringSliceFlag(t, &scanProblemsProblemSeverities, []string{"INFO"})

	runCommand(t, problemsCmd, nil)
	runCommand(t, profileProblemsCmd, nil)
	runCommand(t, scansProblemsCmd, nil)

	problems := lastCommandQuery(t, seen, "/v3/profiles/problems")
	// profiles problems runs second and hits the same path, so the last
	// /profiles/problems request is that command. The top-level problems
	// command is the earlier request.
	var first url.Values
	for _, request := range *seen {
		if request.path == "/v3/profiles/problems" {
			first = request.query

			break
		}
	}

	if first.Get("problemCodes") != "AUTH_FAILURE" || first.Get("problemSeverities") != "ERROR" {
		t.Fatalf("problems query = %v", first)
	}

	if problems.Get("problemCodes") != "TIMEOUT" || problems.Get("problemSeverities") != "WARNING" {
		t.Fatalf("profiles problems query = %v", problems)
	}

	scans := lastCommandQuery(t, seen, "/v3/scans/problems")
	if scans.Get("problemCodes") != "UNREACHABLE_ASSET" || scans.Get("problemSeverities") != "INFO" {
		t.Fatalf("scans problems query = %v", scans)
	}
}
