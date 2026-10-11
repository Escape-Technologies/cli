package escape

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// placeholderGoTypeName is the marker the generated client writes when it
// cannot serialize a union-typed query parameter. No value on the wire may
// contain it.
const placeholderGoTypeName = "Parameter value"

// captureQueries starts a server that records the query of every request and
// answers with a body valid for the requested endpoint. It returns the live
// slice of captured queries.
func captureQueries(t *testing.T) *[]url.Values {
	t.Helper()

	queries := make([]url.Values, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")

		body := `{"data":[]}`
		switch r.URL.Path {
		case "/v3/issues/funnel", "/v3/issues/trends":
			// Both endpoints answer a bare array, not an object.
			body = "[]"
		case "/v3/statistics/trends":
			body = `{"openResolved":[],"averageApiHealth":[]}`
		}

		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_AUTHORIZATION", "Bearer test")

	return &queries
}

// runCapturingQuery runs one wrapper call and returns the query of the single
// request it issued.
func runCapturingQuery(t *testing.T, queries *[]url.Values, call func() error) url.Values {
	t.Helper()

	before := len(*queries)
	if err := call(); err != nil {
		t.Fatalf("call: %v", err)
	}

	if len(*queries) != before+1 {
		t.Fatalf("requests = %d, want exactly one more than %d", len(*queries), before)
	}

	return (*queries)[len(*queries)-1]
}

// assertNoPlaceholder fails if any query value still carries the generated
// union placeholder instead of the real value.
func assertNoPlaceholder(t *testing.T, query url.Values) {
	t.Helper()

	for key, values := range query {
		for _, value := range values {
			if strings.Contains(value, placeholderGoTypeName) {
				t.Fatalf("query %s carries %q, the generated union placeholder", key, value)
			}
		}
	}
}

// TestArrayFiltersAreRepeatedQueryKeys asserts that every slice-valued filter
// the generated client wraps in a union query parameter reaches the wire as one
// repeated key per value, once with one value and once with two. The generated
// client used to send "<Go type> value" instead.
func TestArrayFiltersAreRepeatedQueryKeys(t *testing.T) {
	queries := captureQueries(t)

	tests := []struct {
		name string
		key  string
		call func(values []string) error
	}{
		{
			name: "issues names",
			key:  "names",
			call: func(values []string) error {
				_, _, _, err := ListIssues(context.Background(), "", &ListIssuesFilters{Names: values}, "", "", 1)
				return err
			},
		},
		{
			name: "issues projectIds",
			key:  "projectIds",
			call: func(values []string) error {
				_, _, _, err := ListIssues(context.Background(), "", &ListIssuesFilters{ProjectIDs: values}, "", "", 1)
				return err
			},
		},
		{
			name: "issues targetIds",
			key:  "targetIds",
			call: func(values []string) error {
				_, _, _, err := ListIssues(context.Background(), "", &ListIssuesFilters{TargetIDs: values}, "", "", 1)
				return err
			},
		},
		{
			name: "issues securityTestUids",
			key:  "securityTestUids",
			call: func(values []string) error {
				_, _, _, err := ListIssues(context.Background(), "", &ListIssuesFilters{SecurityTestUids: values}, "", "", 1)
				return err
			},
		},
		{
			name: "issues blacklistedIds",
			key:  "blacklistedIds",
			call: func(values []string) error {
				_, _, _, err := ListIssues(context.Background(), "", &ListIssuesFilters{BlacklistedIDs: values}, "", "", 1)
				return err
			},
		},
		{
			name: "issues blacklistedNames",
			key:  "blacklistedNames",
			call: func(values []string) error {
				_, _, _, err := ListIssues(context.Background(), "", &ListIssuesFilters{BlacklistedNames: values}, "", "", 1)
				return err
			},
		},
		{
			name: "scans projectIds",
			key:  "projectIds",
			call: func(values []string) error {
				_, _, err := ListScans(context.Background(), "", &ListScansFilters{ProjectIDs: &values})
				return err
			},
		},
		{
			name: "scans problems projectIds",
			key:  "projectIds",
			call: func(values []string) error {
				_, _, _, err := ListScanProblems(context.Background(), "", &ListScanProblemsFilters{ProjectIDs: values}, 1)
				return err
			},
		},
		{
			name: "profiles projectIds",
			key:  "projectIds",
			call: func(values []string) error {
				_, _, _, err := ListProfiles(context.Background(), "", &ListProfilesFilters{ProjectIDs: values}, 1)
				return err
			},
		},
		{
			name: "issues funnel projectIds",
			key:  "projectIds",
			call: func(values []string) error {
				_, err := GetIssueFunnel(context.Background(), values)
				return err
			},
		},
		{
			name: "issues trends projectIds",
			key:  "projectIds",
			call: func(values []string) error {
				_, err := GetIssueTrends(context.Background(), "2026-01-01T00:00:00Z", "2026-01-02T00:00:00Z", "", nil, values)
				return err
			},
		},
		{
			name: "issues trends applicationIds",
			key:  "applicationIds",
			call: func(values []string) error {
				_, err := GetIssueTrends(context.Background(), "2026-01-01T00:00:00Z", "2026-01-02T00:00:00Z", "", values, nil)
				return err
			},
		},
		{
			name: "statistics trends projectIds",
			key:  "projectIds",
			call: func(values []string) error {
				_, err := GetStatisticsTrends(context.Background(), "2026-01-01T00:00:00Z", "2026-01-02T00:00:00Z", "", values)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, values := range [][]string{{"value-1"}, {"value-1", "value-2"}} {
				query := runCapturingQuery(t, queries, func() error { return tt.call(values) })
				assertNoPlaceholder(t, query)

				if got := query[tt.key]; !reflect.DeepEqual(got, values) {
					t.Fatalf("%s = %v, want %v", tt.key, got, values)
				}
			}
		})
	}
}

// TestSingleValueUnionFiltersAreSentVerbatim covers the union parameters whose
// filter is a plain string rather than a list. The generated client serialized
// them as a Go type name too; the value must reach the wire unchanged.
func TestSingleValueUnionFiltersAreSentVerbatim(t *testing.T) {
	queries := captureQueries(t)

	tests := []struct {
		name  string
		key   string
		value string
		call  func(value string) error
	}{
		{
			name:  "profiles assetSchemaIds",
			key:   "assetSchemaIds",
			value: "schema-1",
			call: func(value string) error {
				_, _, _, err := ListProfiles(context.Background(), "", &ListProfilesFilters{AssetSchemaIDs: value}, 1)
				return err
			},
		},
		{
			name:  "events eventIds",
			key:   "eventIds",
			value: "event-1",
			call: func(value string) error {
				_, _, _, err := ListEvents(context.Background(), "", &ListEventsFilters{EventIDs: value}, 1)
				return err
			},
		},
		{
			name:  "events targetIds",
			key:   "targetIds",
			value: "target-1",
			call: func(value string) error {
				_, _, _, err := ListEvents(context.Background(), "", &ListEventsFilters{TargetIDs: value}, 1)
				return err
			},
		},
		{
			name:  "events workflowIds",
			key:   "workflowIds",
			value: "workflow-1",
			call: func(value string) error {
				_, _, _, err := ListEvents(context.Background(), "", &ListEventsFilters{WorkflowIDs: value}, 1)
				return err
			},
		},
		{
			name:  "events groups",
			key:   "groups",
			value: "group-1",
			call: func(value string) error {
				_, _, _, err := ListEvents(context.Background(), "", &ListEventsFilters{Groups: value}, 1)
				return err
			},
		},
		{
			name:  "events responseStatusCodes",
			key:   "responseStatusCodes",
			value: "500",
			call: func(value string) error {
				_, _, _, err := ListEvents(context.Background(), "", &ListEventsFilters{ResponseStatusCodes: value}, 1)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := runCapturingQuery(t, queries, func() error { return tt.call(tt.value) })
			assertNoPlaceholder(t, query)

			if got := query[tt.key]; !reflect.DeepEqual(got, []string{tt.value}) {
				t.Fatalf("%s = %v, want [%s]", tt.key, got, tt.value)
			}
		})
	}
}

// TestArrayFiltersAbsentWhenUnset guards the empty case: an unset filter must
// not put an empty key on the wire.
func TestArrayFiltersAbsentWhenUnset(t *testing.T) {
	queries := captureQueries(t)

	query := runCapturingQuery(t, queries, func() error {
		_, _, _, err := ListIssues(context.Background(), "", &ListIssuesFilters{}, "", "", 1)
		return err
	})

	for _, key := range []string{"names", "projectIds", "targetIds", "securityTestUids", "blacklistedIds", "blacklistedNames"} {
		if _, ok := query[key]; ok {
			t.Fatalf("%s = %v, want the key absent", key, query[key])
		}
	}
}
