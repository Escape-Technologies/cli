package main

import (
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
