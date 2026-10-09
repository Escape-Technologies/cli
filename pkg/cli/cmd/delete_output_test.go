package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
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

			commandFor(tt.name).SetContext(context.Background())

			stdout, stderr, err := captureJSONCommand(t, tt.run)
			if err != nil {
				t.Fatalf("run: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
			}

			if hits != 1 {
				t.Fatalf("API calls = %d, want 1", hits)
			}

			assertExactlyOneJSON(t, stdout, tt.dest())
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

func TestSchemaModeDoesNotDelete(t *testing.T) {
	if err := out.SetOutput("schema"); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	orig := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = orig })
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(io.Discard, reader)
		close(done)
	}()
	t.Cleanup(func() {
		_ = writer.Close()
		<-done
	})

	calls := []struct {
		name string
		fn   func() error
	}{
		{"projects delete", func() error { return projectsDeleteCmd.RunE(projectsDeleteCmd, []string{deleteOutputID}) }},
		{"roles delete", func() error { return rolesDeleteCmd.RunE(rolesDeleteCmd, []string{deleteOutputID}) }},
		{"integrations delete", func() error { return integrationsDeleteCmd.RunE(integrationsDeleteCmd, []string{deleteOutputID}) }},
	}
	for _, call := range calls {
		if err := call.fn(); err != nil {
			t.Errorf("%s in schema mode called the API or failed: %v", call.name, err)
		}
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
		"escape-cli users list",
		"escape-cli roles list",
		"escape-cli projects list",
		"escape-cli integrations list",
		"escape-cli workflows list",
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
		"escape-cli events list",
		"escape-cli emails list",
		"escape-cli locations list",
		"escape-cli assets list",
		"escape-cli assets list-activities",
		"escape-cli custom-rules list",
		"escape-cli tags list",
		"escape-cli audit list",
		"escape-cli retests list",
		"escape-cli problems",
		"escape-cli regression-tests list",
		"escape-cli regression-tests history",
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
		"escape-cli profiles create-rest",
		"escape-cli profiles create-webapp",
		"escape-cli profiles create-graphql",
		"escape-cli profiles create-ai-pentest",
		"escape-cli profiles update",
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
	)
	note("a dedicated contract test already decodes this command, or it needs more than one response, a file, or a summary it builds itself",
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
