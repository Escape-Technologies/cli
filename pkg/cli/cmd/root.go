// Package cmd implements the CLI commands
package cmd

import (
	"context"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/env"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/Escape-Technologies/cli/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmdVerbose int
var rootCmdOutputStr string

var asciiLogo = `
 ██████████  █████████    █████████    █████████   ███████████  ██████████      █████████  █████       █████
░░███░░░░░█ ███░░░░░███  ███░░░░░███  ███░░░░░███ ░░███░░░░░███░░███░░░░░█     ███░░░░░███░░███       ░░███ 
 ░███  █ ░ ░███    ░░░  ███     ░░░  ░███    ░███  ░███    ░███ ░███  █ ░     ███     ░░░  ░███        ░███ 
 ░██████   ░░█████████ ░███          ░███████████  ░██████████  ░██████      ░███          ░███        ░███ 
 ░███░░█    ░░░░░░░░███░███          ░███░░░░░███  ░███░░░░░░   ░███░░█      ░███          ░███        ░███ 
 ░███ ░   █ ███    ░███░░███     ███ ░███    ░███  ░███         ░███ ░   █   ░░███     ███ ░███      █ ░███ 
 ██████████░░█████████  ░░█████████  █████   █████ █████        ██████████    ░░█████████  ███████████ █████
░░░░░░░░░░  ░░░░░░░░░    ░░░░░░░░░  ░░░░░   ░░░░░ ░░░░░        ░░░░░░░░░░      ░░░░░░░░░  ░░░░░░░░░░░ ░░░░░                                                                                                        
`

var asciiHeader = "Escape CLI V3"

var rootCmd = &cobra.Command{
	Use: "escape-cli",
	Short: asciiLogo + "\n" + asciiHeader,
    PersistentPreRunE: func(c *cobra.Command, _ []string) error {
        version.WarnIfNotLatestVersion(c.Context())
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
		log.Info("escape-cli version: %s", version.GetVersion().String())
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
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + `
This CLI is based on the Escape API V3.

For additional information, see the documentation: 
https://docs.escape.tech/documentation/tooling/cli
`)

	noColor, err := env.GetColorPreference()
	if err == nil && !noColor {
		rootCmd.Short = "\x1b[38;2;6;226;183m" + asciiLogo + "\x1b[0m" + "\n" + "\x1b[38;2;6;226;183m" + asciiHeader + "\x1b[0m"
	}
}

// Execute the CLI
func Execute(ctx context.Context) error {
	noColor, err := env.GetColorPreference()
	if err != nil {
		return fmt.Errorf("failed to get color preference: %w", err)
	}
	if noColor {
		out.DisableColor()
	}
	cmd, err := rootCmd.ExecuteContextC(ctx)
	if err != nil {
		return fmt.Errorf("command %s failed: %w", cmd.Name(), err)
	}
	return nil
}
