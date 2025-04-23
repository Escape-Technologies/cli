package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/spf13/cobra"
)

var scansCmd = &cobra.Command{
	Use:     "scans",
	Aliases: []string{"sc", "scan"},
	Short:   "View scans results",
}

var scansListCmd = &cobra.Command{
	Use:     "list application-id",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all scans of an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		applicationID := args[0]
		scans, next, err := escape.ListScans(cmd.Context(), applicationID, "")
		if err != nil {
			return fmt.Errorf("unable to list scans: %w", err)
		}
		out.Table(scans, func() []string {
			res := []string{"ID\tSTATUS\tCREATED AT\tPROGRESS"}
			for _, scan := range scans {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%f", scan.GetId(), scan.GetStatus(), scan.GetCreatedAt(), scan.GetProgressRatio()))
			}
			return res
		})
		for next != "" {
			scans, next, err = escape.ListScans(cmd.Context(), applicationID, next)
			if err != nil {
				return fmt.Errorf("unable to list scans: %w", err)
			}
			out.Table(scans, func() []string {
				res := []string{}
				for _, scan := range scans {
					res = append(res, fmt.Sprintf("%s\t%s\t%s\t%f", scan.GetId(), scan.GetStatus(), scan.GetCreatedAt(), scan.GetProgressRatio()))
				}
				return res
			})
		}
		return nil
	},
}

func extractCommitDataFromEnv() {
	log.Trace("Extracting commit data from environment variables")
	if scanStartCmdCommitHash != "" ||
		scanStartCmdCommitLink != "" ||
		scanStartCmdCommitBranch != "" ||
		scanStartCmdCommitAuthor != "" ||
		scanStartCmdCommitAuthorProfilePictureLink != "" {
		log.Info("Commit data already set, skipping environment variables extraction")
		return
	}

	if os.Getenv("GITHUB_SHA") != "" {
		log.Info("Extracting commit data from GitHub environment variables")
		// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
		scanStartCmdCommitHash = os.Getenv("GITHUB_SHA")
		scanStartCmdCommitBranch = os.Getenv("GITHUB_REF_NAME")
		scanStartCmdCommitAuthor = os.Getenv("GITHUB_ACTOR")
		scanStartCmdCommitAuthorProfilePictureLink = "https://avatars.githubusercontent.com/u/" + os.Getenv("GITHUB_ACTOR_ID") + "?v=4"
		scanStartCmdCommitLink = os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + scanStartCmdCommitHash
		return
	}
	if os.Getenv("GITLAB_CI") != "" {
		log.Info("Extracting commit data from GitLab environment variables")
		// https://docs.gitlab.com/ci/variables/predefined_variables/
		scanStartCmdCommitHash = os.Getenv("CI_COMMIT_SHA")
		scanStartCmdCommitBranch = os.Getenv("CI_COMMIT_REF_NAME")
		scanStartCmdCommitAuthor = os.Getenv("GITLAB_USER_EMAIL")
		scanStartCmdCommitLink = os.Getenv("CI_PROJECT_URL") + "/-/commit/" + scanStartCmdCommitHash
		return
	}
	if os.Getenv("CIRCLE_SHA1") != "" {
		log.Info("Extracting commit data from CircleCI environment variables")
		// https://circleci.com/docs/variables/#built-in-environment-variables
		scanStartCmdCommitHash = os.Getenv("CIRCLE_SHA1")
		scanStartCmdCommitBranch = os.Getenv("CIRCLE_BRANCH")
		scanStartCmdCommitAuthor = os.Getenv("CIRCLE_USERNAME")
		return
	}
	if os.Getenv("COMMIT_HASH") != "" {
		log.Info("Extracting commit data from local environment variables")
		scanStartCmdCommitHash = os.Getenv("COMMIT_HASH")
		scanStartCmdCommitLink = os.Getenv("COMMIT_LINK")
		scanStartCmdCommitBranch = os.Getenv("COMMIT_BRANCH")
		scanStartCmdCommitAuthor = os.Getenv("COMMIT_AUTHOR")
		return
	}

	log.Info("No commit data found in environment variables")
}

func debugCommitData() {
	log.Debug("Commit Hash: %s", scanStartCmdCommitHash)
	log.Debug("Commit Link: %s", scanStartCmdCommitLink)
	log.Debug("Commit Branch: %s", scanStartCmdCommitBranch)
	log.Debug("Commit Author: %s", scanStartCmdCommitAuthor)
	log.Debug("Commit AuthorProfilePictureLink: %s", scanStartCmdCommitAuthorProfilePictureLink)
}

var scanStartCmdCommitHash = ""
var scanStartCmdCommitLink = ""
var scanStartCmdCommitBranch = ""
var scanStartCmdCommitAuthor = ""
var scanStartCmdCommitAuthorProfilePictureLink = ""
var scanStartCmdConfigurationOverride = ""
var scanStartCmdWatch bool
var scanStartCmd = &cobra.Command{
	Use: "start application-id",
	Example: `escape-cli scans start 00000000-0000-0000-0000-000000000000
escape-cli scans start 00000000-0000-0000-0000-000000000000 --commit-hash 1234567890
escape-cli scans start 00000000-0000-0000-0000-000000000000 --override '{"scan": {"read_only": true}}'`,
	Args:  cobra.ExactArgs(1),
	Short: "Start a new scan of an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		override := v2.NullableCreateApplicationRequestConfiguration{}
		if scanStartCmdConfigurationOverride != "" {
			err := override.UnmarshalJSON([]byte(scanStartCmdConfigurationOverride))
			if err != nil {
				return fmt.Errorf("unable to unmarshal configuration override: %w", err)
			}
			ovr, _ := override.MarshalJSON()
			log.Info("Configuration override: %s", string(ovr))
		}
		extractCommitDataFromEnv()
		debugCommitData()
		scan, err := escape.StartScan(
			cmd.Context(),
			args[0],
			scanStartCmdCommitHash,
			scanStartCmdCommitLink,
			scanStartCmdCommitBranch,
			scanStartCmdCommitAuthor,
			scanStartCmdCommitAuthorProfilePictureLink,
			override.Get(),
		)
		if err != nil {
			return fmt.Errorf("unable to start scan: %w", err)
		}
		out.Print(scan, "Scan started: "+scan.GetId())
		if scanStartCmdWatch {
			err := watchScan(cmd.Context(), scan.GetId())
			if err != nil {
				return fmt.Errorf("unable to watch scan: %w", err)
			}
		}
		return nil
	},
}

func watchScan(ctx context.Context, scanID string) error {
	ch, err := escape.WatchScan(ctx, scanID)
	if err != nil {
		return fmt.Errorf("unable to watch scan: %w", err)
	}
	firstEvent := <-ch
	out.Table(firstEvent, func() []string {
		return []string{
			"STATUS\tPROGRESS\tCREATED AT\tLEVEL\tTITLE\tDESCRIPTION",
			fmt.Sprintf("%s\t%f\t%s\t%s\t%s\t%s",
				firstEvent.Status,
				firstEvent.ProgressRatio,
				firstEvent.CreatedAt.Format(time.RFC3339),
				firstEvent.Level,
				firstEvent.Title,
				firstEvent.Description,
			),
		}
	})
	for event := range ch {
		out.Table(event, func() []string {
			return []string{
				fmt.Sprintf("%s\t%f\t%s\t%s\t%s\t%s",
					event.Status,
					event.ProgressRatio,
					event.CreatedAt.Format(time.RFC3339),
					event.Level,
					event.Title,
					event.Description,
				),
			}
		})
	}
	return nil
}

var scanWatchCmd = &cobra.Command{
	Use:     "watch scan-id",
	Example: `escape-cli scans watch 00000000-0000-0000-0000-000000000000`,
	Args:    cobra.ExactArgs(1),
	Short:   "Bind the current terminal to a scan, listen for events and print them to the terminal. Exit when the scan is done.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return watchScan(cmd.Context(), args[0])
	},
}

var scanGetCmd = &cobra.Command{
	Use:     "get scan-id",
	Aliases: []string{"describe"},
	Args:    cobra.ExactArgs(1),
	Short:   "Return the scan status",
	RunE: func(cmd *cobra.Command, args []string) error {
		scan, err := escape.GetScan(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get scan: %w", err)
		}
		out.Table(scan, func() []string {
			res := []string{"ID\tSTATUS\tCREATED AT\tPROGRESS"}
			res = append(res, fmt.Sprintf("%s\t%s\t%s\t%f", scan.GetId(), scan.GetStatus(), scan.GetCreatedAt(), scan.GetProgressRatio()))
			return res
		})
		return nil
	},
}

var scanIssuesCmd = &cobra.Command{
	Use:     "issues scan-id",
	Aliases: []string{"results", "res", "result", "issues", "iss"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all issues of a scan",
	RunE: func(cmd *cobra.Command, args []string) error {
		issues, err := escape.GetScanIssues(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get scan issues: %w", err)
		}
		out.Table(issues, func() []string {
			res := []string{"ID\tSEVERITY\tTYPE\tCATEGORY\tNAME\tIGNORED\tURL"}
			for _, i := range issues {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%t\t%s",
					i.GetId(),
					i.GetSeverity(),
					i.GetType(),
					i.GetCategory(),
					i.GetName(),
					i.GetIgnored(),
					i.GetPlatformUrl(),
				))
			}
			return res
		})
		return nil
	},
}

var scanDownloadCmd = &cobra.Command{
	Use:     "download scan-id archive-path",
	Example: "escape-cli scans download 00000000-0000-0000-0000-000000000000 ./exchanges.zip",
	Aliases: []string{"dl", "zip"},
	Args:    cobra.ExactArgs(2), //nolint:mnd
	Short:   "Download a scan result exchange archive (zip export)",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.DownloadScanExchangesZip(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("unable to download scan exchanges zip: %w", err)
		}
		return nil
	},
}

func init() {
	scansCmd.AddCommand(scansListCmd)
	scanStartCmd.PersistentFlags().BoolVarP(&scanStartCmdWatch, "watch", "w", false, "watch for events")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitHash, "commit-hash", "", "commit hash")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitLink, "commit-link", "", "commit link")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitBranch, "commit-branch", "", "commit branch")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitAuthor, "commit-author", "", "commit author")
	scanStartCmd.PersistentFlags().StringVar(&scanStartCmdCommitAuthorProfilePictureLink, "profile-picture", "", "commit author profile picture link")
	scanStartCmd.PersistentFlags().StringVarP(&scanStartCmdConfigurationOverride, "override", "c", "", "configuration override")
	scansCmd.AddCommand(scanStartCmd)
	scansCmd.AddCommand(scanGetCmd)
	scansCmd.AddCommand(scanDownloadCmd)
	scansCmd.AddCommand(scanIssuesCmd)
	scansCmd.AddCommand(scanWatchCmd)
	rootCmd.AddCommand(scansCmd)
}
