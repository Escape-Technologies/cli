package private

import (
	"net"
	"time"
)

type netDialerWithTCPKeepaliveFunc func() *net.Dialer

var netDialerWithTCPKeepaliveImpl netDialerWithTCPKeepaliveFunc = func() *net.Dialer {
	return &net.Dialer{
		KeepAlive: 30 * time.Second,
	}
}

func netDialerWithTCPKeepalive() *net.Dialer {
	return netDialerWithTCPKeepaliveImpl()
}
