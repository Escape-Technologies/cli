package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmdVerbose bool
var rootCmdOutputStr string

var rootCmd = &cobra.Command{
	Use:   "escape-cli",
	Short: "CLI to interact with Escape API",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if rootCmdVerbose {
			log.SetLevel(logrus.TraceLevel)
		}
		err := out.SetOutput(rootCmdOutputStr)
		if err != nil {
			return fmt.Errorf("failed to set output format: %w", err)
		}
		return nil
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		log.Trace("Main cli done, exiting")
	},

	SilenceUsage:  true,
	SilenceErrors: true,
}

// Recursive function to print command names and help
func listCommands(cmd *cobra.Command, prefix string) []string {
	commands := []string{}
	for _, c := range cmd.Commands() {
		if prefix == "" && (c.Name() == "help-all" ||
			c.Name() == "help" ||
			c.Name() == "completion") {
			continue
		}
		line := fmt.Sprintf("%s%s\t%s", prefix, c.Name(), c.Short)
		commands = append(commands, line)
		commands = append(commands, listCommands(c, prefix+"  ")...)
	}
	return commands
}

var helpAllCmd = &cobra.Command{
	Use:   "help-all",
	Short: "List all commands",
	Run: func(cmd *cobra.Command, args []string) {
		commands := listCommands(rootCmd, "")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "COMMAND\tDESCRIPTION")
		for _, command := range commands {
			if !strings.HasPrefix(command, " ") {
				fmt.Fprintln(w, "\t")
			}
			fmt.Fprintln(w, command)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootCmdVerbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&rootCmdOutputStr, "output", "o", "pretty", "output format (pretty|json|yaml)")
	rootCmd.AddCommand(helpAllCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
