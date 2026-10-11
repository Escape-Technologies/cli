package cmd

import (
	"context"
	"net/http"
	"testing"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/spf13/cobra"
)

// TestStatisticsSubResourcesPrintDeclaredDocument walks the three
// sub-resources of the statistics group and checks each prints its declared
// response.
func TestStatisticsSubResourcesPrintDeclaredDocument(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3/statistics/endpoints":
			if got := r.URL.Query().Get("source"); got != "PENTEST" {
				t.Errorf("source = %q, want PENTEST", got)
			}

			writeJSON(t, w, map[string]any{"apiRoutes": 120, "graphqlResolvers": 12, "webPages": 4})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/statistics/issues":
			writeJSON(t, w, map[string]any{
				"open":                 7,
				"remediated":           3,
				"averageTimeToResolve": 12.5,
				"byStatus":             []map[string]any{{"status": "OPEN", "count": 5}},
				"bySeverity":           []map[string]any{{"severity": "HIGH", "count": 4}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v3/statistics/trends":
			writeJSON(t, w, map[string]any{
				"openResolved":     []map[string]any{{"x": "2026-08-01T00:00:00Z", "open": 9, "resolved": 0}},
				"averageApiHealth": []map[string]any{{"x": "2026-08-01T00:00:00Z", "y": 82.5}},
			})
		default:
			http.NotFound(w, r)
		}
	})

	statisticsEndpointsSource = "PENTEST"
	statisticsTrendsStart = "2026-08-01T00:00:00Z"
	statisticsTrendsEnd = "2026-10-01T00:00:00Z"
	t.Cleanup(func() {
		statisticsEndpointsSource = ""
		statisticsTrendsStart = ""
		statisticsTrendsEnd = ""
	})

	tests := []struct {
		name  string
		cmd   *cobra.Command
		args  []string
		check func(t *testing.T, stdout string)
	}{
		{
			name: "endpoints",
			cmd:  statisticsEndpointsCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var endpoints v3.GetStatisticsEndpoints200Response
				mustUnmarshalOne(t, stdout, &endpoints)
				if endpoints.GetApiRoutes() != 120 || endpoints.GetGraphqlResolvers() != 12 || endpoints.GetWebPages() != 4 {
					t.Fatalf("endpoints = %#v", endpoints)
				}
			},
		},
		{
			name: "issues",
			cmd:  statisticsIssuesCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var issues v3.GetStatisticsIssues200Response
				mustUnmarshalOne(t, stdout, &issues)
				if issues.GetOpen() != 7 || issues.GetRemediated() != 3 {
					t.Fatalf("issues = %#v", issues)
				}

				mttr, _ := issues.GetAverageTimeToResolveOk()
				if mttr == nil || *mttr != 12.5 {
					t.Fatalf("mttr = %#v", mttr)
				}
			},
		},
		{
			name: "trends",
			cmd:  statisticsTrendsCmd,
			check: func(t *testing.T, stdout string) {
				t.Helper()

				var trends v3.GetStatisticsTrends200Response
				mustUnmarshalOne(t, stdout, &trends)
				buckets := trends.GetOpenResolved()
				if len(buckets) != 1 || buckets[0].GetOpen() != 9 {
					t.Fatalf("trends = %#v", trends)
				}

				health := trends.GetAverageApiHealth()
				if len(health) != 1 {
					t.Fatalf("health = %#v", health)
				}

				score, _ := health[0].GetYOk()
				if score == nil || *score != 82.5 {
					t.Fatalf("score = %#v", score)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetContext(context.Background())

			stdout, stderr, err := captureJSONCommand(t, func() error {
				return tt.cmd.RunE(tt.cmd, tt.args)
			})
			if err != nil {
				t.Fatalf("%s: %v\nstderr: %s", tt.name, err, stderr)
			}

			tt.check(t, stdout)
		})
	}
}

func TestStatisticsTrendsRequiresTheWindow(t *testing.T) {
	for _, name := range []string{"start", "end"} {
		flag := statisticsTrendsCmd.Flags().Lookup(name)
		if flag == nil || !isRequiredFlag(flag) {
			t.Fatalf("--%s must be marked required so the MCP schema advertises it", name)
		}
	}
}

// TestStatisticsRootStillPrintsThePosture guards the alias that became the
// group: "statistics" used to be an alias of "stats", so the bare group command
// keeps printing the organization posture instead of the help page.
func TestStatisticsRootStillPrintsThePosture(t *testing.T) {
	serveJSON(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/statistics" {
			http.NotFound(w, r)

			return
		}

		writeJSON(t, w, map[string]any{
			"applications": map[string]any{"count": 3},
			"assets":       map[string]any{"total": 42},
			"issues":       map[string]any{"critical": 1, "high": 5, "medium": 4, "low": 3, "info": 2},
		})
	})

	statisticsCmd.SetContext(context.Background())

	stdout, stderr, err := captureJSONCommand(t, func() error {
		return statisticsCmd.RunE(statisticsCmd, nil)
	})
	if err != nil {
		t.Fatalf("statistics: %v\nstderr: %s", err, stderr)
	}

	var stats v3.GetStatistics200Response
	mustUnmarshalOne(t, stdout, &stats)
	if stats.Applications.GetCount() != 3 || stats.Assets.GetTotal() != 42 {
		t.Fatalf("stats = %#v", stats)
	}
}
