package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

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
	Short:   "List scans",
	Long: `List all scans of an application.

Example output:
ID                                      STATUS      CREATED AT                           PROGRESS
00000000-0000-0000-0000-000000000001    FINISHED    2025-02-05 08:34:47.541 +0000 UTC    0.000000
00000000-0000-0000-0000-000000000002    FINISHED    2025-02-02 08:27:23.919 +0000 UTC    0.000000
00000000-0000-0000-0000-000000000003    FINISHED    2025-01-31 18:35:48.477 +0000 UTC    0.000000
00000000-0000-0000-0000-000000000004    FINISHED    2025-01-30 08:25:49.656 +0000 UTC    0.000000`,
	Example: `escape-cli scans list 00000000-0000-0000-0000-000000000000`,
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
	Short: "Start a scan",
	Long:  "Start a new scan of an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		override := v2.NullableFrontendConfiguration{}
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
	var status *v2.ListScans200ResponseDataInner
	isFirst := true
	for event := range ch {
		if event == nil {
			continue
		}
		status = event
		out.Table(event, func() []string {
			res := []string{}
			if isFirst {
				res = append(res, "STATUS\tPROGRESS")
				isFirst = false
			}
			res = append(
				res,
				fmt.Sprintf("%s\t%d%%", event.Status, int(event.ProgressRatio*100)), //nolint:mnd
			)
			return res
		})
	}
	if status == nil {
		return errors.New("unable to watch scan")
	} else if status.Status == v2.ENUME48DD51FE8A350A4154904ABF16320D7_CANCELED {
		out.Log("Scan canceled")
	} else if status.Status == v2.ENUME48DD51FE8A350A4154904ABF16320D7_FAILED {
		out.Log("Scan failed")
	} else {
		err := printScanIssues(ctx, scanID)
		if err != nil {
			return fmt.Errorf("unable to fetch scan issues: %w", err)
		}
	}
	return nil
}

var scanWatchCmd = &cobra.Command{
	Use:     "watch scan-id",
	Example: `escape-cli scans watch 00000000-0000-0000-0000-000000000000`,
	Args:    cobra.ExactArgs(1),
	Short:   "Watch a scan",
	Long:    "Bind the current terminal to a scan, listen for events and print them to the terminal. Exit when the scan is done.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return watchScan(cmd.Context(), args[0])
	},
}

var scanGetCmd = &cobra.Command{
	Use:     "get scan-id",
	Aliases: []string{"describe"},
	Args:    cobra.ExactArgs(1),
	Short:   "Get scan status",
	Long: `Return the scan status.

Example output:
ID                                      STATUS      CREATED AT                           PROGRESS
00000000-0000-0000-0000-000000000001    FINISHED    2024-11-27 08:06:59.576 +0000 UTC    1.000000`,
	Example: `escape-cli scans get 00000000-0000-0000-0000-000000000000`,
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

func printScanIssues(ctx context.Context, scanID string) error {
	issues, err := escape.GetScanIssues(ctx, scanID)
	if err != nil {
		return fmt.Errorf("unable to fetch scan issues: %w", err)
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
}

var scanIssuesCmd = &cobra.Command{
	Use:     "issues scan-id",
	Aliases: []string{"results", "res", "result", "issues", "iss"},
	Args:    cobra.ExactArgs(1),
	Short:   "List scan issues",
	Long: `List all issues of a scan.

Example output:
ID                                      SEVERITY    TYPE    CATEGORY                  NAME                                         IGNORED    URL
00000000-0000-0000-0000-000000000001    MEDIUM      API     PROTOCOL                  Insecure Security Policy header              false      https://app.escape.tech/scan/00000000-0000-0000-0000-000000000005/issues/00000000-0000-0000-0000-000000000001/overview/
00000000-0000-0000-0000-000000000002    LOW         API     INFORMATION_DISCLOSURE    Debug mode enabled                           false      https://app.escape.tech/scan/00000000-0000-0000-0000-000000000005/issues/00000000-0000-0000-0000-000000000002/overview/`,
	Example: `escape-cli scans issues 00000000-0000-0000-0000-000000000000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := printScanIssues(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get scan issues: %w", err)
		}
		return nil
	},
}

var scanDownloadCmd = &cobra.Command{
	Use:     "download scan-id archive-path",
	Example: "escape-cli scans download 00000000-0000-0000-0000-000000000000 ./exchanges.zip",
	Aliases: []string{"dl", "zip"},
	Args:    cobra.ExactArgs(2), //nolint:mnd
	Short:   "Download scan results",
	Long:    "Download a scan result exchange archive (zip export)",
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
