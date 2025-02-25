package private

import (
	"net"
	"syscall"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/sys/unix"
)

func netDialerWithTCPKeepalive() *net.Dialer {
	return &net.Dialer{
		KeepAlive: time.Duration(-1),
		Control: func(_, _ string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1)
				if err != nil {
					log.Error("failed to set SO_KEEPALIVE: %v", err)
				}
			})
		},
	}
}