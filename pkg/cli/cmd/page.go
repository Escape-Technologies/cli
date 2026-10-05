package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
)

// maxPageSize is the public API pagination cap. A larger --size is rejected
// here so the CLI fails before the server does.
const maxPageSize = 100

// pagedListAnnotation marks a command whose --size and --cursor flags are a
// cursor page. pageFlags.bind sets it. pagedListCommand reads it, so a command
// that merely has both flag names (retests get) is not treated as a list tool.
const pagedListAnnotation = "mcp.pagedList"

// listPage is the document a list command prints when the caller asks for one
// page. Pretty mode still prints the table of items and, when another page
// exists, the next cursor on stderr.
//
// nextCursor is always present: an empty string means there is no further page.
// yaml tags keep the camelCase names; without them YAML prints nextcursor.
type listPage struct {
	Items      any    `json:"items" yaml:"items"`
	NextCursor string `json:"nextCursor" yaml:"nextCursor"`
	TotalCount int    `json:"totalCount" yaml:"totalCount"`
}

// Page is the shape capabilities and MCP advertise for a paged list.
// MCP always requests one page, so the tool result is this object rather
// than the bare item slice a full CLI listing prints.
type Page[T any] struct {
	Items      []T    `json:"items" yaml:"items"`
	NextCursor string `json:"nextCursor" yaml:"nextCursor"`
	TotalCount int    `json:"totalCount" yaml:"totalCount"`
}

// pageFlags are the --size and --cursor values bound on a paged list command.
// Leaving both unset keeps the historical "fetch every page" CLI behaviour.
type pageFlags struct {
	size   int
	cursor string
}

func (flags *pageFlags) bind(command *cobra.Command) {
	if command.Annotations == nil {
		command.Annotations = map[string]string{}
	}

	command.Annotations[pagedListAnnotation] = "true"
	command.Flags().IntVar(
		&flags.size,
		"size",
		0,
		fmt.Sprintf("Page size, from 1 to %d. With --size or --cursor, return one page instead of every row. JSON output is {items, nextCursor, totalCount}.", maxPageSize),
	)
	command.Flags().StringVar(
		&flags.cursor,
		"cursor",
		"",
		"Page cursor (nextCursor from the previous page). With --size or --cursor, return that single page.",
	)
}

func singlePageRequested(command *cobra.Command) bool {
	return command.Flags().Changed("size") || command.Flags().Changed("cursor")
}

func validatePageFlags(command *cobra.Command, size int) error {
	if !command.Flags().Changed("size") {
		return nil
	}

	if size < 1 || size > maxPageSize {
		return fmt.Errorf("--size must be between 1 and %d", maxPageSize)
	}

	return nil
}

// pageFetcher loads one API page. size 0 means "use the wrapper's existing
// default page size" so a full listing does not change how many rows each
// request asks for.
type pageFetcher[T any] func(ctx context.Context, cursor string, size int) ([]T, *string, int, error)

// resolveList returns every page, or exactly one page when single is set.
// The single-page result includes that response's next cursor and totalCount.
func resolveList[T any](
	ctx context.Context,
	single bool,
	cursor string,
	size int,
	fetch pageFetcher[T],
) ([]T, *string, int, error) {
	if single {
		items, next, total, err := fetch(ctx, cursor, size)
		if err != nil {
			return nil, nil, 0, err
		}

		if items == nil {
			items = []T{}
		}

		return items, next, total, nil
	}

	// A non-nil slice keeps an empty listing printed as [] in JSON, not null.
	all := []T{}
	nextCursor := ""
	// A server that returns the same cursor forever would otherwise loop.
	// Stop with the rows already collected, the same way retest paging does.
	seen := map[string]struct{}{}
	for {
		page, pageNext, _, err := fetch(ctx, nextCursor, 0)
		if err != nil {
			return nil, nil, 0, err
		}

		all = append(all, page...)
		if pageNext == nil || *pageNext == "" {
			break
		}

		if _, dup := seen[*pageNext]; dup {
			break
		}

		seen[*pageNext] = struct{}{}
		nextCursor = *pageNext
	}

	return all, nil, 0, nil
}

func emitList[T any](single bool, items []T, next *string, total int, table func() []string) {
	if items == nil {
		items = []T{}
	}

	if !single {
		out.Table(items, table)
		return
	}

	nextCursor := ""
	if next != nil {
		nextCursor = *next
	}

	out.Table(listPage{
		Items:      items,
		NextCursor: nextCursor,
		TotalCount: total,
	}, table)
	// Pretty mode prints only the rows. The cursor has to be visible or the
	// caller cannot ask for the next page. JSON and YAML already carry it.
	if nextCursor != "" && out.IsPretty() {
		fmt.Fprintf(os.Stderr, "Next cursor: %s (total %d)\n", nextCursor, total)
	}
}

// runPagedList prints every page, or exactly one page when --size or --cursor
// is set. One page is a JSON object {items, nextCursor, totalCount}; pretty
// mode stays a table of that page's rows.
func runPagedList[T any](
	command *cobra.Command,
	flags pageFlags,
	fetch pageFetcher[T],
	table func(items []T) []string,
) error {
	if err := validatePageFlags(command, flags.size); err != nil {
		return err
	}

	single := singlePageRequested(command)
	items, next, total, err := resolveList(command.Context(), single, flags.cursor, flags.size, fetch)
	if err != nil {
		return err
	}

	emitList(single, items, next, total, func() []string {
		return table(items)
	})

	return nil
}
