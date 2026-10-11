package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestKnowledgeFilesUploadRunsTheThreeStepFlow drives the whole upload: the
// signed URL, the PUT of the bytes, then registering the file with the id that
// URL returned. The API describes the three steps and leaves them to the
// caller, so pinning them here is what keeps the command working.
func TestKnowledgeFilesUploadRunsTheThreeStepFlow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "architecture.md")
	if err := os.WriteFile(path, []byte("# Architecture\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var (
		signedURL string
		putBytes  string
		putType   string
		register  map[string]any
	)
	// Declared up front: the handler refers to it, and a short declaration
	// would sit outside its own scope until the statement ends.
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/upload/signed-url":
			signedURL = server.URL + "/signed"
			_, _ = io.WriteString(w, `{"id":"11111111-1111-1111-1111-111111111111","url":"`+signedURL+`"}`)
		case r.Method == http.MethodPut && r.URL.Path == "/signed":
			data, _ := io.ReadAll(r.Body)
			putBytes = string(data)
			putType = r.Header.Get("Content-Type")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/v3/knowledge-files":
			if err := json.NewDecoder(r.Body).Decode(&register); err != nil {
				t.Errorf("register body: %v", err)
			}

			_, _ = io.WriteString(w, `{"id":"`+parityID+`","name":"architecture.md","kind":"MD","status":"PENDING","createdAt":"2026-01-01T00:00:00Z"}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	previousFile, previousName, previousKind := knowledgeFileUploadFile, knowledgeFileUploadName, knowledgeFileUploadKind
	t.Cleanup(func() {
		knowledgeFileUploadFile = previousFile
		knowledgeFileUploadName = previousName
		knowledgeFileUploadKind = previousKind
	})
	knowledgeFileUploadFile = path
	knowledgeFileUploadName = ""
	knowledgeFileUploadKind = ""

	knowledgeFilesUploadCmd.SetIn(strings.NewReader(""))
	knowledgeFilesUploadCmd.SetOut(io.Discard)
	knowledgeFilesUploadCmd.SetErr(io.Discard)
	knowledgeFilesUploadCmd.SetContext(context.Background())

	if err := knowledgeFilesUploadCmd.RunE(knowledgeFilesUploadCmd, nil); err != nil {
		t.Fatalf("upload: %v", err)
	}

	if putBytes != "# Architecture\n" {
		t.Fatalf("PUT body = %q, want the file bytes", putBytes)
	}

	if putType != "text/markdown" {
		t.Fatalf("PUT content type = %q, want text/markdown", putType)
	}

	// The kind comes from the extension and the object key from the signed
	// upload, neither of them from the caller.
	if register["name"] != "architecture.md" || register["kind"] != "MD" {
		t.Fatalf("register = %#v, want the file name and the kind derived from it", register)
	}

	if register["temporaryObjectKey"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("temporaryObjectKey = %#v, want the id of the signed upload", register["temporaryObjectKey"])
	}
}

// TestKnowledgeFilesUploadRequiresANameOnStdin keeps the piped flow usable: the
// stored name cannot be derived from a path that does not exist.
func TestKnowledgeFilesUploadRequiresANameOnStdin(t *testing.T) {
	previousFile, previousName := knowledgeFileUploadFile, knowledgeFileUploadName
	t.Cleanup(func() {
		knowledgeFileUploadFile = previousFile
		knowledgeFileUploadName = previousName
	})
	knowledgeFileUploadFile = ""
	knowledgeFileUploadName = ""

	knowledgeFilesUploadCmd.SetIn(strings.NewReader("report.pdf"))
	knowledgeFilesUploadCmd.SetOut(io.Discard)
	knowledgeFilesUploadCmd.SetErr(io.Discard)
	knowledgeFilesUploadCmd.SetContext(context.Background())

	err := knowledgeFilesUploadCmd.RunE(knowledgeFilesUploadCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--name is required") {
		t.Fatalf("err = %v, want the --name requirement", err)
	}
}
