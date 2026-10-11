package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	clischema "github.com/Escape-Technologies/cli/pkg/cli/schema"
	"github.com/spf13/cobra"
)

var (
	integrationsKind        string
	integrationsSearch      string
	integrationsProjectIDs  []string
	integrationsLocationIDs []string
	integrationListPage     pageFlags
	integrationRunsPage     pageFlags

	integrationsReposIntegrationID string
	integrationsReposSearch        string
	integrationsReposFullNames     []string
	integrationReposPage           pageFlags

	integrationsVaultsKind   string
	integrationsVaultsSearch string

	integrationsJiraProjectsSearch      string
	integrationsJiraIssueTypesProjectID string
	integrationsAzureAreasProject       string
	integrationsAzureAreasWorkItemType  string
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Manage integrations",
}

var integrationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List integrations for a given kind",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema([]map[string]interface{}{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		filters := &escape.ListIntegrationsFilters{
			ProjectIDs:  integrationsProjectIDs,
			LocationIDs: integrationsLocationIDs,
			Search:      integrationsSearch,
		}
		if err := runPagedList(cmd, integrationListPage, func(ctx context.Context, cursor string, size int) ([]map[string]interface{}, *string, int, error) {
			return escape.ListIntegrations(ctx, integrationsKind, cursor, filters, size)
		}, func(items []map[string]interface{}) []string {
			res := []string{"ID\tNAME\tKIND\tVALID\tUPDATED AT"}
			for _, item := range items {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%v\t%s",
					stringValue(item["id"]),
					stringValue(item["name"]),
					stringValue(item["kind"]),
					item["valid"],
					stringValue(item["updatedAt"]),
				))
			}

			return res
		}); err != nil {
			return fmt.Errorf("failed to list integrations: %w", err)
		}

		return nil
	},
}

var integrationsGetCmd = &cobra.Command{
	Use:     "get integration-id",
	Aliases: []string{"describe", "show"},
	Short:   "Get an integration by ID",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(map[string]interface{}{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		item, err := escape.GetIntegration(cmd.Context(), integrationsKind, args[0])
		if err != nil {
			return fmt.Errorf("failed to get integration: %w", err)
		}

		out.Table(item, func() []string {
			return []string{
				"ID\tNAME\tKIND\tVALID\tUPDATED AT\tLOCATION ID\tPROJECTS",
				fmt.Sprintf("%s\t%s\t%s\t%v\t%s\t%s\t%s",
					stringValue(item["id"]),
					stringValue(item["name"]),
					stringValue(item["kind"]),
					item["valid"],
					stringValue(item["updatedAt"]),
					nestedStringValue(item, "location", "id"),
					joinMapField(item["projects"], "name"),
				),
			}
		})

		return nil
	},
}

var integrationsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an integration from JSON stdin",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		printed, err := integrationInputSchema(integrationCreateInputSchema)
		if err != nil || printed {
			return err
		}

		if out.Schema(map[string]interface{}{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		body, err := readIntegrationBody(cmd)
		if err != nil {
			return err
		}

		item, err := escape.CreateIntegration(cmd.Context(), integrationsKind, body)
		if err != nil {
			return fmt.Errorf("failed to create integration: %w", err)
		}

		out.Print(item, "Integration created: "+stringValue(item["id"]))

		return nil
	},
}

var integrationsUpdateCmd = &cobra.Command{
	Use:   "update integration-id",
	Short: "Update an integration from JSON stdin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		printed, err := integrationInputSchema(integrationUpdateInputSchema)
		if err != nil || printed {
			return err
		}

		if out.Schema(map[string]interface{}{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		body, err := readIntegrationBody(cmd)
		if err != nil {
			return err
		}

		item, err := escape.UpdateIntegration(cmd.Context(), integrationsKind, args[0], body)
		if err != nil {
			return fmt.Errorf("failed to update integration: %w", err)
		}

		out.Print(item, "Integration updated: "+stringValue(item["id"]))

		return nil
	},
}

var integrationsDeleteCmd = &cobra.Command{
	Use:     "delete integration-id",
	Aliases: []string{"del", "remove"},
	Short:   "Delete an integration",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(integrationDeleteOutput) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		proceed, err := confirmDestructive(cmd, fmt.Sprintf("delete %s integration %s", integrationsKind, args[0]), func(ctx context.Context) (any, error) {
			return escape.GetIntegration(ctx, integrationsKind, args[0])
		})
		if err != nil || !proceed {
			return err
		}

		item, err := escape.DeleteIntegration(cmd.Context(), integrationsKind, args[0])
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}

		out.Print(item, "Integration deleted")

		return nil
	},
}

var integrationsTestCmd = &cobra.Command{
	Use:   "test integration-id",
	Short: "Test an integration's credentials",
	Long: `Test Integration - Check the Stored Credentials

Ask the integration to check its credentials. The API returns the run it
scheduled, with its status.`,
	Example: `  escape-cli integrations test <integration-id> --kind JIRA`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.SyncAkamaiIntegration200Response{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		run, err := escape.TestIntegration(cmd.Context(), integrationsKind, args[0])
		if err != nil {
			return fmt.Errorf("failed to test integration: %w", err)
		}

		printIntegrationRun(run)

		return nil
	},
}

var integrationsSyncCmd = &cobra.Command{
	Use:   "sync integration-id",
	Short: "Schedule a synchronization run of an integration",
	Long: `Sync Integration - Pull the Provider's Latest State

Schedule a run that pulls the latest state from the provider (repositories,
projects, tickets). The API returns the run it scheduled, with its status.`,
	Example: `  escape-cli integrations sync <integration-id> --kind JIRA`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema(v3.SyncAkamaiIntegration200Response{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		run, err := escape.SyncIntegration(cmd.Context(), integrationsKind, args[0])
		if err != nil {
			return fmt.Errorf("failed to sync integration: %w", err)
		}

		printIntegrationRun(run)

		return nil
	},
}

var integrationsRunsCmd = &cobra.Command{
	Use:     "runs integration-id",
	Aliases: []string{"run", "history"},
	Short:   "List the runs of an integration",
	Long: `List the runs of an integration: the synchronizations and tests the
platform executed, with their status, timing and progress.`,
	Example: `  escape-cli integrations runs <integration-id> --kind JIRA`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]map[string]interface{}{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		if err := runPagedList(cmd, integrationRunsPage, func(ctx context.Context, cursor string, size int) ([]map[string]interface{}, *string, int, error) {
			return escape.ListIntegrationRuns(ctx, integrationsKind, args[0], cursor, size)
		}, func(runs []map[string]interface{}) []string {
			res := []string{"ID\tKIND\tSTATUS\tCREATED AT\tFINISHED AT\tPROGRESS"}
			for _, run := range runs {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%v",
					stringValue(run["id"]),
					stringValue(run["kind"]),
					stringValue(run["status"]),
					stringValue(run["createdAt"]),
					stringValue(run["finishedAt"]),
					run["progressRatio"],
				))
			}

			return res
		}); err != nil {
			return fmt.Errorf("failed to list integration runs: %w", err)
		}

		return nil
	},
}

var integrationsRepositoriesCmd = &cobra.Command{
	Use:     "repositories",
	Aliases: []string{"repos"},
	Short:   "List the repositories the GitHub integrations can see",
	Long: `List GitHub Repositories - Pick a Target for Code Scanning

List the repositories the organization's GitHub credentials can reach, the
coordinates ` + "`escape-cli assets create --asset-type GITHUB_REPOSITORY`" + `
needs. Without --integration-id the listing spans every GitHub integration;
with it, only that integration's repositories.`,
	Example: `  # Every repository across the GitHub integrations
  escape-cli integrations repositories

  # One integration's repositories, matching a name
  escape-cli integrations repositories --integration-id <integration-id> --search api`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if out.Schema(Page[escape.GithubRepository]{}) {
			return nil
		}

		filters := &escape.ListGithubRepositoriesFilters{
			Search:    integrationsReposSearch,
			FullNames: integrationsReposFullNames,
		}

		if err := runPagedList(cmd, integrationReposPage, func(ctx context.Context, cursor string, size int) ([]escape.GithubRepository, *string, int, error) {
			return escape.ListGithubRepositories(ctx, integrationsReposIntegrationID, cursor, size, filters)
		}, integrationsRepositoriesTable); err != nil {
			return fmt.Errorf("failed to list GitHub repositories: %w", err)
		}

		return nil
	},
}

var integrationsVaultsCmd = &cobra.Command{
	Use:   "vaults integration-id",
	Short: "List the vaults an integration sees at its provider (Beta)",
	Long: `List Provider Vaults - Discover What Can Be Linked

List the vaults an integration can see at its provider, before any of them is
linked to the organization. Each row carries the externalId
'escape-cli vault link' needs. Only names are returned: the values never leave
the provider.`,
	Example: `  # What the Azure integration can see
  escape-cli integrations vaults <integration-id> --kind azure-key-vault

  # Narrow the discovery by name
  escape-cli integrations vaults <integration-id> --kind aws --search prod`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]escape.RemoteVault{}) {
			return nil
		}

		if integrationsVaultsKind == "" {
			return errors.New("--kind is required")
		}

		vaults, err := escape.ListIntegrationVaults(cmd.Context(), integrationsVaultsKind, args[0], integrationsVaultsSearch)
		if err != nil {
			return fmt.Errorf("failed to list provider vaults: %w", err)
		}

		out.Table(vaults, func() []string {
			rows := []string{"NAME\tEXTERNAL ID"}
			for _, vault := range vaults {
				rows = append(rows, fmt.Sprintf("%s\t%s", vault.Name, vault.ExternalID))
			}

			return rows
		})

		return nil
	},
}

var integrationsJiraProjectsCmd = &cobra.Command{
	Use:   "jira-projects integration-id",
	Short: "List Jira projects an integration can file tickets against (Beta)",
	Long: `List Jira Projects - Ticket Workflow Lookup

List the Jira projects an integration can see, so a workflow can target one of them.`,
	Example: `  escape-cli integrations jira-projects <integration-id>
  escape-cli integrations jira-projects <integration-id> --search core`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]escape.JiraProject{}) {
			return nil
		}

		projects, err := escape.ListJiraProjects(cmd.Context(), args[0], integrationsJiraProjectsSearch)
		if err != nil {
			return fmt.Errorf("failed to list Jira projects: %w", err)
		}

		out.Table(projects, func() []string {
			rows := []string{"ID\tNAME\tPROJECT TYPE\tENTITY ID"}
			for _, project := range projects {
				entityID := ""
				if project.EntityID != nil {
					entityID = *project.EntityID
				}

				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s", project.ID, project.Name, project.ProjectTypeKey, entityID))
			}

			return rows
		})

		return nil
	},
}

var integrationsJiraIssueTypesCmd = &cobra.Command{
	Use:   "jira-issue-types integration-id",
	Short: "List Jira issue types for a project (Beta)",
	Long: `List Jira Issue Types - Ticket Workflow Lookup

List the issue types of one Jira project. --project-id is the Jira project id.`,
	Example: `  escape-cli integrations jira-issue-types <integration-id> --project-id CT`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]escape.JiraIssueType{}) {
			return nil
		}

		if integrationsJiraIssueTypesProjectID == "" {
			return errors.New("--project-id is required")
		}

		issueTypes, err := escape.ListJiraIssueTypes(cmd.Context(), args[0], integrationsJiraIssueTypesProjectID)
		if err != nil {
			return fmt.Errorf("failed to list Jira issue types: %w", err)
		}

		out.Table(issueTypes, func() []string {
			rows := []string{"ID\tNAME\tSUBTASK\tDESCRIPTION"}
			for _, issueType := range issueTypes {
				subtask := ""
				if issueType.Subtask != nil {
					subtask = strconv.FormatBool(*issueType.Subtask)
				}

				description := ""
				if issueType.Description != nil {
					description = *issueType.Description
				}

				rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s", issueType.ID, issueType.Name, subtask, description))
			}

			return rows
		})

		return nil
	},
}

var integrationsAzureAreasCmd = &cobra.Command{
	Use:   "azure-areas integration-id",
	Short: "List Azure Boards area paths (Beta)",
	Long: `List Azure Boards Areas - Ticket Workflow Lookup

List the area paths of one Azure Boards project and work item type.`,
	Example: `  escape-cli integrations azure-areas <integration-id> --project MyProject --work-item-type Bug`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if out.Schema([]string{}) {
			return nil
		}

		if integrationsAzureAreasProject == "" {
			return errors.New("--project is required")
		}

		if integrationsAzureAreasWorkItemType == "" {
			return errors.New("--work-item-type is required")
		}

		areas, err := escape.ListAzureBoardsAreas(cmd.Context(), args[0], integrationsAzureAreasProject, integrationsAzureAreasWorkItemType)
		if err != nil {
			return fmt.Errorf("failed to list Azure Boards areas: %w", err)
		}

		out.Table(areas, func() []string {
			rows := []string{"AREA"}
			rows = append(rows, areas...)

			return rows
		})

		return nil
	},
}

func integrationsRepositoriesTable(repositories []escape.GithubRepository) []string {
	rows := []string{"FULL NAME\tDEFAULT BRANCH\tPRIVATE\tARCHIVED\tCLONE URL"}
	for _, repository := range repositories {
		rows = append(rows, fmt.Sprintf("%s\t%s\t%t\t%t\t%s",
			repository.FullName,
			repository.DefaultBranch,
			repository.Private,
			repository.Archived,
			repository.CloneURL,
		))
	}

	return rows
}

// percentScale converts a 0..1 ratio to a percentage.
const percentScale = 100

func printIntegrationRun(run *v3.SyncAkamaiIntegration200Response) {
	out.Table(run, func() []string {
		return []string{
			"ID\tKIND\tSTATUS\tPROGRESS",
			fmt.Sprintf("%s\t%s\t%s\t%.0f%%", run.GetId(), run.GetKind(), run.GetStatus(), run.GetProgressRatio()*percentScale),
		}
	})
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}

		return string(data)
	}
}

func nestedStringValue(item map[string]interface{}, keys ...string) string {
	var current interface{} = item
	for _, key := range keys {
		next, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}

		current = next[key]
	}

	return stringValue(current)
}

func joinMapField(value interface{}, key string) string {
	items, ok := value.([]interface{})
	if !ok {
		return ""
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		object, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if field := stringValue(object[key]); field != "" {
			values = append(values, field)
		}
	}

	return strings.Join(values, ",")
}

func readIntegrationBody(cmd *cobra.Command) ([]byte, error) {
	body, err := readPipedStdin(cmd.InOrStdin())
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	if body == nil {
		return []byte{}, nil
	}

	return body, nil
}

func integrationInputSchema(build func() (*clischema.JSONSchema, error)) (bool, error) {
	if !rootCmdInputSchema {
		return false, nil
	}

	schema, err := build()
	if err != nil {
		return false, fmt.Errorf("input schema: %w", err)
	}

	out.SetInputSchema(true)

	return out.InputSchema(providedSchema{schema: schema}), nil
}

func init() {
	integrationsCmd.AddCommand(integrationsListCmd, integrationsGetCmd, integrationsCreateCmd, integrationsUpdateCmd, integrationsDeleteCmd, integrationsTestCmd, integrationsSyncCmd, integrationsRunsCmd, integrationsRepositoriesCmd, integrationsVaultsCmd, integrationsJiraProjectsCmd, integrationsJiraIssueTypesCmd, integrationsAzureAreasCmd)
	kinds := integrationKinds()
	kindUsage := "integration kind: " + strings.Join(kinds, ", ")
	for _, subcommand := range []*cobra.Command{
		integrationsListCmd,
		integrationsGetCmd,
		integrationsCreateCmd,
		integrationsUpdateCmd,
		integrationsDeleteCmd,
		integrationsTestCmd,
		integrationsSyncCmd,
		integrationsRunsCmd,
	} {
		subcommand.Flags().StringVar(&integrationsKind, "kind", "", kindUsage)
		if err := subcommand.Flags().SetAnnotation("kind", flagEnumAnnotation, kinds); err != nil {
			panic(err)
		}
	}

	// Provider vault discovery is GitHub-independent but still kind-addressed:
	// the --kind value is the URL slug the vaults route uses.
	integrationsVaultsCmd.Flags().StringVar(&integrationsVaultsKind, "kind", "", kindUsage)
	if err := integrationsVaultsCmd.Flags().SetAnnotation("kind", flagEnumAnnotation, kinds); err != nil {
		panic(err)
	}

	markMCPRequired(integrationsVaultsCmd.Flags(), "kind")
	integrationsVaultsCmd.Flags().StringVar(&integrationsVaultsSearch, "search", "", "search term to filter vaults by name")

	integrationsJiraProjectsCmd.Flags().StringVar(&integrationsJiraProjectsSearch, "search", "", "search term to filter Jira projects by name")
	integrationsJiraIssueTypesCmd.Flags().StringVar(&integrationsJiraIssueTypesProjectID, "project-id", "", "Jira project id")
	markMCPRequired(integrationsJiraIssueTypesCmd.Flags(), "project-id")
	integrationsAzureAreasCmd.Flags().StringVar(&integrationsAzureAreasProject, "project", "", "Azure Boards project")
	markMCPRequired(integrationsAzureAreasCmd.Flags(), "project")
	integrationsAzureAreasCmd.Flags().StringVar(&integrationsAzureAreasWorkItemType, "work-item-type", "", "Azure Boards work item type")
	markMCPRequired(integrationsAzureAreasCmd.Flags(), "work-item-type")

	integrationsRepositoriesCmd.Flags().StringVar(&integrationsReposIntegrationID, "integration-id", "", "list only this integration's repositories (default: every GitHub integration)")
	integrationsRepositoriesCmd.Flags().StringVar(&integrationsReposSearch, "search", "", "search term to filter repositories by full name")
	integrationsRepositoriesCmd.Flags().StringSliceVar(&integrationsReposFullNames, "full-name", []string{}, "filter by repository full names (owner/name)")
	integrationReposPage.bind(integrationsRepositoriesCmd)

	integrationsListCmd.Flags().StringVar(&integrationsSearch, "search", "", "search integrations by name")
	integrationsListCmd.Flags().StringSliceVar(&integrationsProjectIDs, "project-id", []string{}, "filter by project ID")
	integrationsListCmd.Flags().StringSliceVar(&integrationsLocationIDs, "location-id", []string{}, "filter by location ID")
	integrationListPage.bind(integrationsListCmd)
	integrationRunsPage.bind(integrationsRunsCmd)
	bindDestructiveGate(integrationsDeleteCmd, true)
	rootCmd.AddCommand(integrationsCmd)
}
