package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		if out.Schema([]map[string]interface{}{}) {
			return nil
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
		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		if out.Schema(map[string]interface{}{}) {
			return nil
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
		if out.Schema(v3.CreateakamaiIntegration200Response{}) {
			return nil
		}

		if integrationsKind == "" {
			return errors.New("--kind is required")
		}

		item, err := escape.DeleteIntegration(cmd.Context(), integrationsKind, args[0])
		if err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}

		out.Print(item, "Integration deleted")

		return nil
	},
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
	integrationsCmd.AddCommand(integrationsListCmd, integrationsGetCmd, integrationsCreateCmd, integrationsUpdateCmd, integrationsDeleteCmd)
	kinds := integrationKinds()
	kindUsage := "integration kind: " + strings.Join(kinds, ", ")
	for _, subcommand := range []*cobra.Command{
		integrationsListCmd,
		integrationsGetCmd,
		integrationsCreateCmd,
		integrationsUpdateCmd,
		integrationsDeleteCmd,
	} {
		subcommand.Flags().StringVar(&integrationsKind, "kind", "", kindUsage)
		if err := subcommand.Flags().SetAnnotation("kind", flagEnumAnnotation, kinds); err != nil {
			panic(err)
		}
	}

	integrationsListCmd.Flags().StringVar(&integrationsSearch, "search", "", "search integrations by name")
	integrationsListCmd.Flags().StringSliceVar(&integrationsProjectIDs, "project-id", []string{}, "filter by project ID")
	integrationsListCmd.Flags().StringSliceVar(&integrationsLocationIDs, "location-id", []string{}, "filter by location ID")
	integrationListPage.bind(integrationsListCmd)
	rootCmd.AddCommand(integrationsCmd)
}
