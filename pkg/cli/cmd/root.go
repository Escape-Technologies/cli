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

func getHelpTemplate(colorEnabled bool) string {
	logo := asciiLogo
	header := asciiHeader
	if colorEnabled {
		logo = "\x1b[38;2;6;226;183m" + asciiLogo + "\x1b[0m"
		header = "\x1b[38;2;6;226;183m" + asciiHeader + "\x1b[0m"
	}

	return logo + "\n" + header + `

{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}{{if .HasAvailableSubCommands}}

Command Categories:
  Scanning:       scans     - Run security scans and view results
  Security:       issues    - Manage security vulnerabilities
  Monitoring:     problems  - View scan problems and failures
  Configuration:  profiles  - Configure scan targets and settings
  Assets:         assets    - Manage your API inventory
  Infrastructure: locations - Deploy private scanning locations
  Organization:   audit     - Review activity logs and events
  Customization:  custom-rules, tags - Extend and organize
{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

Environment Variables:
  ESCAPE_API_KEY              - Your API key for authentication
  ESCAPE_COLOR_DISABLED       - Disable colored output (set to \"true\")
  
  CI/CD Auto-Detection (commit information):
  - GitHub Actions: GITHUB_SHA, GITHUB_REF_NAME, GITHUB_ACTOR
  - GitLab CI: CI_COMMIT_SHA, CI_COMMIT_REF_NAME, GITLAB_USER_EMAIL
  - CircleCI: CIRCLE_SHA1, CIRCLE_BRANCH, CIRCLE_USERNAME

For additional information, see: https://docs.escape.tech/documentation/tooling/cli
`
}

var rootCmd = &cobra.Command{
	Use:   "escape-cli",
	Short: "Escape CLI - Comprehensive API Security Testing Platform",
	Long: `Escape CLI - Comprehensive API Security Testing Platform

Escape is the most advanced API security platform, helping you discover, test,
and secure your APIs with cutting-edge DAST (Dynamic Application Security Testing)
capabilities.

Use this CLI to:
  • Start security scans on REST, GraphQL, and Web APIs
  • Monitor and track security issues across your API ecosystem
  • Manage security profiles, assets, and test configurations
  • Review audit logs and security events
  • Deploy private scanning locations for internal APIs

For more information, see: https://docs.escape.tech/documentation/tooling/cli`,
	PersistentPreRunE: func(c *cobra.Command, _ []string) error {
		version.WarnIfNotLatestVersion(c.Context())

		verbosityFrom := "command line argument"
		if envVerbosity := env.GetVerbosity(); envVerbosity > rootCmdVerbose {
			rootCmdVerbose = envVerbosity
			verbosityFrom = "environment variable ESCAPE_VERBOSITY"
		}

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
		log.Info("Verbose mode: %d from %s", rootCmdVerbose, verbosityFrom)
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
	rootCmd.PersistentFlags().CountVarP(&rootCmdVerbose, "verbose", "v", "verbose output: -v (info), -vv (debug), -vvv (trace), -vvvv (http debug)")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOutputStr, "output", "o", "pretty", "output format: pretty (human-readable tables), json (machine-readable), yaml (configuration files)")

	isColorDisabled := env.GetColorPreference()
	helpTemplate := getHelpTemplate(!isColorDisabled)
	rootCmd.SetHelpTemplate(helpTemplate)
}

// Execute the CLI
func Execute(ctx context.Context) error {
	isColorDisabled := env.GetColorPreference()

	if isColorDisabled {
		out.DisableColor()
	}
	cmd, err := rootCmd.ExecuteContextC(ctx)
	if err != nil {
		return fmt.Errorf("command %s failed: %w", cmd.Name(), err)
	}
	return nil
}
