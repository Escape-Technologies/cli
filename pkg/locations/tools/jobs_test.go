package tools

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerLimitsConcurrentExecutions(t *testing.T) {
	m := newManager(context.Background(), 1, 20, time.Minute)
	defer m.stop()

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var running int32
	executor := func(ctx context.Context, _ map[string]any, _ func(jobResult)) (jobResult, error) {
		current := atomic.AddInt32(&running, 1)
		if current > 1 {
			t.Errorf("expected one concurrent executor, got %d", current)
		}

		defer atomic.AddInt32(&running, -1)

		select {
		case firstStarted <- struct{}{}:
			<-releaseFirst
		default:
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return jobResult{"done": true}, nil
		}
	}

	first, err := m.create(context.Background(), nil, time.Second, executor)
	if err != nil {
		t.Fatalf("create first job: %v", err)
	}

	second, err := m.create(context.Background(), nil, time.Second, executor)
	if err != nil {
		t.Fatalf("create second job: %v", err)
	}

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("first job did not start")
	}

	snapshot, ok := m.get(second)
	if !ok {
		t.Fatal("second job missing")
	}

	if snapshot.Status != jobStatusRunning {
		t.Fatalf("expected queued job to appear running, got %s", snapshot.Status)
	}

	close(releaseFirst)
	assertJobStatus(t, m, first, jobStatusFinished)
	assertJobStatus(t, m, second, jobStatusFinished)
}

func TestManagerTimeoutStartsAfterSemaphoreWait(t *testing.T) {
	m := newManager(context.Background(), 1, 20, time.Minute)
	defer m.stop()

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	executor := func(ctx context.Context, _ map[string]any, _ func(jobResult)) (jobResult, error) {
		select {
		case firstStarted <- struct{}{}:
			<-releaseFirst
			return jobResult{"first": true}, nil
		default:
			select {
			case <-time.After(20 * time.Millisecond):
				return jobResult{"second": true}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if _, err := m.create(context.Background(), nil, 10*time.Millisecond, executor); err != nil {
		t.Fatalf("create first job: %v", err)
	}

	second, err := m.create(context.Background(), nil, 100*time.Millisecond, executor)
	if err != nil {
		t.Fatalf("create second job: %v", err)
	}

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("first job did not start")
	}

	time.Sleep(30 * time.Millisecond)
	close(releaseFirst)
	assertJobStatus(t, m, second, jobStatusFinished)
}

func TestManagerRejectsWhenPendingJobsFull(t *testing.T) {
	m := newManager(context.Background(), 1, 1, time.Minute)
	defer m.stop()

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var calls int32
	executor := func(ctx context.Context, _ map[string]any, _ func(jobResult)) (jobResult, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			close(firstStarted)
			<-releaseFirst

			return jobResult{"first": true}, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Millisecond):
			return jobResult{"queued": true}, nil
		}
	}

	if _, err := m.create(context.Background(), nil, time.Second, executor); err != nil {
		t.Fatalf("create first job: %v", err)
	}

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("first job did not start")
	}

	second, err := m.create(context.Background(), nil, time.Second, executor)
	if err != nil {
		t.Fatalf("create queued job: %v", err)
	}

	if _, err := m.create(context.Background(), nil, time.Second, executor); !errors.Is(err, errPendingJobsFull) {
		t.Fatalf("expected pending jobs full error, got %v", err)
	}

	close(releaseFirst)
	assertJobStatus(t, m, second, jobStatusFinished)
}

func TestManagerCancelsStaleJobs(t *testing.T) {
	m := newManager(context.Background(), 1, 20, 20*time.Millisecond)
	defer m.stop()

	started := make(chan struct{})
	id, err := m.create(context.Background(), nil, time.Second, func(ctx context.Context, _ map[string]any, _ func(jobResult)) (jobResult, error) {
		close(started)
		<-ctx.Done()

		return nil, ctx.Err()
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}

	assertEventually(t, func() bool {
		m.mu.Lock()
		_, ok := m.jobs[id]
		m.mu.Unlock()

		return !ok
	})
}

func TestManagerReportsFailedJobWithProgress(t *testing.T) {
	m := newManager(context.Background(), 1, 20, time.Minute)
	defer m.stop()

	id, err := m.create(context.Background(), nil, time.Second, func(_ context.Context, _ map[string]any, progress func(jobResult)) (jobResult, error) {
		progress(jobResult{"ports": []int{80}})
		return jobResult{"ports": []int{80}}, errors.New("boom")
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	assertJobStatus(t, m, id, jobStatusFailed)
	snapshot, _ := m.get(id)
	if snapshot.Error == "" {
		t.Fatal("expected error to be reported")
	}
}

func assertJobStatus(t *testing.T, m *manager, id string, status jobStatus) {
	t.Helper()

	assertEventually(t, func() bool {
		snapshot, ok := m.get(id)
		return ok && snapshot.Status == status
	})
}
