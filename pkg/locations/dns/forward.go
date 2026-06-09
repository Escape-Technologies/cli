package dns

import (
	"github.com/Escape-Technologies/cli/pkg/log"

	"github.com/miekg/dns"
)

// newHandler builds a forwarding handler that relays each query to the given
// upstreams in order, returning the first successful response. A panic in the
// forwarding path is recovered and turned into SERVFAIL so a single malformed
// query cannot take the server down.
func newHandler(upstreams []string) dns.HandlerFunc {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("[DNS] recovered from panic while forwarding: %v", rec)
				_ = w.WriteMsg(servfail(r))
			}
		}()

		for _, upstream := range upstreams {
			resp, err := dns.Exchange(r, upstream)
			if err == nil && resp != nil {
				_ = w.WriteMsg(resp)
				return
			}
			log.Debug("[DNS] error forwarding to %s: %v", upstream, err)
		}
		_ = w.WriteMsg(servfail(r))
	}
}

func servfail(r *dns.Msg) *dns.Msg {
	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeServerFailure)
	return m
}
