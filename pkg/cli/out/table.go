package out

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func Table(data any, tableMaker func() []string) {
	if output != outputPretty {
		print(output, data, "")
		return
	}

	table := tableMaker()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	for _, row := range table {
		fmt.Fprintln(w, row)
	}
	w.Flush()
}
