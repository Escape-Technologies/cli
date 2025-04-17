package out

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// Table prints a table of data
func Table(data any, tableMaker func() []string) {
	if output != outputPretty {
		print(output, data, "")
		return
	}

	table := tableMaker()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0) //nolint:mnd
	for _, row := range table {
		fmt.Fprintln(w, row) //nolint:errcheck
	}
	w.Flush() //nolint:errcheck
}
