package cli

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		v := version.GetVersion()
		print(v, func() {
			fmt.Println(v.String())
		})
	},
}
