package main

import (
	"os"

	"github.com/Escape-Technologies/cli/pkg/cli"
	"github.com/Escape-Technologies/cli/pkg/log"
)

func main() {
	if err := cli.Run(); err != nil {
		log.Error("Error running cli: %s", err.Error())
		os.Exit(1)
	}
}
