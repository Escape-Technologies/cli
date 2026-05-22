package out

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/log"
)

// PrintError prints an error stack trace
func PrintError(err error) {
	if escape.IsInvalidAPIKey(err) {
		fmt.Println("Error:")
		fmt.Printf("  %s\n", escape.InvalidAPIKeyMessage)
		fmt.Printf("  %s\n", escape.InvalidAPIKeyHint)
		if log.IsVerbose() {
			printError(err)
		}
		return
	}
	fmt.Println("Error:")
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
	fmt.Printf("  %s\n", errString)
	printError(parent)
}
