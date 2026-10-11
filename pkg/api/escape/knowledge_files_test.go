package escape

import "testing"

func TestKnowledgeFileKindFromExtensionIgnoresCase(t *testing.T) {
	for name, want := range map[string]string{
		"spec.pdf":    "PDF",
		"SPEC.PDF":    "PDF",
		"Notes.Md":    "MD",
		"archive.TGZ": "TAR_GZ",
		"unknown.XYZ": "FILE",
	} {
		got, ok := knowledgeFileKindFromExtension(name)
		if !ok || got != want {
			t.Errorf("knowledgeFileKindFromExtension(%q) = %q, %v; want %q", name, got, ok, want)
		}
	}

	if _, ok := knowledgeFileKindFromExtension("README"); ok {
		t.Error("a name without an extension must not resolve to a kind")
	}
}
