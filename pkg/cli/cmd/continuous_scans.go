package cmd

import (
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	continuousScanWatch            bool
	continuousScanCommitHash       string
	continuousScanTipCommitHash    string
	continuousScanTipRepositoryURL string
	continuousScanInitiator        string
)

var continuousScansCmd = &cobra.Command{
	Use:     "continuous-scans",
	Aliases: []string{"cs"},
	Short:   "Start continuous pentest runs (Beta)",
	Long: `Start Continuous Pentest Runs

Start a CONTINUOUS_PENTEST on a profile that already has continuous pentest
enabled. Use scans start for a full AI pentest on the same profile.

BETA: Continuous Pentesting is gated by the ENABLE_CONTINUOUS_PENTESTING
organization setting.`,
}

var continuousScansStartCmd = &cobra.Command{
	Use:   "start profile-id",
	Short: "Start a continuous pentest run (Beta)",
	Long: `Start Continuous Scan - Review Diffs Since The Last Continuous Run

Starts a CONTINUOUS_PENTEST for a profile that already has continuous set up.
The returned id is the scan id. Defaults the initiator to CI.

WATCH MODE:
  Use --watch to block until the run ends. Progress and exit codes follow the
  same contract as scans start --watch. --fail-on-severity requires --watch.`,
	Example: `  # Start a continuous run after deploy
  escape-cli continuous-scans start <profile-id>

  # Start and watch (recommended for CI/CD)
  escape-cli continuous-scans start <profile-id> --watch

  # Tip one watched repository at a specific commit
  escape-cli continuous-scans start <profile-id> \
    --continuous-commit-hash abc123 \
    --repository-url https://github.com/org/repo`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.ScanDetailed1{}) {
			return nil
		}

		if err := validateScanFailOnSeverity(continuousScanWatch); err != nil {
			return err
		}

		fillContinuousScanCommitFromEnv()

		initiator := v3.ENUMPROPERTIESINITIATOR_CI
		if continuousScanInitiator != "" {
			initiator = v3.ENUMPROPERTIESINITIATOR(continuousScanInitiator)
		}

		scan, err := escape.StartContinuousScan(
			cmd.Context(),
			args[0],
			continuousScanCommitHash,
			continuousScanTipCommitHash,
			continuousScanTipRepositoryURL,
			initiator,
		)
		if err != nil {
			return fmt.Errorf("unable to start continuous scan: %w", err)
		}

		started := "Continuous scan started\n  ID:   " + scan.GetId()
		if link := scan.GetLinks().ScanIssues; link != "" {
			started += "\n  View: " + out.LinkURL(link)
		}

		if continuousScanWatch && out.IsJSON() {
			return watchScan(cmd.Context(), scan.GetId(), watchJSONStatus)
		}

		out.Print(scan, started)
		if continuousScanWatch {
			return watchScan(cmd.Context(), scan.GetId(), watchJSONSilent)
		}

		return nil
	},
}

func fillContinuousScanCommitFromEnv() {
	if continuousScanCommitHash != "" {
		return
	}

	if hash := os.Getenv("GITHUB_SHA"); hash != "" {
		continuousScanCommitHash = hash
		return
	}

	if hash := os.Getenv("CI_COMMIT_SHA"); hash != "" {
		continuousScanCommitHash = hash
		return
	}

	if hash := os.Getenv("CIRCLE_SHA1"); hash != "" {
		continuousScanCommitHash = hash
		return
	}

	if hash := os.Getenv("COMMIT_HASH"); hash != "" {
		continuousScanCommitHash = hash
	}
}

func init() {
	continuousScansStartCmd.Flags().BoolVarP(&continuousScanWatch, "watch", "w", false, "watch scan progress until completion")
	markMCPSkip(continuousScansStartCmd.Flags(), "watch")
	continuousScansStartCmd.Flags().StringVar(&scanFailOnSeverity, "fail-on-severity", "", failOnSeverityUsage)
	markMCPSkip(continuousScansStartCmd.Flags(), "fail-on-severity")
	continuousScansStartCmd.Flags().StringVar(&continuousScanCommitHash, "commit-hash", "", "indicative git commit SHA stored as a tag (auto-detected in CI/CD)")
	continuousScansStartCmd.Flags().StringVar(&continuousScanTipCommitHash, "continuous-commit-hash", "", "tip commit SHA for one watched repository")
	continuousScansStartCmd.Flags().StringVar(&continuousScanTipRepositoryURL, "repository-url", "", "watched repository URL for --continuous-commit-hash when more than one repo is watched")
	continuousScansStartCmd.Flags().StringVar(&continuousScanInitiator, "initiator", "", "who started the run: CI (default), MERGE, MANUAL, or SCHEDULED")

	continuousScansCmd.AddCommand(continuousScansStartCmd)
	rootCmd.AddCommand(continuousScansCmd)
}
