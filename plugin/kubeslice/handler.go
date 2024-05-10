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
	log.Debug("Question type", r.Question)

	state := request.Request{W: w, Req: r}
	zone := "slice.local"

	var (
		records, extra []dns.RR
		truncated      bool
		err            error
	)

	switch state.QType() {
	case dns.TypeA:
		records, truncated, err = plugin.A(ctx, &ks, zone, state, nil, plugin.Options{})
		if err != nil {
			return dns.RcodeServerFailure, err
		}
	case dns.TypeSRV:
		records, extra, err = plugin.SRV(ctx, &ks, zone, state, plugin.Options{})
		if err != nil {
			return dns.RcodeServerFailure, err
		}
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Truncated = truncated
	m.Answer = records
	m.Extra = extra

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (e Kubeslice) Name() string { return "kubeslice" }
