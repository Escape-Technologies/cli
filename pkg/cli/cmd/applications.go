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
	Short:   "Interact with applications",
	Long:    "Interact with your escape applications",
}

var applicationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List applications",
	Long:    `List all applications.

Example output:
ID                                      TYPE         NAME                                                           CREATED AT              HAS CI    CRON
00000000-0000-0000-0000-000000000001    REST         Example-Application-1                                          2025-02-21T11:15:07Z    false     0 11 * * 5
00000000-0000-0000-0000-000000000002    REST         Example-Application-2                                          2025-03-12T19:19:08Z    false     0 19 * * 3`,
	Example: `escape-cli applications list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		applications, err := escape.ListApplications(cmd.Context())
		if err != nil {
			return fmt.Errorf("unable to list applications: %w", err)
		}
		out.Table(applications, func() []string {
			result := []string{"ID\tTYPE\tNAME\tCREATED AT\tHAS CI\tCRON"}
			for _, app := range applications {
				result = append(result, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s", app.Id, app.Type, app.Name, app.CreatedAt.Format(time.RFC3339), app.HasCI, app.Cron))
			}
			return result
		})
		return nil
	},
}

var applicationGetCmd = &cobra.Command{
	Use:     "get application-id",
	Aliases: []string{"describe"},
	Short:   "Get application details",
	Long:    `Get details about an application's current configuration.

Example output:
Name: Example-Application
Id: 00000000-0000-0000-0000-000000000000
CreatedAt: 2025-02-01T13:17:44Z
HasCI: false
Cron: 0 13 * * 6
Configuration:
{"authentication":{"users":[{"name":"public"}]},"scan":{"blocklist":{"routes":[{"method":"POST","path":".*logout.*"},{"method":"POST","path":".*password.*"},{"method":"PUT","path":".*password.*"},{"method":"POST","path":".*2fa.*"},{"method":"POST","path":".*(refresh|revoke).*token.*"},{"method":"DELETE","path":".*user"},{"method":"DELETE","path":".*token.*"},{"method":"PUT","path":".*role.*"},{"method":"PUT","path":".*permissions.*"},{"method":"DELETE","path":".*role.*"},{"method":"DELETE","path":".*permissions.*"},{"method":"GET","path":".*createdb.*"},{"method":"DELETE","path":".*user.*{user(name)?}.*"}]},"profile":"surface","read_only":true}}`,
	Args:    cobra.ExactArgs(1),
	Example: `escape-cli applications get 00000000-0000-0000-0000-000000000000`,
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
	Use:     "update-schema application-id schema-path|schema-url",
	Short:   "Update application schema",
	Long:    "Update the schema of an application based on a configuration file (yaml or json)",
	Args:    cobra.ExactArgs(2), //nolint:mnd
	Example: `escape-cli applications update-schema 00000000-0000-0000-0000-000000000000 schema.json
escape-cli applications update-schema 00000000-0000-0000-0000-000000000000 schema.yaml
escape-cli applications update-schema 00000000-0000-0000-0000-000000000000 https://example.com/schema.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := escape.UpdateApplicationSchema(cmd.Context(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("unable to update schema: %w", err)
		}
		return nil
	},
}

var applicationUpdateConfigCmd = &cobra.Command{
	Use:     "update-config application-id config-path",
	Short:   "Update application config",
	Long:    "Update the configuration of an application based on a configuration file (yaml or json)",
	Example: `escape-cli applications update-config 00000000-0000-0000-0000-000000000000 config.json
escape-cli applications update-config 00000000-0000-0000-0000-000000000000 config.yaml`,
	Args:    cobra.ExactArgs(2), //nolint:mnd
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
