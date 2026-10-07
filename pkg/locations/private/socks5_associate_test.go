package private

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestSocks5AssociateRelaysFramedUDP(t *testing.T) {
	t.Setenv("ESCAPE_BACKEND_PROXY_URL", "socks5://127.0.0.1:1") // UDP must bypass the TCP-only backend proxy

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var lc net.ListenConfig
	echo, err := lc.ListenPacket(ctx, "udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer echo.Close() //nolint:errcheck
	go func() {
		buf := make([]byte, 1500)
		for {
			n, addr, err := echo.ReadFrom(buf)
			if err != nil {
				return
			}

			echo.WriteTo(buf[:n], addr) //nolint:errcheck
		}
	}()

	ln, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	go startSocks5Server(ctx, ln, &atomic.Bool{}) //nolint:errcheck

	var d net.Dialer
	c, err := d.DialContext(ctx, "tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()                                //nolint:errcheck
	c.SetDeadline(time.Now().Add(5 * time.Second)) //nolint:errcheck

	c.Write([]byte{5, 1, 0})                      //nolint:errcheck
	io.ReadFull(c, make([]byte, 2))               //nolint:errcheck
	c.Write([]byte{5, 3, 0, 1, 0, 0, 0, 0, 0, 0}) //nolint:errcheck
	reply := make([]byte, 10)
	if _, err := io.ReadFull(c, reply); err != nil || reply[1] != 0 {
		t.Fatalf("associate rejected: %v %v", reply, err)
	}

	dst := echo.LocalAddr().(*net.UDPAddr)
	payload := []byte("ping")
	frame := []byte{0, byte(len(payload)), 0, 1}
	frame = append(frame, dst.IP.To4()...)
	frame = append(frame, byte(dst.Port>>8), byte(dst.Port))
	frame = append(frame, payload...)
	c.Write(frame) //nolint:errcheck

	resp := make([]byte, len(frame))
	if _, err := io.ReadFull(c, resp); err != nil {
		t.Fatal(err)
	}

	if !bytes.HasSuffix(resp, payload) {
		t.Fatalf("unexpected frame: %v", resp)
	}
}
