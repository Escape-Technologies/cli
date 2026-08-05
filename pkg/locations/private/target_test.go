package private

import "testing"

func TestSSHTarget(t *testing.T) {
	t.Setenv("ESCAPE_PRIVATE_LOCATION_URL", "custom.example:2222")
	if got := sshTarget(); got != "custom.example:2222" {
		t.Fatalf("sshTarget() = %q, want custom override", got)
	}

	t.Setenv("ESCAPE_PRIVATE_LOCATION_URL", "")
	if got := sshTarget(); got != "private-location.escape.tech:2222" {
		t.Fatalf("sshTarget() = %q, want default target", got)
	}
}
