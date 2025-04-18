package cmd

import (
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
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
			res := []string{"ID\tSTATUS\tCREATED AT\tPROGRESS RATIO"}
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

// TODO(quentin@escape.tech): var scanStartCmdConfigurationOverride *v2.CreateApplicationRequestConfiguration

func extractCommitDataFromEnv() {
	log.Trace("Extracting commit data from environment variables")
	if scanStartCmdCommitHash != nil ||
		scanStartCmdCommitLink != nil ||
		scanStartCmdCommitBranch != nil ||
		scanStartCmdCommitAuthor != nil ||
		scanStartCmdCommitAuthorProfilePictureLink != nil {
		log.Debug("Commit data already set, skipping environment variables extraction")
		return
	}

	if os.Getenv("GITHUB_SHA") != "" {
		log.Info("Extracting commit data from GitHub environment variables")
		// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
		localVarCommitHash := os.Getenv("GITHUB_SHA")
		scanStartCmdCommitHash = &localVarCommitHash
		localVarCommitBranch := os.Getenv("GITHUB_REF_NAME")
		scanStartCmdCommitBranch = &localVarCommitBranch
		localVarCommitAuthor := os.Getenv("GITHUB_ACTOR")
		scanStartCmdCommitAuthor = &localVarCommitAuthor
		localVarCommitAuthorProfilePictureLink := "https://avatars.githubusercontent.com/u/" + os.Getenv("GITHUB_ACTOR_ID") + "?v=4"
		scanStartCmdCommitAuthorProfilePictureLink = &localVarCommitAuthorProfilePictureLink
		localVarCommitLink := os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + localVarCommitHash
		scanStartCmdCommitLink = &localVarCommitLink
		return
	}
	if os.Getenv("GITLAB_CI") != "" {
		log.Info("Extracting commit data from GitLab environment variables")
		// https://docs.gitlab.com/ci/variables/predefined_variables/
		localVarCommitHash := os.Getenv("CI_COMMIT_SHA")
		scanStartCmdCommitHash = &localVarCommitHash
		localVarCommitBranch := os.Getenv("CI_COMMIT_REF_NAME")
		scanStartCmdCommitBranch = &localVarCommitBranch
		localVarCommitAuthor := os.Getenv("GITLAB_USER_EMAIL")
		scanStartCmdCommitAuthor = &localVarCommitAuthor
		localVarCommitLink := os.Getenv("CI_PROJECT_URL") + "/-/commit/" + localVarCommitHash
		scanStartCmdCommitLink = &localVarCommitLink
		return
	}
	if os.Getenv("CIRCLE_SHA1") != "" {
		log.Info("Extracting commit data from CircleCI environment variables")
		// https://circleci.com/docs/variables/#built-in-environment-variables
		localVarCommitHash := os.Getenv("CIRCLE_SHA1")
		scanStartCmdCommitHash = &localVarCommitHash
		localVarCommitBranch := os.Getenv("CIRCLE_BRANCH")
		scanStartCmdCommitBranch = &localVarCommitBranch
		localVarCommitAuthor := os.Getenv("CIRCLE_USERNAME")
		scanStartCmdCommitAuthor = &localVarCommitAuthor
		return
	}
	if os.Getenv("COMMIT_HASH") != "" {
		log.Info("Extracting commit data from local environment variables")
		localVarCommitHash := os.Getenv("COMMIT_HASH")
		scanStartCmdCommitHash = &localVarCommitHash
		localVarCommitLink := os.Getenv("COMMIT_LINK")
		scanStartCmdCommitLink = &localVarCommitLink
		localVarCommitBranch := os.Getenv("COMMIT_BRANCH")
		scanStartCmdCommitBranch = &localVarCommitBranch
		localVarCommitAuthor := os.Getenv("COMMIT_AUTHOR")
		scanStartCmdCommitAuthor = &localVarCommitAuthor
		return
	}

	log.Info("No commit data found in environment variables")
}

var scanStartCmdCommitHash *string
var scanStartCmdCommitLink *string
var scanStartCmdCommitBranch *string
var scanStartCmdCommitAuthor *string
var scanStartCmdCommitAuthorProfilePictureLink *string
var scanStartCmdWatch bool
var scanStartCmd = &cobra.Command{
	Use:   "start application-id",
	Args:  cobra.ExactArgs(1),
	Short: "Start a new scan of an application",
	RunE: func(cmd *cobra.Command, args []string) error {
		extractCommitDataFromEnv()
		scan, err := escape.StartScan(
			cmd.Context(),
			args[0],
			scanStartCmdCommitHash,
			scanStartCmdCommitLink,
			scanStartCmdCommitBranch,
			scanStartCmdCommitAuthor,
			scanStartCmdCommitAuthorProfilePictureLink,
			nil,
		)
		if err != nil {
			return fmt.Errorf("unable to start scan: %w", err)
		}
		out.Print(scan, "Scan started: "+scan.GetId())
		// TODO(quentin@escape.tech): watch scan
		return nil
	},
}

var scanGetCmd = &cobra.Command{
	Use:     "get scan-id",
	Aliases: []string{"describe", "results", "res", "result", "issues", "iss"},
	Args:    cobra.ExactArgs(1),
	Short:   "List all results (issues) of a scan",
	Run: func(_ *cobra.Command, _ []string) {
		// TODO(quentin@escape.tech): Implement this
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
	scanStartCmd.PersistentFlags().StringVarP(scanStartCmdCommitHash, "commit-hash", "c", "", "commit hash")
	scanStartCmd.PersistentFlags().StringVarP(scanStartCmdCommitLink, "commit-link", "l", "", "commit link")
	scanStartCmd.PersistentFlags().StringVarP(scanStartCmdCommitBranch, "commit-branch", "b", "", "commit branch")
	scanStartCmd.PersistentFlags().StringVarP(scanStartCmdCommitAuthor, "commit-author", "a", "", "commit author")
	scanStartCmd.PersistentFlags().StringVarP(scanStartCmdCommitAuthorProfilePictureLink, "profile-picture", "p", "", "commit author profile picture link")
	scansCmd.AddCommand(scanStartCmd)
	scansCmd.AddCommand(scanGetCmd)
	scansCmd.AddCommand(scanDownloadCmd)
	rootCmd.AddCommand(scansCmd)
}
