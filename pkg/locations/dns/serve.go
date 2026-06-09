package dns

import "github.com/miekg/dns"

// serve runs a TCP DNS server bound to addr until it stops. It blocks and
// returns the error that caused ListenAndServe to exit.
func serve(addr string, handler dns.Handler) error {
	server := &dns.Server{
		Addr:    addr,
		Net:     "tcp",
		Handler: handler,
	}
	return server.ListenAndServe()
}
