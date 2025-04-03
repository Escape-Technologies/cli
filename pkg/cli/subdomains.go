package cli

import (
	"fmt"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	"github.com/spf13/cobra"
)

var subdomainsCmd = &cobra.Command{
	Use:   "subdomains",
	Short: "Interact with your escape subdomains",
}

var subdomainsList = &cobra.Command{
	Use:   "list",
	Short: "List subdomains",
	RunE: func(cmd *cobra.Command, args []string) error {
		size := 3
		subdomains, nextCursor, err := escape.GetSubdomains(cmd.Context(), &size, nil)
		var toAdd []escape.Subdomain
		for nextCursor != nil {
			toAdd, nextCursor, err = escape.GetSubdomains(cmd.Context(), &size, nextCursor)
			subdomains = append(subdomains, toAdd...)
		}
		if err != nil {
			return err
		}
		print(
			subdomains,
			func() {
				for _, d := range subdomains {
					fmt.Printf("%s\t%s\t%d\n", d.Fqdn, d.Id, d.ServicesCount)
				}
			},
		)
		return nil
	},
}
