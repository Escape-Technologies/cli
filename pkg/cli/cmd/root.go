// Package cmd implements the CLI commands
package cmd

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmdVerbose int
var rootCmdOutputStr string

var rootCmd = &cobra.Command{
	Use:   "escape-cli",
	Short: "CLI to interact with Escape API",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if rootCmdVerbose > 0 { //nolint:mnd
			log.SetLevel(logrus.InfoLevel)
		}
		if rootCmdVerbose > 1 { //nolint:mnd
			log.SetLevel(logrus.DebugLevel)
		}
		if rootCmdVerbose > 2 { //nolint:mnd
			log.SetLevel(logrus.TraceLevel)
		}
		if rootCmdVerbose > 3 { //nolint:mnd
			escape.Debug = true
		}
		log.Info("Verbose mode: %d", rootCmdVerbose)
		err := out.SetOutput(rootCmdOutputStr)
		if err != nil {
			return fmt.Errorf("failed to set output format: %w", err)
		}
		return nil
	},
	PostRun: func(_ *cobra.Command, _ []string) {
		log.Trace("Main cli done, exiting")
	},

	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().CountVarP(&rootCmdVerbose, "verbose", "v", "enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOutputStr, "output", "o", "pretty", "output format (pretty|json|yaml)")
}

// Execute the CLI
func Execute(ctx context.Context) error {
	cmd, err := rootCmd.ExecuteContextC(ctx)
	if err != nil {
		return fmt.Errorf("command %s failed: %w", cmd.Name(), err)
	}
	return nil
}
