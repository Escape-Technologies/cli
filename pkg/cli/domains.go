package cli

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Interact with your escape domains",
}

var domainsList = &cobra.Command{
	Use:   "list",
	Short: "List domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		domains, err := escape.GetDomains(cmd.Context())
		if err != nil {
			return err
		}
		print(
			domains,
			func() {
				for _, d := range domains {
					fmt.Println(d.String())
				}
			},
		)
		return nil
	},
}
