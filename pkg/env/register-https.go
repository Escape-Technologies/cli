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

func (s *httpsProxy) Dial(network, addr string) (net.Conn, error) {
	c, err := tls.Dial("tcp", s.host, &tls.Config{})
	if err != nil {
		return nil, err
	}

	reqURL, err := url.Parse("https://" + addr)
	if err != nil {
		c.Close()
		return nil, err
	}
	reqURL.Scheme = ""

	req, err := http.NewRequest("CONNECT", reqURL.String(), nil)
	if err != nil {
		c.Close()
		return nil, err
	}
	req.Close = false
	if s.haveAuth {
		req.SetBasicAuth(s.username, s.password)
	}

	err = req.Write(c)
	if err != nil {
		c.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(c), req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		c.Close()
		return nil, err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		c.Close()
		err = fmt.Errorf("connect server using proxy error, StatusCode [%d]", resp.StatusCode)
		return nil, err
	}

	return c, nil
}

func init() {
	proxy.RegisterDialerType("https", newHTTPSProxy)
} 