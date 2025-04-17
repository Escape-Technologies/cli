package cmd

import (
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of the CLI",
	Run: func(_ *cobra.Command, _ []string) {
		v := version.GetVersion()
		out.Print(v, v.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
