package kubeslice

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"

	// "github.com/coredns/coredns/plugin/etcd/msg"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("kubeslice")

// ServeDNS implements the plugin.Handler interface.
func (ks Kubeslice) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	log.Debug("Received request", r)

	if r.Question[0].Qtype != dns.TypeA {
		return plugin.NextOrFailure(ks.Name(), ks.Next, ctx, w, r)
	}

	state := request.Request{W: w, Req: r}
	zone := "slice.local"

	records, truncated, err := plugin.A(ctx, &ks, zone, state, nil, plugin.Options{})

	if err != nil {
		return dns.RcodeServerFailure, err
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Truncated = truncated
	m.Answer = records

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil

}

// Name implements the Handler interface.
func (e Kubeslice) Name() string { return "kubeslice" }
