package metrics

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// pinBaseLabels pins the process env the rendered base labels read from, so
// expectations do not depend on the machine running the tests.
func pinBaseLabels(t *testing.T) {
	t.Helper()
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("SERVICE", "")
	t.Setenv("HOSTNAME", "")
}

func findLine(payload string, needle string) string {
	for line := range strings.SplitSeq(payload, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}

	return ""
}

func TestCounterRendersLabelValues(t *testing.T) {
	pinBaseLabels(t)
	counter := NewCounter("test_counter_total", "tool", "result")
	counter.Inc("scans_get", "ok")
	counter.Inc("scans_get", "ok")
	counter.Inc("scans_get", "error")

	payload := new(bytes.Buffer)
	WriteText(payload)

	want := `test_counter_total{container="unknown",k8s_cluster_name="unknown",pod="unknown",tool="scans_get",result="error"} 1`
	if got := findLine(payload.String(), want); got != want {
		t.Fatalf("counter line mismatch:\n got  %s\n want %s", got, want)
	}
}

func TestMetricRejectsMismatchedLabelArity(t *testing.T) {
	counter := NewCounter("test_arity_total", "tool")
	counter.Inc("a", "b")

	payload := new(bytes.Buffer)
	WriteText(payload)
	if strings.Contains(payload.String(), "test_arity_total") {
		t.Fatal("mismatched label arity must not create a series")
	}
}

func TestGaugeSetReplacesValue(t *testing.T) {
	gauge := NewGauge("test_gauge")
	gauge.Set(3)
	gauge.Set(7)
	gauge.Dec()

	payload := new(bytes.Buffer)
	WriteText(payload)
	if got := findLine(payload.String(), "test_gauge{"); !strings.HasSuffix(got, " 6") {
		t.Fatalf("gauge should end at 6, got %q", got)
	}
}

func TestHistogramRendersCumulativeBuckets(t *testing.T) {
	pinBaseLabels(t)
	histogram := NewHistogram("test_latency_seconds", []float64{0.5, 1, 10})
	histogram.Observe(0.2)
	histogram.Observe(0.7)
	histogram.Observe(20)

	payload := new(bytes.Buffer)
	WriteText(payload)
	text := payload.String()

	base := `container="unknown",k8s_cluster_name="unknown",pod="unknown"`
	for _, want := range []string{
		`test_latency_seconds_bucket{` + base + `,le="0.5"} 1`,
		`test_latency_seconds_bucket{` + base + `,le="1"} 2`,
		`test_latency_seconds_bucket{` + base + `,le="10"} 2`,
		`test_latency_seconds_bucket{` + base + `,le="+Inf"} 3`,
		`test_latency_seconds_count{` + base + `} 3`,
		`test_latency_seconds_sum{` + base + `} 20.9`,
	} {
		if got := findLine(text, want); got != want {
			t.Fatalf("histogram line mismatch:\n got  %s\n want %s", got, want)
		}
	}
}

func TestEscapeLabelValue(t *testing.T) {
	pinBaseLabels(t)
	counter := NewCounter("test_escape_total", "reason")
	counter.Inc(`say "hi"\new`)

	payload := new(bytes.Buffer)
	WriteText(payload)
	if !strings.Contains(payload.String(), `reason="say \"hi\"\\new"`) {
		t.Fatalf("label value not escaped: %s", payload.String())
	}
}

func TestPushLoopSendsAndFlushesOnCancel(t *testing.T) {
	received := make(chan []byte, 4)
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		body := new(bytes.Buffer)
		_, _ = body.ReadFrom(r.Body)
		received <- body.Bytes()
	}))
	defer server.Close()

	t.Setenv("VICTORIA_METRICS_PUSH_URL", server.URL)
	originalInterval := pushInterval
	pushInterval = 10 * time.Millisecond
	defer func() { pushInterval = originalInterval }()

	counter := NewCounter("test_push_total", "result")
	counter.Inc("ok")

	ctx, cancel := context.WithCancel(t.Context())
	Start(ctx)
	select {
	case payload := <-received:
		if !strings.Contains(string(payload), "test_push_total") {
			t.Fatalf("push payload missing metric: %s", payload)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no periodic push received")
	}

	counter.Inc("ok")
	cancel()
	select {
	case payload := <-received:
		if !strings.Contains(string(payload), `result="ok"`) {
			t.Fatalf("final push payload missing series: %s", payload)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no final push after cancel")
	}
}

func TestStartWithoutURLOnlyRegisters(t *testing.T) {
	t.Setenv("VICTORIA_METRICS_PUSH_URL", "")
	// Must return immediately and not panic; no goroutine, no push.
	Start(t.Context())
}

func TestHandlerServesExposition(t *testing.T) {
	pinBaseLabels(t)
	counter := NewCounter("test_handler_total", "result")
	counter.Inc("ok")

	recorder := httptest.NewRecorder()
	Handler().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", http.NoBody))
	if recorder.Code != http.StatusOK {
		t.Fatalf("handler status %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), `test_handler_total{`) {
		t.Fatalf("handler payload missing metric: %s", recorder.Body.String())
	}
}

func TestGaugeFuncComputesAtRender(t *testing.T) {
	pinBaseLabels(t)
	value := 0.0
	NewGaugeFunc("test_gaugefunc", func() float64 { return value })

	payload := new(bytes.Buffer)
	WriteText(payload)
	if got := findLine(payload.String(), "test_gaugefunc{"); !strings.HasSuffix(got, " 0") {
		t.Fatalf("gaugefunc should render 0 first, got %q", got)
	}

	value = 7
	payload = new(bytes.Buffer)
	WriteText(payload)
	if got := findLine(payload.String(), "test_gaugefunc{"); !strings.HasSuffix(got, " 7") {
		t.Fatalf("gaugefunc should re-evaluate at render, got %q", got)
	}
}
