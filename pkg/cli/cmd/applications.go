package cmd

import (
	"fmt"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var applicationsCmd = &cobra.Command{
	Use:     "applications",
	Aliases: []string{"app", "apps", "application"},
	Short:   "Interact with your escape applications",
}

var applicationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		applications, err := escape.ListApplications(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to list applications: %w", err)
		}
		out.Table(applications, func() []string {
			result := []string{"ID\tNAME\tCREATED AT\tHAS CI\tCRON"}
			for _, app := range applications {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%t\t%s", app.Id, app.Name, app.CreatedAt.Format(time.RFC3339), app.HasCI, app.Cron))
			}
			return result
		})
		return nil
	},
}

var applicationGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"describe"},
	Short:   "Get details about an application current configuration",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := escape.GetApplication(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("unable to get application: %w", err)
		}
		config, _ := application.GetConfiguration().MarshalJSON()
		out.Print(application, fmt.Sprintf(
			"Name: %s\nId: %s\nCreatedAt: %s\nHasCI: %t\nCron: %s\nConfiguration:\n%s",
			application.GetName(),
			application.GetId(),
			application.GetCreatedAt().Format(time.RFC3339),
			application.GetHasCI(),
			application.GetCron(),
			string(config),
		))
		return nil
	},
}

var applicationUpdateSchemaCmd = &cobra.Command{
	Use:   "update-schema",
	Short: "Update the schema of an application",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.UpdateApplicationSchema(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("unable to update schema: %w", err)
		}
		return nil
	},
}

var applicationUpdateConfigCmd = &cobra.Command{
	Use:   "update-config",
	Short: "Update the configuration of an application",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.UpdateApplicationConfig(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("unable to update config: %w", err)
		}
		return nil
	},
}

func init() {
	applicationsCmd.AddCommand(applicationsListCmd)
	applicationsCmd.AddCommand(applicationGetCmd)
	applicationsCmd.AddCommand(applicationUpdateSchemaCmd)
	applicationsCmd.AddCommand(applicationUpdateConfigCmd)
	rootCmd.AddCommand(applicationsCmd)
}
