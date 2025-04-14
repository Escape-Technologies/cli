package cli

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload files to escape",
}

var uploadUrlCmd = &cobra.Command{
	Use:   "url",
	Short: "Get upload url",
	RunE: func(cmd *cobra.Command, args []string) error {
		url, err := escape.GetUrl(cmd.Context())
		if err != nil {
			return err
		}
		print(url, func() { fmt.Println(url.String()) })
		return nil
	},
}
