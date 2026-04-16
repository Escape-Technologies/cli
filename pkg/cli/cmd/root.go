// Package cmd implements the CLI commands
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
var rootCmdInputSchema bool

const escapeBrandColor = "\x1b[38;2;6;226;183m"
const boldYellowColor = "\x1b[1;33m"
const dimColor = "\x1b[90m"
const resetColor = "\x1b[0m"

const startupUpdateTimeout = 200 * time.Millisecond

var rootCmd = &cobra.Command{
	Use:   "escape-cli",
	Short: buildHelpHeader(),
	Long: `Replace legacy scanners and manual offensive security processes with AI agents
that discover, test, and remediate directly in your engineering workflows.

QUICK START:

  $ escape-cli profiles list                           List scan profiles
  $ escape-cli scans start <profile-id> --watch        Launch a scan
  $ escape-cli emails list --email <scan-inbox>        Inspect scan inbox emails
  $ escape-cli issues list --severity HIGH,CRITICAL    Review findings
  $ escape-cli asm trigger                             Trigger attack surface discovery

AGENT INTEGRATION:

  $ escape-cli <command> -o json            Machine-readable output
  $ escape-cli <command> -o schema          Output JSON Schema
  $ escape-cli <command> --input-schema     Input JSON Schema

DOCUMENTATION:

  https://docs.escape.tech/documentation/tooling/cli`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		verbosityFrom := "command line argument"
		if envVerbosity := env.GetVerbosity(); envVerbosity > rootCmdVerbose {
			rootCmdVerbose = envVerbosity
			verbosityFrom = "environment variable ESCAPE_VERBOSITY"
		}

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
		log.Info("escape-cli version: %s", version.GetVersion().LogString())
		err := out.SetOutput(rootCmdOutputStr)
		if err != nil {
			return fmt.Errorf("failed to set output format: %w", err)
		}
		out.SetInputSchema(rootCmdInputSchema)
		printStartupHeader(cmd.Context())
		return nil
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
	PostRun: func(_ *cobra.Command, _ []string) {
		log.Trace("Main cli done, exiting")
	},

	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().CountVarP(&rootCmdVerbose, "verbose", "v", "verbose output: -v (debug), -vv (trace), -vvv (http/raw debug)")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOutputStr, "output", "o", "pretty", "output format: pretty (human-readable tables), json (machine-readable), yaml (configuration files), schema (JSON Schema for AI agents)")
	rootCmd.PersistentFlags().BoolVar(&rootCmdInputSchema, "input-schema", false, "print JSON Schema for stdin input format (for create/update commands)")
	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + `
COMMAND CATEGORIES:
  Offensive Testing:    scans, profiles, authentications   Scan targets and configurations
  Scan Inbox:           emails                             Inspect scan inbox messages
  Attack Surface:       asm, assets                        Discovery and inventory
  Findings:             issues, problems, events           Vulnerabilities and diagnostics
  Infrastructure:       locations                          Private scanning locations
  Organization:         users, roles, projects, audit      Access control and audit trail
  Automation:           workflows, jobs                    CI/CD triggers and exports
  Customization:        custom-rules, tags, integrations   Rules, labels, and integrations

ENVIRONMENT VARIABLES:
  ESCAPE_API_KEY              API key for authentication
  ESCAPE_APPLICATION_URL      Platform URL (default: https://public.escape.tech)
  ESCAPE_COLOR_DISABLED       Set to "true" to disable colored output
  HTTP_PROXY, HTTPS_PROXY     Proxy configuration

  CI/CD commit metadata is auto-detected from GitHub Actions, GitLab CI, and CircleCI.
`)
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

func buildHelpHeader() string {
	v := version.GetVersion()
	return brandText("Escape CLI "+v.DisplayVersion()) + "\n" + dimText("Offensive Security Engineering Platform")
}

func printStartupHeader(ctx context.Context) {
	if !isPrettyOutput() {
		return
	}

	v := version.GetVersion()

	logo := [3]string{
		" ▐▛██▜▌ ",
		"▗█▛  ▜█▖",
		" ▝▜██▛▘ ",
	}

	fmt.Fprintf(os.Stderr, "%s  %s\n",
		brandText(logo[0]),
		brandText("Escape CLI"),
	)
	fmt.Fprintf(os.Stderr, "%s  %s\n",
		brandText(logo[1]),
		dimText("Offensive Security Engineering Platform"),
	)

	versionLine := dimText(v.DisplayVersion())
	if upgrade := resolveUpgrade(ctx); upgrade != "" {
		versionLine += "  " + upgrade
	}
	fmt.Fprintf(os.Stderr, "%s  %s\n", brandText(logo[2]), versionLine)

	fmt.Fprintln(os.Stderr)
}

func resolveUpgrade(ctx context.Context) string {
	checkCtx, cancel := context.WithTimeout(ctx, startupUpdateTimeout)
	defer cancel()

	type result struct {
		update *version.UpdateInfo
		method version.InstallMethod
	}

	ch := make(chan result, 1)
	go func() {
		ch <- result{
			update: version.CheckForUpdate(checkCtx),
			method: version.GetInstallInfo().Method,
		}
	}()

	select {
	case r := <-ch:
		if r.update == nil || !r.update.Available {
			return ""
		}
		cmd := version.UpgradeCommand(r.method, r.update.Latest)
		if cmd != "" {
			return boldYellowText("Update v" + r.update.Latest + " · " + cmd)
		}
		return boldYellowText("Update available: v" + r.update.Latest)
	case <-checkCtx.Done():
		return ""
	}
}

func isPrettyOutput() bool {
	output := strings.ToLower(strings.TrimSpace(rootCmdOutputStr))
	return output == "" || output == "pretty"
}

func brandText(value string) string {
	return styleText(value, escapeBrandColor)
}

func boldYellowText(value string) string {
	return styleText(value, boldYellowColor)
}

func dimText(value string) string {
	return styleText(value, dimColor)
}

func styleText(value string, color string) string {
	if env.GetColorPreference() {
		return value
	}

	return color + value + resetColor
}
