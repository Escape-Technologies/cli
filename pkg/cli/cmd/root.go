package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmdVerbose bool
var rootCmdOutputStr string

var rootCmd = &cobra.Command{
	Use:   "escape-cli",
	Short: "CLI to interact with Escape API",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if rootCmdVerbose {
			log.SetLevel(logrus.TraceLevel)
		}
		err := out.SetOutput(rootCmdOutputStr)
		if err != nil {
			return fmt.Errorf("failed to set output format: %w", err)
		}
		return nil
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		log.Trace("Main cli done, exiting")
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdVerbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOutputStr, "output", "o", "pretty", "output format (pretty|json|yaml)")
}

func Execute() error {
	return rootCmd.Execute()
}
