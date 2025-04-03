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
		size := 100
		subdomains, nextCursor, err := escape.GetSubdomains(cmd.Context(), &size, nil)
		if err != nil {
			return err
		}
		for _, d := range subdomains {
			print(d, func() { fmt.Println(d.String()) })
		}

		for nextCursor != nil {
			subdomains, nextCursor, err = escape.GetSubdomains(cmd.Context(), &size, nextCursor)
			if err != nil {
				return err
			}
			for _, d := range subdomains {
				print(d, func() { fmt.Println(d.String()) })
			}
		}
		return nil
	},
}
