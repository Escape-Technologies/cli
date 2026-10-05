package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Escape-Technologies/cli/pkg/cli/out"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func TestListPageJSONShape(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(listPage{
		Items:      []string{"a"},
		NextCursor: "cursor-2",
		TotalCount: 4,
	})
	if err != nil {
		t.Fatalf("marshal page: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal page: %v", err)
	}

	items, ok := decoded["items"].([]any)
	if !ok || len(items) != 1 || items[0] != "a" {
		t.Fatalf("items = %#v", decoded["items"])
	}

	if decoded["nextCursor"] != "cursor-2" {
		t.Fatalf("nextCursor = %#v", decoded["nextCursor"])
	}

	if decoded["totalCount"] != float64(4) {
		t.Fatalf("totalCount = %#v", decoded["totalCount"])
	}

	empty, err := json.Marshal(listPage{Items: []string{}})
	if err != nil {
		t.Fatalf("marshal empty page: %v", err)
	}

	if !strings.Contains(string(empty), `"items":[]`) || !strings.Contains(string(empty), `"nextCursor":""`) || !strings.Contains(string(empty), `"totalCount":0`) {
		t.Fatalf("empty page = %s", empty)
	}
}

func TestListPageYAMLUsesCamelCase(t *testing.T) {
	t.Parallel()

	raw, err := yaml.Marshal(listPage{
		Items:      []string{"a"},
		NextCursor: "cursor-2",
		TotalCount: 4,
	})
	if err != nil {
		t.Fatalf("marshal yaml: %v", err)
	}

	text := string(raw)
	for _, want := range []string{"nextCursor: cursor-2", "totalCount: 4", "items:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("yaml %q missing %q", text, want)
		}
	}

	if strings.Contains(text, "nextcursor:") || strings.Contains(text, "totalcount:") {
		t.Fatalf("yaml folded the field names: %s", text)
	}
}

func TestResolveListStopsWhenCursorRepeats(t *testing.T) {
	t.Parallel()

	calls := 0
	items, _, _, err := resolveList(context.Background(), false, "", 0, func(_ context.Context, cursor string, size int) ([]string, *string, int, error) {
		calls++
		if size != 0 {
			t.Fatalf("full listing size = %d", size)
		}

		again := "again"
		if cursor == "" || cursor == "again" {
			return []string{cursor}, &again, 1, nil
		}

		t.Fatalf("unexpected cursor %q", cursor)

		return nil, nil, 0, nil
	})
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected the repeated cursor to stop the walk after 2 calls, got %d", calls)
	}

	if strings.Join(items, ",") != ",again" {
		t.Fatalf("items = %#v", items)
	}
}

func TestResolveListDrainsEveryPage(t *testing.T) {
	t.Parallel()

	calls := 0
	items, next, total, err := resolveList(context.Background(), false, "ignored", 50, func(_ context.Context, cursor string, size int) ([]string, *string, int, error) {
		calls++
		if size != 0 {
			t.Fatalf("full listing must keep the API default page size, got %d", size)
		}

		switch cursor {
		case "":
			more := "page-2"
			return []string{"a"}, &more, 3, nil
		case "page-2":
			return []string{"b", "c"}, nil, 3, nil
		default:
			t.Fatalf("unexpected cursor %q", cursor)
			return nil, nil, 0, nil
		}
	})
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 pages, got %d", calls)
	}

	if next != nil || total != 0 {
		t.Fatalf("full listing should not surface page metadata, next=%v total=%d", next, total)
	}

	if strings.Join(items, ",") != "a,b,c" {
		t.Fatalf("items = %#v", items)
	}
}

func TestResolveListStopsAtOnePage(t *testing.T) {
	t.Parallel()

	calls := 0
	more := "page-3"
	items, next, total, err := resolveList(context.Background(), true, "page-2", 25, func(_ context.Context, cursor string, size int) ([]string, *string, int, error) {
		calls++
		if cursor != "page-2" || size != 25 {
			t.Fatalf("cursor=%q size=%d", cursor, size)
		}

		return nil, &more, 9, nil
	})
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}

	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	if len(items) != 0 {
		t.Fatalf("nil page should become an empty slice, got %#v", items)
	}

	if next == nil || *next != "page-3" || total != 9 {
		t.Fatalf("next=%v total=%d", next, total)
	}
}

func TestPageFlagsSelectSinglePage(t *testing.T) {
	t.Parallel()

	unset := &cobra.Command{Use: "list"}
	var unsetFlags pageFlags
	unsetFlags.bind(unset)
	if singlePageRequested(unset) {
		t.Fatal("expected a full listing when size and cursor are omitted")
	}

	sized := &cobra.Command{Use: "list"}
	var sizedFlags pageFlags
	sizedFlags.bind(sized)
	if err := sized.Flags().Set("size", "25"); err != nil {
		t.Fatal(err)
	}

	if !singlePageRequested(sized) {
		t.Fatal("expected --size to select one page")
	}

	if err := validatePageFlags(sized, sizedFlags.size); err != nil {
		t.Fatal(err)
	}

	if sizedFlags.size != 25 {
		t.Fatalf("size = %d", sizedFlags.size)
	}

	cursored := &cobra.Command{Use: "list"}
	var cursorFlags pageFlags
	cursorFlags.bind(cursored)
	if err := cursored.Flags().Set("cursor", "abc"); err != nil {
		t.Fatal(err)
	}

	if !singlePageRequested(cursored) {
		t.Fatal("expected --cursor to select one page")
	}

	if err := validatePageFlags(cursored, cursorFlags.size); err != nil {
		t.Fatal(err)
	}

	invalid := &cobra.Command{Use: "list"}
	var invalidFlags pageFlags
	invalidFlags.bind(invalid)
	if err := invalid.Flags().Set("size", "0"); err != nil {
		t.Fatal(err)
	}

	if err := validatePageFlags(invalid, invalidFlags.size); err == nil {
		t.Fatal("expected --size 0 to be rejected")
	}

	over := &cobra.Command{Use: "list"}
	var overFlags pageFlags
	overFlags.bind(over)
	if err := over.Flags().Set("size", "101"); err != nil {
		t.Fatal(err)
	}

	if err := validatePageFlags(over, overFlags.size); err == nil {
		t.Fatal("expected --size 101 to be rejected")
	}

	if err := over.Flags().Set("size", "100"); err != nil {
		t.Fatal(err)
	}

	if err := validatePageFlags(over, overFlags.size); err != nil {
		t.Fatal(err)
	}

	usage := over.Flags().Lookup("size").Usage
	if !strings.Contains(usage, "1 to 100") {
		t.Fatalf("size usage %q does not state the range", usage)
	}
}

func TestListCommandsBindPageFlags(t *testing.T) {
	t.Parallel()

	commands := []*cobra.Command{
		issueListCmd,
		assetsListCmd,
		eventsListCmd,
		profilesListCmd,
		projectsListCmd,
		integrationsListCmd,
		locationsListCmd,
		workflowsListCmd,
		scanIssuesCmd,
		problemsCmd,
		profileProblemsCmd,
		scansProblemsCmd,
		retestsListCmd,
	}
	for _, command := range commands {
		if command.Flags().Lookup("size") == nil || command.Flags().Lookup("cursor") == nil {
			t.Fatalf("%s is missing --size or --cursor", command.Name())
		}

		if command.Annotations[pagedListAnnotation] != "true" {
			t.Fatalf("%s is missing %s", command.CommandPath(), pagedListAnnotation)
		}
	}

	if pagedListCommand(retestsGetCmd) {
		t.Fatal("retests get has --size and --cursor but is not a paged list")
	}

	if scanTargetsCmd.Flags().Lookup("cursor") != nil {
		t.Fatal("scans targets --size is a result cap, not a page cursor")
	}
}

func TestEmitListPrintsPageObjectInJSONAndTableInPretty(t *testing.T) {
	restore := out.SetOutput
	t.Cleanup(func() {
		if err := restore("pretty"); err != nil {
			t.Errorf("restore output: %v", err)
		}
	})

	if err := out.SetOutput("json"); err != nil {
		t.Fatal(err)
	}

	jsonOut := captureStdout(t, func() {
		more := "next"
		emitList(true, []string{"one"}, &more, 8, func() []string {
			return []string{"ID", "one"}
		})
	})
	var decoded listPage
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Fatalf("json output %q: %v", jsonOut, err)
	}

	items, ok := decoded.Items.([]any)
	if !ok || len(items) != 1 || items[0] != "one" || decoded.NextCursor != "next" || decoded.TotalCount != 8 {
		t.Fatalf("page = %#v", decoded)
	}

	if err := out.SetOutput("pretty"); err != nil {
		t.Fatal(err)
	}

	prettyOut := captureStdout(t, func() {
		emitList(true, []string{"one"}, nil, 8, func() []string {
			return []string{"ID", "one"}
		})
	})
	if strings.Contains(prettyOut, "nextCursor") || strings.Contains(prettyOut, "{") {
		t.Fatalf("pretty mode should stay a table, got %q", prettyOut)
	}

	if !strings.Contains(prettyOut, "ID") || !strings.Contains(prettyOut, "one") {
		t.Fatalf("pretty table = %q", prettyOut)
	}

	stderr := captureStderr(t, func() {
		more := "cursor-2"
		emitList(true, []string{"one"}, &more, 8, func() []string {
			return []string{"ID", "one"}
		})
	})
	if stderr != "Next cursor: cursor-2 (total 8)\n" {
		t.Fatalf("stderr = %q", stderr)
	}

	if err := out.SetOutput("json"); err != nil {
		t.Fatal(err)
	}

	jsonErr := captureStderr(t, func() {
		more := "cursor-2"
		emitList(true, []string{"one"}, &more, 8, func() []string {
			return []string{"ID", "one"}
		})
	})
	if jsonErr != "" {
		t.Fatalf("json mode should keep the cursor in the document, stderr = %q", jsonErr)
	}
}

// The e2e suite asserts that an empty full listing prints exactly [].
// A nil slice would print null, which breaks `escape issues list -o json | jq`.
func TestFullListingWithNoRowsPrintsEmptyJSONArray(t *testing.T) {
	restore := out.SetOutput
	t.Cleanup(func() {
		if err := restore("pretty"); err != nil {
			t.Errorf("restore output: %v", err)
		}
	})
	if err := out.SetOutput("json"); err != nil {
		t.Fatal(err)
	}

	items, next, total, err := resolveList(context.Background(), false, "", 0, func(_ context.Context, _ string, _ int) ([]string, *string, int, error) {
		// The API body {"nextCursor":null,"totalCount":0,"data":[]}.
		return []string{}, nil, 0, nil
	})
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}

	if items == nil {
		t.Fatal("an empty full listing must be a non-nil slice")
	}

	stdout := captureStdout(t, func() {
		emitList(false, items, next, total, func() []string { return []string{"ID"} })
	})
	if strings.TrimSpace(stdout) != "[]" {
		t.Fatalf("stdout = %q, want []", stdout)
	}

	nilStdout := captureStdout(t, func() {
		emitList[string](false, nil, nil, 0, func() []string { return []string{"ID"} })
	})
	if strings.TrimSpace(nilStdout) != "[]" {
		t.Fatalf("nil items stdout = %q, want []", nilStdout)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	return captureFD(t, &os.Stdout, fn)
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	return captureFD(t, &os.Stderr, fn)
}

func captureFD(t *testing.T, target **os.File, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	original := *target
	*target = writer
	fn()
	closeErr := writer.Close()
	*target = original
	if closeErr != nil {
		t.Fatal(closeErr)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}
