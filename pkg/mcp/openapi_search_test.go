package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func loadFixtureSpec(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "openapi_fixture.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return body
}

func newFixtureIndex(t *testing.T) *OpenAPISearchIndex {
	t.Helper()
	body := loadFixtureSpec(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return NewOpenAPISearchIndex(OpenAPISearchIndexOptions{SpecURL: srv.URL, TTL: time.Minute})
}

func TestOpenAPISearch_PicksScansForLastDaysQuestion(t *testing.T) {
	t.Parallel()

	index := newFixtureIndex(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	matches, servers, err := index.search(ctx, "how do I list scans that ran in the last 3 days", 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Method != "GET" || matches[0].Path != "/scans" {
		t.Fatalf("expected GET /scans, got %s %s", matches[0].Method, matches[0].Path)
	}
	if len(servers) == 0 || !strings.Contains(servers[0], "public.escape.tech") {
		t.Fatalf("expected fixture server URL, got %v", servers)
	}
}

func TestOpenAPISearch_PicksScanProblems(t *testing.T) {
	t.Parallel()

	index := newFixtureIndex(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	matches, _, err := index.search(ctx, "how to get all scan problems", 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected at least one match")
	}
	if matches[0].Path != "/scans/problems" {
		t.Fatalf("expected /scans/problems, got %s", matches[0].Path)
	}
}

func TestOpenAPISearch_PicksCreateAsset(t *testing.T) {
	t.Parallel()

	index := newFixtureIndex(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	matches, _, err := index.search(ctx, "create an asset", 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected match")
	}
	if matches[0].Method != "POST" || matches[0].Path != "/assets" {
		t.Fatalf("expected POST /assets, got %s %s", matches[0].Method, matches[0].Path)
	}
	if matches[0].RequestBody == nil || matches[0].RequestBody.Schema == nil {
		t.Fatalf("expected resolved request body schema")
	}
	if _, ok := matches[0].RequestBody.Schema.Properties["name"]; !ok {
		t.Fatalf("expected $ref to be resolved into Properties[name]")
	}
}

func TestParseOpenAPISpec_AcceptsOpenAPI31SchemaForms(t *testing.T) {
	t.Parallel()

	body := []byte(`{
	  "openapi": "3.1.0",
	  "servers": [{"url": "https://public.escape.tech/v3"}],
	  "paths": {
	    "/items/{id}": {
	      "parameters": [
	        {"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}
	      ],
	      "post": {
	        "operationId": "createItem",
	        "summary": "Create item",
	        "parameters": [
	          {"name": "after", "in": "query", "schema": {"anyOf": [{"type": "string", "format": "date-time"}, {"type": "null"}]}}
	        ],
	        "requestBody": {
	          "content": {
	            "application/json": {
	              "schema": {"$ref": "#/components/schemas/CreateItem"}
	            }
	          }
	        }
	      }
	    }
	  },
	  "components": {
	    "schemas": {
	      "CreateItem": {
	        "type": "object",
	        "properties": {
	          "name": {"type": ["string", "null"]},
	          "owner": {"$ref": "#/components/schemas/User"}
	        }
	      },
	      "User": {
	        "type": "object",
	        "properties": {
	          "id": {"type": "string"},
	          "manager": {"$ref": "#/components/schemas/User"}
	        }
	      }
	    }
	  }
	}`)

	ops, servers, err := parseOpenAPISpec(body)
	if err != nil {
		t.Fatalf("parseOpenAPISpec: %v", err)
	}
	if len(servers) != 1 || servers[0] != "https://public.escape.tech/v3" {
		t.Fatalf("unexpected servers: %v", servers)
	}
	if len(ops) != 1 {
		t.Fatalf("expected one operation, got %d", len(ops))
	}
	if len(ops[0].Parameters) != 2 || ops[0].Parameters[0].Name != "id" {
		t.Fatalf("expected path-level parameter to be included first, got %+v", ops[0].Parameters)
	}
	name := ops[0].RequestBody.Schema.Properties["name"]
	if name.Type != "string" {
		t.Fatalf("expected nullable type array to choose string, got %q", name.Type)
	}
	owner := ops[0].RequestBody.Schema.Properties["owner"]
	if owner == nil || owner.Properties["id"].Type != "string" {
		t.Fatalf("expected owner ref resolved, got %+v", owner)
	}
}

func TestOpenAPISearch_LimitClamping(t *testing.T) {
	t.Parallel()

	index := newFixtureIndex(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	matches, _, err := index.search(ctx, "scans", 99)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(matches) > openapiMaxResultsPerQuery {
		t.Fatalf("expected limit clamped to %d, got %d", openapiMaxResultsPerQuery, len(matches))
	}

	matches, _, err = index.search(ctx, "scans", 0)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected limit clamped up to 1, got %d", len(matches))
	}
}

func TestOpenAPISearch_EmptyQuestionReturnsNoMatch(t *testing.T) {
	t.Parallel()

	index := newFixtureIndex(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	matches, _, err := index.search(ctx, "   ", 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no matches for blank question, got %d", len(matches))
	}
}

func TestOpenAPISearch_FetchErrorPropagates(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	t.Cleanup(srv.Close)

	index := NewOpenAPISearchIndex(OpenAPISearchIndexOptions{SpecURL: srv.URL, TTL: time.Minute})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := index.search(ctx, "list scans", 1)
	if err == nil {
		t.Fatalf("expected error on 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}
