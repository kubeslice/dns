package kubeslice

import (
	"context"
	"net"

	"github.com/coredns/coredns/plugin"
	// "github.com/coredns/coredns/plugin/etcd/msg"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("kubeslice")

type Kubeslice struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (ks Kubeslice) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	log.Debug("Received request", r)

	if r.Question[0].Qtype != dns.TypeA {
		return plugin.NextOrFailure(ks.Name(), ks.Next, ctx, w, r)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	hdr := dns.RR_Header{Name: r.Question[0].Name, Ttl: 8482, Class: dns.ClassINET, Rrtype: dns.TypeA}
	ip := net.ParseIP("10.20.10.2")
	m.Answer = []dns.RR{&dns.A{Hdr: hdr, A: ip}}

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil

}

// Name implements the Handler interface.
func (e Kubeslice) Name() string { return "kubeslice" }
