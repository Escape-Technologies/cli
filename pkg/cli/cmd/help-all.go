package cmd

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

type helpAllCmdData struct {
	Parent      *helpAllCmdData
	Name        string
	Description string
}

func (h *helpAllCmdData) String() string {
	prefix := ""
	if h.Parent != nil {
		prefix = "  "
	}
	return fmt.Sprintf("%s%s\t%s", prefix, h.Name, h.Description)
}

// Recursive function to print command names and help
func listCommands(cmd *cobra.Command, parent *helpAllCmdData) []*helpAllCmdData {
	commands := []*helpAllCmdData{}
	for _, c := range cmd.Commands() {
		if parent == nil && (c.Name() == "help-all" ||
			c.Name() == "help" ||
			c.Name() == "completion") {
			continue
		}
		cmd := &helpAllCmdData{
			Parent:      parent,
			Name:        c.Name(),
			Description: c.Short,
		}
		commands = append(commands, cmd)
		commands = append(commands, listCommands(c, cmd)...)
	}
	return commands
}

var helpAllCmd = &cobra.Command{
	Use:   "help-all",
	Short: "List all commands",
	Run: func(_ *cobra.Command, _ []string) {
		commands := listCommands(rootCmd, nil)
		out.Table(commands, func() []string {
			res := []string{"COMMAND\tDESCRIPTION"}
			for _, c := range commands {
				res = append(res, c.String())
			}
			return res
		})
	},
}

func init() {
	rootCmd.AddCommand(helpAllCmd)
}
