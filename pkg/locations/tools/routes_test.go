package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRoutesCreateAndPollJob(t *testing.T) {
	m := newManager(context.Background(), 1, 20, time.Minute)
	defer m.stop()

	handler := newMux(m, map[string]tool{
		"fake": {
			name:             "fake",
			executionTimeout: func(map[string]any) (time.Duration, error) { return time.Second, nil },
			executor: func(_ context.Context, _ map[string]any, progress func(jobResult)) (jobResult, error) {
				progress(jobResult{"step": "running"})
				return jobResult{"ports": []int{443}}, nil
			},
		},
	})

	body := bytes.NewBufferString(`{"target":"example.com"}`)
	res := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tools/fake", body)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, res.Code, res.Body.String())
	}

	var created createJobResponse
	if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	if created.JobID == "" {
		t.Fatal("expected job_id")
	}

	assertEventually(t, func() bool {
		res = httptest.NewRecorder()
		req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/tools/jobs/"+created.JobID, nil)
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
		}

		var snapshot jobSnapshot
		if err := json.Unmarshal(res.Body.Bytes(), &snapshot); err != nil {
			t.Fatalf("decode snapshot: %v", err)
		}

		return snapshot.Status == jobStatusFinished
	})
}

func TestRoutesDeleteCancelsAndDropsJob(t *testing.T) {
	m := newManager(context.Background(), 1, 20, time.Minute)
	defer m.stop()

	blocked := make(chan struct{})
	handler := newMux(m, map[string]tool{
		"fake": {
			name:             "fake",
			executionTimeout: func(map[string]any) (time.Duration, error) { return time.Minute, nil },
			executor: func(ctx context.Context, _ map[string]any, _ func(jobResult)) (jobResult, error) {
				close(blocked)
				<-ctx.Done()

				return jobResult{"partial": true}, ctx.Err()
			},
		},
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tools/fake", bytes.NewBufferString(`{}`))
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, res.Code)
	}

	var created createJobResponse
	if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("executor did not start")
	}

	res = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/tools/jobs/"+created.JobID, nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, res.Code, res.Body.String())
	}

	res = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/tools/jobs/"+created.JobID, nil)
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected deleted job to be missing, got %d", res.Code)
	}
}

func TestRoutesReturnJSONTooManyRequestsWhenPendingJobsFull(t *testing.T) {
	m := newManager(context.Background(), 1, 0, time.Minute)
	defer m.stop()

	blocked := make(chan struct{})
	handler := newMux(m, map[string]tool{
		"fake": {
			name:             "fake",
			executionTimeout: func(map[string]any) (time.Duration, error) { return time.Minute, nil },
			executor: func(ctx context.Context, _ map[string]any, _ func(jobResult)) (jobResult, error) {
				close(blocked)
				<-ctx.Done()

				return nil, ctx.Err()
			},
		},
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tools/fake", bytes.NewBufferString(`{}`))
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, res.Code, res.Body.String())
	}

	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("executor did not start")
	}

	res = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tools/fake", bytes.NewBufferString(`{}`))
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d: %s", http.StatusTooManyRequests, res.Code, res.Body.String())
	}

	var body errorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	if body.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestRoutesRejectUnknownTool(t *testing.T) {
	m := newManager(context.Background(), 1, 20, time.Minute)
	defer m.stop()

	handler := newMux(m, map[string]tool{})
	res := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tools/missing", bytes.NewBufferString(`{}`))
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func assertEventually(t *testing.T, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if check() {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition was not met")
}
