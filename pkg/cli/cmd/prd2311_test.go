package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestPRD2311CommandsSendTheDocumentedRequest drives each command added for
// PRD-2311 against one route and pins the method, path and body it sends.
func TestPRD2311CommandsSendTheDocumentedRequest(t *testing.T) {
	securityTestBody := func(t *testing.T) string {
		t.Helper()

		raw, err := json.Marshal(securityTestFixture())
		if err != nil {
			t.Fatalf("fixture: %v", err)
		}

		return string(raw)
	}

	t.Run("issues bulk-workflow", func(t *testing.T) {
		restoreBulkWorkflowFlags(t)
		clearBulkWorkflowFlags()
		bulkWorkflowIDs = []string{parityID}
		bulkWorkflowSeverities = []string{"HIGH"}
		bulkWorkflowDryRun = true

		runParityBatch2Case(t, parityBatch2Case{
			name:    "issues bulk-workflow",
			command: issueBulkWorkflowCmd,
			args:    func() []string { return nil },
			method:  http.MethodPost,
			path:    "/v3/issues/bulk-workflow",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if !reflect.DeepEqual(got["workflowIds"], []any{parityID}) || got["dryRun"] != true {
					t.Fatalf("body = %#v", got)
				}

				where, _ := got["where"].(map[string]any)
				if !reflect.DeepEqual(where["severities"], []any{"HIGH"}) {
					t.Fatalf("where = %#v", got["where"])
				}
			},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"ids":["`+parityID+`"],"count":1,"dryRun":true,"workflows":[]}`)
			},
		})
	})

	t.Run("profiles bulk-update", func(t *testing.T) {
		runParityBatch2Case(t, parityBatch2Case{
			name:    "profiles bulk-update",
			command: profileBulkUpdateCmd,
			args:    func() []string { return nil },
			method:  http.MethodPost,
			path:    "/v3/profiles/bulk-update",
			stdin:   `{"profileIds":["` + parityID + `"],"tagIds":["` + parityID + `"]}`,
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if !reflect.DeepEqual(got["profileIds"], []any{parityID}) || !reflect.DeepEqual(got["tagIds"], []any{parityID}) {
					t.Fatalf("body = %#v", got)
				}
			},
			respond: func(w http.ResponseWriter) { writeParityJSON(w, `{"success":true}`) },
		})
	})

	t.Run("security-tests update sends enabled and severity", func(t *testing.T) {
		restoreSecurityTestUpdateFlags(t)
		securityTestsUpdateEnabled = "false"
		securityTestsUpdateSeverity = "HIGH"
		securityTestsUpdateResetSeverity = false

		runParityBatch2Case(t, parityBatch2Case{
			name:    "security-tests update",
			command: securityTestsUpdateCmd,
			args:    func() []string { return []string{parityID} },
			method:  http.MethodPut,
			path:    "/v3/security-tests/" + parityID,
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if got["enabled"] != false || got["customSeverity"] != "HIGH" {
					t.Fatalf("body = %#v, want both fields", got)
				}
			},
			respond: func(w http.ResponseWriter) { writeParityJSON(w, securityTestBody(t)) },
		})
	})

	t.Run("security-tests update resets severity with null", func(t *testing.T) {
		restoreSecurityTestUpdateFlags(t)
		securityTestsUpdateEnabled = ""
		securityTestsUpdateSeverity = ""
		securityTestsUpdateResetSeverity = true

		runParityBatch2Case(t, parityBatch2Case{
			name:    "security-tests update reset",
			command: securityTestsUpdateCmd,
			args:    func() []string { return []string{parityID} },
			method:  http.MethodPut,
			path:    "/v3/security-tests/" + parityID,
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				value, ok := got["customSeverity"]
				if !ok || value != nil {
					t.Fatalf("customSeverity = %#v, want JSON null", value)
				}

				if _, present := got["enabled"]; present {
					t.Fatalf("enabled = %#v, want it omitted", got["enabled"])
				}
			},
			respond: func(w http.ResponseWriter) { writeParityJSON(w, securityTestBody(t)) },
		})
	})

	t.Run("security-tests bulk-update", func(t *testing.T) {
		restoreSecurityTestBulkFlags(t)
		clearSecurityTestBulkFlags()
		securityTestsBulkEnabled = "false"
		securityTestsBulkSeverity = "HIGH"
		securityTestsBulkDryRun = true
		securityTestsBulkIDs = []string{parityID}

		runParityBatch2Case(t, parityBatch2Case{
			name:    "security-tests bulk-update",
			command: securityTestsBulkUpdateCmd,
			args:    func() []string { return nil },
			method:  http.MethodPost,
			path:    "/v3/security-tests/bulk-update",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if got["enabled"] != false || got["customSeverity"] != "HIGH" || got["dryRun"] != true {
					t.Fatalf("body = %#v", got)
				}

				where, _ := got["where"].(map[string]any)
				if !reflect.DeepEqual(where["ids"], []any{parityID}) {
					t.Fatalf("where = %#v", got["where"])
				}
			},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"ids":["`+parityID+`"],"count":1,"dryRun":true}`)
			},
		})
	})

	t.Run("security-tests bulk-update resets severity", func(t *testing.T) {
		restoreSecurityTestBulkFlags(t)
		clearSecurityTestBulkFlags()
		securityTestsBulkResetSeverity = true
		securityTestsBulkIDs = []string{parityID}

		runParityBatch2Case(t, parityBatch2Case{
			name:    "security-tests bulk-update reset",
			command: securityTestsBulkUpdateCmd,
			args:    func() []string { return nil },
			method:  http.MethodPost,
			path:    "/v3/security-tests/bulk-update",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				value, ok := got["customSeverity"]
				if !ok || value != nil {
					t.Fatalf("customSeverity = %#v, want JSON null", value)
				}
			},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"ids":["`+parityID+`"],"count":1,"dryRun":false}`)
			},
		})
	})

	t.Run("compliance enable", func(t *testing.T) {
		previous := append([]string(nil), complianceEnableFrameworks...)
		t.Cleanup(func() { complianceEnableFrameworks = previous })
		complianceEnableFrameworks = []string{"OWASP", "PCI_DSS"}

		runParityBatch2Case(t, parityBatch2Case{
			name:    "compliance enable",
			command: complianceEnableCmd,
			args:    func() []string { return nil },
			method:  http.MethodPut,
			path:    "/v3/compliance/enabled",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if !reflect.DeepEqual(got["frameworks"], []any{"OWASP", "PCI_DSS"}) {
					t.Fatalf("frameworks = %#v", got["frameworks"])
				}
			},
			respond: func(w http.ResponseWriter) { writeParityJSON(w, `{"data":["OWASP","PCI_DSS"]}`) },
		})
	})

	t.Run("jobs rename", func(t *testing.T) {
		previous := jobsRenameName
		t.Cleanup(func() { jobsRenameName = previous })
		jobsRenameName = "Q3 penetration test"

		runParityBatch2Case(t, parityBatch2Case{
			name:    "jobs rename",
			command: jobsRenameCmd,
			args:    func() []string { return []string{parityID} },
			method:  http.MethodPut,
			path:    "/v3/jobs/" + parityID + "/name",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if got["name"] != "Q3 penetration test" {
					t.Fatalf("name = %#v", got["name"])
				}
			},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"id":"`+parityID+`","name":"Q3 penetration test"}`)
			},
		})
	})

	t.Run("jobs send-artefacts", func(t *testing.T) {
		previousEmails := append([]string(nil), jobsArtefactsEmails...)
		previousSubject := jobsArtefactsSubject
		t.Cleanup(func() {
			jobsArtefactsEmails = previousEmails
			jobsArtefactsSubject = previousSubject
		})
		jobsArtefactsEmails = []string{"a@example.com", "b@example.com"}
		jobsArtefactsSubject = "Q3 report"

		runParityBatch2Case(t, parityBatch2Case{
			name:    "jobs send-artefacts",
			command: jobsSendArtefactsCmd,
			args:    func() []string { return []string{parityID} },
			method:  http.MethodPost,
			path:    "/v3/jobs/" + parityID + "/send-artefacts",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if !reflect.DeepEqual(got["emails"], []any{"a@example.com", "b@example.com"}) || got["subject"] != "Q3 report" {
					t.Fatalf("body = %#v", got)
				}
			},
			respond: func(w http.ResponseWriter) { writeParityJSON(w, `{"sent":true}`) },
		})
	})

	t.Run("scanner-configuration create", func(t *testing.T) {
		runParityBatch2Case(t, parityBatch2Case{
			name:    "scanner-configuration create",
			command: orgSettingsScannerConfigurationCreateCmd,
			args:    func() []string { return nil },
			method:  http.MethodPost,
			path:    "/v3/organization-settings/scanner-configuration",
			stdin:   `{"content":{"mode":"read_only"}}`,
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				content, _ := got["content"].(map[string]any)
				if content["mode"] != "read_only" {
					t.Fatalf("content = %#v", got["content"])
				}
			},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"id":"`+parityID+`","createdAt":"2026-01-01T00:00:00Z"}`)
			},
		})
	})

	t.Run("integrations repositories", func(t *testing.T) {
		restoreIntegrationRepositoryFlags(t)
		integrationsReposSearch = "api"
		integrationsReposIntegrationID = ""
		integrationsReposFullNames = nil

		runParityBatch2Case(t, parityBatch2Case{
			name:    "integrations repositories",
			command: integrationsRepositoriesCmd,
			args:    func() []string { return nil },
			method:  http.MethodGet,
			path:    "/v3/integrations/github/repositories",
			query:   map[string]string{"search": "api"},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"data":[{"fullName":"org/api","defaultBranch":"main","private":false,"archived":false,"cloneUrl":"https://github.com/org/api"}],"nextCursor":null,"totalCount":1}`)
			},
		})
	})

	t.Run("integrations repositories one integration", func(t *testing.T) {
		restoreIntegrationRepositoryFlags(t)
		integrationsReposSearch = ""
		integrationsReposIntegrationID = parityID
		integrationsReposFullNames = nil

		runParityBatch2Case(t, parityBatch2Case{
			name:    "integrations repositories one integration",
			command: integrationsRepositoriesCmd,
			args:    func() []string { return nil },
			method:  http.MethodGet,
			path:    "/v3/integrations/github/" + parityID + "/repositories",
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"data":[],"nextCursor":null,"totalCount":0}`)
			},
		})
	})
}

// TestPRD2311CommandsRejectIncompleteInput keeps a write from reaching the API
// when the flags do not describe a complete request.
func TestBulkWorkflowKeepsStdinWhereFields(t *testing.T) {
	restoreBulkWorkflowFlags(t)
	clearBulkWorkflowFlags()
	bulkWorkflowIDs = []string{parityID}
	bulkWorkflowSeverities = []string{"HIGH"}
	bulkWorkflowDryRun = true

	runParityBatch2Case(t, parityBatch2Case{
		name:    "issues bulk-workflow merges severity onto stdin where",
		command: issueBulkWorkflowCmd,
		args:    func() []string { return nil },
		method:  http.MethodPost,
		path:    "/v3/issues/bulk-workflow",
		stdin:   `{"workflowIds":["` + parityID + `"],"where":{"assetIds":["asset-a"]}}`,
		body: func(t *testing.T, got map[string]any) {
			t.Helper()
			where, _ := got["where"].(map[string]any)
			if !reflect.DeepEqual(where["assetIds"], []any{"asset-a"}) || !reflect.DeepEqual(where["severities"], []any{"HIGH"}) {
				t.Fatalf("where = %#v", got["where"])
			}
		},
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"ids":["`+parityID+`"],"count":1,"dryRun":true,"workflows":[]}`)
		},
	})
}

func TestComplianceEnableEmptyListDisablesEveryFramework(t *testing.T) {
	previous := append([]string(nil), complianceEnableFrameworks...)
	flag := complianceEnableCmd.Flags().Lookup("framework")
	previousChanged := flag.Changed
	t.Cleanup(func() {
		complianceEnableFrameworks = previous
		flag.Changed = previousChanged
	})
	resetFlag(t, complianceEnableCmd, "dry-run")
	resetFlag(t, complianceEnableCmd, "yes")
	// A string slice that was set earlier appends on the next Set. Replace
	// clears it so `--framework=` is the only value applied. That argv is
	// how the MCP executor renders an empty list, and pflag reads it as
	// []string{}. The PUT body is {"frameworks":[]} rather than null.
	replacer, ok := flag.Value.(interface{ Replace([]string) error })
	if !ok {
		t.Fatal("framework flag cannot replace its slice")
	}

	if err := replacer.Replace(nil); err != nil {
		t.Fatalf("replace framework: %v", err)
	}

	if err := complianceEnableCmd.ParseFlags([]string{"--framework="}); err != nil {
		t.Fatalf("parse --framework=: %v", err)
	}

	if len(complianceEnableFrameworks) != 0 {
		t.Fatalf("frameworks = %#v, want an empty list", complianceEnableFrameworks)
	}

	runParityBatch2Case(t, parityBatch2Case{
		name:    "compliance enable disables every framework",
		command: complianceEnableCmd,
		args:    func() []string { return nil },
		method:  http.MethodPut,
		path:    "/v3/compliance/enabled",
		body: func(t *testing.T, got map[string]any) {
			t.Helper()
			frameworks, ok := got["frameworks"].([]any)
			if !ok || len(frameworks) != 0 {
				t.Fatalf("frameworks = %#v, want []", got["frameworks"])
			}
		},
		respond: func(w http.ResponseWriter) { writeParityJSON(w, `{"data":[]}`) },
	})
}

func TestPRD2311CommandsRejectIncompleteInput(t *testing.T) {
	t.Run("bulk-workflow needs a workflow", func(t *testing.T) {
		restoreBulkWorkflowFlags(t)
		clearBulkWorkflowFlags()
		err := runCommandExpectingFailure(t, issueBulkWorkflowCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "--workflow-id") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("bulk-workflow rejects all with a filter", func(t *testing.T) {
		restoreBulkWorkflowFlags(t)
		clearBulkWorkflowFlags()
		bulkWorkflowIDs = []string{parityID}
		bulkWorkflowAll = true
		bulkWorkflowSeverities = []string{"HIGH"}
		err := runCommandExpectingFailure(t, issueBulkWorkflowCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "cannot be combined") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("security-tests update needs a change", func(t *testing.T) {
		restoreSecurityTestUpdateFlags(t)
		securityTestsUpdateEnabled = ""
		securityTestsUpdateSeverity = ""
		securityTestsUpdateResetSeverity = false
		err := runCommandExpectingFailure(t, securityTestsUpdateCmd, []string{parityID})
		if err == nil || !strings.Contains(err.Error(), "at least one") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("security-tests update rejects reset with a severity", func(t *testing.T) {
		restoreSecurityTestUpdateFlags(t)
		securityTestsUpdateEnabled = ""
		securityTestsUpdateSeverity = "HIGH"
		securityTestsUpdateResetSeverity = true
		err := runCommandExpectingFailure(t, securityTestsUpdateCmd, []string{parityID})
		if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("compliance enable rejects an unknown framework", func(t *testing.T) {
		previous := append([]string(nil), complianceEnableFrameworks...)
		t.Cleanup(func() { complianceEnableFrameworks = previous })
		complianceEnableFrameworks = []string{"NOT_A_FRAMEWORK"}
		err := runCommandExpectingFailure(t, complianceEnableCmd, nil)
		if err == nil || !strings.Contains(err.Error(), "invalid framework") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("jobs rename needs a name", func(t *testing.T) {
		previous := jobsRenameName
		t.Cleanup(func() { jobsRenameName = previous })
		jobsRenameName = ""
		err := runCommandExpectingFailure(t, jobsRenameCmd, []string{parityID})
		if err == nil || !strings.Contains(err.Error(), "--name") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("jobs send-artefacts needs a recipient", func(t *testing.T) {
		previousEmails := append([]string(nil), jobsArtefactsEmails...)
		previousSubject := jobsArtefactsSubject
		t.Cleanup(func() {
			jobsArtefactsEmails = previousEmails
			jobsArtefactsSubject = previousSubject
		})
		jobsArtefactsEmails = nil
		jobsArtefactsSubject = ""
		err := runCommandExpectingFailure(t, jobsSendArtefactsCmd, []string{parityID})
		if err == nil || !strings.Contains(err.Error(), "--email") {
			t.Fatalf("error = %v", err)
		}
	})
}

func runCommandExpectingFailure(t *testing.T, command *cobra.Command, args []string) error {
	t.Helper()

	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetContext(context.Background())

	if err := command.RunE(command, args); err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	return nil
}

func restoreBulkWorkflowFlags(t *testing.T) {
	t.Helper()

	ids := append([]string(nil), bulkWorkflowIDs...)
	issues := append([]string(nil), bulkWorkflowIssueIDs...)
	assets := append([]string(nil), bulkWorkflowAssetIDs...)
	severities := append([]string(nil), bulkWorkflowSeverities...)
	profiles := append([]string(nil), bulkWorkflowProfileIDs...)
	tags := append([]string(nil), bulkWorkflowTagIDs...)
	kinds := append([]string(nil), bulkWorkflowScannerKinds...)
	all, dry := bulkWorkflowAll, bulkWorkflowDryRun
	t.Cleanup(func() {
		bulkWorkflowIDs = ids
		bulkWorkflowIssueIDs = issues
		bulkWorkflowAssetIDs = assets
		bulkWorkflowSeverities = severities
		bulkWorkflowProfileIDs = profiles
		bulkWorkflowTagIDs = tags
		bulkWorkflowScannerKinds = kinds
		bulkWorkflowAll = all
		bulkWorkflowDryRun = dry
	})
}

func clearBulkWorkflowFlags() {
	bulkWorkflowIDs = nil
	bulkWorkflowIssueIDs = nil
	bulkWorkflowAssetIDs = nil
	bulkWorkflowSeverities = nil
	bulkWorkflowProfileIDs = nil
	bulkWorkflowTagIDs = nil
	bulkWorkflowScannerKinds = nil
	bulkWorkflowAll = false
	bulkWorkflowDryRun = false
}

func restoreSecurityTestUpdateFlags(t *testing.T) {
	t.Helper()

	enabled, severity, reset := securityTestsUpdateEnabled, securityTestsUpdateSeverity, securityTestsUpdateResetSeverity
	t.Cleanup(func() {
		securityTestsUpdateEnabled = enabled
		securityTestsUpdateSeverity = severity
		securityTestsUpdateResetSeverity = reset
	})
}

func restoreSecurityTestBulkFlags(t *testing.T) {
	t.Helper()

	enabled, severity := securityTestsBulkEnabled, securityTestsBulkSeverity
	reset, all, dry := securityTestsBulkResetSeverity, securityTestsBulkAll, securityTestsBulkDryRun
	ids := append([]string(nil), securityTestsBulkIDs...)
	search := securityTestsBulkSearch
	whereEnabled := securityTestsBulkWhereEnabled
	kinds := append([]string(nil), securityTestsBulkScannerKinds...)
	categories := append([]string(nil), securityTestsBulkCategories...)
	compliances := append([]string(nil), securityTestsBulkCompliances...)
	assetTypes := append([]string(nil), securityTestsBulkAssetTypes...)
	t.Cleanup(func() {
		securityTestsBulkEnabled = enabled
		securityTestsBulkSeverity = severity
		securityTestsBulkResetSeverity = reset
		securityTestsBulkAll = all
		securityTestsBulkDryRun = dry
		securityTestsBulkIDs = ids
		securityTestsBulkSearch = search
		securityTestsBulkWhereEnabled = whereEnabled
		securityTestsBulkScannerKinds = kinds
		securityTestsBulkCategories = categories
		securityTestsBulkCompliances = compliances
		securityTestsBulkAssetTypes = assetTypes
	})
}

func clearSecurityTestBulkFlags() {
	securityTestsBulkEnabled = ""
	securityTestsBulkSeverity = ""
	securityTestsBulkResetSeverity = false
	securityTestsBulkAll = false
	securityTestsBulkDryRun = false
	securityTestsBulkIDs = nil
	securityTestsBulkSearch = ""
	securityTestsBulkWhereEnabled = ""
	securityTestsBulkScannerKinds = nil
	securityTestsBulkCategories = nil
	securityTestsBulkCompliances = nil
	securityTestsBulkAssetTypes = nil
}

func restoreIntegrationRepositoryFlags(t *testing.T) {
	t.Helper()

	search, integrationID := integrationsReposSearch, integrationsReposIntegrationID
	fullNames := append([]string(nil), integrationsReposFullNames...)
	page := integrationReposPage
	t.Cleanup(func() {
		integrationsReposSearch = search
		integrationsReposIntegrationID = integrationID
		integrationsReposFullNames = fullNames
		integrationReposPage = page
	})
	quietPageFlags(t, integrationsRepositoriesCmd)
}

func quietPageFlags(t *testing.T, command *cobra.Command) {
	t.Helper()

	for _, name := range []string{"size", "cursor"} {
		flag := command.Flags().Lookup(name)
		previous := flag.Value.String()
		changed := flag.Changed
		t.Cleanup(func() {
			_ = flag.Value.Set(previous)
			flag.Changed = changed
		})

		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatalf("reset %s: %v", name, err)
		}

		flag.Changed = false
	}
}
