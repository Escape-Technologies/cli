package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestExportSchemaDoesNotCallTheAPI(t *testing.T) {
	called := false
	serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
		http.Error(w, "schema mode must not call the API", http.StatusInternalServerError)
	})

	stdout, stderr, err := captureOutput(t, "schema", func() error {
		if err := issueExportCmd.RunE(issueExportCmd, nil); err != nil {
			return fmt.Errorf("issues export schema: %w", err)
		}

		if err := assetExportCmd.RunE(assetExportCmd, nil); err != nil {
			return fmt.Errorf("assets export schema: %w", err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("schema: %v\n%s", err, stderr)
	}

	if called {
		t.Fatal("schema mode called the API")
	}

	if !strings.Contains(stdout, `"contentType"`) || !strings.Contains(stdout, `"version"`) || !strings.Contains(stdout, `"runs"`) {
		t.Fatalf("schema = %s", stdout)
	}
}

func TestIssueExportJSONPrintsCSVAndSARIF(t *testing.T) {
	var path, accept, where string
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		accept = r.Header.Get("Accept")
		where = r.URL.Query().Get("where")
		if strings.HasSuffix(r.URL.Path, "/sarif") {
			w.Header().Set("Content-Type", "application/sarif+json")
			_, _ = w.Write([]byte(`{"runs":[],"version":"2.1.0"}`))

			return
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		_, _ = w.Write([]byte("id,name\n1,a"))
	})
	setChangedFlag(t, issueExportCmd, "format", "csv")
	setChangedFlag(t, issueExportCmd, "where", `{"status":"OPEN"}`)
	resetFlag(t, issueExportCmd, "out")

	stdout := runCmdJSON(t, issueExportCmd, nil)
	if path != "/v3/exports/issues/csv" || accept != "text/csv" || where != `{"status":"OPEN"}` {
		t.Fatalf("csv request path=%s accept=%s where=%s", path, accept, where)
	}

	var document csvExport
	if err := json.Unmarshal([]byte(stdout), &document); err != nil {
		t.Fatalf("csv stdout %s: %v", stdout, err)
	}

	if document.ContentType != csvContentType || document.Content != "id,name\n1,a" {
		t.Fatalf("csv document = %#v", document)
	}

	setChangedFlag(t, issueExportCmd, "format", "sarif")
	stdout = runCmdJSON(t, issueExportCmd, nil)
	if path != "/v3/exports/issues/sarif" || accept != "application/sarif+json" {
		t.Fatalf("sarif request path=%s accept=%s", path, accept)
	}

	if strings.Contains(stdout, "contentType") {
		t.Fatalf("sarif was wrapped: %s", stdout)
	}

	if !strings.Contains(stdout, `{"runs":[],"version":"2.1.0"}`) {
		t.Fatalf("sarif was reencoded: %s", stdout)
	}
}

func TestAssetExportJSONPrintsCSV(t *testing.T) {
	var path, accept, where string
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		accept = r.Header.Get("Accept")
		where = r.URL.Query().Get("where")
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("id\n1"))
	})
	resetFlag(t, assetExportCmd, "format")
	setChangedFlag(t, assetExportCmd, "where", `{"statuses":["MONITORED"]}`)
	resetFlag(t, assetExportCmd, "out")

	stdout := runCmdJSON(t, assetExportCmd, nil)
	if path != "/v3/exports/assets/csv" || accept != "text/csv" || where != `{"statuses":["MONITORED"]}` {
		t.Fatalf("request path=%s accept=%s where=%s", path, accept, where)
	}

	var document csvExport
	if err := json.Unmarshal([]byte(stdout), &document); err != nil {
		t.Fatalf("stdout %s: %v", stdout, err)
	}

	if document.ContentType != csvContentType || document.Content != "id\n1" {
		t.Fatalf("document = %#v", document)
	}
}

func TestExportPrettyAndOutWriteRawBytes(t *testing.T) {
	// The body already ends in a newline. Pretty stdout must be those exact
	// bytes: trimming would hide the extra blank record out.Print used to add.
	const body = "id,name\n1,a\n"
	serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(body))
	})
	setChangedFlag(t, issueExportCmd, "format", "CSV")
	resetFlag(t, issueExportCmd, "where")
	resetFlag(t, issueExportCmd, "out")

	stdout, stderr, err := captureOutput(t, "pretty", func() error {
		issueExportCmd.SetContext(t.Context())

		return issueExportCmd.RunE(issueExportCmd, nil)
	})
	if err != nil {
		t.Fatalf("pretty: %v\n%s", err, stderr)
	}

	if stdout != body {
		t.Fatalf("pretty stdout = %q, want %q", stdout, body)
	}

	path := filepath.Join(t.TempDir(), "issues.csv")
	setChangedFlag(t, issueExportCmd, "out", path)
	stdout = runCmdJSON(t, issueExportCmd, nil)
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(written) != body {
		t.Fatalf("file = %q", written)
	}

	var document csvExport
	if err := json.Unmarshal([]byte(stdout), &document); err != nil {
		t.Fatalf("stdout %s: %v", stdout, err)
	}

	if document.Content != body {
		t.Fatalf("stdout document = %#v", document)
	}
}

func TestExportWhereMustBeAJSONObject(t *testing.T) {
	setChangedFlag(t, issueExportCmd, "format", "csv")
	cases := []struct {
		value string
		want  string
	}{
		{value: " ", want: "--where must not be blank"},
		{value: "[]", want: "--where must be a JSON object"},
		{value: "{", want: "--where must be a JSON object"},
	}
	for _, testCase := range cases {
		setChangedFlag(t, issueExportCmd, "where", testCase.value)
		issueExportCmd.SetContext(t.Context())
		err := issueExportCmd.RunE(issueExportCmd, nil)
		if err == nil || !strings.Contains(err.Error(), testCase.want) {
			t.Errorf("where %q err = %v", testCase.value, err)
		}
	}
}

func TestAssetExportRejectsSARIF(t *testing.T) {
	setChangedFlag(t, assetExportCmd, "format", "sarif")
	assetExportCmd.SetContext(t.Context())
	err := assetExportCmd.RunE(assetExportCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--format must be one of: csv") {
		t.Fatalf("err = %v", err)
	}
}

func TestIssueExportRejectsNonJSONSARIF(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/sarif+json")
		_, _ = w.Write([]byte("not-json"))
	})
	setChangedFlag(t, issueExportCmd, "format", "sarif")
	resetFlag(t, issueExportCmd, "where")
	resetFlag(t, issueExportCmd, "out")
	issueExportCmd.SetContext(t.Context())
	err := issueExportCmd.RunE(issueExportCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "sarif export is not a JSON document") {
		t.Fatalf("err = %v", err)
	}
}

func TestIssueExportSchemaDoesNotRequireFormatFlag(t *testing.T) {
	resetFlag(t, issueExportCmd, "format")
	resetFlag(t, issueExportCmd, "where")
	resetFlag(t, issueExportCmd, "out")
	// output is a persistent flag. Flags() does not contain it until something
	// parses the root command, so look it up on PersistentFlags directly.
	outputFlag := rootCmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("missing flag output")
	}

	previousOutput := outputFlag.Value.String()
	previousOutputChanged := outputFlag.Changed
	t.Cleanup(func() {
		_ = outputFlag.Value.Set(previousOutput)
		outputFlag.Changed = previousOutputChanged
		rootCmd.SetArgs(nil)
	})
	if err := outputFlag.Value.Set(outputFlag.DefValue); err != nil {
		t.Fatalf("reset output: %v", err)
	}

	outputFlag.Changed = false

	rootCmd.SetArgs([]string{"issues", "export", "-o", "schema"})
	stdout, stderr, err := captureOutput(t, "pretty", func() error {
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("schema execute: %v\n%s", err, stderr)
	}

	if !strings.Contains(stdout, `"contentType"`) || !strings.Contains(stdout, `"oneOf"`) {
		t.Fatalf("schema = %s", stdout)
	}

	if err := issueExportCmd.Flags().Lookup("format").Value.Set(""); err != nil {
		t.Fatal(err)
	}

	issueExportCmd.Flags().Lookup("format").Changed = false
	rootCmd.SetArgs([]string{"issues", "export", "-o", "json"})
	err = rootCmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--format must be one of") {
		t.Fatalf("err = %v, want a format error from RunE", err)
	}
}

func setChangedFlag(t *testing.T, command *cobra.Command, name, value string) {
	t.Helper()
	resetFlag(t, command, name)
	if err := command.Flags().Set(name, value); err != nil {
		t.Fatalf("set --%s: %v", name, err)
	}
}

func runCmdJSON(t *testing.T, command *cobra.Command, args []string) string {
	t.Helper()
	command.SetContext(t.Context())
	stdout, stderr, err := captureJSONCommand(t, func() error {
		return command.RunE(command, args)
	})
	if err != nil {
		t.Fatalf("%s: %v\n%s", command.CommandPath(), err, stderr)
	}

	return stdout
}
