package private

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
)

const (
	// pollSessionHeader carries the session id on the opening GET response.
	pollSessionHeader = "X-Session-Id"
	// pollGetTimeout bounds one held GET. The distributor answers within ~20s (204 when idle),
	// so a slower answer means a proxy buffers the response and the session cannot work.
	pollGetTimeout = 45 * time.Second
	// pollPostTimeout bounds one POST of buffered SSH bytes.
	pollPostTimeout = 30 * time.Second
	// pollPostMax caps one POST body. Larger writes go out over several ordered POSTs.
	pollPostMax = 256 * 1024
	// pollReadChunk is the read size when moving a GET body into the incoming buffer.
	pollReadChunk = 32 * 1024
	// pollBufferMax caps each direction's buffer. A full outgoing buffer blocks Write; a full
	// incoming buffer delays the next GET until Read drains it.
	pollBufferMax = 1 << 20
)

// errPollSession marks a long-poll session the distributor no longer knows (expired, or
// the request reached another pod because the affinity cookie was lost).
var errPollSession = errors.New("long-poll session lost")

// dialPoll opens a long-poll session on u and returns it as a byte stream.
//
// The opening GET (no sid) creates the session on one distributor pod. Every later request
// must reach that same pod: the load balancer pins it with an affinity cookie, so all
// requests of the session share one cookie jar, which also keeps renewed cookies.
func dialPoll(ctx context.Context, u *url.URL, dial dialFunc, tlsConfig *tls.Config) (net.Conn, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	httpTransport := &http.Transport{
		DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
			return dial(ctx, addr)
		},
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: sshDialTimeout,
		DisableCompression:  true,
		MaxIdleConnsPerHost: 2, //nolint:mnd // One held GET plus one POST.
	}
	client := &http.Client{Transport: httpTransport, Jar: jar}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build long-poll request: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		httpTransport.CloseIdleConnections()

		return nil, fmt.Errorf("failed to open long-poll session: %w", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close() //nolint:errcheck

	sid := resp.Header.Get(pollSessionHeader)
	if resp.StatusCode != http.StatusOK || sid == "" {
		httpTransport.CloseIdleConnections()

		return nil, fmt.Errorf("failed to open long-poll session: %s", resp.Status)
	}

	log.Debug("HTTP long-poll session opened to %s", u)

	sessionURL := *u
	query := sessionURL.Query()
	query.Set("sid", sid)
	sessionURL.RawQuery = query.Encode()

	// The session outlives the dial context, which only bounds the opening request.
	sessionCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	c := &pollConn{
		client:    client,
		url:       sessionURL.String(),
		in:        newPipe(),
		out:       newPipe(),
		cancel:    cancel,
		closeOnce: sync.Once{},
		remote:    pollAddr(u.Host),
	}

	go c.receive(sessionCtx)
	go c.send(sessionCtx)

	return c, nil
}

// pollConn is a long-poll session as a byte stream. A receive loop keeps one GET held and
// feeds Read; a send loop drains Write into POSTs, one at a time so the distributor sees
// bytes in order. Read and Write never wait on each other, which SSH key exchange and
// reverse forwarding require.
type pollConn struct {
	client    *http.Client
	url       string
	in        *pipe
	out       *pipe
	cancel    context.CancelFunc
	closeOnce sync.Once
	remote    net.Addr
}

func (c *pollConn) receive(ctx context.Context) {
	for {
		err := c.in.waitRoom()
		if err != nil {
			return
		}

		err = c.get(ctx)
		if err != nil {
			c.fail(err)

			return
		}
	}
}

func (c *pollConn) get(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pollGetTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to build long-poll GET: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("long-poll GET failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode {
	case http.StatusOK:
		return c.drain(resp.Body)
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("%w: GET %s", errPollSession, resp.Status)
	default:
		return fmt.Errorf("long-poll GET: %s", resp.Status)
	}
}

func (c *pollConn) drain(body io.Reader) error {
	buf := make([]byte, pollReadChunk)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			pushErr := c.in.push(buf[:n])
			if pushErr != nil {
				return pushErr
			}
		}

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("long-poll GET body: %w", err)
		}
	}
}

func (c *pollConn) send(ctx context.Context) {
	for {
		data, err := c.out.take(pollPostMax)
		if err != nil {
			return
		}

		err = c.post(ctx, data)
		if err != nil {
			c.fail(err)

			return
		}
	}
}

func (c *pollConn) post(ctx context.Context, data []byte) error {
	ctx, cancel := context.WithTimeout(ctx, pollPostTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to build long-poll POST: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("long-poll POST failed: %w", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close() //nolint:errcheck

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("%w: POST %s", errPollSession, resp.Status)
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return fmt.Errorf("long-poll POST: %s", resp.Status)
	default:
		return nil
	}
}

// fail ends the session: pending Read returns buffered bytes then err, Write returns err.
func (c *pollConn) fail(err error) {
	c.closeOnce.Do(func() {
		// Close out first: a caller that observed the Read error must also see Write fail.
		c.out.close(err)
		c.in.close(err)
		c.cancel()
		c.client.CloseIdleConnections()
	})
}

func (c *pollConn) Read(p []byte) (int, error)  { return c.in.Read(p) }
func (c *pollConn) Write(p []byte) (int, error) { return c.out.Write(p) }

// Close drops the session locally. The distributor expires it once requests stop.
func (c *pollConn) Close() error {
	c.fail(net.ErrClosed)

	return nil
}

func (c *pollConn) LocalAddr() net.Addr  { return pollAddr("local") }
func (c *pollConn) RemoteAddr() net.Addr { return c.remote }

func (c *pollConn) SetDeadline(t time.Time) error {
	c.in.setDeadline(t)
	c.out.setDeadline(t)

	return nil
}

func (c *pollConn) SetReadDeadline(t time.Time) error {
	c.in.setDeadline(t)

	return nil
}

func (c *pollConn) SetWriteDeadline(t time.Time) error {
	c.out.setDeadline(t)

	return nil
}

type pollAddr string

func (pollAddr) Network() string  { return "poll" }
func (a pollAddr) String() string { return string(a) }

// pipe is one direction of a pollConn: a bounded byte buffer between a net.Conn caller,
// subject to a deadline, and a loop goroutine, which only stops when the pipe closes.
type pipe struct {
	mu       sync.Mutex
	cond     *sync.Cond
	buf      []byte
	err      error
	deadline time.Time
}

func newPipe() *pipe {
	p := &pipe{mu: sync.Mutex{}, cond: nil, buf: nil, err: nil, deadline: time.Time{}}
	p.cond = sync.NewCond(&p.mu)

	return p
}

// Read is the caller side of the incoming pipe. It returns buffered bytes before the close error.
func (p *pipe) Read(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.buf) == 0 {
		if p.err != nil {
			return 0, p.err
		}

		err := p.waitLocked(true)
		if err != nil {
			return 0, err
		}
	}

	n := copy(b, p.buf)
	p.buf = p.buf[n:]
	p.cond.Broadcast()

	return n, nil
}

// Write is the caller side of the outgoing pipe. It blocks while the buffer is full.
func (p *pipe) Write(b []byte) (int, error) {
	return p.write(b, true)
}

// push is the loop side of the incoming pipe. The caller's read deadline does not apply.
func (p *pipe) push(b []byte) error {
	_, err := p.write(b, false)

	return err
}

func (p *pipe) write(b []byte, useDeadline bool) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	written := 0
	for len(b) > 0 {
		if p.err != nil {
			return written, p.err
		}

		room := pollBufferMax - len(p.buf)
		if room <= 0 {
			err := p.waitLocked(useDeadline)
			if err != nil {
				return written, err
			}

			continue
		}

		n := min(room, len(b))
		p.buf = append(p.buf, b[:n]...)
		b = b[n:]
		written += n
		p.cond.Broadcast()
	}

	return written, nil
}

// take is the loop side of the outgoing pipe: it waits for bytes and removes up to limit.
func (p *pipe) take(limit int) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.buf) == 0 {
		if p.err != nil {
			return nil, p.err
		}

		_ = p.waitLocked(false)
	}

	n := min(limit, len(p.buf))
	data := append([]byte(nil), p.buf[:n]...)
	p.buf = p.buf[n:]
	p.cond.Broadcast()

	return data, nil
}

// waitRoom is the loop side of the incoming pipe: it waits until Read made room.
func (p *pipe) waitRoom() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.buf) >= pollBufferMax {
		if p.err != nil {
			return p.err
		}

		_ = p.waitLocked(false)
	}

	return p.err
}

func (p *pipe) close(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err == nil {
		p.err = err
	}

	p.cond.Broadcast()
}

func (p *pipe) setDeadline(t time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.deadline = t
	p.cond.Broadcast()
}

// waitLocked waits for any state change. With useDeadline it fails once the deadline passed.
func (p *pipe) waitLocked(useDeadline bool) error {
	if !useDeadline || p.deadline.IsZero() {
		p.cond.Wait()

		return nil
	}

	remaining := time.Until(p.deadline)
	if remaining <= 0 {
		return os.ErrDeadlineExceeded
	}

	// The timer takes mu, so its broadcast cannot fire before Wait released it.
	timer := time.AfterFunc(remaining, func() {
		p.mu.Lock()
		p.cond.Broadcast()
		p.mu.Unlock()
	})
	p.cond.Wait()
	timer.Stop()

	return nil
}

var _ net.Conn = (*pollConn)(nil)
