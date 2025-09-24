package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)


var uploadCmd = &cobra.Command{
	Use:     "upload",
	Aliases: []string{"tag"},
	Short:   "Interact with upload",
	Long:    "Interact with your escape upload",
}

var uploadSignedURLCmd = &cobra.Command{
	Use:     "signed-url",
	Aliases: []string{"su"},
	Short:   "Get a signed url to upload a file",
	Long: `Get a signed url to upload a file.

Example output:
ID
00000000-0000-0000-0000-000000000000
`,
	Example: `escape-cli upload signed-url<schema.json`, RunE: func(cmd *cobra.Command, _ []string) error {
		upload, err := escape.GetUploadSignedURL(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to get signed url: %w", err)
		}

		id, url := upload.GetId(), upload.GetUrl()

		var data []byte

		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		data = b
		
		var schema map[string]interface{}
		if err := json.Unmarshal(data, &schema); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		err = escape.UploadSchema(cmd.Context(), url, data)
		if err != nil {
			return fmt.Errorf("failed to upload schema: %w", err)
		}

		fields := []string{"ID"}
		out.Table(id, func() []string {
			fields = append(fields, id)
			return fields
		})

		return nil
	},
}

func init() {
	uploadCmd.AddCommand(uploadSignedURLCmd)

	rootCmd.AddCommand(uploadCmd)
}
