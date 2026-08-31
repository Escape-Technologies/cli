package cmd

import (
	"strings"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func coverageOK() v3.ENUMPROPERTIESDATAITEMSPROPERTIESAPIROUTEPROPERTIESCOVERAGE {
	return v3.ENUMPROPERTIESDATAITEMSPROPERTIESAPIROUTEPROPERTIESCOVERAGE_OK
}

func coverageUnauthorized() v3.ENUMPROPERTIESDATAITEMSPROPERTIESAPIROUTEPROPERTIESCOVERAGE {
	return v3.ENUMPROPERTIESDATAITEMSPROPERTIESAPIROUTEPROPERTIESCOVERAGE_UNAUTHORIZED
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
}
