package private

import (
	"net"
	"time"
)

// DefaultKeepAliveDuration TCP keepalive timeout.
const DefaultKeepAliveDuration = 30 * time.Second

type netDialerWithTCPKeepaliveFunc func() *net.Dialer

var netDialerWithTCPKeepaliveImpl netDialerWithTCPKeepaliveFunc = func() *net.Dialer {
	return &net.Dialer{
		KeepAlive: DefaultKeepAliveDuration,
	}
}

func netDialerWithTCPKeepalive() *net.Dialer {
	return netDialerWithTCPKeepaliveImpl()
}
