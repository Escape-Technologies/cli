package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

var (
	integrationsKind        string
	integrationsSearch      string
	integrationsProjectIDs  []string
	integrationsLocationIDs []string
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

		items, next, err := escape.ListIntegrations(cmd.Context(), integrationsKind, "", &escape.ListIntegrationsFilters{
			ProjectIDs:  integrationsProjectIDs,
			LocationIDs: integrationsLocationIDs,
			Search:      integrationsSearch,
		})
		if err != nil {
			return fmt.Errorf("failed to list integrations: %w", err)
		}
		all := items
		for next != nil && *next != "" {
			items, next, err = escape.ListIntegrations(cmd.Context(), integrationsKind, *next, &escape.ListIntegrationsFilters{
				ProjectIDs:  integrationsProjectIDs,
				LocationIDs: integrationsLocationIDs,
				Search:      integrationsSearch,
			})
			if err != nil {
				return fmt.Errorf("failed to list integrations: %w", err)
			}
			all = append(all, items...)
		}

		out.Table(all, func() []string {
			res := []string{"ID\tNAME\tKIND\tVALID\tUPDATED AT"}
			for _, item := range all {
				res = append(res, fmt.Sprintf("%s\t%s\t%s\t%v\t%s",
					stringValue(item["id"]),
					stringValue(item["name"]),
					stringValue(item["kind"]),
					item["valid"],
					stringValue(item["updatedAt"]),
				))
			}
			return res
		})
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
		if integrationsKind == "" {
			return errors.New("--kind is required")
		}
		if out.InputSchema(map[string]interface{}{}) {
			return nil
		}
		if out.Schema(map[string]interface{}{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
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
		if integrationsKind == "" {
			return errors.New("--kind is required")
		}
		if out.InputSchema(map[string]interface{}{}) {
			return nil
		}
		if out.Schema(map[string]interface{}{}) {
			return nil
		}

		body, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
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
		if integrationsKind == "" {
			return errors.New("--kind is required")
		}
		if err := escape.DeleteIntegration(cmd.Context(), integrationsKind, args[0]); err != nil {
			return fmt.Errorf("failed to delete integration: %w", err)
		}
		out.Log("Integration deleted")
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

func init() {
	integrationsCmd.AddCommand(integrationsListCmd, integrationsGetCmd, integrationsCreateCmd, integrationsUpdateCmd, integrationsDeleteCmd)
	for _, subcommand := range []*cobra.Command{
		integrationsListCmd,
		integrationsGetCmd,
		integrationsCreateCmd,
		integrationsUpdateCmd,
		integrationsDeleteCmd,
	} {
		subcommand.Flags().StringVar(&integrationsKind, "kind", "", "integration kind")
	}
	integrationsListCmd.Flags().StringVar(&integrationsSearch, "search", "", "search integrations by name")
	integrationsListCmd.Flags().StringSliceVar(&integrationsProjectIDs, "project-id", []string{}, "filter by project ID")
	integrationsListCmd.Flags().StringSliceVar(&integrationsLocationIDs, "location-id", []string{}, "filter by location ID")
	rootCmd.AddCommand(integrationsCmd)
}
