// Package main is the entry point for the CLI
package main

import (
	"context"
	"os"

	"github.com/Escape-Technologies/cli/pkg/cli"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
)

func main() {
	err := cli.Run(context.Background())
	if err != nil {
		out.PrintError(err)
		os.Exit(1)
	}
}
