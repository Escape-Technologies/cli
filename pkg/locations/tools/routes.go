package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

const (
	apiBasePath = "/api/tools"
	apiPrefix   = apiBasePath + "/"
	jobPrefix   = apiBasePath + "/jobs/"
)

type tool struct {
	name             string
	executionTimeout func(map[string]any) (time.Duration, error)
	executor         jobExecutor
}

type createJobResponse struct {
	JobID string `json:"job_id"`
}

type deleteJobResponse struct {
	Status  string    `json:"status"`
	Results jobResult `json:"results"`
	Error   string    `json:"error,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func buildHandler(ctx context.Context, maxJobs int, maxPendingJobs int, ttl time.Duration) http.Handler {
	m := newManager(ctx, maxJobs, maxPendingJobs, ttl)
	tools := []tool{
		buildPortScannerTool(),
	}

	toolMap := make(map[string]tool)
	for _, tool := range tools {
		toolMap[tool.name] = tool
	}

	return newMux(m, toolMap)
}

func newMux(m *manager, tools map[string]tool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(apiBasePath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != apiBasePath {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		names := make([]string, 0, len(tools))
		for name := range tools {
			names = append(names, name)
		}

		writeJSON(w, http.StatusOK, map[string]any{"tools": names})
	})

	mux.HandleFunc(apiPrefix, func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, apiPrefix)
		if name == "" || strings.Contains(name, "/") {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tool, ok := tools[name]
		if !ok {
			http.NotFound(w, r)
			return
		}

		payload, err := decodePayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		timeout, err := tool.executionTimeout(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id, err := m.create(r.Context(), payload, timeout, tool.executor)
		if err != nil {
			if errors.Is(err, errPendingJobsFull) {
				writeJSON(w, http.StatusTooManyRequests, errorResponse{Error: err.Error()})
				return
			}

			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		log.Debug("Tools server created job %s for tool %s", id, name)
		writeJSON(w, http.StatusAccepted, createJobResponse{JobID: id})
	})

	mux.HandleFunc(jobPrefix, func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, jobPrefix)
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			snapshot, ok := m.get(id)
			if !ok {
				http.NotFound(w, r)
				return
			}

			writeJSON(w, http.StatusOK, snapshot)
		case http.MethodDelete:
			snapshot, ok := m.delete(id)
			if !ok {
				http.NotFound(w, r)
				return
			}

			response := deleteJobResponse{Status: "success", Results: snapshot.Results, Error: snapshot.Error}
			writeJSON(w, http.StatusOK, response)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

func decodePayload(r *http.Request) (map[string]any, error) {
	defer func() { _ = r.Body.Close() }()

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode request body: %w", err)
	}

	if payload == nil {
		return nil, errors.New("request body must be a JSON object")
	}

	return payload, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
