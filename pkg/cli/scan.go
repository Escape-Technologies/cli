package cli

import (
	"encoding/json"
	"fmt"
	"os"

	v1 "github.com/Escape-Technologies/cli/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var startScanCmd = &cobra.Command{
	Use:     "start-scan [applicationId]",
	Short:   "Trigger a scan on an application",
	Args:    cobra.ExactArgs(1),
	Example: "escape-cli start-scan 123e4567-e89b-12d3-a456-426614174001",
	RunE: func(cmd *cobra.Command, args []string) error {
		applicationIdString := args[0]
		fmt.Printf("Triggering scan for application %s\n\n", applicationIdString)
		applicationId, err := uuid.Parse(applicationIdString)
		if err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}

		client, err := v1.NewAPIClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		body := v1.PostApplicationsIdStartScanJSONRequestBody{}
		scan, err := client.PostApplicationsIdStartScanWithResponse(cmd.Context(), applicationId, body)
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
		os.Exit(1)
		return err
	},
}
