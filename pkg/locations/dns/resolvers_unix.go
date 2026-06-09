//go:build !windows

package dns

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

// systemResolvers reads the OS resolver configuration from /etc/resolv.conf.
// Covers Linux and macOS, where the file exists and tracks the active resolver.
func systemResolvers() ([]string, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/resolv.conf: %w", err)
	}
	port := config.Port
	if port == "" {
		port = defaultPort
	}
	servers := make([]string, 0, len(config.Servers))
	for _, s := range config.Servers {
		servers = append(servers, net.JoinHostPort(s, port))
	}
	return servers, nil
}
