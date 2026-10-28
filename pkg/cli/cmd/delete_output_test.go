package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"
)

const deleteOutputID = "00000000-0000-0000-0000-0000000000aa"

func TestDestructiveDeletesPrintTheirDeclaredDocument(t *testing.T) {
	const workflowBody = `{
		"id":"` + deleteOutputID + `",
		"name":"nightly",
		"trigger":"MANUAL",
		"paused":false,
		"filter":null,
		"projects":[],
		"createdAt":"2026-01-01T00:00:00Z",
		"updatedAt":"2026-01-01T00:00:00Z",
		"filters":[],
		"actions":[]
	}`
	const integrationBody = `{
		"id":"` + deleteOutputID + `",
		"name":"jira",
		"createdAt":"2026-01-01T00:00:00Z",
		"updatedAt":"2026-01-01T00:00:00Z",
		"kind":"JIRA",
		"valid":true,
		"validationErrors":[],
		"organizationId":"00000000-0000-0000-0000-000000000002",
		"projects":[],
		"tags":[]
	}`

	tests := []struct {
		name    string
		method  string
		path    string
		body    string
		prepare func(t *testing.T)
		run     func() error
		dest    func() any
		// wantMsg is set when the command prints {"msg":...}. {"message":...}
		// also unmarshals into out.Message, leaving Msg empty, so the text
		// has to be checked.
		wantMsg string
	}{
		{
			name:   "assets delete",
			method: http.MethodDelete,
			path:   "/v3/assets/" + deleteOutputID,
			body:   `{"message":"deleted"}`,
			run: func() error {
				return assetDeleteCmd.RunE(assetDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.DeleteProfile200Response{} },
		},
		{
			// The live route answers a bare true instead of its declared
			// {"message": string}. The CLI still prints the declared object.
			name:   "assets delete bare true",
			method: http.MethodDelete,
			path:   "/v3/assets/" + deleteOutputID,
			body:   `true`,
			run: func() error {
				return assetDeleteCmd.RunE(assetDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.DeleteProfile200Response{} },
		},
		{
			name:   "assets bulk-update",
			method: http.MethodPost,
			path:   "/v3/assets/bulk-update",
			body:   `{"success":true}`,
			prepare: func(t *testing.T) {
				restoreBulkAssetFlags(t)
				bulkAssetIDs = []string{deleteOutputID}
			},
			run: func() error {
				return assetBulkUpdateCmd.RunE(assetBulkUpdateCmd, nil)
			},
			dest: func() any { return &escape.BulkOperationResult{} },
		},
		{
			name:   "assets bulk-delete",
			method: http.MethodPost,
			path:   "/v3/assets/bulk-delete",
			body:   `{"success":true}`,
			prepare: func(t *testing.T) {
				restoreBulkAssetFlags(t)
				bulkAssetIDs = []string{deleteOutputID}
			},
			run: func() error {
				return assetBulkDeleteCmd.RunE(assetBulkDeleteCmd, nil)
			},
			dest: func() any { return &escape.BulkOperationResult{} },
		},
		{
			name:   "tags delete",
			method: http.MethodDelete,
			path:   "/v3/tags/" + deleteOutputID,
			body:   `{"message":"deleted"}`,
			run: func() error {
				return tagsDeleteCmd.RunE(tagsDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.DeleteProfile200Response{} },
		},
		{
			name:   "profiles delete",
			method: http.MethodDelete,
			path:   "/v3/profiles/" + deleteOutputID,
			body:   `{"message":"deleted"}`,
			run: func() error {
				return profileDeleteCmd.RunE(profileDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.DeleteProfile200Response{} },
		},
		{
			name:   "locations delete",
			method: http.MethodDelete,
			path:   "/v3/locations/" + deleteOutputID,
			body:   `{"message":"Location deleted successfully"}`,
			run: func() error {
				return locationsDeleteCmd.RunE(locationsDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.DeleteLocation200Response{} },
		},
		{
			name:   "workflows delete",
			method: http.MethodDelete,
			path:   "/v3/workflows/" + deleteOutputID,
			body:   workflowBody,
			run: func() error {
				return workflowsDeleteCmd.RunE(workflowsDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.CreateWorkflow200Response{} },
		},
		{
			name:   "custom-rules delete",
			method: http.MethodDelete,
			path:   "/v3/custom-rules/" + deleteOutputID,
			body:   `{"deleted":true}`,
			run: func() error {
				return customRulesDeleteCmd.RunE(customRulesDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.DeleteCustomRule200Response{} },
		},
		{
			name:   "integrations delete",
			method: http.MethodDelete,
			path:   "/v3/integrations/jira/" + deleteOutputID,
			body:   integrationBody,
			prepare: func(t *testing.T) {
				prev := integrationsKind
				t.Cleanup(func() {
					integrationsKind = prev
					flag := integrationsDeleteCmd.Flags().Lookup("kind")
					_ = flag.Value.Set(flag.DefValue)
					flag.Changed = false
				})
				if err := integrationsDeleteCmd.Flags().Set("kind", "jira"); err != nil {
					t.Fatal(err)
				}
			},
			run: func() error {
				return integrationsDeleteCmd.RunE(integrationsDeleteCmd, []string{deleteOutputID})
			},
			dest: func() any { return &v3.CreateakamaiIntegration200Response{} },
		},
		{
			name:   "views delete",
			method: http.MethodDelete,
			path:   "/v3/views/" + deleteOutputID,
			body:   `{"message":"deleted"}`,
			prepare: func(t *testing.T) {
				resetFlag(t, viewsDeleteCmd, "dry-run")
			},
			run: func() error {
				return viewsDeleteCmd.RunE(viewsDeleteCmd, []string{deleteOutputID})
			},
			dest:    func() any { return &out.Message{} },
			wantMsg: "View " + deleteOutputID + " deleted",
		},
		{
			name:   "views reorder",
			method: http.MethodPost,
			path:   "/v3/views/reorder",
			body:   `{"success":true}`,
			run: func() error {
				// The harness clears stdin after prepare. Set the body here.
				viewsReorderCmd.SetIn(strings.NewReader(`{"target":"ISSUES","order":["` + deleteOutputID + `"]}`))

				return viewsReorderCmd.RunE(viewsReorderCmd, nil)
			},
			dest:    func() any { return &out.Message{} },
			wantMsg: "Views reordered",
		},
		{
			name:   "vault variables delete",
			method: http.MethodDelete,
			path:   "/v3/vault-variables/" + deleteOutputID,
			body:   `{"message":"deleted"}`,
			prepare: func(t *testing.T) {
				resetFlag(t, vaultVariablesDeleteCmd, "dry-run")
			},
			run: func() error {
				return vaultVariablesDeleteCmd.RunE(vaultVariablesDeleteCmd, []string{deleteOutputID})
			},
			dest:    func() any { return &out.Message{} },
			wantMsg: "Vault variable " + deleteOutputID + " deleted",
		},
		{
			name:   "knowledge-files delete",
			method: http.MethodDelete,
			path:   "/v3/knowledge-files/" + deleteOutputID,
			body:   `{"message":"deleted"}`,
			prepare: func(t *testing.T) {
				resetFlag(t, knowledgeFilesDeleteCmd, "dry-run")
			},
			run: func() error {
				return knowledgeFilesDeleteCmd.RunE(knowledgeFilesDeleteCmd, []string{deleteOutputID})
			},
			dest:    func() any { return &out.Message{} },
			wantMsg: "Knowledge file " + deleteOutputID + " deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hits int
			serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
				hits++
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tt.method, tt.path)
					http.NotFound(w, r)

					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, tt.body)
			})
			if tt.prepare != nil {
				tt.prepare(t)
			}

			// A terminal stdin would trigger the confirmation prompt; the
			// prompt itself is covered by confirm_test.go.
			commandFor(tt.name).SetIn(strings.NewReader(""))
			commandFor(tt.name).SetContext(context.Background())

			stdout, stderr, err := captureJSONCommand(t, tt.run)
			if err != nil {
				t.Fatalf("run: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			if hits != 1 {
				t.Fatalf("API calls = %d, want 1", hits)
			}

			decoded := tt.dest()
			assertExactlyOneJSON(t, stdout, decoded)
			if tt.wantMsg != "" {
				msg, ok := decoded.(*out.Message)
				if !ok || msg.Msg != tt.wantMsg {
					t.Fatalf("msg = %#v, want %q", decoded, tt.wantMsg)
				}
			}
		})
	}
}

func commandFor(name string) *cobra.Command {
	switch name {
	case "assets delete", "assets delete bare true":
		return assetDeleteCmd
	case "assets bulk-update":
		return assetBulkUpdateCmd
	case "assets bulk-delete":
		return assetBulkDeleteCmd
	case "tags delete":
		return tagsDeleteCmd
	case "profiles delete":
		return profileDeleteCmd
	case "locations delete":
		return locationsDeleteCmd
	case "workflows delete":
		return workflowsDeleteCmd
	case "custom-rules delete":
		return customRulesDeleteCmd
	case "views delete":
		return viewsDeleteCmd
	case "views reorder":
		return viewsReorderCmd
	case "vault variables delete":
		return vaultVariablesDeleteCmd
	case "knowledge-files delete":
		return knowledgeFilesDeleteCmd
	default:
		return integrationsDeleteCmd
	}
}

func restoreBulkAssetFlags(t *testing.T) {
	t.Helper()
	prevIDs := bulkAssetIDs
	prevTypes := bulkAssetTypes
	prevStatuses := bulkAssetStatuses
	prevTags := bulkAssetTagIDs
	prevProjects := bulkAssetProjIDs
	prevStatus := bulkAssetStatus
	t.Cleanup(func() {
		bulkAssetIDs = prevIDs
		bulkAssetTypes = prevTypes
		bulkAssetStatuses = prevStatuses
		bulkAssetTagIDs = prevTags
		bulkAssetProjIDs = prevProjects
		bulkAssetStatus = prevStatus
	})
	bulkAssetIDs = nil
	bulkAssetTypes = nil
	bulkAssetStatuses = nil
	bulkAssetTagIDs = nil
	bulkAssetProjIDs = nil
	bulkAssetStatus = ""
}

func TestViewsDeleteDryRunPrintsMsg(t *testing.T) {
	resetFlag(t, viewsDeleteCmd, "dry-run")
	if err := viewsDeleteCmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}

	viewsDeleteCmd.SetIn(strings.NewReader(""))
	viewsDeleteCmd.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return viewsDeleteCmd.RunE(viewsDeleteCmd, []string{deleteOutputID})
	})
	if err != nil {
		t.Fatalf("dry run: %v\nstderr: %s", err, stderr)
	}

	var msg out.Message
	assertExactlyOneJSON(t, stdout, &msg)
	if !strings.HasPrefix(msg.Msg, "Dry run:") || msg.Msg == "" {
		t.Fatalf("msg = %q, want a Dry run status", msg.Msg)
	}
}

func assertExactlyOneJSON(t *testing.T, stdout string, dest any) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	if err := dec.Decode(dest); err != nil {
		t.Fatalf("stdout is not one JSON document of the declared type: %v\n%s", err, stdout)
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout has more than one JSON document (%v): %s", err, stdout)
	}
}

// schemaModeCommandTimeout bounds one command exercised with -o schema.
// A command that skips out.Schema and then watches or retries must fail the
// test instead of hanging it. 1s flakes on a loaded CI runner.
const schemaModeCommandTimeout = 5 * time.Second //nolint:mnd

// schemaModeMCPServe is not executed: it binds a port and blocks.
const schemaModeMCPServe = "escape-cli mcp serve"

// schemaModeUtilityCommands do not advertise an output schema. The walk still
// checks that they neither call the API nor read stdin, and does not require
// a printed schema or a nil error. mcp serve is listed and is not run.
var schemaModeUtilityCommands = []string{
	"escape-cli",
	"escape-cli help-all",
	"escape-cli version",
	schemaModeMCPServe,
}

// schemaModeOriginalCommands are the commands the schema-mode tests named.
// They stay under the strict check. projects delete, roles delete, and
// upload schema are not MCP tools, so the registry check alone would drop them.
var schemaModeOriginalCommands = []string{
	"escape-cli projects delete",
	"escape-cli roles delete",
	"escape-cli integrations delete",
	"escape-cli scans cancel",
	"escape-cli scans ignore",
	"escape-cli upload schema",
	"escape-cli validations start",
	"escape-cli ai-chats create",
	"escape-cli ai-chats message",
	"escape-cli issues bulk-workflow",
	"escape-cli profiles bulk-update",
	"escape-cli security-tests update",
	"escape-cli security-tests bulk-update",
	"escape-cli compliance enable",
	"escape-cli jobs rename",
	"escape-cli jobs send-artefacts",
	"escape-cli org-settings scanner-configuration create",
	"escape-cli integrations repositories",
	"escape-cli integrations jira-projects",
	"escape-cli integrations jira-issue-types",
	"escape-cli integrations azure-areas",
	"escape-cli scans scenarios",
	"escape-cli issues export",
	"escape-cli assets export",
}

var (
	errSchemaModePanic   = errors.New("panic in schema mode")
	errSchemaModeTimeout = errors.New("schema mode did not return")
)

// stdinReadTracker counts bytes read from the reader passed to cmd.SetIn.
// Commands that read os.Stdin directly are caught by the file offset instead.
type stdinReadTracker struct {
	reader io.Reader
	read   int64
}

func (t *stdinReadTracker) Read(p []byte) (int, error) {
	n, err := t.reader.Read(p)
	t.read += int64(n)

	// io.Reader callers compare this error to io.EOF directly, so it stays unwrapped.
	return n, err //nolint:wrapcheck
}

func schemaModeUtility(path string) bool {
	return slices.Contains(schemaModeUtilityCommands, path)
}

// schemaModeStdin is a body every stdin command can parse far enough to reach
// its API call. Schema mode has to return before that read.
const schemaModeStdin = `{"configuration":{"mode":"read_only"},"name":"schema","stdinProbe":"body"}`

// schemaModeArgs satisfies commands that index positional arguments before
// returning. RunE is invoked directly, so cobra's Args validator does not run.
var schemaModeArgs = []string{deleteOutputID, deleteOutputID, deleteOutputID, deleteOutputID}

// TestSchemaModeRunsNoMutation runs every command with `--output schema`.
// A registered command, and each command the original test named, must return
// nil and print a schema before any API call or stdin read. Utility commands
// are the named exception: they have no output schema.
func TestSchemaModeRunsNoMutation(t *testing.T) {
	if err := out.SetOutput("schema"); err != nil {
		t.Fatal(err)
	}

	out.SetInputSchema(false)
	t.Cleanup(func() {
		_ = out.SetOutput("pretty")
		out.SetInputSchema(false)
	})

	var requestsMu sync.Mutex
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsMu.Lock()
		requests = append(requests, r.Method+" "+r.URL.Path)
		requestsMu.Unlock()
		http.Error(w, "schema mode must not call the API", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", deleteOutputID)

	stdin, err := os.CreateTemp(t.TempDir(), "schema-stdin")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := stdin.WriteString(schemaModeStdin); err != nil {
		t.Fatal(err)
	}

	if _, err := stdin.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	origStdin := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() { os.Stdin = origStdin })

	if !commandTreeHas(rootCmd, schemaModeMCPServe) {
		t.Fatalf("schema mode skip lost %s", schemaModeMCPServe)
	}

	commands := commandsExercisedInSchemaMode(rootCmd)
	seen := make(map[string]struct{}, len(commands))
	for _, command := range commands {
		seen[command.CommandPath()] = struct{}{}
	}

	for path := range CommandSchemaRegistry() {
		if _, ok := seen[path]; !ok {
			t.Errorf("schema mode walk missed registered command %s", path)
		}
	}

	for _, path := range schemaModeOriginalCommands {
		if _, ok := seen[path]; !ok {
			t.Errorf("schema mode walk missed original command %s", path)
		}
	}

	for _, command := range commands {
		path := command.CommandPath()
		if _, err := stdin.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		before := len(recordedRequests(&requestsMu, &requests))
		stdout, stdinReads, err := runCommandInSchemaMode(command)
		got := recordedRequests(&requestsMu, &requests)
		if len(got) > before {
			t.Errorf("%s in schema mode called the API: %s", path, strings.Join(got[before:], ", "))
		}

		offset, seekErr := stdin.Seek(0, io.SeekCurrent)
		if seekErr != nil {
			t.Fatal(seekErr)
		}

		if offset != 0 {
			t.Errorf("%s in schema mode read os.Stdin (%d bytes)", path, offset)
		}

		if stdinReads != 0 {
			t.Errorf("%s in schema mode read stdin (%d bytes)", path, stdinReads)
		}

		if schemaModeUtility(path) {
			if errors.Is(err, errSchemaModePanic) || errors.Is(err, errSchemaModeTimeout) {
				t.Errorf("%s in schema mode: %v", path, err)
			}

			continue
		}

		if err != nil {
			t.Errorf("%s in schema mode: %v", path, err)
		}

		if !strings.Contains(stdout, `"$schema"`) {
			t.Errorf("%s in schema mode did not print a schema: %q", path, stdout)
		}
	}
}

// commandsExercisedInSchemaMode returns every command that can run, including
// group commands whose RunE still calls the API (for example `statistics`).
// mcp serve is omitted: it binds a port and blocks, and it does not call the API.
func commandsExercisedInSchemaMode(root *cobra.Command) []*cobra.Command {
	var found []*cobra.Command
	var walk func(*cobra.Command)
	walk = func(command *cobra.Command) {
		if command.CommandPath() == schemaModeMCPServe {
			return
		}

		if command.RunE != nil || command.Run != nil {
			found = append(found, command)
		}

		for _, child := range command.Commands() {
			walk(child)
		}
	}
	walk(root)
	sort.Slice(found, func(i, j int) bool {
		return found[i].CommandPath() < found[j].CommandPath()
	})

	return found
}

func commandTreeHas(root *cobra.Command, path string) bool {
	if root.CommandPath() == path {
		return true
	}

	for _, child := range root.Commands() {
		if commandTreeHas(child, path) {
			return true
		}
	}

	return false
}

func recordedRequests(mu *sync.Mutex, requests *[]string) []string {
	mu.Lock()
	defer mu.Unlock()

	return append([]string(nil), (*requests)...)
}

func runCommandInSchemaMode(command *cobra.Command) (string, int64, error) {
	previousOut := command.OutOrStdout()
	previousErr := command.ErrOrStderr()
	previousIn := command.InOrStdin()
	previousCtx := command.Context()
	tracker := &stdinReadTracker{reader: strings.NewReader(schemaModeStdin)}
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetIn(tracker)
	ctx, cancel := context.WithCancel(context.Background())
	command.SetContext(ctx)
	defer func() {
		cancel()
		command.SetOut(previousOut)
		command.SetErr(previousErr)
		command.SetIn(previousIn)
		command.SetContext(previousCtx)
	}()

	reader, writer, err := os.Pipe()
	if err != nil {
		return "", 0, fmt.Errorf("stdout pipe: %w", err)
	}

	captured := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(reader)
		captured <- data
	}()

	origStdout := os.Stdout
	os.Stdout = writer
	restored := false
	restoreStdout := func() {
		if restored {
			return
		}

		restored = true
		os.Stdout = origStdout
		_ = writer.Close()
	}
	defer restoreStdout()

	finished := make(chan error, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				finished <- fmt.Errorf("%w: %v", errSchemaModePanic, recovered)
			}
		}()

		if command.RunE != nil {
			finished <- command.RunE(command, schemaModeArgs)

			return
		}

		command.Run(command, schemaModeArgs)
		finished <- nil
	}()

	var runErr error
	var reads int64
	select {
	case runErr = <-finished:
		reads = tracker.read
	case <-time.After(schemaModeCommandTimeout):
		runErr = fmt.Errorf("did not return within %s: %w", schemaModeCommandTimeout, errSchemaModeTimeout)
	}

	restoreStdout()

	return string(<-captured), reads, runErr
}

func TestRolesAndProjectsDeleteDryRunMatchDeclaredSchema(t *testing.T) {
	const roleBody = `{
		"id":"` + deleteOutputID + `",
		"name":"auditor",
		"permissions":["VIEW_ONLY"],
		"createdAt":"2026-01-01T00:00:00Z",
		"bindings":[]
	}`
	const projectBody = `{
		"id":"` + deleteOutputID + `",
		"name":"payments",
		"slug":"payments",
		"createdAt":"2026-01-01T00:00:00Z",
		"bindings":[]
	}`

	tests := []struct {
		name     string
		path     string
		body     string
		command  *cobra.Command
		declared any
	}{
		{
			name:     "roles delete",
			path:     "/v3/roles/" + deleteOutputID,
			body:     roleBody,
			command:  rolesDeleteCmd,
			declared: roleDeleteOutput,
		},
		{
			name:     "projects delete",
			path:     "/v3/projects/" + deleteOutputID,
			body:     projectBody,
			command:  projectsDeleteCmd,
			declared: projectDeleteOutput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, r.Method+" "+r.URL.Path)
				if r.Method != http.MethodGet || r.URL.Path != tt.path {
					http.NotFound(w, r)

					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, tt.body)
			})

			resetFlag(t, tt.command, "dry-run")
			if err := tt.command.Flags().Set("dry-run", "true"); err != nil {
				t.Fatal(err)
			}

			tt.command.SetIn(strings.NewReader(""))
			tt.command.SetContext(context.Background())

			stdout, stderr, err := captureJSONCommand(t, func() error {
				return tt.command.RunE(tt.command, []string{deleteOutputID})
			})
			if err != nil {
				t.Fatalf("dry run: %v\nstderr: %s", err, stderr)
			}

			if len(calls) != 1 || calls[0] != http.MethodGet+" "+tt.path {
				t.Fatalf("calls = %v, want one GET %s", calls, tt.path)
			}

			var document map[string]any
			assertExactlyOneJSON(t, stdout, &document)
			if document["id"] != deleteOutputID {
				t.Fatalf("dry-run document = %s", stdout)
			}

			if _, ok := document["msg"]; ok {
				t.Fatalf("dry-run document changed to a message: %s", stdout)
			}

			assertMatchesDeclaredSchema(t, tt.declared, stdout)
			assertMatchesDeclaredSchema(t, tt.declared, `{"msg":"Dry run: would delete"}`)
		})
	}
}

func assertMatchesDeclaredSchema(t *testing.T, declared any, document string) {
	t.Helper()

	rawSchema, err := json.Marshal(clischema.Generate(declared))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	var schemaDoc any
	if err := json.Unmarshal(rawSchema, &schemaDoc); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	compiler := jsonschema.NewCompiler()
	const schemaURL = "https://escape.tech/cli/output.schema.json"
	if err := compiler.AddResource(schemaURL, schemaDoc); err != nil {
		t.Fatalf("add schema: %v", err)
	}

	compiled, err := compiler.Compile(schemaURL)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}

	var instance any
	if err := json.Unmarshal([]byte(document), &instance); err != nil {
		t.Fatalf("unmarshal document: %v\n%s", err, document)
	}

	if err := compiled.Validate(instance); err != nil {
		t.Fatalf("document does not match declared schema: %v\n%s", err, document)
	}
}

// outputSchemaNotExercised lists MCP tools whose output schema cannot be
// checked by one httptest response. A tool with an output schema must appear
// here or in TestDestructiveDeletesPrintTheirDeclaredDocument. The reasons
// are why a single fixture does not drive the command.
func outputSchemaNotExercised() map[string]string {
	reasons := map[string]string{}
	note := func(reason string, paths ...string) {
		for _, path := range paths {
			reasons[path] = reason
		}
	}
	note("prints a page or array after unmarshalling a different wire document, often across more than one request",
		"escape-cli security-tests list",
		"escape-cli ai-chats list",
		"escape-cli ai-chats messages",
		"escape-cli jobs list",
		"escape-cli users list",
		"escape-cli roles list",
		"escape-cli projects list",
		"escape-cli integrations list",
		"escape-cli integrations runs",
		"escape-cli workflows list",
		"escape-cli workflows executions",
		"escape-cli profiles list",
		"escape-cli profiles problems",
		"escape-cli issues list",
		"escape-cli issues list-activities",
		"escape-cli issues funnel",
		"escape-cli issues trends",
		"escape-cli scans list",
		"escape-cli scans issues",
		"escape-cli scans targets",
		"escape-cli scans problems",
		"escape-cli scans scenarios",
		"escape-cli events list",
		"escape-cli emails list",
		"escape-cli locations list",
		"escape-cli locations logs",
		"escape-cli assets list",
		"escape-cli assets list-activities",
		"escape-cli custom-rules list",
		"escape-cli tags list",
		"escape-cli audit list",
		"escape-cli retests list",
		"escape-cli problems",
		"escape-cli regression-tests list",
		"escape-cli regression-tests history",
		"escape-cli knowledge-files list",
		"escape-cli knowledge-files upload",
		"escape-cli views list",
		"escape-cli views issues",
		"escape-cli vault variables list",
		"escape-cli integrations repositories",
		"escape-cli knowledge-files list",
	)
	note("needs a valid stdin body; TestAdvertisedBodiesReachTheAPI checks the body reaches the API",
		"escape-cli roles create",
		"escape-cli roles update",
		"escape-cli projects create",
		"escape-cli projects update",
		"escape-cli integrations create",
		"escape-cli integrations update",
		"escape-cli workflows create",
		"escape-cli workflows update",
		"escape-cli views create",
		"escape-cli views update",
		"escape-cli org-settings update",
		"escape-cli profiles create-rest",
		"escape-cli profiles create-webapp",
		"escape-cli profiles create-graphql",
		"escape-cli profiles create-ai-pentest",
		"escape-cli profiles update",
		"escape-cli profiles update-configuration",
		"escape-cli assets create",
		"escape-cli assets bulk-import",
		"escape-cli custom-rules create",
		"escape-cli custom-rules update",
		"escape-cli issues comment",
		"escape-cli issues remediation generate",
		"escape-cli issues remediation feedback",
		"escape-cli retests start",
		"escape-cli authentications start",
		"escape-cli regression-tests create",
		"escape-cli regression-tests update",
		"escape-cli regression-tests answer",
		"escape-cli validations start",
		"escape-cli issues bulk-workflow",
		"escape-cli profiles bulk-update",
		"escape-cli org-settings scanner-configuration create",
	)
	note("a dedicated contract test already decodes this command, or it needs more than one response, a file, or a summary it builds itself",
		"escape-cli ai-chats create",
		"escape-cli security-tests get",
		"escape-cli compliance list",
		"escape-cli compliance enabled",
		"escape-cli compliance per-application",
		"escape-cli scans compliance",
		"escape-cli statistics endpoints",
		"escape-cli statistics issues",
		"escape-cli statistics trends",
		"escape-cli permissions list",
		"escape-cli validations get",
		"escape-cli ai-chats get",
		"escape-cli ai-chats message",
		"escape-cli me",
		"escape-cli users me",
		"escape-cli users get",
		"escape-cli users invite",
		"escape-cli roles get",
		"escape-cli roles bind",
		"escape-cli roles unbind",
		"escape-cli projects get",
		"escape-cli integrations get",
		"escape-cli workflows get",
		"escape-cli profiles get",
		"escape-cli profiles get-schema",
		"escape-cli issues get",
		"escape-cli issues get-with-events",
		"escape-cli issues remediation get",
		"escape-cli issues update",
		"escape-cli issues bulk-update",
		"escape-cli issues notify",
		"escape-cli issues trigger-workflow",
		"escape-cli scans get",
		"escape-cli scans start",
		"escape-cli scans cancel",
		"escape-cli scans ignore",
		"escape-cli scans coverage",
		"escape-cli scans reasoning",
		"escape-cli scans agents",
		"escape-cli events get",
		"escape-cli emails read",
		"escape-cli authentications get",
		"escape-cli jobs trigger-export",
		"escape-cli jobs get",
		"escape-cli issues export",
		"escape-cli assets export",
		"escape-cli locations get",
		"escape-cli locations create",
		"escape-cli locations update",
		"escape-cli assets get",
		"escape-cli assets update",
		"escape-cli assets comment",
		"escape-cli asm trigger",
		"escape-cli custom-rules get",
		"escape-cli tags get",
		"escape-cli tags create",
		"escape-cli tags update",
		"escape-cli stats",
		"escape-cli retests get",
		"escape-cli regression-tests get",
		"escape-cli regression-tests delete",
		"escape-cli regression-tests run",
		"escape-cli regression-tests stop",
		"escape-cli security-tests update",
		"escape-cli security-tests bulk-update",
		"escape-cli compliance enable",
		"escape-cli jobs rename",
		"escape-cli jobs send-artefacts",
		"escape-cli integrations jira-projects",
		"escape-cli integrations jira-issue-types",
		"escape-cli integrations azure-areas",
	)

	return reasons
}

func TestEveryOutputSchemaToolIsExercisedOrExplained(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	exercised := map[string]struct{}{
		"escape-cli assets delete":                       {},
		"escape-cli assets bulk-update":                  {},
		"escape-cli assets bulk-delete":                  {},
		"escape-cli tags delete":                         {},
		"escape-cli profiles delete":                     {},
		"escape-cli locations delete":                    {},
		"escape-cli workflows delete":                    {},
		"escape-cli custom-rules delete":                 {},
		"escape-cli integrations delete":                 {},
		"escape-cli scans configuration":                 {},
		"escape-cli scans statistics":                    {},
		"escape-cli profiles continuous-pentest create":  {},
		"escape-cli profiles continuous-pentest update":  {},
		"escape-cli profiles continuous-pentest enable":  {},
		"escape-cli profiles continuous-pentest disable": {},
		"escape-cli continuous-scans start":              {},
		"escape-cli views delete":                        {},
		"escape-cli views reorder":                       {},
		"escape-cli vault variables create":              {},
		"escape-cli vault variables update":              {},
		"escape-cli vault variables delete":              {},
		"escape-cli knowledge-files delete":              {},
		"escape-cli org-settings get":                    {},
		"escape-cli users deactivate":                    {},
		"escape-cli workflows pause":                     {},
		"escape-cli locations tags":                      {},
		"escape-cli integrations test":                   {},
		"escape-cli integrations sync":                   {},
	}
	explained := outputSchemaNotExercised()
	seen := map[string]struct{}{}
	var missing []string
	for _, spec := range specs {
		if len(spec.Tool.RawOutputSchema) == 0 {
			continue
		}

		seen[spec.Path] = struct{}{}
		if _, ok := exercised[spec.Path]; ok {
			continue
		}

		if _, ok := explained[spec.Path]; ok {
			continue
		}

		missing = append(missing, spec.Path)
	}

	if len(missing) > 0 {
		t.Fatalf("tools with an output schema are neither exercised nor explained: %s", strings.Join(missing, ", "))
	}

	for path := range exercised {
		if _, ok := seen[path]; !ok {
			t.Errorf("exercised path %s has no output schema", path)
		}
	}

	for path := range explained {
		if _, ok := seen[path]; !ok {
			t.Errorf("explained path %s has no output schema", path)
		}
	}
}
