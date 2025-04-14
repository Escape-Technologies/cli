package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Integrations commands",
}

var integrationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		integrations, err := escape.GetIntegrations(context.Background())
		if err != nil {
			return fmt.Errorf("unable to get integrations: %w", err)
		}
		print(
			integrations,
			func() {
				fmt.Println(escape.FormatIntegrationsTable(integrations))
			},
		)
		return nil
	},
}

var integrationsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get integration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		integrationId, err := uuid.Parse(args[0])
		if err != nil {
			return fmt.Errorf("unable to get integration: %w", err)
		}
		integration, err := escape.GetIntegration(context.Background(), integrationId)
		if err != nil {
			return fmt.Errorf("unable to get integration: %w", err)
		}
		print(
			integration,
			func() {
				fmt.Println(escape.FormatIntegrationTable([]escape.IntegrationWithParameters{*integration}))
			},
		)
		return nil
	},
}

var integrationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create integration",
	Args:  cobra.ExactArgs(1),
	Example: "escape integrations create akamai.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		integrationParameters, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("unable to read integration file: %w", err)
		}

		var integrationInput escape.IntegrationInput
		
		err = json.Unmarshal(integrationParameters, &integrationInput)
		if err != nil {
			return fmt.Errorf("unable to unmarshal integration: %w", err)
		}
		
		integration, err := escape.CreateIntegration(context.Background(), integrationInput)
		if err != nil {
			return fmt.Errorf("unable to create integration: %w", err)
		}
		print(
			integration,
			func() {
				fmt.Println(escape.FormatIntegrationTable([]escape.IntegrationWithParameters{*integration}))
			},
		)
		return nil
	},
}