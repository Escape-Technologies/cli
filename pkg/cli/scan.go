package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/spf13/cobra"
)

var startScanCmd = &cobra.Command{
	Use:     "scan [applicationId]",
	Short:   "Trigger a scan on an application",
	Args:    cobra.ExactArgs(1),
	Example: "escape-cli scan 123e4567-e89b-12d3-a456-426614174001",
	RunE: func(cmd *cobra.Command, args []string) error {
		applicationID := args[0]
		fmt.Printf("Triggering scan for application %s\n\n", applicationID)

		client, err := api.NewAPIClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		parsedUUID, err := uuid.Parse(applicationID)
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
		applicationId := types.UUID(parsedUUID)
		params := &api.PostApplicationsIdStartScanParams{
			ContentType: api.PostApplicationsIdStartScanParamsContentTypeApplicationjson,
		}

		body := api.PostApplicationsIdStartScanJSONRequestBody{}
		scan, err := client.PostApplicationsIdStartScanWithResponse(context.Background(), applicationId, params, body)
		if err != nil {
			return err
		}
		// Handle response
		var data interface{}
		if scan.JSON200 != nil {
			print(
				scan.JSON200,
				func() {
					fmt.Printf("-> Scan successfully launched\n")
					fmt.Printf("Scan ID: %s\n", scan.JSON200.Id)
				},
			)
			return nil
		} else if scan.JSON400 != nil {
			data = scan.JSON400
		} else {
			data = scan.JSON500
		}
		print(
			data,
			func() {
				var responseMessage map[string]interface{}
				json.Unmarshal(scan.Body, &responseMessage)

				// Print status code and error message
				fmt.Println(scan.HTTPResponse.Status)
				fmt.Printf("%s\n\n", responseMessage["message"])
			},
		)
		return nil
	},
}
