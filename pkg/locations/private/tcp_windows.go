//go:build windows
// +build windows

package private

import (
	"net"
)

func init() {
	netDialerWithTCPKeepaliveImpl = func() *net.Dialer {
		return &net.Dialer{
			KeepAlive: DefaultKeepAliveDuration,
		}
	}
} 