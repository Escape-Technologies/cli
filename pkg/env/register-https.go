package env

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

type httpsProxy struct {
	host     string
	haveAuth bool
	username string
	password string
	forward  proxy.Dialer
}

func newHTTPSProxy(uri *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	s := new(httpsProxy)
	s.host = uri.Host
	s.forward = forward
	if uri.User != nil {
		s.haveAuth = true
		s.username = uri.User.Username()
		s.password, _ = uri.User.Password()
	}
	return s, nil
}

func (s *httpsProxy) Dial(_, addr string) (net.Conn, error) {
	c, err := tls.Dial("tcp", s.host, &tls.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", s.host, err)
	}

	reqURL, err := url.Parse("https://" + addr)
	if err != nil {
		c.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	reqURL.Scheme = ""

	req, err := http.NewRequest("CONNECT", reqURL.String(), nil) //nolint:noctx
	if err != nil {
		c.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Close = false
	if s.haveAuth {
		req.SetBasicAuth(s.username, s.password)
	}

	err = req.Write(c)
	if err != nil {
		c.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(c), req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close() //nolint:errcheck
		}
		c.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		c.Close() //nolint:errcheck
		return nil, fmt.Errorf("connect server using proxy error, StatusCode [%d]", resp.StatusCode)
	}

	return c, nil
}

func init() {
	proxy.RegisterDialerType("https", newHTTPSProxy)
}
