package out

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

var (
	greenText  = color.New(color.FgGreen).SprintFunc()
	linkText   = color.New(color.FgBlue).SprintFunc()
	cyanText   = color.New(color.FgCyan).SprintFunc()
	yellowText = color.New(color.FgYellow).SprintFunc()
	redText    = color.New(color.FgRed).SprintFunc()
	grayText   = color.New(color.FgHiBlack).SprintFunc()
	boldText   = color.New(color.Bold).SprintFunc()
	noColor    = color.New(color.Reset).SprintFunc()
	idText     = color.New(color.FgHiMagenta).SprintFunc()
)

func colorizeBool(value string) string {
	switch strings.ToLower(value) {
	case "true":
		return greenText(value)
	case "false":
		return redText(value)
	default:
		return value
	}
}

func colorizeSeverity(value string) string {
	switch strings.ToLower(value) {
	case "info":
		return grayText(value)
	case "low":
		return greenText(value)
	case "medium":
		return yellowText(value)
	case "high":
		return redText(value)
	default:
		return value
	}
}

func colorizeProgress(value string) string {
	if strings.HasPrefix(value, "100") {
		return greenText(value)
	}
	return yellowText(value)
}

func colorizeLevel(value string) string {
	switch strings.ToLower(value) {
	case "info":
		return grayText(value)
	case "warning":
		return yellowText(value)
	case "error":
		return redText(value)
	default:
		return noColor(value)
	}
}

func colorizeStatus(value string) string {
	switch strings.ToLower(value) {
	case "open":
		return redText(value)
	case "resolved":
		return greenText(value)
	case "manual_review":
		return yellowText(value)
	case "ignored":
		return grayText(value)
	case "running":
		return yellowText(value)
	case "finished":
		return greenText(value)
	case "failed":
		return redText(value)
	case "canceled":
		return grayText(value)
	case "deprecated":
		return yellowText(value)
	case "monitored":
		return greenText(value)
	case "false_positive":
		return yellowText(value)
	case "out_of_scope":
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

func colorizeLastSeen(value string) string {
	lastSeen, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return grayText(value)
	}
	//if less than 15 days, return green
	if time.Since(lastSeen) < 15*24*time.Hour {
		return greenText(value)
	}
	//if less than 30 days, return yellow
	if time.Since(lastSeen) < 30*24*time.Hour {
		return yellowText(value)
	}

	return redText(value)
}

func colorizeValue(value string, columnName string) string {
	if value == "[]" || value == "" {
		return boldText("-")
	}
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
	case "CATEGORY", "KIND", "STAGE", "TYPE", "RISKS":
		return colorizeEnum(value)
	case "CREATED AT", "UPDATED AT":
		return colorizeDate(value)
	case "LAST SEEN":
		return colorizeLastSeen(value)
	case "PROGRESS":
		return colorizeProgress(value)
	case "LEVEL":
		return colorizeLevel(value)
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

