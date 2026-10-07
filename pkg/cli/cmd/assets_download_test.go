package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func runAssetDownload(t *testing.T) error {
	t.Helper()
	assetDownloadCmd.SetContext(context.Background())
	t.Cleanup(func() { assetDownloadCmd.SetContext(context.TODO()) })

	if err := assetDownloadCmd.RunE(assetDownloadCmd, []string{assetDownloadTestID}); err != nil {
		return fmt.Errorf("run asset download: %w", err)
	}

	return nil
}

const assetDownloadTestID = "00000000-0000-0000-0000-000000000001"

// assetDownloadTestServer serves the content endpoint and a fake signed URL
// blob. The blob body is what the download command is expected to persist.
func assetDownloadTestServer(t *testing.T, body string) *httptest.Server {
	t.Helper()

	var signedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/assets/" + assetDownloadTestID + "/content":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"signedUrl":"` + signedURL + `","contentType":"application/yaml","filename":"schema.yaml"}`))
		case "/blob":
			_, _ = w.Write([]byte(body))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	signedURL = server.URL + "/blob"

	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	return server
}

func TestAssetDownloadWritesSignedURLToFile(t *testing.T) {
	const body = "openapi: 3.0.0\n"
	assetDownloadTestServer(t, body)

	dest := filepath.Join(t.TempDir(), "schema.yaml")
	prevOutput := assetDownloadOutputFile
	assetDownloadOutputFile = dest
	t.Cleanup(func() { assetDownloadOutputFile = prevOutput })

	if err := runAssetDownload(t); err != nil {
		t.Fatalf("download: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read destination: %v", err)
	}

	if string(got) != body {
		t.Fatalf("content = %q, want %q", got, body)
	}
}

func TestAssetDownloadStreamsSignedURLToStdout(t *testing.T) {
	const body = "openapi: 3.0.0\n"
	assetDownloadTestServer(t, body)

	prevOutput := assetDownloadOutputFile
	assetDownloadOutputFile = ""
	t.Cleanup(func() { assetDownloadOutputFile = prevOutput })

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	original := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = original })

	runErr := runAssetDownload(t)
	os.Stdout = original

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if runErr != nil {
		t.Fatalf("download: %v", runErr)
	}

	if string(got) != body {
		t.Fatalf("stdout = %q, want %q", got, body)
	}
}

func TestAssetDownloadFailureLeavesNoFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v3/assets/"+assetDownloadTestID+"/content" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"signedUrl":"` + "http://" + r.Host + `/missing","contentType":"application/yaml","filename":"schema.yaml"}`))

			return
		}

		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	t.Setenv("ESCAPE_API_URL", server.URL)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")

	dir := t.TempDir()
	dest := filepath.Join(dir, "schema.yaml")
	prevOutput := assetDownloadOutputFile
	assetDownloadOutputFile = dest
	t.Cleanup(func() { assetDownloadOutputFile = prevOutput })

	if err := runAssetDownload(t); err == nil {
		t.Fatal("expected a download error")
	}

	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("destination exists after a failed download: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("failed download left temp files behind: %v", entries)
	}
}
