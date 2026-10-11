package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// parityBatch2Case is one command of the write groups, driven against a test
// server that answers a single route. Every case pins the request the command
// sends and decodes the one document it prints.
type parityBatch2Case struct {
	name    string
	command *cobra.Command
	args    func() []string
	prepare func(t *testing.T)
	method  string
	path    string
	query   map[string]string
	body    func(t *testing.T, got map[string]any)
	respond func(w http.ResponseWriter)
}

func runParityBatch2Case(t *testing.T, tt parityBatch2Case) {
	t.Helper()

	var hits int
	var request map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.Method != tt.method || r.URL.Path != tt.path {
			t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tt.method, tt.path)
			http.NotFound(w, r)

			return
		}

		for name, want := range tt.query {
			if got := r.URL.Query().Get(name); got != want {
				t.Errorf("query %s = %q, want %q", name, got, want)
			}
		}

		if data, _ := io.ReadAll(r.Body); len(data) > 0 {
			if err := json.Unmarshal(data, &request); err != nil {
				t.Errorf("body is not JSON: %v", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		tt.respond(w)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	if tt.prepare != nil {
		tt.prepare(t)
	}

	tt.command.SetIn(strings.NewReader(""))
	tt.command.SetOut(io.Discard)
	tt.command.SetErr(io.Discard)
	tt.command.SetContext(context.Background())

	if err := tt.command.RunE(tt.command, tt.args()); err != nil {
		t.Fatalf("%s: %v", tt.name, err)
	}

	if hits != 1 {
		t.Fatalf("%s: API calls = %d, want 1", tt.name, hits)
	}

	if tt.body != nil {
		tt.body(t, request)
	}
}

const parityID = "00000000-0000-0000-0000-0000000000aa"

func writeParityJSON(w http.ResponseWriter, body string) {
	_, _ = io.WriteString(w, body)
}

// TestViewsDeleteprints its declared document and deletes once.
func TestViewsDeletePrintsItsDeclaredDocument(t *testing.T) {
	runParityBatch2Case(t, parityBatch2Case{
		name:    "views delete",
		command: viewsDeleteCmd,
		args:    func() []string { return []string{parityID} },
		method:  http.MethodDelete,
		path:    "/v3/views/" + parityID,
		respond: func(w http.ResponseWriter) { writeParityJSON(w, `{"message":"deleted"}`) },
	})
}

// TestViewsDeleteIsMCPDestructive pins the confirmation gate on the view
// delete tool, the same one the older deletes carry.
func TestViewsDeleteIsMCPDestructive(t *testing.T) {
	if !isDestructiveMCPCommand("escape-cli views delete") {
		t.Fatal("views delete must be in the MCP confirmation list")
	}
}

func TestVaultVariablesCommands(t *testing.T) {
	previousKey, previousType, previousContent, previousVault := vaultVariablesCreateKey, vaultVariableType, vaultVariableContent, vaultVariableVaultID
	t.Cleanup(func() {
		vaultVariablesCreateKey, vaultVariableType, vaultVariableContent, vaultVariableVaultID = previousKey, previousType, previousContent, previousVault
	})

	t.Run("create sends the write-only value once", func(t *testing.T) {
		vaultVariablesCreateKey = "API_TOKEN"
		vaultVariableType = "SECRET"
		vaultVariableContent = "tok_123"
		vaultVariableVaultID = ""

		runParityBatch2Case(t, parityBatch2Case{
			name:    "vault variables create",
			command: vaultVariablesCreateCmd,
			args:    func() []string { return nil },
			method:  http.MethodPost,
			path:    "/v3/vault-variables",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				for key, want := range map[string]any{
					"key":     "API_TOKEN",
					"type":    "SECRET",
					"content": "tok_123",
				} {
					if got[key] != want {
						t.Fatalf("%s = %#v, want %#v", key, got[key], want)
					}
				}

				if _, set := got["vaultId"]; set {
					t.Fatalf("vaultId must stay unset for a SECRET variable: %#v", got)
				}
			},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"id":"`+parityID+`","key":"API_TOKEN","type":"SECRET","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`)
			},
		})
	})

	t.Run("rotate sends the new value", func(t *testing.T) {
		vaultVariableContent = "tok_456"

		runParityBatch2Case(t, parityBatch2Case{
			name:    "vault variables rotate",
			command: vaultVariablesRotateCmd,
			args:    func() []string { return []string{parityID} },
			method:  http.MethodPost,
			path:    "/v3/vault-variables/" + parityID + "/rotate",
			body: func(t *testing.T, got map[string]any) {
				t.Helper()
				if got["content"] != "tok_456" {
					t.Fatalf("content = %#v", got["content"])
				}
			},
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"id":"`+parityID+`","key":"API_TOKEN","type":"SECRET","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`)
			},
		})
	})

	t.Run("delete is gated and destructive", func(t *testing.T) {
		if !isDestructiveMCPCommand("escape-cli vault variables delete") {
			t.Fatal("vault variables delete must be in the MCP confirmation list")
		}

		runParityBatch2Case(t, parityBatch2Case{
			name:    "vault variables delete",
			command: vaultVariablesDeleteCmd,
			args:    func() []string { return []string{parityID} },
			method:  http.MethodDelete,
			path:    "/v3/vault-variables/" + parityID,
			respond: func(w http.ResponseWriter) { writeParityJSON(w, `{"message":"deleted"}`) },
		})
	})
}

func TestVaultLinkSendsTheProviderCoordinates(t *testing.T) {
	previousIntegration, previousExternal := vaultLinkIntegrationID, vaultLinkExternalID
	t.Cleanup(func() { vaultLinkIntegrationID, vaultLinkExternalID = previousIntegration, previousExternal })
	vaultLinkIntegrationID = parityID
	vaultLinkExternalID = "prod/keyvault"

	runParityBatch2Case(t, parityBatch2Case{
		name:    "vault link",
		command: vaultLinkCmd,
		args:    func() []string { return nil },
		method:  http.MethodPost,
		path:    "/v3/vaults",
		body: func(t *testing.T, got map[string]any) {
			t.Helper()
			if got["integrationId"] != parityID || got["externalId"] != "prod/keyvault" {
				t.Fatalf("link body = %#v", got)
			}
		},
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"id":"`+parityID+`","name":"prod","type":"AZURE","externalId":"prod/keyvault","integrationId":"`+parityID+`","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`)
		},
	})
}

func TestKnowledgeFilesDeleteIsDestructiveAndDeletesOnce(t *testing.T) {
	if !isDestructiveMCPCommand("escape-cli knowledge-files delete") {
		t.Fatal("knowledge-files delete must be in the MCP confirmation list")
	}

	runParityBatch2Case(t, parityBatch2Case{
		name:    "knowledge-files delete",
		command: knowledgeFilesDeleteCmd,
		args:    func() []string { return []string{parityID} },
		method:  http.MethodDelete,
		path:    "/v3/knowledge-files/" + parityID,
		respond: func(w http.ResponseWriter) { writeParityJSON(w, `{"message":"deleted"}`) },
	})
}

func TestOrgSettingsGetReadsTheSettings(t *testing.T) {
	runParityBatch2Case(t, parityBatch2Case{
		name:    "org-settings get",
		command: orgSettingsGetCmd,
		args:    func() []string { return nil },
		method:  http.MethodGet,
		path:    "/v3/organization-settings",
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"slas":[{"severity":"HIGH","timeMs":86400000,"enabled":true}],"settings":[],"killSwitches":{"dast":true,"aiPentesting":false,"asm":true}}`)
		},
	})
}

func TestUsersDeactivateSendsTheFlagAndIsDestructive(t *testing.T) {
	if !isDestructiveMCPCommand("escape-cli users deactivate") {
		t.Fatal("users deactivate must be in the MCP confirmation list")
	}

	runParityBatch2Case(t, parityBatch2Case{
		name:    "users deactivate",
		command: usersDeactivateCmd,
		args:    func() []string { return []string{parityID} },
		method:  http.MethodPut,
		path:    "/v3/users/" + parityID,
		body: func(t *testing.T, got map[string]any) {
			t.Helper()
			if got["deactivated"] != true {
				t.Fatalf("deactivated = %#v, want true", got["deactivated"])
			}
		},
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"id":"`+parityID+`","email":"user@example.com","roleBindings":[],"createdAt":"2026-01-01T00:00:00Z","deactivatedAt":"2026-01-01T00:00:00Z"}`)
		},
	})
}

func TestWorkflowsPauseSendsThePauseFlag(t *testing.T) {
	previousResume := workflowsResume
	t.Cleanup(func() { workflowsResume = previousResume })

	runParityBatch2Case(t, parityBatch2Case{
		name:    "workflows pause",
		command: workflowsPauseCmd,
		args:    func() []string { return []string{parityID} },
		method:  http.MethodPut,
		path:    "/v3/workflows/" + parityID,
		body: func(t *testing.T, got map[string]any) {
			t.Helper()
			if got["paused"] != true {
				t.Fatalf("paused = %#v, want true", got["paused"])
			}
		},
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"id":"`+parityID+`","name":"nightly","trigger":"MANUAL","paused":true,"filter":null,"projects":[],"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","filters":[],"actions":[]}`)
		},
	})

	workflowsResume = true
	runParityBatch2Case(t, parityBatch2Case{
		name:    "workflows resume",
		command: workflowsPauseCmd,
		args:    func() []string { return []string{parityID} },
		method:  http.MethodPut,
		path:    "/v3/workflows/" + parityID,
		body: func(t *testing.T, got map[string]any) {
			t.Helper()
			if got["paused"] != false {
				t.Fatalf("paused = %#v, want false on --resume", got["paused"])
			}
		},
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"id":"`+parityID+`","name":"nightly","trigger":"MANUAL","paused":false,"filter":null,"projects":[],"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","filters":[],"actions":[]}`)
		},
	})
}

func TestLocationsTagsReplaceTheFullList(t *testing.T) {
	previousIDs, previousClear := locationsTagsIDs, locationsTagsClear
	t.Cleanup(func() { locationsTagsIDs, locationsTagsClear = previousIDs, previousClear })
	locationsTagsIDs = []string{"00000000-0000-0000-0000-000000000001"}
	locationsTagsClear = false

	runParityBatch2Case(t, parityBatch2Case{
		name:    "locations tags",
		command: locationsTagsCmd,
		args:    func() []string { return []string{parityID} },
		method:  http.MethodPut,
		path:    "/v3/locations/" + parityID + "/tags",
		body: func(t *testing.T, got map[string]any) {
			t.Helper()
			ids, _ := got["tagIds"].([]any)
			if len(ids) != 1 || ids[0] != "00000000-0000-0000-0000-000000000001" {
				t.Fatalf("tagIds = %#v", got["tagIds"])
			}
		},
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"id":"`+parityID+`","name":"prod-vpc","type":"PRIVATE","enabled":true,"tags":[],"links":{"locationOverview":"https://app.escape.tech/locations"}}`)
		},
	})
}

func TestIntegrationsActionsRequireAKind(t *testing.T) {
	for _, command := range []*cobra.Command{integrationsTestCmd, integrationsSyncCmd, integrationsRunsCmd} {
		previousKind := integrationsKind
		t.Cleanup(func() { integrationsKind = previousKind })
		integrationsKind = ""

		command.SetIn(strings.NewReader(""))
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		command.SetContext(context.Background())

		if err := command.RunE(command, []string{parityID}); err == nil || !strings.Contains(err.Error(), "--kind is required") {
			t.Fatalf("%s without --kind: err = %v, want --kind is required", command.Name(), err)
		}
	}
}

func TestIntegrationsTestAndSyncScheduleARun(t *testing.T) {
	previousKind := integrationsKind
	t.Cleanup(func() { integrationsKind = previousKind })
	integrationsKind = "jira"

	for _, command := range []*cobra.Command{integrationsTestCmd, integrationsSyncCmd} {
		runParityBatch2Case(t, parityBatch2Case{
			name:    command.Name(),
			command: command,
			args:    func() []string { return []string{parityID} },
			method:  http.MethodPost,
			path:    "/v3/integrations/jira/" + parityID + "/" + command.Name(),
			respond: func(w http.ResponseWriter) {
				writeParityJSON(w, `{"id":"run-1","kind":"JIRA","status":"STARTING","createdAt":"2026-01-01T00:00:00Z","duration":0,"progressRatio":0}`)
			},
		})
	}
}

// TestIntegrationsRunsListsOnePage pins the kind-generic runs route: the CLI
// must call the run listing of the kind it was given, not a fixed one.
func TestIntegrationsRunsListsOnePage(t *testing.T) {
	previousKind := integrationsKind
	t.Cleanup(func() { integrationsKind = previousKind })
	integrationsKind = "jira"

	command := integrationsRunsCmd
	resetFlag(t, command, "size")
	if err := command.Flags().Set("size", "10"); err != nil {
		t.Fatal(err)
	}

	runParityBatch2Case(t, parityBatch2Case{
		name:    "integrations runs",
		command: command,
		args:    func() []string { return []string{parityID} },
		method:  http.MethodGet,
		path:    "/v3/integrations/jira/" + parityID + "/runs",
		query:   map[string]string{"size": "10"},
		respond: func(w http.ResponseWriter) {
			writeParityJSON(w, `{"data":[{"id":"run-1","kind":"JIRA","status":"FINISHED","createdAt":"2026-01-01T00:00:00Z","progressRatio":1}],"nextCursor":null,"totalCount":1}`)
		},
	})
}

// TestBetaMarkersCoverTheFlaggedGroups keeps the Beta badge aligned with the
// server-side beta policy: the vault and the knowledge library are the two
// flagged surfaces this batch exposes.
func TestBetaMarkersCoverTheFlaggedGroups(t *testing.T) {
	for _, path := range []string{
		"escape-cli vault variables list",
		"escape-cli vault variables create",
		"escape-cli vault variables delete",
		"escape-cli vault variables rotate",
		"escape-cli vault list",
		"escape-cli vault link",
		"escape-cli vault entries",
		"escape-cli knowledge-files list",
		"escape-cli knowledge-files upload",
		"escape-cli knowledge-files delete",
	} {
		if betaMCPToolNoteFor(path) == "" {
			t.Errorf("%s wraps a Beta operation and must carry the marker", path)
		}
	}

	for _, path := range []string{
		"escape-cli views list",
		"escape-cli org-settings get",
		"escape-cli users deactivate",
		"escape-cli workflows pause",
		"escape-cli locations logs",
	} {
		if betaMCPToolNoteFor(path) != "" {
			t.Errorf("%s is not a Beta operation", path)
		}
	}
}

// TestNewWriteCommandsAreRegistered makes sure every new leaf command is an
// MCP tool with schemas, or is deliberately absent from the registry.
func TestNewWriteCommandsAreRegistered(t *testing.T) {
	registry := CommandSchemaRegistry()
	for _, path := range []string{
		"escape-cli views list",
		"escape-cli views issues",
		"escape-cli views create",
		"escape-cli views update",
		"escape-cli views reorder",
		"escape-cli views delete",
		"escape-cli vault variables list",
		"escape-cli vault variables create",
		"escape-cli vault variables delete",
		"escape-cli vault variables rotate",
		"escape-cli vault list",
		"escape-cli vault link",
		"escape-cli vault entries",
		"escape-cli knowledge-files list",
		"escape-cli knowledge-files delete",
		"escape-cli knowledge-files upload",
		"escape-cli org-settings get",
		"escape-cli org-settings update",
		"escape-cli users deactivate",
		"escape-cli workflows pause",
		"escape-cli workflows executions",
		"escape-cli locations logs",
		"escape-cli locations tags",
		"escape-cli integrations test",
		"escape-cli integrations sync",
		"escape-cli integrations runs",
	} {
		if _, ok := registry[path]; !ok {
			t.Errorf("%s is not in CommandSchemaRegistry", path)
		}
	}
}

// TestVaultVariablesCreateRejectsAnExternalWithoutAVault keeps the flag
// contract the API documents: vaultId is required for EXTERNAL and forbidden
// otherwise.
func TestVaultVariablesCreateRejectsAnExternalWithoutAVault(t *testing.T) {
	previousKey, previousType, previousContent, previousVault := vaultVariablesCreateKey, vaultVariableType, vaultVariableContent, vaultVariableVaultID
	t.Cleanup(func() {
		vaultVariablesCreateKey, vaultVariableType, vaultVariableContent, vaultVariableVaultID = previousKey, previousType, previousContent, previousVault
	})
	vaultVariablesCreateKey = "DB_PASSWORD"
	vaultVariableType = "EXTERNAL"
	vaultVariableContent = "production/db"
	vaultVariableVaultID = ""

	command := vaultVariablesCreateCmd
	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetContext(context.Background())

	if err := command.RunE(command, nil); err == nil || !strings.Contains(err.Error(), "--vault-id is required") {
		t.Fatalf("err = %v, want the vault-id requirement", err)
	}
}
