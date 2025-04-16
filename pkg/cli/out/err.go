package out

import (
	"errors"
	"fmt"
	"strings"
)

func PrintError(err error) {
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
