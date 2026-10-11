package escape

import "testing"

func TestFileContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "report.pdf", want: "application/pdf"},
		{name: "notes.MD", want: "text/markdown"},
		{name: "spec.yaml", want: "application/yaml"},
		{name: "openapi.json", want: "application/json"},
		{name: "findings.csv", want: "text/csv"},
		{name: "readme.txt", want: "text/plain"},
		{name: "archive.unknownext", want: "application/octet-stream"},
		{name: "noextension", want: "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fileContentType(tt.name); got != tt.want {
				t.Fatalf("fileContentType(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestSchemaUploadContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "json object",
			data: []byte(`{"openapi":"3.0.0"}`),
			want: "application/json",
		},
		{
			name: "yaml document",
			data: []byte("openapi: 3.0.3\ninfo:\n  title: org\n"),
			want: "application/yaml",
		},
		{
			name: "yaml flow mapping",
			data: []byte("{openapi: 3.0.3, info: {title: org, version: \"1.0.0\"}}\n"),
			want: "application/yaml",
		},
		{
			name: "leading whitespace json",
			data: []byte("\n{\"openapi\":\"3.0.0\"}\n"),
			want: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := schemaUploadContentType(tt.data); got != tt.want {
				t.Fatalf("schemaUploadContentType() = %q, want %q", got, tt.want)
			}
		})
	}
}
