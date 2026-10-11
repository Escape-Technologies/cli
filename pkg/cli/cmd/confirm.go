package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// errConfirmationDeclined is returned when an interactive operator answers no
// at the confirmation prompt. The command then exits non-zero without having
// sent the request.
var errConfirmationDeclined = errors.New("confirmation declined")

// destructiveGate is the shared confirmation state of one destructive command.
type destructiveGate struct {
	// ownsDryRun records that the gate registered --dry-run as a
	// send-nothing report. Without it the command's --dry-run belongs to the
	// API itself (issues bulk-update forwards the API's own dry run), so the
	// gate leaves that flag alone.
	ownsDryRun bool
}

// destructiveGates lists the commands guarded by confirmDestructive. A command
// joins through bindDestructiveGate; membership is what turns the shared
// confirmation on, so a new destructive command cannot forget it.
var destructiveGates = map[*cobra.Command]destructiveGate{}

// bindDestructiveGate marks a destructive command for confirmDestructive and
// registers the shared flags. --yes skips the prompt; --dry-run reports what
// the command would do and sends nothing, except on commands whose dry run the
// API executes itself.
func bindDestructiveGate(command *cobra.Command, ownsDryRun bool) {
	destructiveGates[command] = destructiveGate{ownsDryRun: ownsDryRun}

	if command.Flags().Lookup("yes") == nil {
		command.Flags().BoolP("yes", "y", false, "confirm the action without prompting; a non-interactive stdin never prompts")
	}

	if ownsDryRun && command.Flags().Lookup("dry-run") == nil {
		command.Flags().Bool("dry-run", false, "report what the command would change and send no request")
	}

	// The MCP catalog guards destructive tools with its own confirm property.
	// A CLI --yes would let a model satisfy that guard by passing a flag, so
	// the flag stays off the tool schema.
	markMCPSkip(command.Flags(), "yes")
}

// confirmDestructive applies the shared destructive gate. --dry-run reports and
// stops, an interactive stdin asks once, and every other stdin (pipes, CI, the
// MCP executor) proceeds without a prompt so scripts never block.
//
// fetch is optional: when set, a dry run loads the resource the command would
// change and prints it. The caller proceeds only when the returned flag is
// true; a dry-run report and a declined prompt both stop the command.
func confirmDestructive(cmd *cobra.Command, action string, fetch func(context.Context) (any, error)) (bool, error) {
	gate, ok := destructiveGates[cmd]
	if !ok {
		return true, nil
	}

	if gate.ownsDryRun {
		if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
			return false, reportDryRun(cmd, action, fetch)
		}
	}

	if yes, _ := cmd.Flags().GetBool("yes"); yes || !isInteractive(cmd.InOrStdin()) {
		return true, nil
	}

	confirmed, err := promptConfirm(cmd.ErrOrStderr(), cmd.InOrStdin(), action)
	if err != nil {
		return false, err
	}

	if !confirmed {
		return false, errConfirmationDeclined
	}

	return true, nil
}

// reportDryRun prints what the command would change, fetching the target
// resource when the command can. A fetch that fails is reported as an error:
// the caller asked to inspect a resource the command could not read, and
// writing "would delete" over that would be a lie.
func reportDryRun(cmd *cobra.Command, action string, fetch func(context.Context) (any, error)) error {
	if fetch == nil {
		out.Log("Dry run: would " + action)

		return nil
	}

	value, err := fetch(cmd.Context())
	if err != nil {
		return fmt.Errorf("dry run: %w", err)
	}

	out.Print(value, "Dry run: would "+action)

	return nil
}

// isInteractive reports whether r is a terminal. A character device is not
// enough: /dev/null is one, and that is exactly what `go test` and a CI job
// attach to stdin. Only a real terminal prompts, so a pipe, a redirected
// stdin and the MCP executor never block on a question nobody can answer.
func isInteractive(r io.Reader) bool {
	file, ok := r.(*os.File)
	if !ok {
		return false
	}

	return term.IsTerminal(int(file.Fd()))
}

// promptConfirm asks for a y/N confirmation on the terminal. The prompt goes to
// stderr so stdout stays a single structured document.
func promptConfirm(stderr io.Writer, stdin io.Reader, action string) (bool, error) {
	if _, err := fmt.Fprintf(stderr, "About to %s [y/N]: ", action); err != nil {
		return false, fmt.Errorf("write confirmation prompt: %w", err)
	}

	line, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
