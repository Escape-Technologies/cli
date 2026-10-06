package private

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// fakeDistributor echoes bytes over the distributor's /v2/connect protocol: WebSocket, or a
// long-poll session pinned by an affinity cookie, as the load balancer does in front of it.
type fakeDistributor struct {
	refuseUpgrade bool

	mu       sync.Mutex
	sessions map[string]*fakeSession
	opened   atomic.Int32
}

type fakeSession struct {
	mu   sync.Mutex
	buf  []byte
	data chan struct{}
}

const fakeAffinityCookie = "AWSALB"

func (d *fakeDistributor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != connectPath {
		http.NotFound(w, r)

		return
	}

	if websocket.IsWebSocketUpgrade(r) {
		d.echoWebSocket(w, r)

		return
	}

	sid := r.URL.Query().Get("sid")
	if sid == "" {
		d.open(w)

		return
	}

	sess := d.session(r, sid)
	if sess == nil {
		http.Error(w, "unknown sid", http.StatusNotFound)

		return
	}

	switch r.Method {
	case http.MethodPost:
		body, _ := io.ReadAll(r.Body)
		sess.mu.Lock()
		sess.buf = append(sess.buf, body...)
		sess.mu.Unlock()
		select {
		case sess.data <- struct{}{}:
		default:
		}

		w.WriteHeader(http.StatusNoContent)
	default:
		select {
		case <-sess.data:
		case <-time.After(100 * time.Millisecond):
		}

		sess.mu.Lock()
		out := sess.buf
		sess.buf = nil
		sess.mu.Unlock()
		if len(out) == 0 {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		_, _ = w.Write(out)
	}
}

func (d *fakeDistributor) echoWebSocket(w http.ResponseWriter, r *http.Request) {
	if d.refuseUpgrade {
		http.Error(w, "upgrade refused", http.StatusBadRequest)

		return
	}

	upgrader := websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close() //nolint:errcheck

	for {
		msgType, data, err := ws.ReadMessage()
		if err != nil {
			return
		}

		if err := ws.WriteMessage(msgType, data); err != nil {
			return
		}
	}
}

func (d *fakeDistributor) open(w http.ResponseWriter) {
	sid := strconv.Itoa(int(d.opened.Add(1)))
	d.mu.Lock()
	if d.sessions == nil {
		d.sessions = map[string]*fakeSession{}
	}

	d.sessions[sid] = &fakeSession{mu: sync.Mutex{}, buf: nil, data: make(chan struct{}, 1)}
	d.mu.Unlock()

	http.SetCookie(w, &http.Cookie{Name: fakeAffinityCookie, Value: "pod-" + sid})
	w.Header().Set(pollSessionHeader, sid)
	w.WriteHeader(http.StatusOK)
}

// session returns nil without the affinity cookie: the request landed on another pod.
func (d *fakeDistributor) session(r *http.Request, sid string) *fakeSession {
	cookie, err := r.Cookie(fakeAffinityCookie)
	if err != nil || cookie.Value != "pod-"+sid {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	return d.sessions[sid]
}

func (d *fakeDistributor) drop(sid string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.sessions, sid)
}

// connectProxy is a frontend HTTP proxy that only speaks CONNECT, like a corporate proxy.
type connectProxy struct {
	targets []string
	mu      sync.Mutex
}

func (p *connectProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodConnect {
		http.Error(w, "CONNECT only", http.StatusMethodNotAllowed)

		return
	}

	p.mu.Lock()
	p.targets = append(p.targets, r.Host)
	p.mu.Unlock()

	upstream, err := (&net.Dialer{}).DialContext(r.Context(), "tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)

		return
	}

	client, _, err := http.NewResponseController(w).Hijack()
	if err != nil {
		upstream.Close() //nolint:errcheck

		return
	}

	_, _ = io.WriteString(client, "HTTP/1.1 200 Connection established\r\n\r\n")

	go func() {
		_, _ = io.Copy(upstream, client)
		upstream.Close() //nolint:errcheck
	}()

	_, _ = io.Copy(client, upstream)
	client.Close() //nolint:errcheck
}

// startDistributor serves d over TLS and points the CLI at it with a trusted CA bundle.
func startDistributor(t *testing.T, d *fakeDistributor, transport string) *httptest.Server {
	t.Helper()

	srv := httptest.NewTLSServer(d)
	t.Cleanup(srv.Close)

	certPath := filepath.Join(t.TempDir(), "ca.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ESCAPE_SSL_CERT_PATH", certPath)
	t.Setenv("ESCAPE_SSL_INSECURE", "")
	t.Setenv("ESCAPE_TRANSPORT", transport)
	t.Setenv("ESCAPE_PRIVATE_LOCATION_URL", srv.URL)

	return srv
}

func dialEndpoint(t *testing.T, proxyURL *url.URL) net.Conn {
	t.Helper()

	ep, err := resolveEndpoint()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := getConn(ctx, ep, proxyURL)
	if err != nil {
		t.Fatalf("getConn(%s): %v", ep.transport, err)
	}

	t.Cleanup(func() { conn.Close() }) //nolint:errcheck

	return conn
}

// assertEcho writes a payload far larger than one POST or buffer while reading the echo
// concurrently: a conn that serialises Read and Write deadlocks here.
func assertEcho(t *testing.T, conn net.Conn) {
	t.Helper()

	payload := make([]byte, 3*pollBufferMax+123)
	_, _ = rand.Read(payload)

	writeErr := make(chan error, 1)
	go func() {
		for chunk := payload; len(chunk) > 0; {
			n := min(len(chunk), 40*1024)
			if _, err := conn.Write(chunk[:n]); err != nil {
				writeErr <- err

				return
			}

			chunk = chunk[n:]
		}

		writeErr <- nil
	}()

	_ = conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	got := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("read echo: %v", err)
	}

	if err := <-writeErr; err != nil {
		t.Fatalf("write: %v", err)
	}

	if !bytes.Equal(got, payload) {
		t.Fatal("echo differs from payload")
	}
}

func TestGetConnHTTPSTransports(t *testing.T) {
	tests := []struct {
		name          string
		transport     string
		refuseUpgrade bool
		viaProxy      bool
		wantType      net.Conn
	}{
		{name: "wss", transport: "wss", wantType: &wsConn{}},
		{name: "https long-poll", transport: "https", wantType: &pollConn{}},
		{name: "auto with Upgrade allowed", transport: "auto", wantType: &wsConn{}},
		{name: "auto falls back when Upgrade is refused", transport: "auto", refuseUpgrade: true, wantType: &pollConn{}},
		{name: "wss through CONNECT proxy", transport: "wss", viaProxy: true, wantType: &wsConn{}},
		{name: "https through CONNECT proxy", transport: "https", viaProxy: true, wantType: &pollConn{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := startDistributor(t, &fakeDistributor{refuseUpgrade: tt.refuseUpgrade}, tt.transport)

			var proxyURL *url.URL
			proxy := &connectProxy{}
			if tt.viaProxy {
				proxySrv := httptest.NewServer(proxy)
				t.Cleanup(proxySrv.Close)
				proxyURL, _ = url.Parse(proxySrv.URL)
			}

			conn := dialEndpoint(t, proxyURL)
			if gotType, wantType := typeName(conn), typeName(tt.wantType); gotType != wantType {
				t.Fatalf("getConn returned %s, want %s", gotType, wantType)
			}

			assertEcho(t, conn)

			if tt.viaProxy {
				proxy.mu.Lock()
				defer proxy.mu.Unlock()
				if len(proxy.targets) == 0 || proxy.targets[0] != srv.Listener.Addr().String() {
					t.Fatalf("proxy CONNECT targets = %v, want %s", proxy.targets, srv.Listener.Addr())
				}
			}
		})
	}
}

func TestGetConnWSSUpgradeRefused(t *testing.T) {
	startDistributor(t, &fakeDistributor{refuseUpgrade: true}, "wss")

	ep, err := resolveEndpoint()
	if err != nil {
		t.Fatal(err)
	}

	_, err = getConn(context.Background(), ep, nil)
	if !errors.Is(err, errUpgradeRefused) {
		t.Fatalf("getConn error = %v, want errUpgradeRefused", err)
	}
}

func TestPollConnDeadline(t *testing.T) {
	startDistributor(t, &fakeDistributor{}, "https")
	conn := dialEndpoint(t, nil)

	_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	_, err := conn.Read(make([]byte, 1))
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("Read error = %v, want a timeout", err)
	}

	_ = conn.SetReadDeadline(time.Time{})
	assertEcho(t, conn)
}

func TestPollConnSessionLost(t *testing.T) {
	d := &fakeDistributor{}
	startDistributor(t, d, "https")
	conn := dialEndpoint(t, nil)

	d.drop("1")

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Read(make([]byte, 1))
	if !errors.Is(err, errPollSession) {
		t.Fatalf("Read error = %v, want errPollSession", err)
	}

	if _, err := conn.Write([]byte("x")); !errors.Is(err, errPollSession) {
		t.Fatalf("Write error = %v, want errPollSession", err)
	}
}

func typeName(c net.Conn) string {
	switch c.(type) {
	case *wsConn:
		return "wsConn"
	case *pollConn:
		return "pollConn"
	default:
		return "other"
	}
}
