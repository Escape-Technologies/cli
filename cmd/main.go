package main

import (
	"os"

	"github.com/Escape-Technologies/cli/pkg/cli"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
)

func main() {
	err := cli.Run()
	if err != nil {
		out.PrintError(err)
		os.Exit(1)
	}
}
