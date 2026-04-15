package cmd

import (
	"encoding/json"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

type commandCapability struct {
	Path         string   `json:"path"`
	Use          string   `json:"use"`
	Short        string   `json:"short,omitempty"`
	Aliases      []string `json:"aliases,omitempty"`
	HasSub       bool     `json:"hasSubcommands"`
	HasFlags     bool     `json:"hasFlags"`
	HasInSchema  bool     `json:"hasInputSchema"`
	HasOutSchema bool     `json:"hasOutputSchema"`
}

var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Describe CLI commands in a machine-readable format",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if out.Schema([]commandCapability{}) {
			return nil
		}

		capabilities := make([]commandCapability, 0)
		var walk func(command *cobra.Command)
		walk = func(command *cobra.Command) {
			if !command.IsAvailableCommand() || command.Hidden {
				return
			}
			capabilities = append(capabilities, commandCapability{
				Path:         command.CommandPath(),
				Use:          command.Use,
				Short:        command.Short,
				Aliases:      command.Aliases,
				HasSub:       command.HasAvailableSubCommands(),
				HasFlags:     command.LocalFlags().HasAvailableFlags() || command.PersistentFlags().HasAvailableFlags(),
				HasInSchema:  !command.HasAvailableSubCommands(),
				HasOutSchema: true,
			})
			for _, child := range command.Commands() {
				walk(child)
			}
		}
		walk(rootCmd)

		pretty, _ := json.MarshalIndent(capabilities, "", "  ")
		out.Print(capabilities, string(pretty))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}
