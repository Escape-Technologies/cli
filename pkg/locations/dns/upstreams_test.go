package dns

import (
	"reflect"
	"testing"
)

func TestIsUsable(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"1.1.1.1", true},
		{"8.8.8.8", true},
		{"::1", true},
		{"169.254.1.1", true}, // IPv4 link-local is not filtered
		{"", false},
		{"0.0.0.0", false},
		{"::", false},
		{"fe80::1", false},      // IPv6 link-local
		{"fe80::1%eth0", false}, // zone makes ParseIP fail
		{"not-an-ip", false},
	}
	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			if got := isUsable(c.host); got != c.want {
				t.Fatalf("isUsable(%q) = %v, want %v", c.host, got, c.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"empty", nil, []string{}},
		{"bare ip gets port", []string{"1.1.1.1"}, []string{"1.1.1.1:53"}},
		{"dedup identical", []string{"1.1.1.1", "1.1.1.1"}, []string{"1.1.1.1:53"}},
		{"dedup across port forms", []string{"8.8.8.8:53", "8.8.8.8"}, []string{"8.8.8.8:53"}},
		{"drop link-local", []string{"fe80::1", "1.1.1.1"}, []string{"1.1.1.1:53"}},
		{"drop empty", []string{"", "1.1.1.1"}, []string{"1.1.1.1:53"}},
		{"ipv6 bare", []string{"::1"}, []string{"[::1]:53"}},
		{"ipv6 bracketed kept", []string{"[::1]:53"}, []string{"[::1]:53"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := normalize(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("normalize(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

// resolveUpstreams must always yield at least one upstream thanks to the
// fallback chain, regardless of the host's resolver configuration.
func TestResolveUpstreamsNeverEmpty(t *testing.T) {
	if got := resolveUpstreams(); len(got) == 0 {
		t.Fatal("resolveUpstreams returned no upstreams")
	}
}

func TestFallbackServersAreUsable(t *testing.T) {
	if got := normalize(fallbackServers); len(got) != len(fallbackServers) {
		t.Fatalf("fallbackServers should all be usable, normalize gave %v", got)
	}
}
