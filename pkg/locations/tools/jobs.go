package tools

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/google/uuid"
)

type jobStatus string

const (
	jobStatusRunning  jobStatus = "running"
	jobStatusFailed   jobStatus = "failed"
	jobStatusCanceled jobStatus = "canceled"
	jobStatusFinished jobStatus = "finished"
	ttlCheckDivisor             = 2
	minTickerTTL                = 2 * time.Second
)

type jobResult map[string]any

type jobExecutor func(context.Context, map[string]any, func(jobResult)) (jobResult, error)

var errPendingJobsFull = errors.New("too many pending jobs")

type job struct {
	id     string
	cancel context.CancelFunc

	mu       sync.Mutex
	status   jobStatus
	results  jobResult
	err      string
	lastPing time.Time
}

func (j *job) snapshot() jobSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()

	return jobSnapshot{
		Status:  j.status,
		Results: copyResult(j.results),
		Error:   j.err,
	}
}

func (j *job) touch(now time.Time) {
	j.mu.Lock()
	j.lastPing = now
	j.mu.Unlock()
}

func (j *job) setStatus(status jobStatus, results jobResult, err error) {
	log.Debug("Setting job %s status to %s", j.id, status)
	j.mu.Lock()
	defer j.mu.Unlock()

	j.status = status
	if results != nil {
		j.results = copyResult(results)
	}

	if err != nil {
		j.err = err.Error()
	}
}

func (j *job) stale(now time.Time, ttl time.Duration) bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	return now.Sub(j.lastPing) > ttl
}

func (j *job) done() bool {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.status == jobStatusFailed || j.status == jobStatusCanceled || j.status == jobStatusFinished
}

type jobSnapshot struct {
	Status  jobStatus `json:"status"`
	Results jobResult `json:"results"`
	Error   string    `json:"error,omitempty"`
}

type manager struct {
	cancel       context.CancelFunc
	done         <-chan struct{}
	sem          chan struct{}
	ttl          time.Duration
	pendingLimit int

	mu      sync.Mutex
	pending int
	jobs    map[string]*job
}

func newManager(ctx context.Context, maxJobs int, maxPendingJobs int, ttl time.Duration) *manager {
	if maxJobs < 1 {
		maxJobs = 1
	}

	if maxPendingJobs < 0 {
		maxPendingJobs = 0
	}

	childCtx, cancel := context.WithCancel(ctx)
	m := &manager{
		cancel:       cancel,
		done:         childCtx.Done(),
		sem:          make(chan struct{}, maxJobs),
		ttl:          ttl,
		pendingLimit: maxPendingJobs,
		jobs:         map[string]*job{},
	}

	go m.cleanupLoop()

	return m
}

func (m *manager) stop() {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range m.jobs {
		j.cancel()
	}
}

func (m *manager) create(ctx context.Context, payload map[string]any, timeout time.Duration, executor jobExecutor) (string, error) {
	if timeout <= 0 {
		return "", errors.New("timeout must be positive")
	}

	hasSlot := false
	isPending := false
	select {
	case m.sem <- struct{}{}:
		hasSlot = true
	default:
		m.mu.Lock()
		if m.pending >= m.pendingLimit {
			m.mu.Unlock()
			return "", errPendingJobsFull
		}

		m.pending++
		isPending = true
		m.mu.Unlock()
	}

	id := uuid.NewString()
	jobCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	j := &job{
		id:       id,
		cancel:   cancel,
		status:   jobStatusRunning,
		results:  jobResult{},
		lastPing: time.Now(),
	}

	m.mu.Lock()
	m.jobs[id] = j
	m.mu.Unlock()

	go func() {
		select {
		case <-m.done:
		case <-jobCtx.Done():
		}

		cancel()
	}()
	go func() {
		defer cancel()
		m.run(jobCtx, j, payload, timeout, executor, hasSlot, isPending)
	}()

	return id, nil
}

func (m *manager) run(ctx context.Context, j *job, payload map[string]any, timeout time.Duration, executor jobExecutor, hasSlot bool, isPending bool) {
	if !hasSlot {
		select {
		case m.sem <- struct{}{}:
			hasSlot = true
			if isPending {
				m.releasePending()
			}
		case <-ctx.Done():
			if isPending {
				m.releasePending()
			}

			j.setStatus(jobStatusCanceled, nil, ctx.Err())

			return
		case <-m.done:
			if isPending {
				m.releasePending()
			}

			j.setStatus(jobStatusCanceled, nil, errors.New("job manager stopped"))

			return
		}
	}

	if hasSlot {
		defer func() { <-m.sem }()
	}

	log.Debug("Creating context for job %s with timeout %s", j.id, timeout)
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Debug("Executing job %s", j.id)
	results, err := executor(runCtx, payload, func(progress jobResult) {
		j.setStatus(jobStatusRunning, progress, nil)
	})
	// Check if the context was canceled or timed out
	if err == nil {
		err = runCtx.Err()
	}

	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(runCtx.Err(), context.Canceled) {
			j.setStatus(jobStatusCanceled, results, err)
			return
		}

		j.setStatus(jobStatusFailed, results, err)

		return
	}

	log.Debug("Job %s finished, results: %v", j.id, results)

	j.setStatus(jobStatusFinished, results, nil)
}

func (m *manager) releasePending() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pending--
}

func (m *manager) get(id string) (jobSnapshot, bool) {
	m.mu.Lock()
	j, ok := m.jobs[id]
	m.mu.Unlock()
	if !ok {
		return jobSnapshot{}, false
	}

	j.touch(time.Now())

	return j.snapshot(), true
}

func (m *manager) delete(id string) (jobSnapshot, bool) {
	m.mu.Lock()
	j, ok := m.jobs[id]
	if ok {
		delete(m.jobs, id)
	}

	m.mu.Unlock()
	if !ok {
		return jobSnapshot{}, false
	}

	j.cancel()
	if !j.done() {
		j.setStatus(jobStatusCanceled, nil, errors.New("job canceled"))
	}

	return j.snapshot(), true
}

func (m *manager) cleanupLoop() {
	ticker := time.NewTicker(m.ttl / ttlCheckDivisor)
	if m.ttl < minTickerTTL {
		ticker.Reset(m.ttl)
	}

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cancelStale(time.Now())
		case <-m.done:
			return
		}
	}
}

func (m *manager) cancelStale(now time.Time) {
	m.mu.Lock()
	var stale []*job
	for id, j := range m.jobs {
		if j.stale(now, m.ttl) {
			delete(m.jobs, id)
			stale = append(stale, j)
		}
	}

	m.mu.Unlock()

	for _, j := range stale {
		j.cancel()
		j.setStatus(jobStatusCanceled, nil, errors.New("job canceled after missing status checks"))
	}
}

func copyResult(in jobResult) jobResult {
	if in == nil {
		return jobResult{}
	}

	out := make(jobResult, len(in))
	for k, v := range in {
		out[k] = v
	}

	return out
}
