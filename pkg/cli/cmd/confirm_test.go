package cmd

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func gatedCommand(t *testing.T, ownsDryRun bool) *cobra.Command {
	t.Helper()
	command := &cobra.Command{Use: "probe", RunE: func(*cobra.Command, []string) error { return nil }}
	bindDestructiveGate(command, ownsDryRun)

	return command
}

// A pipe or a redirected stdin is not a terminal: the gate must let those
// through, or every scripted delete would block on a prompt.
func TestConfirmDestructiveProceedsOnNonTerminalStdin(t *testing.T) {
	command := gatedCommand(t, true)
	command.SetIn(strings.NewReader(""))

	proceed, err := confirmDestructive(command, "delete probe", nil)
	if err != nil {
		t.Fatalf("confirmDestructive: %v", err)
	}

	if !proceed {
		t.Fatal("a non-terminal stdin must not prompt")
	}
}

func TestIsInteractiveRejectsCharacterDevicesThatAreNotTerminals(t *testing.T) {
	if isInteractive(strings.NewReader("")) {
		t.Error("a plain reader is not a terminal")
	}

	if isInteractive(nil) {
		t.Error("a nil reader is not a terminal")
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	t.Cleanup(func() {
		_ = reader.Close()
		_ = writer.Close()
	})

	if isInteractive(reader) {
		t.Error("a pipe is not a terminal")
	}

	// /dev/null is a character device, and it is what a test binary and a CI
	// job attach to stdin. Treating it as interactive would hang both.
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}

	t.Cleanup(func() { _ = null.Close() })

	if isInteractive(null) {
		t.Errorf("%s is a character device, not a terminal", os.DevNull)
	}
}

func TestConfirmDestructiveDryRunSendsNothing(t *testing.T) {
	command := gatedCommand(t, true)
	command.SetIn(strings.NewReader(""))
	if err := command.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("set dry-run: %v", err)
	}

	proceed, err := confirmDestructive(command, "delete probe", nil)
	if err != nil {
		t.Fatalf("confirmDestructive: %v", err)
	}

	if proceed {
		t.Fatal("a dry run must stop the command")
	}
}

func TestConfirmDestructiveDryRunFailureSurfaces(t *testing.T) {
	command := gatedCommand(t, true)
	command.SetIn(strings.NewReader(""))
	if err := command.Flags().Set("dry-run", "true"); err != nil {
		t.Fatalf("set dry-run: %v", err)
	}

	proceed, err := confirmDestructive(command, "delete probe", func(context.Context) (any, error) {
		return nil, errors.New("not found")
	})
	if err == nil || !strings.Contains(err.Error(), "dry run") {
		t.Fatalf("err = %v, want the dry run to report its fetch failure", err)
	}

	if proceed {
		t.Fatal("a dry run must stop the command")
	}
}

func TestConfirmDestructiveYesSkipsThePrompt(t *testing.T) {
	command := gatedCommand(t, false)
	command.SetIn(strings.NewReader(""))
	if err := command.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes: %v", err)
	}

	proceed, err := confirmDestructive(command, "delete probe", nil)
	if err != nil || !proceed {
		t.Fatalf("proceed = %v, err = %v, want true and nil", proceed, err)
	}
}

// A command the gate never registered keeps its old behaviour rather than
// silently losing the prompt it was written to have.
func TestConfirmDestructiveIgnoresUnregisteredCommands(t *testing.T) {
	command := &cobra.Command{Use: "unregistered", RunE: func(*cobra.Command, []string) error { return nil }}
	command.SetIn(strings.NewReader(""))

	proceed, err := confirmDestructive(command, "delete probe", nil)
	if err != nil || !proceed {
		t.Fatalf("proceed = %v, err = %v, want true and nil", proceed, err)
	}
}

func TestBindDestructiveGateRegistersFlagsOnce(t *testing.T) {
	command := &cobra.Command{Use: "probe", RunE: func(*cobra.Command, []string) error { return nil }}
	command.Flags().Bool("yes", false, "already declared")

	bindDestructiveGate(command, true)
	bindDestructiveGate(command, true)

	if flag := command.Flags().Lookup("yes"); flag == nil {
		t.Fatal("--yes was not registered")
	}

	if flag := command.Flags().Lookup("dry-run"); flag == nil {
		t.Fatal("--dry-run was not registered")
	}
}

// issues bulk-update owns the API's own dry run, so the gate must not shadow
// it with a send-nothing report.
func TestBindDestructiveGateLeavesAnAPIOwnedDryRunAlone(t *testing.T) {
	command := &cobra.Command{Use: "probe", RunE: func(*cobra.Command, []string) error { return nil }}
	command.Flags().Bool("dry-run", false, "the API runs this dry run")

	bindDestructiveGate(command, false)

	if flag := command.Flags().Lookup("dry-run"); flag.Usage != "the API runs this dry run" {
		t.Fatalf("dry-run usage = %q, want the API-owned flag untouched", flag.Usage)
	}
}

func TestPromptConfirmAcceptsOnlyAnExplicitYes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "y", in: "y\n", want: true},
		{name: "yes", in: "YES\n", want: true},
		{name: "blank", in: "\n", want: false},
		{name: "other", in: "nope\n", want: false},
		{name: "eof", in: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirmed, err := promptConfirm(io.Discard, strings.NewReader(tt.in), "delete probe")
			if err != nil {
				t.Fatalf("promptConfirm: %v", err)
			}

			if confirmed != tt.want {
				t.Fatalf("confirmed = %v, want %v", confirmed, tt.want)
			}
		})
	}
}

// TestPromptConfirmPropagatesReadErrors checks a broken stdin surfaces as an
// error instead of reading as a declined confirmation.
func TestPromptConfirmPropagatesReadErrors(t *testing.T) {
	_, err := promptConfirm(io.Discard, failingReader{}, "delete probe")
	if err == nil || !strings.Contains(err.Error(), "read confirmation") {
		t.Fatalf("err = %v, want the read to surface", err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

// TestConfirmDestructiveRunsTheWiredCommand drives the gate through a real
// delete command: the unit probes above check the gate, this checks the
// command actually consults it.
func TestConfirmDestructiveRunsTheWiredCommand(t *testing.T) {
	var hits int
	serveTagDelete(t, &hits)

	command := tagsDeleteCmd
	command.SetIn(strings.NewReader(""))
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetContext(context.Background())

	if err := command.RunE(command, []string{deleteOutputID}); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if hits != 1 {
		t.Fatalf("API calls = %d, want 1 (no prompt, no dry run)", hits)
	}
}

// TestConfirmDestructiveDryRunReportsTheWiredTarget checks a dry run on a real
// command fetches the resource it would delete and never sends the delete.
func TestConfirmDestructiveDryRunReportsTheWiredTarget(t *testing.T) {
	var hits int
	serveTagDelete(t, &hits)

	command := tagsDeleteCmd
	resetFlag(t, command, "dry-run")
	if err := command.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}

	command.SetIn(strings.NewReader(""))
	command.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return command.RunE(command, []string{deleteOutputID})
	})
	if err != nil {
		t.Fatalf("dry run: %v\nstderr: %s", err, stderr)
	}

	if hits != 1 {
		t.Fatalf("API calls = %d, want 1 (the dry-run fetch only)", hits)
	}

	if !strings.Contains(stdout, deleteOutputID) {
		t.Fatalf("stdout = %s, want the tag that would be deleted", stdout)
	}
}

// serveTagDelete serves the tag get and delete routes and counts requests.
func serveTagDelete(t *testing.T, hits *int) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*hits++
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"id":"`+deleteOutputID+`","name":"prod","color":"ff0000"}`)

			return
		}

		_, _ = io.WriteString(w, `{"message":"deleted"}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
}
