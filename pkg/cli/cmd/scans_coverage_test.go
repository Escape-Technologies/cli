package cmd

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func coverageOK() v3.ENUMPROPERTIESISSUEPROPERTIESCOVERAGEITEMS {
	return v3.ENUMPROPERTIESISSUEPROPERTIESCOVERAGEITEMS_OK
}

func coverageUnauthorized() v3.ENUMPROPERTIESISSUEPROPERTIESCOVERAGEITEMS {
	return v3.ENUMPROPERTIESISSUEPROPERTIESCOVERAGEITEMS_UNAUTHORIZED
}

func testAPIRoute(displayName string, method string, requestCount float32) *v3.ApiRouteDetailed {
	route := v3.NewApiRouteDetailedWithDefaults()
	route.SetDisplayName(displayName)
	route.SetName(displayName)
	route.SetOperation(method)
	route.SetRequestCount(requestCount)

	return route
}

func TestMatchCoverageFilter(t *testing.T) {
	t.Parallel()

	row := ScanCoverageTarget{
		Coverage: "OK",
		CoverageByUser: []ScanCoverageUserRow{
			{Name: "company_b_initiator", Coverage: "OK"},
			{Name: "company_a", Coverage: "UNAUTHORIZED"},
		},
	}

	if !matchCoverageFilter(row, "", "") {
		t.Fatal("expected unfiltered row to match")
	}

	if !matchCoverageFilter(row, "ok", "") {
		t.Fatal("expected coverage filter to be case-insensitive")
	}

	if matchCoverageFilter(row, "UNAUTHORIZED", "") {
		t.Fatal("expected overall UNAUTHORIZED filter to miss OK row")
	}

	if !matchCoverageFilter(row, "", "company_a") {
		t.Fatal("expected user filter to match coverageByUser name")
	}

	if matchCoverageFilter(row, "", "missing") {
		t.Fatal("expected unknown user to miss")
	}

	if matchCoverageFilter(row, "OK", "company_a") {
		t.Fatal("expected --coverage OK --user company_a to miss UNAUTHORIZED user")
	}

	if !matchCoverageFilter(row, "UNAUTHORIZED", "company_a") {
		t.Fatal("expected --coverage UNAUTHORIZED --user company_a to match that user's status")
	}

	if !matchCoverageFilter(row, "OK", "company_b_initiator") {
		t.Fatal("expected --coverage OK --user company_b_initiator to match that user's status")
	}
}

func TestSummarizeCoverageByUserCountsEveryStatus(t *testing.T) {
	t.Parallel()

	ok := coverageOK()
	unauthorized := coverageUnauthorized()

	routeA := testAPIRoute("/transfers", "GET", 1)
	routeA.SetCoverage(ok)
	routeA.SetCoverageByUser([]v3.CoverageByUserEntry{
		*v3.NewCoverageByUserEntry("alice", ok),
		*v3.NewCoverageByUserEntry("bob", unauthorized),
	})
	routeB := testAPIRoute("/transfers", "POST", 4)
	routeB.SetCoverage(ok)
	routeB.SetCoverageByUser([]v3.CoverageByUserEntry{
		*v3.NewCoverageByUserEntry("alice", ok),
	})

	targetA := v3.NewTargetDetailed("2026-01-01T00:00:00Z", "t-a")
	targetA.SetApiRoute(*routeA)
	targetB := v3.NewTargetDetailed("2026-01-01T00:00:00Z", "t-b")
	targetB.SetApiRoute(*routeB)

	got := summarizeCoverageByUser([]ScanCoverageTarget{
		compactScanTarget(*targetA),
		compactScanTarget(*targetB),
	})

	alice := got["alice"]
	if alice.Targets != 2 || alice.Statuses["OK"] != 2 {
		t.Fatalf("alice summary = %+v", alice)
	}

	if len(alice.OKExamples) != 2 ||
		alice.OKExamples[0] != "GET /transfers" ||
		alice.OKExamples[1] != "POST /transfers" {
		t.Fatalf("alice okExamples = %v", alice.OKExamples)
	}

	bob := got["bob"]
	if bob.Targets != 1 || bob.Statuses["UNAUTHORIZED"] != 1 {
		t.Fatalf("bob summary = %+v", bob)
	}

	if len(bob.OKExamples) != 0 {
		t.Fatalf("bob should have no OK examples, got %v", bob.OKExamples)
	}
}

func TestCompactScanTargetUsesCoverageByUser(t *testing.T) {
	t.Parallel()

	ok := coverageOK()
	route := testAPIRoute("/btl/v4/transfers", "GET", 471)
	route.SetCoverage(ok)
	route.SetCoverageByUser([]v3.CoverageByUserEntry{
		*v3.NewCoverageByUserEntry("company_b_initiator", ok),
	})
	target := v3.NewTargetDetailed("2026-01-01T00:00:00Z", "target-1")
	target.SetApiRoute(*route)

	got := compactScanTarget(*target)
	if got.Type != "API_ROUTE" || got.Method != "GET" || got.Name != "/btl/v4/transfers" {
		t.Fatalf("compact row = %+v", got)
	}

	if got.Coverage != "OK" || got.RequestCount != 471 {
		t.Fatalf("coverage/count = %+v", got)
	}

	if len(got.CoverageByUser) != 1 || got.CoverageByUser[0].Name != "company_b_initiator" {
		t.Fatalf("coverageByUser = %+v", got.CoverageByUser)
	}
}

// serveScanTargets fakes GET /v3/scans/{id}/targets with offset cursors.
func serveScanTargets(t *testing.T, targets []map[string]any) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start, _ := strconv.Atoi(r.URL.Query().Get("cursor"))
		size, _ := strconv.Atoi(r.URL.Query().Get("size"))
		end := min(start+size, len(targets))
		var next any
		if end < len(targets) {
			next = strconv.Itoa(end)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":       targets[start:end],
			"nextCursor": next,
			"totalCount": len(targets),
		})
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
}

func coverageTestRoute(index int, coverage any, blacklisted bool, byUser []map[string]any) map[string]any {
	id := fmt.Sprintf("00000000-0000-0000-0000-%012d", index)

	return map[string]any{
		"id":        id,
		"createdAt": "2026-01-01T00:00:00Z",
		"apiRoute": map[string]any{
			"id":             id,
			"blacklisted":    blacklisted,
			"createdAt":      "2026-01-01T00:00:00Z",
			"displayName":    fmt.Sprintf("GET /r%d", index),
			"name":           fmt.Sprintf("/r%d", index),
			"operation":      "GET",
			"requestCount":   1,
			"source":         "SOURCE_INFERRED",
			"coverage":       coverage,
			"coverageByUser": byUser,
		},
	}
}

func TestBuildScanCoverageCountsOverallStatusesOnce(t *testing.T) {
	alice := func(coverage string) map[string]any { return map[string]any{"name": "alice", "coverage": coverage} }
	bob := func(coverage string) map[string]any { return map[string]any{"name": "bob", "coverage": coverage} }

	// 250 routes over three pages: 120 OK (alice and bob overlap on 60), 80 UNAUTHORIZED,
	// 10 blocklisted, 40 never requested, then 50 non-API targets.
	targets := make([]map[string]any, 0, 300)
	for i := range 250 {
		switch {
		case i < 40:
			targets = append(targets, coverageTestRoute(i, "OK", false, []map[string]any{alice("OK"), bob("UNAUTHORIZED")}))
		case i < 100:
			targets = append(targets, coverageTestRoute(i, "OK", false, []map[string]any{alice("OK"), bob("OK")}))
		case i < 120:
			targets = append(targets, coverageTestRoute(i, "OK", false, []map[string]any{bob("OK")}))
		case i < 200:
			targets = append(targets, coverageTestRoute(i, "UNAUTHORIZED", false, []map[string]any{alice("UNAUTHORIZED")}))
		case i < 210:
			targets = append(targets, coverageTestRoute(i, nil, true, nil))
		default:
			targets = append(targets, coverageTestRoute(i, nil, false, nil))
		}
	}

	for i := range 50 {
		id := fmt.Sprintf("10000000-0000-0000-0000-%012d", i)
		target := map[string]any{"id": id, "createdAt": "2026-01-01T00:00:00Z"}
		if i < 30 {
			target["webPage"] = map[string]any{"id": id, "url": "https://app.example.com/", "visits": 1}
		} else {
			target["codeFile"] = map[string]any{"id": id, "language": "go", "path": "main.go"}
		}

		targets = append(targets, target)
	}

	serveScanTargets(t, targets)

	got, err := buildScanCoverage(t.Context(), "scan-1", "", "", "", defaultCoverageTargetListSize)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Complete || got.TotalCount != 300 || !got.TargetsTruncated {
		t.Fatalf("complete=%t total=%d truncated=%t", got.Complete, got.TotalCount, got.TargetsTruncated)
	}

	want := map[string]int{"OK": 120, "UNAUTHORIZED": 80, coverageBlocklisted: 10, coverageSkipped: 40}
	if got.Overall.Targets != 250 || !maps.Equal(got.Overall.Statuses, want) {
		t.Fatalf("overall = %+v", got.Overall)
	}

	if got.Overall.OKExamples[0] != "GET /r0" {
		t.Fatalf("okExamples should not repeat the method: %v", got.Overall.OKExamples)
	}

	if got.ByUser["alice"].Statuses["OK"] != 100 || got.ByUser["bob"].Statuses["OK"] != 80 {
		t.Fatalf("byUser = %+v", got.ByUser)
	}

	if !maps.Equal(got.ByType, map[string]int{"API_ROUTE": 250, "WEB_PAGE": 30, "CODE_FILE": 20}) {
		t.Fatalf("byType = %v", got.ByType)
	}

	pretty := formatScanCoveragePretty(got)
	for _, line := range []string{
		"Targets in scan  300",
		"  API_ROUTE  250",
		"Coverage of 250 API routes and GraphQL resolvers",
		"  OK            120  48.0%",
		"  USER   ROUTES  OK   UNAUTHORIZED\n",
		"  alice  180     100  80\n",
	} {
		if !strings.Contains(pretty, line) {
			t.Fatalf("pretty output misses %q:\n%s", line, pretty)
		}
	}
}

func TestBuildScanCoverageFiltersSampleNotTotals(t *testing.T) {
	serveScanTargets(t, []map[string]any{
		coverageTestRoute(0, "OK", false, nil),
		coverageTestRoute(1, nil, false, nil),
		coverageTestRoute(2, "SERVER_ERROR", false, nil),
	})

	got, err := buildScanCoverage(t.Context(), "scan-1", "", coverageSkipped, "", 0)
	if err != nil {
		t.Fatal(err)
	}

	if got.MatchedCount != 1 || got.Targets[0].Name != "/r1" {
		t.Fatalf("SKIPPED filter matched %d: %+v", got.MatchedCount, got.Targets)
	}

	if got.Overall.Targets != 3 || got.Overall.Statuses["OK"] != 1 {
		t.Fatalf("overall must ignore the sample filter: %+v", got.Overall)
	}
}

func TestBuildScanCoverageRejectsInvalidSize(t *testing.T) {
	t.Parallel()

	_, err := buildScanCoverage(t.Context(), "scan-1", "", "", "", -1)
	if err == nil {
		t.Fatal("expected negative size error")
	}

	_, err = buildScanCoverage(t.Context(), "scan-1", "", "", "", maxCoverageTargetListSize+1)
	if err == nil {
		t.Fatal("expected size cap error")
	}
}

func TestCapCoverageTargetsZeroMeansUnlimited(t *testing.T) {
	t.Parallel()

	matched := []ScanCoverageTarget{
		{ID: "a"},
		{ID: "b"},
		{ID: "c"},
	}

	got, truncated := capCoverageTargets(matched, 0)
	if truncated || len(got) != 3 {
		t.Fatalf("size 0 should return all matching, got len=%d truncated=%t", len(got), truncated)
	}

	got, truncated = capCoverageTargets(matched, 2)
	if !truncated || len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" {
		t.Fatalf("size 2 should cap, got %+v truncated=%t", got, truncated)
	}

	got, truncated = capCoverageTargets(matched, 5)
	if truncated || len(got) != 3 {
		t.Fatalf("size above len should not truncate, got len=%d truncated=%t", len(got), truncated)
	}
}

func TestGuidanceMentionsCompleteByUser(t *testing.T) {
	t.Parallel()

	if !strings.Contains(coverageToolGuidance, "complete=true") {
		t.Fatal("guidance must tell the model when byUser is exhaustive")
	}

	if !strings.Contains(coverageToolGuidance, "scans_reasoning") {
		t.Fatal("guidance must warn against using reasoning logs for coverage")
	}

	if !strings.Contains(coverageToolGuidance, "byUser.statuses") {
		t.Fatal("guidance must point at byUser.statuses for per-user answers")
	}

	if !strings.Contains(coverageToolGuidance, "overall.statuses") {
		t.Fatal("guidance must point at overall.statuses for route totals")
	}
}
