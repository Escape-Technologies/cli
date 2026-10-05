package cmd

import (
	"strings"
	"testing"
)

func TestAuditDateFromUsageStatesDefault(t *testing.T) {
	flag := auditListCmd.Flags().Lookup("date-from")
	if flag == nil {
		t.Fatal("missing --date-from")
	}

	const want = "default: 12 hours ago"
	if !strings.Contains(flag.Usage, want) {
		t.Fatalf("--date-from usage %q does not contain %q", flag.Usage, want)
	}

	spec := mcpSpecByName(t, mcpSpecs(t), "audit_list")
	raw, ok := schemaProperties(t, spec)["date_from"]
	if !ok {
		t.Fatal("audit_list schema missing date_from")
	}

	prop, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("date_from schema = %#v", raw)
	}

	description, _ := prop["description"].(string)
	if !strings.Contains(description, want) {
		t.Fatalf("date_from description %q does not contain %q", description, want)
	}
}
