package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
)

func TestProblemsSchemaDocumentsSummaryAndDetail(t *testing.T) {
	if err := out.SetOutput("schema"); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = out.SetOutput("pretty") })

	stdout := captureStdout(t, func() {
		if err := problemsCmd.RunE(problemsCmd, nil); err != nil {
			t.Fatal(err)
		}
	})

	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("schema output is not JSON: %v\n%s", err, stdout)
	}

	props, _ := doc["properties"].(map[string]any)
	items, _ := props["items"].(map[string]any)
	oneOf, _ := items["oneOf"].([]any)
	if len(oneOf) != 2 {
		t.Fatalf("items oneOf = %#v", items["oneOf"])
	}

	encoded := string(stdout)
	if !strings.Contains(encoded, "ProblemCount") || !strings.Contains(encoded, "Message") {
		t.Fatalf("schema does not name both item shapes:\n%s", encoded)
	}
}
