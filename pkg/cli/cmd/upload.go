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
	Use:   "upload",
	Short: "Upload large files to Escape platform",
	Long: `Upload Files - Temporary Storage for Large Assets

Upload large files (like API schemas) to temporary Escape storage. Returns an upload ID
that can be used in other commands when creating profiles or assets.

USE CASES:
  • Large OpenAPI/Swagger specifications
  • GraphQL schemas
  • Postman collections
  • Any schema file referenced in profile creation`,
}

var uploadSchemaCmd = &cobra.Command{
	Use:     "schema",
	Aliases: []string{"s"},
	Short:   "Upload an API schema file",
	Long: `Upload API Schema - Store Large Schema Files

Upload API schema files to temporary storage and receive an upload ID. Use this ID
when creating profiles that reference large schema files.`,
	Example: `  # Upload OpenAPI schema
  escape-cli upload schema < openapi-spec.json

  # Upload and capture ID
  UPLOAD_ID=$(escape-cli upload schema < schema.json -o json | jq -r '.id')

  # Use in profile creation
  echo '{"schema_upload_id": "'$UPLOAD_ID'", ...}' | escape-cli profiles create-rest`, RunE: func(cmd *cobra.Command, _ []string) error {
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
	uploadCmd.AddCommand(uploadSchemaCmd)

	rootCmd.AddCommand(uploadCmd)
}
