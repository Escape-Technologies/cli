package out

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	greenBool = color.New(color.FgGreen).SprintFunc()
	linkText = color.New(color.FgBlue).SprintFunc()
	cyanText  = color.New(color.FgCyan).SprintFunc()
	yellowText = color.New(color.FgYellow).SprintFunc()
	redBool   = color.New(color.FgRed).SprintFunc()
	grayText    = color.New(color.FgHiBlack).SprintFunc()
	boldText  = color.New(color.Bold).SprintFunc()
	noColor   = color.New(color.Reset).SprintFunc()
	idText = color.New(color.FgHiMagenta).SprintFunc()
)

func colorizeBool(value string) string {
	switch strings.ToLower(value) {
	case "true":
		return greenBool(value)
	case "false":
		return redBool(value)
	default:
		return value
	}
}

func colorizeSeverity(value string) string {
	switch strings.ToLower(value) {
	case "info":
		return grayText(value)
	case "low":
		return greenBool(value)
	case "medium":
		return yellowText(value)
	case "high":
		return redBool(value)
	default:
		return value
	}
}

func colorizeStatus(value string) string {
	switch strings.ToLower(value) {
	case "open":
		return redBool(value)
	case "resolved":
		return greenBool(value)
	case "manual_review":
		return yellowText(value)
	case "ignored":
		return grayText(value)
	default:
		return noColor(value)
	}
}

func colorizeEnum(value string) string {
	return cyanText(value)
}

func colorizeDate(value string) string {
	return grayText(value)
}

func colorizeValue(value string, columnName string) string {
	// handle links
	urlRegex := regexp.MustCompile(`\b(?:(?:https?|grpc):\/\/)?(?:localhost|(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}|(?:\d{1,3}\.){3}\d{1,3})(?::\d+)?\b`)
	if urlRegex.MatchString(strings.ToLower(value)) {
		return linkText(value)
	}

	// handle based on column name
	switch strings.ToUpper(columnName) {
	case "NAME":
		return boldText(value)
	case "ID":
		return idText(value)
	case "STATUS":
		return colorizeStatus(value)
	case "SEVERITY":
		return colorizeSeverity(value)
	case "CATEGORY", "KIND":
		return colorizeEnum(value)
	case "CREATED AT", "UPDATED AT":
		return colorizeDate(value)
	}

	// handle boolean values
	switch strings.ToLower(value) {
	case "true", "false":
		return colorizeBool(value)
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
