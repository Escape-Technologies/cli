// Package dns provides the DNS integration for private locations
package dns

import (
	"strings"

	"github.com/Escape-Technologies/cli/pkg/log"

	"github.com/miekg/dns"
)

const (
	listenAddr = "127.0.0.1:1053"
)

// Start starts the DNS server
func Start() {
	// Load system resolvers from /etc/resolv.conf
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		log.Error("[DNS] Failed to load resolv.conf: %v", err)
		return
	}

	// TCP-only DNS server
	server := &dns.Server{
		Addr: listenAddr,
		Net:  "tcp",
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			for _, s := range config.Servers {
				resp, err := dns.Exchange(r, s+":"+config.Port)
				if err != nil {
					_ = w.WriteMsg(resp)
					return
				}
				log.Debug("[DNS] Error forwarding: %v", err)
			}
			m := new(dns.Msg)
			m.SetRcode(r, dns.RcodeServerFailure)
			_ = w.WriteMsg(m)
		}),
	}

	log.Debug("[DNS] Starting TCP DNS server on %s, forwarding to %s", listenAddr, strings.Join(config.Servers, ", "))
	if err := server.ListenAndServe(); err != nil {
		log.Error("[DNS] Failed to start TCP DNS server: %v", err)
	}
}
