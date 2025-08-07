package out

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	greenBool = color.New(color.FgGreen).SprintFunc()
	redBool   = color.New(color.FgRed).SprintFunc()
	grayIDLink    = color.New(color.FgHiBlack).SprintFunc()
	boldText  = color.New(color.Bold).SprintFunc()
	noColor   = color.New(color.Reset).SprintFunc()
)

func colorizeValue(value string, columnName string) string {
	// handle based on column name
	switch strings.ToUpper(columnName) {
	case "NAME":
		return boldText(value)
	case "ID":
		return grayIDLink(value)
	case "LINK":
		return grayIDLink(value)
	}

	// handle boolean values
	switch strings.ToLower(value) {
	case "true":
		return greenBool(value)
	case "false":
		return redBool(value)
	default:
		return noColor(value)
	}
}

// Table prints a table of data
func Table(data any, tableMaker func() []string) {
	if output != outputPretty {
		pprint(output, data, "")
		return
	}

	table := tableMaker()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0) //nolint:mnd

	if len(table) > 0 {
		// Make headers bold
		headers := strings.Split(table[0], "\t")
		boldHeaders := make([]string, len(headers))
		for i, header := range headers {
			boldHeaders[i] = boldText(header)
		}
		fmt.Fprintln(w, strings.Join(boldHeaders, "\t")) //nolint:errcheck

		for i := 1; i < len(table); i++ {
			fields := strings.Split(table[i], "\t")
			for j, field := range fields {
				if j < len(headers) {
					fields[j] = colorizeValue(field, strings.TrimSpace(headers[j]))
				} else {
					fields[j] = colorizeValue(field, "")
				}
			}
			fmt.Fprintln(w, strings.Join(fields, "\t")) //nolint:errcheck
		}
	}
	
	w.Flush() //nolint:errcheck
}
