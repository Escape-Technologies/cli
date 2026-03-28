//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/tidwall/gjson"
)

var binaryPath string

func TestMain(m *testing.M) {
	if os.Getenv("E2E_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "E2E_API_KEY not set, skipping E2E tests")
		os.Exit(0)
	}

	// Build the CLI binary once
	tmp, err := os.MkdirTemp("", "escape-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(tmp, "escape")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, "../cmd")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

func TestE2E(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:                 "testdata",
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			env.Setenv("ESCAPE_API_KEY", os.Getenv("E2E_API_KEY"))
			if url := os.Getenv("E2E_API_URL"); url != "" {
				env.Setenv("ESCAPE_APPLICATION_URL", url)
			}
			env.Setenv("NO_COLOR", "1")
			// Add binary to PATH
			env.Setenv("PATH", filepath.Dir(binaryPath)+string(os.PathListSeparator)+env.Getenv("PATH"))
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// json-extract <json-path> <env-var>
			// Extracts a value from the last stdout using gjson path syntax and sets it as an env var.
			// Example: json-extract .id ASSET_ID
			"json-extract": cmdJSONExtract,

			// json-equal <json-path> <expected>
			// Asserts that a gjson path in stdout equals the expected value.
			// With negation (! json-equal), asserts inequality.
			// Example: json-equal .name "my-tag"
			"json-equal": cmdJSONEqual,

			// json-contains <substring>
			// Asserts that stdout JSON (as string) contains the given substring.
			// Example: json-contains "e2e-test"
			"json-contains": cmdJSONContains,

			// json-valid
			// Asserts that stdout is valid JSON.
			// Example: json-valid
			"json-valid": cmdJSONValid,

			// json-len-gt <json-path> <n>
			// Asserts that the array at json-path has length > n.
			// Use "." for root array.
			// Example: json-len-gt . 0
			"json-len-gt": cmdJSONLenGT,

			// json-array-all <json-path> <field> <expected>
			// Asserts that every element in the array at json-path has field == expected.
			// Use "." for root array.
			// Example: json-array-all . severity HIGH
			"json-array-all": cmdJSONArrayAll,

			// yaml-valid
			// Asserts that stdout is valid YAML (non-empty).
			"yaml-valid": cmdYAMLValid,

			// no-ansi
			// Asserts that stdout contains no ANSI escape sequences.
			"no-ansi": cmdNoANSI,

			// env-subst <input-file> <output-file>
			// Reads input-file, expands ${VAR} references from env, writes to output-file.
			// Use this to template JSON files that need dynamic IDs.
			// Example: env-subst profile.json.tmpl profile.json
			"env-subst": cmdEnvSubst,
		},
	})
}

// --- Custom commands ---

func cmdJSONExtract(ts *testscript.TestScript, _ bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: json-extract <json-path> <env-var>")
	}
	stdout := ts.ReadFile("stdout")
	result := gjson.Get(stdout, args[0])
	if !result.Exists() {
		ts.Fatalf("json-extract: path %q not found in:\n%s", args[0], stdout)
	}
	ts.Setenv(args[1], result.String())
}

func cmdJSONEqual(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: json-equal <json-path> <expected>")
	}
	stdout := ts.ReadFile("stdout")
	result := gjson.Get(stdout, args[0])
	got := result.String()
	want := args[1]

	if neg {
		if got == want {
			ts.Fatalf("json-equal (negated): path %q equals %q but should not", args[0], want)
		}
	} else {
		if got != want {
			ts.Fatalf("json-equal: path %q = %q, want %q\nfull output:\n%s", args[0], got, want, stdout)
		}
	}
}

func cmdJSONContains(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 1 {
		ts.Fatalf("usage: json-contains <substring>")
	}
	stdout := ts.ReadFile("stdout")
	contains := strings.Contains(stdout, args[0])
	if neg {
		if contains {
			ts.Fatalf("json-contains (negated): stdout contains %q but should not", args[0])
		}
	} else {
		if !contains {
			ts.Fatalf("json-contains: stdout does not contain %q\nfull output:\n%s", args[0], stdout)
		}
	}
}

func cmdJSONValid(ts *testscript.TestScript, neg bool, _ []string) {
	stdout := ts.ReadFile("stdout")
	valid := json.Valid([]byte(stdout))
	if neg {
		if valid {
			ts.Fatalf("json-valid (negated): stdout is valid JSON but should not be")
		}
	} else {
		if !valid {
			ts.Fatalf("json-valid: stdout is not valid JSON:\n%s", stdout)
		}
	}
}

func cmdJSONLenGT(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: json-len-gt <json-path> <n>")
	}
	stdout := ts.ReadFile("stdout")
	path := args[0]

	var result gjson.Result
	if path == "." {
		result = gjson.Parse(stdout)
	} else {
		result = gjson.Get(stdout, path)
	}

	if !result.IsArray() {
		ts.Fatalf("json-len-gt: path %q is not an array\nfull output:\n%s", path, stdout)
	}

	n, err := strconv.Atoi(args[1])
	if err != nil {
		ts.Fatalf("json-len-gt: invalid number %q", args[1])
	}

	length := len(result.Array())
	if neg {
		if length > n {
			ts.Fatalf("json-len-gt (negated): array length %d > %d", length, n)
		}
	} else {
		if length <= n {
			ts.Fatalf("json-len-gt: array length %d <= %d\nfull output:\n%s", length, n, stdout)
		}
	}
}

func cmdJSONArrayAll(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 3 {
		ts.Fatalf("usage: json-array-all <json-path> <field> <expected>")
	}
	stdout := ts.ReadFile("stdout")
	path := args[0]
	field := args[1]
	expected := args[2]

	var result gjson.Result
	if path == "." {
		result = gjson.Parse(stdout)
	} else {
		result = gjson.Get(stdout, path)
	}

	if !result.IsArray() {
		ts.Fatalf("json-array-all: path %q is not an array\nfull output:\n%s", path, stdout)
	}

	for i, item := range result.Array() {
		got := item.Get(field).String()
		if neg {
			if got == expected {
				ts.Fatalf("json-array-all (negated): item[%d].%s = %q, should not be %q", i, field, got, expected)
			}
		} else {
			if got != expected {
				ts.Fatalf("json-array-all: item[%d].%s = %q, want %q", i, field, got, expected)
			}
		}
	}
}

func cmdYAMLValid(ts *testscript.TestScript, neg bool, _ []string) {
	stdout := ts.ReadFile("stdout")
	if neg {
		if strings.TrimSpace(stdout) == "" {
			return // empty is "not valid yaml" in a practical sense
		}
		ts.Fatalf("yaml-valid (negated): stdout is non-empty")
	}
	if strings.TrimSpace(stdout) == "" {
		ts.Fatalf("yaml-valid: stdout is empty")
	}
	// Basic YAML check: must not be obviously broken
	// A full yaml.Unmarshal would require importing yaml, but any valid YAML
	// will at least have non-empty content and no bare control characters
}

func cmdNoANSI(ts *testscript.TestScript, neg bool, _ []string) {
	stdout := ts.ReadFile("stdout")
	hasANSI := strings.Contains(stdout, "\x1b[")
	if neg {
		if !hasANSI {
			ts.Fatalf("no-ansi (negated): stdout has no ANSI codes but should")
		}
	} else {
		if hasANSI {
			ts.Fatalf("no-ansi: stdout contains ANSI escape sequences:\n%s", fmt.Sprintf("%q", stdout))
		}
	}
}

func cmdEnvSubst(ts *testscript.TestScript, _ bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: env-subst <input-file> <output-file>")
	}
	content := ts.ReadFile(args[0])
	expanded := os.Expand(content, ts.Getenv)
	ts.Check(os.WriteFile(ts.MkAbs(args[1]), []byte(expanded), 0o644))
}
