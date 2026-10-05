package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	climcp "github.com/Escape-Technologies/cli/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestMCPCatalogDropsUnboundedWatch(t *testing.T) {
	if _, ok := CommandSchemaRegistry()["escape-cli scans watch"]; ok {
		t.Fatal("scans watch must stay out of the MCP registry")
	}

	specs := mcpSpecsByName(t)
	if _, ok := specs["scans_watch"]; ok {
		t.Fatal("scans_watch must not be an MCP tool")
	}

	for name, spec := range specs {
		if _, exposed := toolProperties(t, spec)["watch"]; exposed {
			t.Fatalf("%s still exposes the watch flag", name)
		}

		if _, exposed := toolProperties(t, spec)["fail_on_severity"]; exposed {
			t.Fatalf("%s still exposes --fail-on-severity", name)
		}

		for _, binding := range spec.FlagBindings {
			if binding.FlagName == "watch" {
				t.Fatalf("%s still binds --watch", name)
			}

			if binding.FlagName == "fail-on-severity" {
				t.Fatalf("%s still binds --fail-on-severity", name)
			}
		}
	}

	// The CLI itself keeps --watch for humans and CI.
	for _, command := range []*cobra.Command{
		scanStartCmd,
		jobsGetCmd,
		jobsTriggerExportCmd,
		authenticationsStartCmd,
		authenticationsGetCmd,
	} {
		if command.Flags().Lookup("watch") == nil && command.PersistentFlags().Lookup("watch") == nil {
			t.Fatalf("%s is missing --watch (local=%v persistent=%v)", command.CommandPath(), flagNames(command.Flags()), flagNames(command.PersistentFlags()))
		}
	}
}

func flagNames(set *pflag.FlagSet) []string {
	names := []string{}
	set.VisitAll(func(flag *pflag.Flag) {
		names = append(names, flag.Name)
	})

	return names
}

func TestMCPCatalogDescribesPolling(t *testing.T) {
	specs := mcpSpecsByName(t)

	checks := map[string]string{
		"scans_start":           "scans_get",
		"scans_get":             "scans_get",
		"jobs_trigger_export":   "jobs_get",
		"jobs_get":              "jobs_get",
		"authentications_start": "authentications_get",
		"authentications_get":   "authentications_get",
	}
	for name, poll := range checks {
		spec := specs[name]
		if !strings.Contains(spec.Description, poll) {
			t.Fatalf("%s description %q does not name %s", name, spec.Description, poll)
		}

		if spec.Tool.Description != spec.Description {
			t.Fatalf("%s tool description drifted from the catalog description", name)
		}
	}
}

func TestEmailsWaitMCPArgsAreBounded(t *testing.T) {
	specs := mcpSpecsByName(t)
	spec, ok := specs["emails_wait"]
	if !ok {
		t.Fatal("missing emails_wait")
	}

	wantTimeout := climcp.EmailsWaitTimeout.String()
	if strings.Join(spec.ForcedArgs, " ") != "--timeout "+wantTimeout+" --"+emailWaitPendingFlag+"=true" {
		t.Fatalf("forced args = %#v", spec.ForcedArgs)
	}

	properties := toolProperties(t, spec)
	if _, exposed := properties["timeout"]; exposed {
		t.Fatal("emails_wait must not let the model set an unbounded timeout")
	}

	if _, exposed := properties["pending_on_timeout"]; exposed {
		t.Fatal("emails_wait must not expose the pending-on-timeout flag")
	}

	if len(spec.Tool.RawOutputSchema) != 0 {
		t.Fatal("emails_wait must not advertise a ScanEmailDetails output schema")
	}

	if !strings.Contains(spec.Description, emailWaitPendingMessage) {
		t.Fatalf("description %q does not tell the model to call again", spec.Description)
	}

	if !strings.Contains(spec.Description, "emails_list") {
		t.Fatalf("description %q does not name emails_list", spec.Description)
	}

	if !strings.Contains(spec.Description, "pass that after value") {
		t.Fatalf("description %q does not tell the model to pass after back", spec.Description)
	}

	if _, exposed := properties["after"]; !exposed {
		t.Fatal("emails_wait must expose after so the model can resume from the pending baseline")
	}

	if _, exposed := properties["seen_id"]; !exposed {
		t.Fatal("emails_wait must expose seen_id so ids at the after timestamp are not treated as new")
	}

	if timeoutFlag := emailsWaitCmd.Flags().Lookup("timeout"); timeoutFlag == nil || timeoutFlag.DefValue != "0s" {
		t.Fatalf("CLI emails wait timeout default must stay unbounded, got %#v", timeoutFlag)
	}

	prevTimeout := emailsWaitTimeout
	prevPending := emailsWaitPendingOnTimeout
	t.Cleanup(func() {
		_ = emailsWaitCmd.Flags().Set("timeout", prevTimeout.String())
		if prevPending {
			_ = emailsWaitCmd.Flags().Set(emailWaitPendingFlag, "true")
		} else {
			_ = emailsWaitCmd.Flags().Set(emailWaitPendingFlag, "false")
		}
	})
	if err := emailsWaitCmd.ParseFlags(spec.ForcedArgs); err != nil {
		t.Fatalf("CLI rejected MCP emails wait args: %v", err)
	}

	if emailsWaitTimeout != climcp.EmailsWaitTimeout {
		t.Fatalf("timeout = %s, want %s", emailsWaitTimeout, climcp.EmailsWaitTimeout)
	}

	if !emailsWaitPendingOnTimeout {
		t.Fatal("pending-on-timeout was not applied")
	}
}

func mcpSpecsByName(t *testing.T) map[string]climcp.ToolSpec {
	t.Helper()
	specs, err := buildMCPToolSpecs(rootCmd, CommandSchemaRegistry())
	if err != nil {
		t.Fatalf("buildMCPToolSpecs: %v", err)
	}

	byName := make(map[string]climcp.ToolSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}

	return byName
}

func toolProperties(t *testing.T, spec climcp.ToolSpec) map[string]any {
	t.Helper()
	var schema struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(spec.Tool.RawInputSchema, &schema); err != nil {
		t.Fatalf("schema: %v", err)
	}

	return schema.Properties
}
