package out

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	ansiReset       = "\x1b[0m"
	ansiFgDefault   = "\x1b[39m"
	targetOverhead = 29
)

var noColorEnabled bool

// DisableColor disables color output called if the ESCAPE_NO_COLOR environment variable is set
func DisableColor() {
	noColorEnabled = true
}

func makeColored(value string, prefix string) string {
	if noColorEnabled {
		return value
	}
	base := prefix + value + ansiReset
	used := len(prefix) + len(ansiReset)
	remaining := max(0, targetOverhead-used)
	var b strings.Builder
	b.WriteString(base)
	for remaining > 0 {
		if remaining%4 == 0 {
			b.WriteString(ansiReset)
			remaining -= 4
			continue
		}
		b.WriteString(ansiFgDefault)
		remaining -= 5
	}
	return b.String()
}

func greenText(value string) string  { return makeColored(value, "\x1b[32m") }
func linkText(value string) string   { return makeColored(value, "\x1b[34m") }
func yellowText(value string) string { return makeColored(value, "\x1b[33m") }
func redText(value string) string    { return makeColored(value, "\x1b[31m") }
func grayText(value string) string   { return makeColored(value, "\x1b[90m") }
func boldText(value string) string   { return makeColored(value, "\x1b[1m") }
func idText(value string) string     { return makeColored(value, "\x1b[95m") }
func escapeText(value string) string {return makeColored(value, "\x1b[38;2;6;226;183m")}
func shortEscapeLink(value string) string { return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", value, linkText("Link")) }
func noColor(value string) string    { return makeColored(value, "") }

func colorizeBool(value string) string {
	if noColorEnabled {
		return value
	}
	switch strings.ToLower(value) {
	case "true":
		return greenText(value)
	case "false":
		return redText(value)
	default:
		return noColor(value)
	}
}

func colorizeSeverity(value string) string {
	if noColorEnabled {
		return value
	}
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
		return noColor(value)
	}
}


func colorizeProgress(value string) string {
	if noColorEnabled {
		return value
	}
	if strings.HasPrefix(value, "100") || strings.HasPrefix(value, "1.000") {
		return greenText(value)
	}
	return yellowText(value)
}

func colorizeLevel(value string) string {
	if noColorEnabled {
		return value
	}
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
	if noColorEnabled {
		return value
	}
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
	if noColorEnabled {
		return value
	}
	return escapeText(value)
}

func colorizeDate(value string) string {
	if noColorEnabled {
		return value
	}
	return grayText(value)
}

func colorizeLastSeen(value string) string {
	if noColorEnabled {
		return value
	}
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

func colorizeHelpAll(value string) string {
	if noColorEnabled {
		return value
	}
	if strings.HasPrefix(value, "  ") {
		return escapeText(boldText(value))
	}
	return redText(boldText(value))
}

func colorizeActor(value string) string {
	if noColorEnabled {
		return value
	}
	if strings.ToLower(value) == "escape" {
		return escapeText(value)
	}
	return idText(value)
}


// colorizeWithHex applies an ANSI 24-bit foreground color to text using a hex RGB string (e.g. "6a63f0").
func colorizeWithHex(text string, hexRGB string) string {
	if noColorEnabled {
		return text
	}
    const expectedHexRGBLen = 6
    if len(hexRGB) == expectedHexRGBLen {
        if r, errR := strconv.ParseInt(hexRGB[0:2], 16, 0); errR == nil {
            if g, errG := strconv.ParseInt(hexRGB[2:4], 16, 0); errG == nil {
                if b, errB := strconv.ParseInt(hexRGB[4:6], 16, 0); errB == nil {
                    seq := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
                    return makeColored(text, seq)
                }
            }
        }
    }
    return grayText(text)
}

// TagText is a helper function to colorize a tag name based on its hexRGB color
func TagText(name string, hexRGB string) string {
    return colorizeWithHex(name, hexRGB)
}

func colorizeValue(value string, columnName string, isLastColumn bool) string {
	if noColorEnabled {
		return value
	}
	if value == "[]" || value == "" {
		return boldText("-")
	}
	if columnName == "ACTION" {
		return escapeText(value)
	}
	if columnName == "ACTOR EMAIL" {
		return yellowText(value)
	}
	// handle links
	urlRegex := regexp.MustCompile(`\b(?:(?:https?|grpc):\/\/)?(?:localhost|(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}|(?:\d{1,3}\.){3}\d{1,3})(?::\d+)?\b`)
	if urlRegex.MatchString(strings.ToLower(value)) {
		if isLastColumn && strings.HasPrefix(value, "https://app.escape") {
			return shortEscapeLink(value)
		}
		if strings.Contains(value, " ") {
			return boldText(value)
		}
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
	case "ACTOR EMAIL":
		return yellowText(value)
	case "ACTOR":
		return colorizeActor(value)
	case "SEVERITY":
		return colorizeSeverity(value)
	case "TITLE":
		return boldText(value)
	case "CATEGORY", "KIND", "STAGE", "TYPE", "RISKS", "INITIATORS", "ASSET TYPE":
		return colorizeEnum(value)
	case "CREATED AT", "UPDATED AT", "DATE":
		return colorizeDate(value)
	case "LAST SEEN":
		return colorizeLastSeen(value)
	case "PROGRESS":
		return colorizeProgress(value)
	case "LEVEL":
		return colorizeLevel(value)
	case "COMMAND":
		return colorizeHelpAll(value)
	case "DESCRIPTION":
		return grayText(value)
	case "CRON":
		return greenText(value)
	case "COLOR":
		return colorizeWithHex(value, value)
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0) //nolint:mnd

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
				isLastColumn := j == len(headers)-1
				if j < len(headers) {
					fields[j] = colorizeValue(field, strings.TrimSpace(headers[j]), isLastColumn)
				} else {
					fields[j] = colorizeValue(field, "", isLastColumn)
				}
			}
			fmt.Fprintln(w, strings.Join(fields, "\t")) //nolint:errcheck
		}
	}

	w.Flush() //nolint:errcheck
}


