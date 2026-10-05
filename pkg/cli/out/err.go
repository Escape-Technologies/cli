package out

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/log"
)

// PrintError writes the error chain to stderr.
// stdout stays the command document. A failed `scans watch -o json` prints
// one scan document there, then main exits 1 with this text on stderr.
func PrintError(err error) {
	if escape.IsInvalidAPIKey(err) {
		fmt.Fprintln(os.Stderr, "Error:")
		fmt.Fprintf(os.Stderr, "  %s\n", escape.InvalidAPIKeyMessage)
		fmt.Fprintf(os.Stderr, "  %s\n", escape.InvalidAPIKeyHint)
		if log.IsVerbose() {
			printError(err)
		}

		return
	}

	fmt.Fprintln(os.Stderr, "Error:")
	printError(err)
}

func printError(err error) {
	if err == nil {
		return
	}

	parent := errors.Unwrap(err)
	errString := err.Error()
	if parent != nil {
		errString = strings.ReplaceAll(errString, parent.Error(), "")
	}

	fmt.Fprintf(os.Stderr, "  %s\n", errString)
	printError(parent)
}
