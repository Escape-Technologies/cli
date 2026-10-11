package escape

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

const parityWriteID = "00000000-0000-0000-0000-0000000000aa"

// securityTestWire is a catalog document the update response unmarshals into.
const securityTestWire = `{"id":"` + parityWriteID + `","name":"SQL injection","category":"INJECTION","description":"Injects SQL payloads into every parameter.","documentationPath":"/docs/tests/sql-injection","scanners":["BLST_REST"],"assetTypes":["REST"],"enabled":true,"compliance":[{"framework":"PCI_DSS","item":"6.5.1"}],"issueCounts":{"critical":1,"high":2,"medium":3,"low":4,"info":5}}`

type capturedWrite struct {
	method string
	path   string
	query  url.Values
	body   []byte
}

func captureWrite(t *testing.T, response string) *capturedWrite {
	t.Helper()

	got := &capturedWrite{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.Query()
		got.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	return got
}

func decodeWriteBody(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("body %s: %v", raw, err)
	}

	return body
}

func TestBulkTriggerIssueWorkflowsPostsTheBody(t *testing.T) {
	got := captureWrite(t, `{"ids":["`+parityWriteID+`"],"count":1,"dryRun":true,"workflows":[]}`)
	request := v3.NewBulkTriggerIssueWorkflowsRequest([]string{parityWriteID})
	request.SetDryRun(true)
	where := v3.BulkTriggerIssueWorkflowsRequestWhere{}
	where.SetIds([]string{parityWriteID})
	request.SetWhere(where)

	if _, err := BulkTriggerIssueWorkflows(context.Background(), *request); err != nil {
		t.Fatalf("bulk trigger: %v", err)
	}

	if got.method != http.MethodPost || got.path != "/v3/issues/bulk-workflow" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	if !reflect.DeepEqual(body["workflowIds"], []any{parityWriteID}) {
		t.Fatalf("workflowIds = %#v", body["workflowIds"])
	}

	if body["dryRun"] != true {
		t.Fatalf("dryRun = %#v", body["dryRun"])
	}

	whereBody, _ := body["where"].(map[string]any)
	if !reflect.DeepEqual(whereBody["ids"], []any{parityWriteID}) {
		t.Fatalf("where = %#v", body["where"])
	}
}

func TestBulkUpdateProfilesPostsTheBody(t *testing.T) {
	got := captureWrite(t, `{"success":true}`)
	raw := []byte(`{"profileIds":["` + parityWriteID + `"],"tagIds":["` + parityWriteID + `"],"stdinProbe":"body"}`)

	if _, err := BulkUpdateProfiles(context.Background(), raw); err != nil {
		t.Fatalf("bulk update profiles: %v", err)
	}

	if got.method != http.MethodPost || got.path != "/v3/profiles/bulk-update" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	if !reflect.DeepEqual(body["profileIds"], []any{parityWriteID}) || body["stdinProbe"] != "body" {
		t.Fatalf("body = %#v", body)
	}
}

func TestUpdateSecurityTestPutsEnabledAndSeverity(t *testing.T) {
	got := captureWrite(t, securityTestWire)
	request := v3.NewUpdateSecurityTestRequest()
	request.SetEnabled(false)
	request.SetCustomSeverity(v3.ENUMPROPERTIESDATAITEMSPROPERTIESASSETPROPERTIESSTATISTICSPROPERTIESISSUESPROPERTIESSEVERITIESITEMSPROPERTIESSEVERITY_HIGH)

	if _, err := UpdateSecurityTest(context.Background(), parityWriteID, *request); err != nil {
		t.Fatalf("update security test: %v", err)
	}

	if got.method != http.MethodPut || got.path != "/v3/security-tests/"+parityWriteID {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	if body["enabled"] != false || body["customSeverity"] != "HIGH" {
		t.Fatalf("body = %#v, want enabled and customSeverity together", body)
	}
}

func TestUpdateSecurityTestPutsANullSeverityReset(t *testing.T) {
	got := captureWrite(t, securityTestWire)
	request := v3.NewUpdateSecurityTestRequest()
	request.AdditionalProperties = map[string]interface{}{"customSeverity": nil}

	if _, err := UpdateSecurityTest(context.Background(), parityWriteID, *request); err != nil {
		t.Fatalf("reset security test: %v", err)
	}

	body := decodeWriteBody(t, got.body)
	value, ok := body["customSeverity"]
	if !ok || value != nil {
		t.Fatalf("customSeverity = %#v, want JSON null", value)
	}
}

func TestBulkUpdateSecurityTestsPostsTheFilter(t *testing.T) {
	got := captureWrite(t, `{"ids":["`+parityWriteID+`"],"count":1,"dryRun":true}`)
	request := v3.NewBulkUpdateSecurityTestsRequest()
	request.SetEnabled(false)
	request.SetDryRun(true)
	request.AdditionalProperties = map[string]interface{}{"customSeverity": nil}
	where := v3.BulkUpdateSecurityTestsRequestWhere{}
	where.SetIds([]string{parityWriteID})
	request.SetWhere(where)

	if _, err := BulkUpdateSecurityTests(context.Background(), *request); err != nil {
		t.Fatalf("bulk update security tests: %v", err)
	}

	if got.method != http.MethodPost || got.path != "/v3/security-tests/bulk-update" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	value, ok := body["customSeverity"]
	if body["enabled"] != false || body["dryRun"] != true || !ok || value != nil {
		t.Fatalf("body = %#v, want enabled, dryRun and a null customSeverity", body)
	}
}

func TestUpdateEnabledCompliancesPutsTheFrameworkList(t *testing.T) {
	got := captureWrite(t, `{"data":["OWASP","PCI_DSS"]}`)

	if _, err := UpdateEnabledCompliances(context.Background(), []string{"OWASP", "PCI_DSS"}); err != nil {
		t.Fatalf("enable compliances: %v", err)
	}

	if got.method != http.MethodPut || got.path != "/v3/compliance/enabled" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	if !reflect.DeepEqual(body["frameworks"], []any{"OWASP", "PCI_DSS"}) {
		t.Fatalf("frameworks = %#v", body["frameworks"])
	}
}

func TestRenameExportPutsTheName(t *testing.T) {
	got := captureWrite(t, `{"id":"`+parityWriteID+`","name":"Q3 penetration test"}`)

	renamed, err := RenameExport(context.Background(), parityWriteID, "Q3 penetration test")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}

	if got.method != http.MethodPut || got.path != "/v3/jobs/"+parityWriteID+"/name" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	if body["name"] != "Q3 penetration test" || renamed.GetName() != "Q3 penetration test" {
		t.Fatalf("name = %#v, renamed = %#v", body["name"], renamed.GetName())
	}
}

func TestSendJobArtefactsPostsRecipientsAndOmitsAnEmptySubject(t *testing.T) {
	got := captureWrite(t, `{"sent":true}`)

	if _, err := SendJobArtefacts(context.Background(), parityWriteID, []string{"a@example.com", "b@example.com"}, ""); err != nil {
		t.Fatalf("send artefacts: %v", err)
	}

	if got.method != http.MethodPost || got.path != "/v3/jobs/"+parityWriteID+"/send-artefacts" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	if !reflect.DeepEqual(body["emails"], []any{"a@example.com", "b@example.com"}) {
		t.Fatalf("emails = %#v", body["emails"])
	}

	if _, ok := body["subject"]; ok {
		t.Fatalf("subject = %#v, want it omitted when empty", body["subject"])
	}
}

func TestSendJobArtefactsPostsTheSubject(t *testing.T) {
	got := captureWrite(t, `{"sent":true}`)

	if _, err := SendJobArtefacts(context.Background(), parityWriteID, []string{"a@example.com"}, "Q3 report"); err != nil {
		t.Fatalf("send artefacts: %v", err)
	}

	body := decodeWriteBody(t, got.body)
	if body["subject"] != "Q3 report" {
		t.Fatalf("subject = %#v", body["subject"])
	}
}

func TestCreateOrganizationScannerConfigurationPostsTheContent(t *testing.T) {
	got := captureWrite(t, `{"id":"`+parityWriteID+`","createdAt":"2026-01-01T00:00:00Z"}`)
	raw := []byte(`{"content":{"mode":"read_only"},"stdinProbe":"body"}`)

	if _, err := CreateOrganizationScannerConfiguration(context.Background(), raw); err != nil {
		t.Fatalf("create scanner configuration: %v", err)
	}

	if got.method != http.MethodPost || got.path != "/v3/organization-settings/scanner-configuration" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	body := decodeWriteBody(t, got.body)
	content, _ := body["content"].(map[string]any)
	if content["mode"] != "read_only" || body["stdinProbe"] != "body" {
		t.Fatalf("body = %#v", body)
	}
}

func TestListGithubRepositoriesRepeatsFullNames(t *testing.T) {
	got := captureWrite(t, `{"data":[{"id":"repo-1","fullName":"org/api","name":"api","ownerLogin":"org","cloneUrl":"https://github.com/org/api.git","htmlUrl":"https://github.com/org/api","defaultBranch":"main","private":true,"archived":false,"description":null,"integrationIds":["inst-1"]}],"nextCursor":null,"totalCount":1}`)

	repos, _, _, err := ListGithubRepositories(context.Background(), "", "", 0, &ListGithubRepositoriesFilters{
		Search:    "api",
		FullNames: []string{"org/api", "org/web"},
	})
	if err != nil {
		t.Fatalf("list repositories: %v", err)
	}

	if len(repos) != 1 || repos[0].FullName != "org/api" || repos[0].CloneURL != "https://github.com/org/api.git" || repos[0].HTMLURL != "https://github.com/org/api" || repos[0].Description != nil || len(repos[0].IntegrationIDs) != 1 || repos[0].IntegrationIDs[0] != "inst-1" {
		t.Fatalf("repos = %#v", repos)
	}

	if got.method != http.MethodGet || got.path != "/v3/integrations/github/repositories" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	if got.query.Get("search") != "api" {
		t.Fatalf("search = %q", got.query.Get("search"))
	}

	if !reflect.DeepEqual(got.query["fullNames"], []string{"org/api", "org/web"}) {
		t.Fatalf("fullNames = %#v, want one query value per repository", got.query["fullNames"])
	}
}

func TestListGithubRepositoriesScopesOneIntegration(t *testing.T) {
	got := captureWrite(t, `{"data":[],"nextCursor":null,"totalCount":0}`)

	if _, _, _, err := ListGithubRepositories(context.Background(), parityWriteID, "", 0, nil); err != nil {
		t.Fatalf("list repositories: %v", err)
	}

	if got.method != http.MethodGet || got.path != "/v3/integrations/github/"+parityWriteID+"/repositories" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}
}

func TestListIntegrationVaultsSendsKindAndSearch(t *testing.T) {
	got := captureWrite(t, `[{"name":"prod","externalId":"vault-1"}]`)

	vaults, err := ListIntegrationVaults(context.Background(), "azure-key-vault", parityWriteID, "prod")
	if err != nil {
		t.Fatalf("list vaults: %v", err)
	}

	if got.method != http.MethodGet || got.path != "/v3/integrations/azure-key-vault/"+parityWriteID+"/vaults" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}

	if got.query.Get("search") != "prod" {
		t.Fatalf("search = %q", got.query.Get("search"))
	}

	if len(vaults) != 1 || vaults[0].ExternalID != "vault-1" || vaults[0].Name != "prod" {
		t.Fatalf("vaults = %#v", vaults)
	}
}
