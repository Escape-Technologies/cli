package stats

import "testing"

func TestSnapshotAndReset(t *testing.T) {
	t.Parallel()

	IncRequest()
	IncRequest()
	IncDNS()

	first := SnapshotAndReset()
	if first.Requests != 2 || first.DNS != 1 {
		t.Fatalf("first snapshot=%+v", first)
	}

	empty := SnapshotAndReset()
	if empty.Requests != 0 || empty.DNS != 0 {
		t.Fatalf("empty snapshot=%+v", empty)
	}

	IncDNS()
	second := SnapshotAndReset()
	if second.Requests != 0 || second.DNS != 1 {
		t.Fatalf("second snapshot=%+v", second)
	}
}
