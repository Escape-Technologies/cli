package cmd

import (
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display CLI version information",
	Long: `Display Version Information

Shows the current version of the Escape CLI, including version number, commit hash,
and build date. Use this to verify your installation and check for updates.`,
	Example: `  # Show version
  escape-cli version

  # Check version in JSON format
  escape-cli version -o json`,
	Run: func(_ *cobra.Command, _ []string) {
		v := version.GetVersion()
		out.Print(v, v.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
