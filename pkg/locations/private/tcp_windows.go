//go:build windows
// +build windows

package private

import (
	"net"
	"time"
)

func init() {
	// Windows-specific implementation
	// For Windows, we'll use the default net.Dialer with a reasonable keepalive value
	netDialerWithTCPKeepaliveImpl = func() *net.Dialer {
		return &net.Dialer{
			KeepAlive: 30 * time.Second,
		}
	}
} 