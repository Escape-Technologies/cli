package cli

import (
	"github.com/Escape-Technologies/cli/pkg/cli/cmd"
)

func Run() error {
	return cmd.Execute()
}
