package dns

import (
	"net"
	"testing"

	"github.com/miekg/dns"
)

// capture is a minimal dns.ResponseWriter that records the written message.
type capture struct {
	msg *dns.Msg
}

func (c *capture) LocalAddr() net.Addr       { return &net.TCPAddr{} } // nolint:exhaustruct
func (c *capture) RemoteAddr() net.Addr      { return &net.TCPAddr{} } // nolint:exhaustruct
func (c *capture) WriteMsg(m *dns.Msg) error { c.msg = m; return nil }
func (c *capture) Write(b []byte) (int, error) {
	return len(b), nil
}
func (c *capture) Close() error        { return nil }
func (c *capture) TsigStatus() error   { return nil }
func (c *capture) TsigTimersOnly(bool) {}
func (c *capture) Hijack()             {}

// startUpstream runs a UDP DNS server answering every A query with answer and
// returns its address. dns.Exchange forwards over UDP, so UDP is enough.
func startUpstream(t *testing.T, answer string) string {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)}) // nolint:exhaustruct
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	handler := dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		for _, q := range r.Question {
			if q.Qtype != dns.TypeA {
				continue
			}
			rr, err := dns.NewRR(q.Name + " 60 IN A " + answer)
			if err != nil {
				t.Errorf("build rr: %v", err)
				continue
			}
			m.Answer = append(m.Answer, rr)
		}
		_ = w.WriteMsg(m)
	})
	server := &dns.Server{PacketConn: conn, Net: "udp", Handler: handler} // nolint:exhaustruct
	go func() { _ = server.ActivateAndServe() }()
	t.Cleanup(func() { _ = server.Shutdown() })
	return conn.LocalAddr().String()
}

// deadAddr returns a UDP address with nothing listening on it.
func deadAddr(t *testing.T) string {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)}) // nolint:exhaustruct
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	addr := conn.LocalAddr().String()
	_ = conn.Close()
	return addr
}

func query(name string) *dns.Msg {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), dns.TypeA)
	return m
}

func TestForwardSuccess(t *testing.T) {
	upstream := startUpstream(t, "1.2.3.4")
	w := &capture{} // nolint:exhaustruct
	newHandler([]string{upstream})(w, query("example.com"))

	if w.msg == nil {
		t.Fatal("no response written")
	}
	if w.msg.Rcode != dns.RcodeSuccess {
		t.Fatalf("rcode = %v, want success", w.msg.Rcode)
	}
	if len(w.msg.Answer) != 1 {
		t.Fatalf("answers = %d, want 1", len(w.msg.Answer))
	}
	a, ok := w.msg.Answer[0].(*dns.A)
	if !ok || a.A.String() != "1.2.3.4" {
		t.Fatalf("answer = %v, want 1.2.3.4", w.msg.Answer[0])
	}
}

func TestForwardFailover(t *testing.T) {
	upstream := startUpstream(t, "5.6.7.8")
	w := &capture{} // nolint:exhaustruct
	newHandler([]string{deadAddr(t), upstream})(w, query("example.com"))

	if w.msg == nil || w.msg.Rcode != dns.RcodeSuccess {
		t.Fatalf("expected success via second upstream, got %v", w.msg)
	}
	a, ok := w.msg.Answer[0].(*dns.A)
	if !ok || a.A.String() != "5.6.7.8" {
		t.Fatalf("answer = %v, want 5.6.7.8", w.msg.Answer)
	}
}

func TestForwardAllDead(t *testing.T) {
	w := &capture{} // nolint:exhaustruct
	newHandler([]string{deadAddr(t), deadAddr(t)})(w, query("example.com"))

	if w.msg == nil {
		t.Fatal("no response written")
	}
	if w.msg.Rcode != dns.RcodeServerFailure {
		t.Fatalf("rcode = %v, want SERVFAIL", w.msg.Rcode)
	}
}
