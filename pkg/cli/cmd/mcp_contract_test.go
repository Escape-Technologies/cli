package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestMCPAdvertisedPropertiesAreConsumed(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	if len(specs) == 0 {
		t.Fatal("expected registered MCP tools")
	}

	commands := indexCommands(rootCmd)

	for _, spec := range specs {
		schema := decodeToolSchema(t, spec.Tool.RawInputSchema)
		properties, _ := schema["properties"].(map[string]any)
		bound := map[string]struct{}{}
		for _, name := range spec.PositionalArgs {
			bound[name] = struct{}{}
		}

		for _, binding := range spec.FlagBindings {
			bound[binding.Property] = struct{}{}
		}

		if spec.BodyProperty != "" {
			bound[spec.BodyProperty] = struct{}{}
		}

		// confirm is consumed by the destructive guard and is never a CLI flag.
		if spec.Destructive {
			bound[climcp.ConfirmProperty] = struct{}{}
		}

		for name := range properties {
			if _, ok := bound[name]; !ok {
				t.Errorf("%s advertises %q but the command does not consume it", spec.Name, name)
			}
		}

		if spec.BodyProperty == "" {
			continue
		}

		if commands[spec.Path] == nil {
			t.Errorf("%s has no cobra command", spec.Name)
		}
	}
}

func TestMCPInputSchemasHaveNoTopLevelCombinator(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	if len(specs) == 0 {
		t.Fatal("expected registered MCP tools")
	}

	for _, spec := range specs {
		assertNoTopLevelCombinator(t, spec.Name, spec.Tool.RawInputSchema)
		stub, err := climcp.BuildStubTool(spec)
		if err != nil {
			t.Errorf("%s stub: %v", spec.Name, err)
			continue
		}

		assertNoTopLevelCombinator(t, spec.Name+" stub", stub.RawInputSchema)
	}
}

func assertNoTopLevelCombinator(t *testing.T, name string, raw json.RawMessage) {
	t.Helper()
	schema := decodeToolSchema(t, raw)
	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		if _, ok := schema[key]; ok {
			t.Errorf("%s input schema has top-level %s", name, key)
		}
	}
}

func TestIssueCommentAcceptsAdvertisedBody(t *testing.T) {
	spec := mcpSpec(t, "issues_comment")
	schema := decodeToolSchema(t, spec.Tool.RawInputSchema)
	assertNoTopLevelCombinator(t, spec.Name, spec.Tool.RawInputSchema)
	if !strings.Contains(spec.Description, "message or body") {
		t.Fatalf("description %q does not say to pass message or body", spec.Description)
	}

	properties := schema["properties"].(map[string]any)
	body := properties["body"].(map[string]any)
	bodyProps := body["properties"].(map[string]any)
	if _, ok := bodyProps["comment"]; !ok {
		t.Fatalf("issues_comment body schema = %#v, want comment", bodyProps)
	}

	stub, err := climcp.BuildStubTool(spec)
	if err != nil {
		t.Fatalf("BuildStubTool: %v", err)
	}

	assertNoTopLevelCombinator(t, spec.Name+" stub", stub.RawInputSchema)

	comment, err := issueCommentText("", []byte(`{"comment":"from body"}`))
	if err != nil || comment != "from body" {
		t.Fatalf("stdin body: comment=%q err=%v", comment, err)
	}

	comment, err = issueCommentText("from flag", []byte(`{"comment":"from body"}`))
	if err != nil || comment != "from flag" {
		t.Fatalf("flag should win: comment=%q err=%v", comment, err)
	}

	if _, err := issueCommentText("  ", nil); err == nil {
		t.Fatal("expected --message is required when both inputs are empty")
	}

	if _, err := issueCommentText("", []byte(`{"comment":"  "}`)); err == nil {
		t.Fatal("expected --message is required when comment is blank")
	}

	resetFlag(t, issueCommentCmd, "message")
	err = runWithoutAPIKey(t, issueCommentCmd, []string{"00000000-0000-0000-0000-000000000001"}, `{"comment":"from body"}`)
	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") {
		t.Fatalf("body was not consumed, got %v", err)
	}

	resetFlag(t, issueCommentCmd, "message")
	message := issueCommentCmd.Flags().Lookup("message")
	if err := message.Value.Set("from flag"); err != nil {
		t.Fatalf("set message: %v", err)
	}

	message.Changed = true
	stdin := &countingReader{r: strings.NewReader("{not-json")}
	issueCommentCmd.SetIn(stdin)
	issueCommentCmd.SetOut(io.Discard)
	issueCommentCmd.SetErr(io.Discard)
	t.Setenv("ESCAPE_API_KEY", "")
	t.Setenv("ESCAPE_AUTHORIZATION", "")
	err = issueCommentCmd.RunE(issueCommentCmd, []string{"00000000-0000-0000-0000-000000000001"})
	if stdin.reads != 0 {
		t.Fatalf("stdin was read %d times while --message was set", stdin.reads)
	}

	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") || strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("flag-only comment should reach the API, got %v", err)
	}
}

func TestLocationCommandsMergeAdvertisedBody(t *testing.T) {
	createSpec := mcpSpec(t, "locations_create")
	createSchema := decodeToolSchema(t, createSpec.Tool.RawInputSchema)
	createBody := createSchema["properties"].(map[string]any)["body"].(map[string]any)
	for _, field := range []string{"name", "sshPublicKey"} {
		if _, ok := createBody["properties"].(map[string]any)[field]; !ok {
			t.Errorf("locations_create body missing %s", field)
		}
	}

	updateSpec := mcpSpec(t, "locations_update")
	updateSchema := decodeToolSchema(t, updateSpec.Tool.RawInputSchema)
	updateBody := updateSchema["properties"].(map[string]any)["body"].(map[string]any)
	if _, ok := updateBody["properties"].(map[string]any)["enabled"]; !ok {
		t.Fatal("locations_update body missing enabled")
	}

	parsed, err := parseLocationBody([]byte(`{"name":"from-body","sshPublicKey":"ssh-from-body"}`))
	if err != nil {
		t.Fatalf("parse create body: %v", err)
	}

	created, err := mergeLocationCreate(parsed, "from-flag", "", true, false)
	if err != nil {
		t.Fatalf("merge create: %v", err)
	}

	if created.Name != "from-flag" || created.SSHPublicKey != "ssh-from-body" {
		t.Fatalf("merge create = %#v", created)
	}

	if _, err := mergeLocationCreate(locationBody{}, "", "", false, false); err == nil {
		t.Fatal("expected --name is required")
	}

	disabled := false
	updated, err := mergeLocationUpdate(locationBody{Enabled: &disabled}, "", "", true, false, false, false)
	if err != nil || updated.Enabled == nil || *updated.Enabled {
		t.Fatalf("stdin enabled=false was dropped: %#v err=%v", updated, err)
	}

	if _, err := mergeLocationUpdate(locationBody{}, "", "", false, false, false, false); err == nil {
		t.Fatal("expected at least one field")
	}

	resetFlag(t, locationsCreateCmd, "name")
	resetFlag(t, locationsCreateCmd, "ssh-public-key")
	err = runWithoutAPIKey(t, locationsCreateCmd, nil, `{"name":"prod-vpc","sshPublicKey":"ssh-ed25519 AAAA"}`)
	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") {
		t.Fatalf("create body was not consumed, got %v", err)
	}

	resetFlag(t, locationsUpdateCmd, "name")
	resetFlag(t, locationsUpdateCmd, "ssh-public-key")
	resetFlag(t, locationsUpdateCmd, "enabled")
	err = runWithoutAPIKey(t, locationsUpdateCmd, []string{"00000000-0000-0000-0000-000000000002"}, `{"enabled":false}`)
	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") {
		t.Fatalf("update body was not consumed, got %v", err)
	}

	resetFlag(t, locationsCreateCmd, "name")
	resetFlag(t, locationsCreateCmd, "ssh-public-key")
	name := locationsCreateCmd.Flags().Lookup("name")
	if err := name.Value.Set("prod-vpc"); err != nil {
		t.Fatalf("set name: %v", err)
	}

	name.Changed = true
	createStdin := &countingReader{r: strings.NewReader("{not-json")}
	locationsCreateCmd.SetIn(createStdin)
	locationsCreateCmd.SetOut(io.Discard)
	locationsCreateCmd.SetErr(io.Discard)
	t.Setenv("ESCAPE_API_KEY", "")
	t.Setenv("ESCAPE_AUTHORIZATION", "")
	err = locationsCreateCmd.RunE(locationsCreateCmd, nil)
	if createStdin.reads != 0 {
		t.Fatalf("create read stdin %d times while --name was set", createStdin.reads)
	}

	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") || strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("flag-only create should reach the API, got %v", err)
	}

	resetFlag(t, locationsUpdateCmd, "name")
	resetFlag(t, locationsUpdateCmd, "ssh-public-key")
	resetFlag(t, locationsUpdateCmd, "enabled")
	enabled := locationsUpdateCmd.Flags().Lookup("enabled")
	if err := enabled.Value.Set("true"); err != nil {
		t.Fatalf("set enabled: %v", err)
	}

	enabled.Changed = true
	updateStdin := &countingReader{r: strings.NewReader("{not-json")}
	locationsUpdateCmd.SetIn(updateStdin)
	locationsUpdateCmd.SetOut(io.Discard)
	locationsUpdateCmd.SetErr(io.Discard)
	err = locationsUpdateCmd.RunE(locationsUpdateCmd, []string{"00000000-0000-0000-0000-000000000002"})
	if updateStdin.reads != 0 {
		t.Fatalf("update read stdin %d times while --enabled was set", updateStdin.reads)
	}

	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") || strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("flag-only update should reach the API, got %v", err)
	}
}

func TestAssetCreateSchemaCoversDispatchedTypes(t *testing.T) {
	schema, err := assetCreateInputSchema()
	if err != nil {
		t.Fatalf("assetCreateInputSchema: %v", err)
	}

	payloads, err := assetCreatePayloads()
	if err != nil {
		t.Fatalf("assetCreatePayloads: %v", err)
	}

	expanded := 0
	for _, payload := range payloads {
		if strings.Contains(payload.Name(), "Comment") {
			t.Fatalf("comment creation is not an asset create payload: %s", payload.Name())
		}

		expanded += len(schemaTypesFor(payload))
	}

	if len(schema.OneOf) != expanded {
		t.Fatalf("oneOf variants = %d, dispatched payloads = %d", len(schema.OneOf), expanded)
	}

	if !strings.Contains(schema.Description, climcp.GetToolSpecToolName) {
		t.Fatalf("description %q does not point at the full schema tool", schema.Description)
	}

	seen := map[string]int{}
	for _, variant := range schema.OneOf {
		prop := variant.Properties["asset_type"]
		if prop == nil || len(prop.Enum) == 0 {
			t.Fatalf("variant %q has no asset_type enum", variant.Description)
		}

		for _, value := range prop.Enum {
			seen[value]++
		}
	}

	for _, want := range []string{"WEBAPP", "REST", "rest", "DNS", "IPV4_RANGE", "SCHEMA", "GITHUB_REPOSITORY", "GRAPHQL", "AWS_LAMBDA"} {
		if seen[want] == 0 {
			t.Errorf("asset_type enum missing %s", want)
		}
	}

	if seen["SCHEMA"] != 3 {
		t.Errorf("SCHEMA variants = %d, want fetch, inline content and upload", seen["SCHEMA"])
	}

	spec := mcpSpec(t, "assets_create")
	toolSchema := decodeToolSchema(t, spec.Tool.RawInputSchema)
	body := toolSchema["properties"].(map[string]any)["body"].(map[string]any)
	oneOf, _ := body["oneOf"].([]any)
	if len(oneOf) != len(schema.OneOf) {
		t.Fatalf("MCP body oneOf = %d, schema oneOf = %d", len(oneOf), len(schema.OneOf))
	}

	err = runWithoutAPIKey(t, createAssetCmd, nil, `{"asset_type":"WEBAPP","url":"https://example.com"}`)
	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") {
		t.Fatalf("WEBAPP body was not dispatched, got %v", err)
	}

	err = runWithoutAPIKey(t, createAssetCmd, nil, `{"asset_type":"NOT_A_REAL_TYPE"}`)
	if err == nil || !strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "ESCAPE_API_KEY") {
		t.Fatalf("unknown asset type should fail before the client, got %v", err)
	}
}

func TestIntegrationSchemaDescribesEachKind(t *testing.T) {
	kinds := integrationKinds()
	for _, want := range []string{"jira", "github", "aws-account", "azure-key-vault", "wiz"} {
		if !containsString(kinds, want) {
			t.Errorf("integration kinds %v missing %s", kinds, want)
		}
	}

	createSchema, err := integrationCreateInputSchema()
	if err != nil {
		t.Fatalf("integrationCreateInputSchema: %v", err)
	}

	updateSchema, err := integrationUpdateInputSchema()
	if err != nil {
		t.Fatalf("integrationUpdateInputSchema: %v", err)
	}

	assertIntegrationVariants(t, createSchema.AnyOf, kinds)
	assertIntegrationVariants(t, updateSchema.AnyOf, kinds)
	if !strings.Contains(createSchema.Description, "--kind does not select a variant") {
		t.Fatalf("description %q still claims the body is selected by --kind", createSchema.Description)
	}

	for _, variant := range createSchema.AnyOf {
		if strings.Contains(variant.Description, `"postman"`) {
			if !strings.Contains(variant.Description, "no request model") || !strings.Contains(variant.Description, "accepts any object") {
				t.Fatalf("postman must say it has no model and accepts any object: %s", variant.Description)
			}

			if variant.AdditionalProperties != nil {
				t.Fatal("postman variant must leave additionalProperties unset")
			}

			continue
		}

		if variant.AdditionalProperties == nil || *variant.AdditionalProperties {
			t.Fatalf("variant %q must set additionalProperties false", variant.Description)
		}
	}

	spec := mcpSpec(t, "integrations_create")
	schema := decodeToolSchema(t, spec.Tool.RawInputSchema)
	kind := schema["properties"].(map[string]any)["kind"].(map[string]any)
	if got := stringSlice(kind["enum"]); !reflect.DeepEqual(got, kinds) {
		t.Fatalf("kind enum = %v, want %v", got, kinds)
	}

	body := schema["properties"].(map[string]any)["body"].(map[string]any)
	anyOf, _ := body["anyOf"].([]any)
	if len(anyOf) != len(kinds) {
		t.Fatalf("body anyOf = %d, kinds = %d", len(anyOf), len(kinds))
	}

	stub, err := climcp.BuildStubTool(spec)
	if err != nil {
		t.Fatalf("BuildStubTool: %v", err)
	}

	stubSchema := decodeToolSchema(t, stub.RawInputSchema)
	stubKind := stubSchema["properties"].(map[string]any)["kind"].(map[string]any)
	if got := stringSlice(stubKind["enum"]); !reflect.DeepEqual(got, kinds) {
		t.Fatalf("stub kind enum = %v, want %v", got, kinds)
	}

	prevKind := integrationsKind
	t.Cleanup(func() { integrationsKind = prevKind })
	integrationsKind = "jira"
	err = runWithoutAPIKey(t, integrationsCreateCmd, nil, `{"name":"jira","parameters":{}}`)
	if err == nil || !strings.Contains(err.Error(), "ESCAPE_API_KEY") {
		t.Fatalf("integration body was not consumed, got %v", err)
	}
}

func TestReadPipedStdinSkipsTerminalAndEmptyInput(t *testing.T) {
	data, err := readPipedStdin(strings.NewReader(" {\"a\":1} "))
	if err != nil || string(data) != " {\"a\":1} " {
		t.Fatalf("reader: data=%q err=%v", data, err)
	}

	data, err = readPipedStdin(strings.NewReader(" \n"))
	if err != nil || data != nil {
		t.Fatalf("blank reader: data=%q err=%v", data, err)
	}

	data, err = readPipedStdin(nil)
	if err != nil || data != nil {
		t.Fatalf("nil reader: data=%q err=%v", data, err)
	}

	tty, err := os.Open("/dev/tty")
	if err != nil {
		return
	}

	t.Cleanup(func() { _ = tty.Close() })
	data, err = readPipedStdin(tty)
	if err != nil || data != nil {
		t.Fatalf("terminal should not be read: data=%q err=%v", data, err)
	}
}

func assertIntegrationVariants(t *testing.T, variants []*clischema.JSONSchema, kinds []string) {
	t.Helper()
	if len(variants) != len(kinds) {
		t.Fatalf("variants = %d, kinds = %d", len(variants), len(kinds))
	}

	for i, variant := range variants {
		if !strings.Contains(variant.Description, kinds[i]) {
			t.Errorf("variant %d description %q does not name kind %s", i, variant.Description, kinds[i])
		}
	}
}

func mcpSpec(t *testing.T, name string) climcp.ToolSpec {
	t.Helper()
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	for _, spec := range specs {
		if spec.Name == name {
			return spec
		}
	}

	t.Fatalf("MCP tool %q not registered", name)

	return climcp.ToolSpec{}
}

func decodeToolSchema(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal tool schema: %v", err)
	}

	return schema
}

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, _ := item.(string)
			out = append(out, text)
		}

		return out
	default:
		return nil
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}

func runWithoutAPIKey(t *testing.T, command *cobra.Command, args []string, stdin string) error {
	t.Helper()
	t.Setenv("ESCAPE_API_KEY", "")
	t.Setenv("ESCAPE_AUTHORIZATION", "")
	command.SetIn(strings.NewReader(stdin))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	if err := command.RunE(command, args); err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	return nil
}

func resetFlag(t *testing.T, command *cobra.Command, name string) {
	t.Helper()
	flag := command.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing flag %s", name)
	}

	previous := flag.Value.String()
	previousChanged := flag.Changed
	t.Cleanup(func() {
		_ = flag.Value.Set(previous)
		flag.Changed = previousChanged
	})
	if err := flag.Value.Set(flag.DefValue); err != nil {
		t.Fatalf("reset %s: %v", name, err)
	}

	flag.Changed = false
}

type countingReader struct {
	r     io.Reader
	reads int
}

func (c *countingReader) Read(p []byte) (int, error) {
	c.reads++
	n, err := c.r.Read(p)

	return n, preserveEOF(err)
}

// preserveEOF wraps read errors but leaves io.EOF bare. io.Copy compares the
// error to io.EOF with ==, so wrapping EOF would surface it as a failure.
func preserveEOF(err error) error {
	if err == nil || err == io.EOF {
		return err
	}

	return fmt.Errorf("read: %w", err)
}

// TestAdvertisedBodiesReachTheAPI pipes a JSON body at every tool that
// advertises one and checks the httptest server received that body. Handlers
// read either os.Stdin or the command's InOrStdin, so both are fed.
func TestAdvertisedBodiesReachTheAPI(t *testing.T) {
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	commands := indexCommands(rootCmd)

	var mu sync.Mutex
	var got []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		mu.Lock()
		got = append(got, string(data))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"00000000-0000-0000-0000-000000000099","message":"ok"}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	for _, spec := range specs {
		if spec.BodyProperty == "" {
			continue
		}

		command := commands[spec.Path]
		if command == nil {
			t.Errorf("%s has no cobra command", spec.Name)
			continue
		}

		body := stdinBodyFor(spec.Path)
		resetLocalFlags(t, command)
		if strings.HasPrefix(spec.Path, "escape-cli integrations ") {
			if err := command.Flags().Set("kind", "jira"); err != nil {
				t.Fatalf("%s kind: %v", spec.Name, err)
			}
		}

		args := make([]string, len(spec.PositionalArgs))
		for i := range args {
			args[i] = "00000000-0000-0000-0000-000000000001"
		}

		before := recordedBodies(t, &mu, &got)
		pipeStdin(t, body)
		command.SetContext(context.Background())
		command.SetIn(strings.NewReader(body))
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		runErr := command.RunE(command, args)
		after := recordedBodies(t, &mu, &got)
		if !bodyWasForwarded(after[len(before):], "stdinProbe") {
			t.Errorf("%s did not forward the piped body; err=%v new requests=%q", spec.Name, runErr, after[len(before):])
		}
	}
}

func stdinBodyFor(path string) string {
	const id = "00000000-0000-0000-0000-000000000001"
	switch path {
	case "escape-cli issues comment":
		return `{"comment":"stdinProbe"}`
	case "escape-cli issues remediation generate":
		return `{"kind":"summary","stdinProbe":"body"}`
	case "escape-cli issues remediation feedback":
		return `{"feedback":true,"stdinProbe":"body"}`
	case "escape-cli locations create":
		return `{"name":"stdinProbe","sshPublicKey":"ssh-ed25519 AAAA"}`
	case "escape-cli locations update":
		return `{"name":"stdinProbe"}`
	case "escape-cli assets create":
		return `{"asset_class":"FRONTEND","asset_type":"WEBAPP","url":"https://example.com/stdinProbe"}`
	case "escape-cli assets bulk-import":
		return `{"assets":[{"asset_type":"WEBAPP","url":"https://example.com/stdinProbe"}]}`
	case "escape-cli retests start":
		return `{"profileId":"` + id + `","issueIds":["00000000-0000-0000-0000-000000000002"],"context":"stdinProbe"}`
	case "escape-cli roles create":
		return `{"name":"stdinProbe","permissions":["VIEW_ONLY"],"bindings":[]}`
	case "escape-cli roles update":
		return `{"role":{"name":"stdinProbe"}}`
	case "escape-cli projects create":
		return `{"name":"stdinProbe","slug":"stdin-probe","bindings":[]}`
	case "escape-cli projects update":
		return `{"project":{"name":"stdinProbe"}}`
	case "escape-cli workflows create":
		return `{"organizationId":"` + id + `","name":"stdinProbe","trigger":"MANUAL","actions":[]}`
	case "escape-cli workflows update":
		return `{"name":"stdinProbe","trigger":"MANUAL"}`
	case "escape-cli profiles create-rest", "escape-cli profiles create-webapp", "escape-cli profiles create-graphql":
		return `{"assetId":"` + id + `","name":"stdinProbe"}`
	case "escape-cli profiles create-ai-pentest":
		return `{"assetIds":["` + id + `"],"name":"stdinProbe"}`
	case "escape-cli profiles update-configuration":
		return `{"configuration":{"mode":"read_only"},"stdinProbe":"body"}`
	case "escape-cli profiles continuous-pentest create",
		"escape-cli profiles continuous-pentest enable":
		return `{"repositories":[{"url":"https://github.com/org/repo","branch":"main"}],"cron":"0 0 * * *","stdinProbe":"body"}`
	case "escape-cli profiles continuous-pentest update":
		return `{"cron":"0 0 * * *","stdinProbe":"body"}`
	case "escape-cli regression-tests create":
		return `{"name":"stdinProbe","additionalContext":"stdinProbe","inputFilename":"report.pdf","temporaryObjectKey":"00000000-0000-0000-0000-000000000001","stdinProbe":"body"}`
	case "escape-cli regression-tests update":
		return `{"name":"stdinProbe","stdinProbe":"body"}`
	case "escape-cli regression-tests answer":
		return `{"content":"stdinProbe","stdinProbe":"body"}`
	case "escape-cli custom-rules create":
		return `{"content":{"rule":{"type":"API","alert":{"severity":"INFO","name":"stdinProbe","context":"probe","category":"CUSTOM"},"detect":[{"if":"scan.type"}]}}}`
	default:
		return `{"stdinProbe":"body","name":"stdinProbe"}`
	}
}

func bodyWasForwarded(requests []string, marker string) bool {
	for _, request := range requests {
		if strings.Contains(request, marker) {
			return true
		}
	}

	return false
}

func recordedBodies(t *testing.T, mu *sync.Mutex, got *[]string) []string {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()

	return append([]string(nil), (*got)...)
}

func resetLocalFlags(t *testing.T, command *cobra.Command) {
	t.Helper()
	command.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		resetFlag(t, command, flag.Name)
	})
}

func pipeStdin(t *testing.T, body string) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	if _, err := writer.Write([]byte(body)); err != nil {
		t.Fatalf("write stdin: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}

	original := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = original
		_ = reader.Close()
	})
}
