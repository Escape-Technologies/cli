package private

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// transport is the outer socket that carries the SSH session to the Escape Platform.
// Everything above it (SSH client, reverse listener, SOCKS5, escape channel) is identical.
type transport string

const (
	// transportSSH is raw TCP to the SSH endpoint.
	transportSSH transport = "ssh"
	// transportWSS is a WebSocket on the HTTPS endpoint.
	transportWSS transport = "wss"
	// transportHTTPS is HTTP long-poll on the HTTPS endpoint.
	transportHTTPS transport = "https"
	// transportAuto is a WebSocket, then long-poll on the same URL when the Upgrade fails.
	transportAuto transport = "auto"
)

const (
	defaultSSHTarget   = "private-location.escape.tech:2222"
	defaultConnectHost = "location.escape.tech"
	connectPath        = "/v2/connect"
)

// endpoint is where the tunnel dials, resolved from ESCAPE_TRANSPORT and ESCAPE_PRIVATE_LOCATION_URL.
type endpoint struct {
	transport transport
	// addr is the host:port TCP destination, and the CONNECT target through a frontend proxy.
	addr string
	// connectURL is the /v2/connect URL shared by WebSocket and long-poll. Nil for ssh.
	connectURL *url.URL
}

// resolveEndpoint reads the transport (default ssh) and its destination.
//
// For ssh, ESCAPE_PRIVATE_LOCATION_URL is the host:port, as before.
// For the HTTPS transports, a bare host or host:port override only replaces the host: the
// HTTPS endpoint is always port 443 on /v2/connect. An override with a scheme (https, wss,
// http, ws) is used as is, which reaches a plaintext distributor in local development.
func resolveEndpoint() (endpoint, error) {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("ESCAPE_TRANSPORT")))
	override := strings.TrimSpace(os.Getenv("ESCAPE_PRIVATE_LOCATION_URL"))

	switch t := transport(raw); t {
	case "", transportSSH:
		addr := override
		if addr == "" {
			addr = defaultSSHTarget
		}

		return endpoint{transport: transportSSH, addr: addr, connectURL: nil}, nil
	case transportWSS, transportHTTPS, transportAuto:
		u, err := connectURL(override)
		if err != nil {
			return endpoint{}, err
		}

		port := u.Port()
		if port == "" {
			port = "443"
			if u.Scheme == "http" {
				port = "80"
			}
		}

		return endpoint{transport: t, addr: net.JoinHostPort(u.Hostname(), port), connectURL: u}, nil
	default:
		return endpoint{}, fmt.Errorf("invalid ESCAPE_TRANSPORT %q: use ssh, wss, https or auto", raw)
	}
}

func connectURL(override string) (*url.URL, error) {
	u := &url.URL{Scheme: "https", Host: defaultConnectHost, Path: connectPath}
	if override == "" {
		return u, nil
	}

	if !strings.Contains(override, "://") {
		host, _, err := net.SplitHostPort(override)
		if err != nil {
			host = override
		}

		u.Host = host

		return u, nil
	}

	parsed, err := url.Parse(override)
	if err != nil {
		return nil, fmt.Errorf("invalid ESCAPE_PRIVATE_LOCATION_URL: %w", err)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "https", "wss":
		u.Scheme = "https"
	case "http", "ws":
		u.Scheme = "http"
	default:
		return nil, fmt.Errorf("invalid ESCAPE_PRIVATE_LOCATION_URL scheme %q: use https or http", parsed.Scheme)
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("invalid ESCAPE_PRIVATE_LOCATION_URL %q: missing host", override)
	}

	u.Host = parsed.Host
	if parsed.Path != "" && parsed.Path != "/" {
		u.Path = parsed.Path
	}

	return u, nil
}
