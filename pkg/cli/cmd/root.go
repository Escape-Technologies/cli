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
	Use:   "escape-cli",
	Short: asciiLogo + "\n" + asciiHeader,
	Long: `Escape CLI - Your Gateway to Comprehensive API Security Testing

Escape is the most advanced API security platform, helping you discover, test,
and secure your APIs with cutting-edge DAST (Dynamic Application Security Testing)
capabilities.

🎯 WHAT YOU CAN DO:
  • Start security scans on your REST, GraphQL, and Web APIs
  • Monitor and track security issues across your API ecosystem
  • Manage security profiles, assets, and test configurations
  • Review audit logs and security events
  • Deploy private scanning locations for internal APIs

📚 GETTING STARTED:
  1. First time? Check your version:
     $ escape-cli version
  
  2. List your API profiles:
     $ escape-cli profiles list
  
  3. Start a security scan:
     $ escape-cli scans start <profile-id> --watch
  
  4. Review discovered issues:
     $ escape-cli issues list --severity HIGH,CRITICAL

💡 PRO TIPS:
  • Use -v for debug, -vv for trace, -vvv for http/raw debug
  • Output in JSON or YAML with -o json or -o yaml
  • Most list commands support powerful filtering options
  • Use --watch flag when starting scans for real-time updates

🔗 RESOURCES:
  • Documentation: https://docs.escape.tech/documentation/tooling/cli
  • API Reference: https://public.escape.tech/v3
  • Support: https://escape.tech/contact`,
	PersistentPreRunE: func(c *cobra.Command, _ []string) error {
		version.WarnIfNotLatestVersion(c.Context())

		verbosityFrom := "command line argument"
		if envVerbosity := env.GetVerbosity(); envVerbosity > rootCmdVerbose {
			rootCmdVerbose = envVerbosity
			verbosityFrom = "environment variable ESCAPE_VERBOSITY"
		}

		// Verbosity levels per Health Monitoring for Private Locations docs:
		// 0 = default (minimal), 1 = debug, 2 = trace, 3 = trace + http/raw
		if rootCmdVerbose > 0 { //nolint:mnd
			log.SetLevel(logrus.DebugLevel)
		}
		if rootCmdVerbose > 1 { //nolint:mnd
			log.SetLevel(logrus.TraceLevel)
		}
		if rootCmdVerbose > 2 { //nolint:mnd
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
	rootCmd.PersistentFlags().CountVarP(&rootCmdVerbose, "verbose", "v", "verbose output: -v (debug), -vv (trace), -vvv (http/raw debug)")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOutputStr, "output", "o", "pretty", "output format: pretty (human-readable tables), json (machine-readable), yaml (configuration files)")
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + `
COMMAND CATEGORIES:
  Scanning:       scans     - Run security scans and view results
  Security:       issues    - Manage security vulnerabilities
  Monitoring:     problems  - View scan problems and failures
  Configuration:  profiles  - Configure scan targets and settings
  Assets:         assets    - Manage your API inventory
  Infrastructure: locations - Deploy private scanning locations
  Organization:   audit     - Review activity logs and events
  Customization:  custom-rules, tags - Extend and organize

ENVIRONMENT VARIABLES:
  ESCAPE_APPLICATION_URL      - Escape platform URL (default: https://public.escape.tech)
  ESCAPE_API_KEY              - Your API key for authentication
  NO_COLOR                    - Disable colored output (set to any value)
  HTTP_PROXY, HTTPS_PROXY     - Configure proxy settings
  
  CI/CD Auto-Detection (commit information):
  - GitHub Actions: GITHUB_SHA, GITHUB_REF_NAME, GITHUB_ACTOR
  - GitLab CI: CI_COMMIT_SHA, CI_COMMIT_REF_NAME, GITLAB_USER_EMAIL
  - CircleCI: CIRCLE_SHA1, CIRCLE_BRANCH, CIRCLE_USERNAME

For additional information, see the documentation: 
https://docs.escape.tech/documentation/tooling/cli
`)

	isColorDisabled := env.GetColorPreference()
	if !isColorDisabled {
		rootCmd.Short = "\x1b[38;2;6;226;183m" + asciiLogo + "\x1b[0m" + "\n" + "\x1b[38;2;6;226;183m" + asciiHeader + "\x1b[0m"
	}
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
