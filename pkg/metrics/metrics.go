// Package metrics is a self-contained Prometheus exporter for the CLI and its
// embedded MCP server.
//
// The public GitHub mirror forbids imports of internal monorepo packages
// (`just cli-check-isolation`), so this package copies the push pattern of
// the internal sdk/go_metrics instead of importing it: counters, gauges, and
// histograms are held in process, rendered as Prometheus text exposition, and
// POSTed to VICTORIA_METRICS_PUSH_URL on a fixed interval. Without that env
// var the push loop never starts and metrics stay pull-only via Handler.
package metrics

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	pushTimeout = 3 * time.Second
	// pushErrorBodyLimit caps how much of a rejected push response is logged.
	pushErrorBodyLimit = 512
)

// pushInterval is how often the registry is flushed to VictoriaMetrics.
// Matches sdk/go_metrics so dashboards see the same cadence everywhere. A var
// so tests can shrink it.
var pushInterval = 15 * time.Second

const (
	environmentEnv     = "ENVIRONMENT"
	serviceEnv         = "OTEL_SERVICE_NAME"
	serviceFallbackEnv = "SERVICE"
	podEnv             = "HOSTNAME"

	unknownLabelValue = "unknown"
)

var pushHTTPClient = &http.Client{Timeout: pushTimeout}

// registry is the process-wide set of exported families.
var (
	registryMu sync.RWMutex
	registry   []family
)

type family interface {
	// Name orders families in the rendered payload.
	Name() string
	// write appends the family's text exposition samples, prefixed with the
	// process base labels.
	write(w *bytes.Buffer, base map[string]string)
}

func register(f family) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, f)
}

// Counter is a monotonically increasing metric split by label values.
type Counter struct {
	name   string
	labels []string

	mu     sync.Mutex
	series map[string]float64
}

// NewCounter registers a counter family. labelNames must match the length of
// every labelValues argument passed to Inc/Add; mismatched calls are dropped.
func NewCounter(name string, labelNames ...string) *Counter {
	counter := &Counter{name: name, labels: labelNames, series: map[string]float64{}}
	register(counter)

	return counter
}

// Inc adds one to the series identified by labelValues.
func (c *Counter) Inc(labelValues ...string) {
	c.Add(1, labelValues...)
}

// Add adds delta to the series identified by labelValues.
func (c *Counter) Add(delta float64, labelValues ...string) {
	if len(labelValues) != len(c.labels) {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.series[labelKey(labelValues)] += delta
}

// Name returns the metric name, used to order families in the payload.
func (c *Counter) Name() string { return c.name }

func (c *Counter) write(w *bytes.Buffer, base map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range sortedKeys(c.series) {
		writeSample(w, c.name, base, c.labels, splitKey(key), "", "")
		fmt.Fprintf(w, " %g\n", c.series[key])
	}
}

// Gauge is a metric that can go up and down, split by label values.
type Gauge struct {
	name   string
	labels []string

	mu     sync.Mutex
	series map[string]float64
}

// NewGauge registers a gauge family. labelNames must match the length of
// every labelValues argument passed to Set/Inc/Dec; mismatched calls are dropped.
func NewGauge(name string, labelNames ...string) *Gauge {
	gauge := &Gauge{name: name, labels: labelNames, series: map[string]float64{}}
	register(gauge)

	return gauge
}

// Set replaces the value of the series identified by labelValues.
func (g *Gauge) Set(value float64, labelValues ...string) {
	if len(labelValues) != len(g.labels) {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.series[labelKey(labelValues)] = value
}

// Inc adds one to the series identified by labelValues.
func (g *Gauge) Inc(labelValues ...string) {
	g.Add(1, labelValues...)
}

// Dec subtracts one from the series identified by labelValues.
func (g *Gauge) Dec(labelValues ...string) {
	g.Add(-1, labelValues...)
}

// Add adds delta to the series identified by labelValues.
func (g *Gauge) Add(delta float64, labelValues ...string) {
	if len(labelValues) != len(g.labels) {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.series[labelKey(labelValues)] += delta
}

// Get returns the current value of the series identified by labelValues.
func (g *Gauge) Get(labelValues ...string) (float64, bool) {
	if len(labelValues) != len(g.labels) {
		return 0, false
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	value, ok := g.series[labelKey(labelValues)]

	return value, ok
}

// GaugeFunc computes its value at every render instead of storing one. Use it
// for counts derived from other state, where a stored value would go stale
// when traffic stops.
type GaugeFunc struct {
	name string
	fn   func() float64
}

// NewGaugeFunc registers a computed gauge family. fn runs on every render
// (push tick or scrape) and must be cheap and lock-safe.
func NewGaugeFunc(name string, fn func() float64) *GaugeFunc {
	gauge := &GaugeFunc{name: name, fn: fn}
	register(gauge)

	return gauge
}

// Name returns the metric name, used to order families in the payload.
func (g *GaugeFunc) Name() string { return g.name }

func (g *GaugeFunc) write(w *bytes.Buffer, base map[string]string) {
	writeSample(w, g.name, base, nil, nil, "", "")
	fmt.Fprintf(w, " %g\n", g.fn())
}

// Name returns the metric name, used to order families in the payload.
func (g *Gauge) Name() string { return g.name }

func (g *Gauge) write(w *bytes.Buffer, base map[string]string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, key := range sortedKeys(g.series) {
		writeSample(w, g.name, base, g.labels, splitKey(key), "", "")
		fmt.Fprintf(w, " %g\n", g.series[key])
	}
}

// defaultBuckets covers sub-millisecond to multi-second request latencies,
// matching the buckets the platform httpapi middleware uses.
var defaultBuckets = []float64{
	0.0001, 0.00025, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1,
	0.25, 0.5, 1, 2.5, 5, 10,
}

// Histogram is a cumulative distribution split by label values. Buckets are
// fixed upper bounds; the implicit +Inf bucket tracks every observation.
type Histogram struct {
	name    string
	labels  []string
	buckets []float64

	mu     sync.Mutex
	series map[string]*histogramSeries
}

type histogramSeries struct {
	counts []uint64
	sum    float64
	count  uint64
}

// NewHistogram registers a histogram family. A nil buckets argument uses
// defaultBuckets. labelNames must match the length of every labelValues
// argument passed to Observe; mismatched calls are dropped.
func NewHistogram(name string, buckets []float64, labelNames ...string) *Histogram {
	if buckets == nil {
		buckets = defaultBuckets
	}

	histogram := &Histogram{
		name:    name,
		labels:  labelNames,
		buckets: buckets,
		series:  map[string]*histogramSeries{},
	}
	register(histogram)

	return histogram
}

// Observe records one value in the series identified by labelValues.
func (h *Histogram) Observe(value float64, labelValues ...string) {
	if len(labelValues) != len(h.labels) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	key := labelKey(labelValues)
	series, ok := h.series[key]
	if !ok {
		series = &histogramSeries{counts: make([]uint64, len(h.buckets))}
		h.series[key] = series
	}

	for i, bound := range h.buckets {
		if value <= bound {
			series.counts[i]++
		}
	}

	series.sum += value
	series.count++
}

// Name returns the metric name, used to order families in the payload.
func (h *Histogram) Name() string { return h.name }

func (h *Histogram) write(w *bytes.Buffer, base map[string]string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, key := range sortedKeys(h.series) {
		series := h.series[key]
		for i, bound := range h.buckets {
			writeSample(w, h.name+"_bucket", base, h.labels, splitKey(key), "le", formatBound(bound))
			fmt.Fprintf(w, " %d\n", series.counts[i])
		}

		writeSample(w, h.name+"_bucket", base, h.labels, splitKey(key), "le", "+Inf")
		fmt.Fprintf(w, " %d\n", series.count)

		writeSample(w, h.name+"_sum", base, h.labels, splitKey(key), "", "")
		fmt.Fprintf(w, " %g\n", series.sum)

		writeSample(w, h.name+"_count", base, h.labels, splitKey(key), "", "")
		fmt.Fprintf(w, " %d\n", series.count)
	}
}

func formatBound(bound float64) string {
	return strconv.FormatFloat(bound, 'f', -1, 64)
}

// pushErrors counts failed pushes so the exporter's own health stays visible.
var pushErrors = NewCounter("metrics_push_errors_total")

// Start starts the push loop when VICTORIA_METRICS_PUSH_URL is set. Without
// the env var it is a no-op, so an unconfigured local `mcp serve` never sends
// anything. A final flush fires when ctx ends, matching sdk/go_metrics.
func Start(ctx context.Context) {
	url := pushURL()
	if url == "" {
		return
	}

	slog.InfoContext(ctx, "metrics: pushing to VictoriaMetrics", "interval", pushInterval)
	go pushLoop(ctx, url)
}

func pushLoop(ctx context.Context, url string) {
	ticker := time.NewTicker(pushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pushOnce(ctx, url)
		case <-ctx.Done():
			// Detach so the final flush is not bounded by an already
			// cancelled server context.
			pushCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), pushTimeout)
			if err := push(pushCtx, url); isPushError(err) {
				slog.WarnContext(ctx, "metrics: final push to VictoriaMetrics failed", "error", err)
			}

			cancel()

			return
		}
	}
}

func pushOnce(ctx context.Context, url string) {
	pushCtx, cancel := context.WithTimeout(ctx, pushTimeout)
	err := push(pushCtx, url)
	cancel()
	if isPushError(err) {
		slog.WarnContext(ctx, "metrics: push to VictoriaMetrics failed", "error", err)
	}
}

func isPushError(err error) bool {
	return err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func pushURL() string {
	return strings.TrimSpace(os.Getenv("VICTORIA_METRICS_PUSH_URL"))
}

func push(ctx context.Context, url string) error {
	payload := new(bytes.Buffer)
	WriteText(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload.Bytes()))
	if err != nil {
		pushErrors.Inc()

		return fmt.Errorf("build push request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if token := strings.TrimSpace(os.Getenv("VICTORIA_METRICS_BASIC_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Basic "+token)
	}

	res, err := pushHTTPClient.Do(req)
	if err != nil {
		pushErrors.Inc()

		return fmt.Errorf("push metrics: %w", err)
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(res.Body, pushErrorBodyLimit))
		pushErrors.Inc()

		return fmt.Errorf("push rejected with status %d: %s", res.StatusCode, string(body))
	}

	return nil
}

// Handler serves the current registry as Prometheus text exposition. It is
// meant for an internal scrape port, not the public MCP endpoint.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		payload := new(bytes.Buffer)
		WriteText(payload)
		_, _ = w.Write(payload.Bytes())
	})
}

// WriteText renders every registered family as Prometheus text exposition,
// including the shared process labels other Go services attach
// (k8s_cluster_name, container, pod).
func WriteText(w *bytes.Buffer) {
	base := map[string]string{
		"k8s_cluster_name": envOr(environmentEnv),
		"container":        serviceName(),
		"pod":              envOr(podEnv),
	}

	registryMu.RLock()
	families := make([]family, len(registry))
	copy(families, registry)
	registryMu.RUnlock()

	sort.Slice(families, func(i, j int) bool { return families[i].Name() < families[j].Name() })
	for _, family := range families {
		family.write(w, base)
	}
}

func envOr(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return unknownLabelValue
	}

	return value
}

func serviceName() string {
	if service := strings.TrimSpace(os.Getenv(serviceEnv)); service != "" {
		return service
	}

	return envOr(serviceFallbackEnv)
}

func writeSample(w *bytes.Buffer, name string, base map[string]string, labelNames, labelValues []string, extraName, extraValue string) {
	w.WriteString(name)
	w.WriteString("{")
	first := true
	for _, key := range sortedKeys(base) {
		writeLabel(w, &first, key, base[key])
	}

	for i, label := range labelNames {
		writeLabel(w, &first, label, labelValues[i])
	}

	if extraName != "" {
		writeLabel(w, &first, extraName, extraValue)
	}

	w.WriteString("}")
}

func writeLabel(w *bytes.Buffer, first *bool, key, value string) {
	if !*first {
		w.WriteString(",")
	}

	*first = false
	w.WriteString(key)
	w.WriteString(`="`)
	w.WriteString(escapeLabelValue(value))
	w.WriteString(`"`)
}

func escapeLabelValue(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(value)
}

func labelKey(labelValues []string) string {
	return strings.Join(labelValues, "\x1f")
}

func splitKey(key string) []string {
	return strings.Split(key, "\x1f")
}

func sortedKeys[V any](series map[string]V) []string {
	keys := make([]string, 0, len(series))
	for key := range series {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
