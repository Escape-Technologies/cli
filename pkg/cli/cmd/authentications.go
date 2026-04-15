package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

const (
	authenticationProgressPercent = 100
	authenticationPollInterval    = 2 * time.Second
)

var authenticationsWatch bool

var authenticationsCmd = &cobra.Command{
	Use:   "authentications",
	Short: "Validate profile authentication configuration",
}

var authenticationsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start an authentication check from JSON stdin",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.StartAuthenticationRequest{}) {
			return nil
		}
		if out.Schema(v3.StartAuthentication200Response{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		auth, err := escape.StartAuthentication(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to start authentication check: %w", err)
		}

		out.Print(auth, "Authentication check started: "+auth.GetId())
		if authenticationsWatch {
			return watchAuthentication(cmd, auth.GetId())
		}
		return nil
	},
}

var authenticationsGetCmd = &cobra.Command{
	Use:     "get authentication-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get authentication check status",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("authentication ID is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.GetAuthentication200Response{}) {
			return nil
		}
		if authenticationsWatch {
			return watchAuthentication(cmd, args[0])
		}

		auth, err := escape.GetAuthentication(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get authentication check: %w", err)
		}

		out.Table(auth, func() []string {
			return []string{
				"ID\tSTATUS\tPROGRESS\tCREATED AT\tEVENTS",
				fmt.Sprintf("%s\t%s\t%d%%\t%s\t%d", auth.GetId(), auth.GetStatus(), int(auth.GetProgressRatio()*authenticationProgressPercent), auth.GetCreatedAt(), len(auth.GetEvents())),
			}
		})
		return nil
	},
}

func watchAuthentication(cmd *cobra.Command, authenticationID string) error {
	for {
		auth, err := escape.GetAuthentication(cmd.Context(), authenticationID)
		if err != nil {
			return fmt.Errorf("failed to get authentication check: %w", err)
		}

		out.Table(auth, func() []string {
			return []string{
				"ID\tSTATUS\tPROGRESS\tCREATED AT\tEVENTS",
				fmt.Sprintf("%s\t%s\t%d%%\t%s\t%d", auth.GetId(), auth.GetStatus(), int(auth.GetProgressRatio()*authenticationProgressPercent), auth.GetCreatedAt(), len(auth.GetEvents())),
			}
		})

		switch auth.GetStatus() {
		case v3.ENUMPROPERTIESSTATUS_FINISHED, v3.ENUMPROPERTIESSTATUS_COMPLETED:
			return nil
		case v3.ENUMPROPERTIESSTATUS_FAILED, v3.ENUMPROPERTIESSTATUS_CANCELED:
			return fmt.Errorf("authentication check ended with status %s", auth.GetStatus())
		case v3.ENUMPROPERTIESSTATUS_PENDING, v3.ENUMPROPERTIESSTATUS_RUNNING, v3.ENUMPROPERTIESSTATUS_STARTING:
		}

		select {
		case <-cmd.Context().Done():
			return fmt.Errorf("authentication watch canceled: %w", cmd.Context().Err())
		case <-time.After(authenticationPollInterval):
		}
	}
}

func init() {
	authenticationsCmd.AddCommand(authenticationsStartCmd, authenticationsGetCmd)
	authenticationsStartCmd.Flags().BoolVarP(&authenticationsWatch, "watch", "w", false, "watch authentication status until completion")
	authenticationsGetCmd.Flags().BoolVarP(&authenticationsWatch, "watch", "w", false, "watch authentication status until completion")
	rootCmd.AddCommand(authenticationsCmd)
}
