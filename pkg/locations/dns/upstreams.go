package dns

import (
	"net"

	"github.com/Escape-Technologies/cli/pkg/log"
)

const defaultPort = "53"

// fallbackServers is the last-resort upstream list when no system resolver can
// be discovered: loopback first (matches Go's net default and the local stub
// resolver on systemd hosts), then public resolvers.
var fallbackServers = []string{
	"127.0.0.1:53",
	"[::1]:53",
	"1.1.1.1:53",
	"8.8.8.8:53",
}

// resolveUpstreams returns the ordered list of host:port resolvers to forward
// to. It discovers the system resolvers, normalizes them, and falls back to
// fallbackServers when discovery yields nothing usable.
func resolveUpstreams() []string {
	servers, err := systemResolvers()
	if err != nil {
		log.Debug("[DNS] system resolver discovery failed: %v", err)
	}
	upstreams := normalize(servers)
	if len(upstreams) == 0 {
		log.Debug("[DNS] no system resolvers found, using fallback resolvers")
		return fallbackServers
	}
	return upstreams
}

// normalize dedupes, drops unusable addresses, and ensures every entry carries
// a port. Inputs are bare IPs or host:port strings.
func normalize(servers []string) []string {
	seen := make(map[string]struct{}, len(servers))
	out := make([]string, 0, len(servers))
	for _, s := range servers {
		host := hostOf(s)
		if !isUsable(host) {
			continue
		}
		addr := withPort(s)
		if _, dup := seen[addr]; dup {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	return out
}

// isUsable reports whether an IP is a sensible upstream: parseable, not
// unspecified, and not IPv6 link-local (fe80::/10), which Windows commonly
// lists but cannot be reached without a zone.
func isUsable(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil || ip.IsUnspecified() {
		return false
	}
	if ip.To4() == nil && ip.IsLinkLocalUnicast() {
		return false
	}
	return true
}

func hostOf(server string) string {
	if host, _, err := net.SplitHostPort(server); err == nil {
		return host
	}
	return server
}

func withPort(server string) string {
	if _, _, err := net.SplitHostPort(server); err == nil {
		return server
	}
	return net.JoinHostPort(server, defaultPort)
}
