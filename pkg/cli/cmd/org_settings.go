package cmd

import (
	"errors"
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var orgSettingsCmd = &cobra.Command{
	Use:     "org-settings",
	Aliases: []string{"organization-settings", "organisation-settings"},
	Short:   "Manage the organization settings",
	Long: `Manage Organization Settings - AI Context and Issue SLAs

Read and update the organization-wide settings: the system prompt every AI
pentest receives as context, and the SLA policy applied per issue severity.`,
}

var orgSettingsGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"show"},
	Short:   "Get the organization settings",
	Long: `Get the organization settings: the AI system prompt, the SLA policy per
severity, the read-only kill switches, and the scanner configuration.`,
	Example: `  escape-cli org-settings get`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(v3.GetOrganizationSettings200Response{}) {
			return nil
		}

		settings, err := escape.GetOrganizationSettings(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to get organization settings: %w", err)
		}

		out.Table(settings, func() []string {
			res := []string{"SEVERITY\tSLA (ms)\tENABLED"}
			for _, sla := range settings.GetSlas() {
				res = append(res, fmt.Sprintf("%s\t%d\t%t", sla.GetSeverity(), sla.GetTimeMs(), sla.GetEnabled()))
			}

			prompt := ""
			if settings.AiSystemPrompt != nil {
				prompt = *settings.AiSystemPrompt
			}

			res = append(res, "AI SYSTEM PROMPT\t"+prompt)

			return res
		})

		return nil
	},
}

var orgSettingsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the organization settings from JSON stdin (Beta)",
	Long: `Update the organization settings from a JSON body.

Only the fields present in the body change; an empty aiSystemPrompt clears the
prompt. Use --input-schema to see the expected shape.`,
	Example: `  # Set the AI system prompt
  echo '{"aiSystemPrompt":"Focus on the public API surface."}' | escape-cli org-settings update

  # See the input schema
  escape-cli org-settings update --input-schema`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.InputSchema(v3.UpdateOrganizationSettingsRequest{}) {
			return nil
		}

		if out.Schema(v3.GetOrganizationSettings200Response{}) {
			return nil
		}

		body, err := readPipedStdin(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}

		if len(body) == 0 {
			return errors.New("JSON body is required on stdin")
		}

		settings, err := escape.UpdateOrganizationSettings(cmd.Context(), body)
		if err != nil {
			return fmt.Errorf("failed to update organization settings: %w", err)
		}

		out.Print(settings, "Organization settings updated")

		return nil
	},
}

func init() {
	orgSettingsCmd.AddCommand(orgSettingsGetCmd, orgSettingsUpdateCmd)
	rootCmd.AddCommand(orgSettingsCmd)
}
