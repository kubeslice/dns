package kubeslice

import (
	"context"
	"errors"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"

	// "github.com/coredns/coredns/plugin/etcd/msg"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("kubeslice")

// ServeDNS implements the plugin.Handler interface.
func (ks Kubeslice) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	zone := "slice.local"

	// kubeslice only support A records for now, so return empty list if request is not A
        if state.QType() != dns.TypeA {
                log.Debug("received invalid request type, only A is supported now", r.Question)
                return dns.RcodeNotImplemented, errors.New("Request type not supported")
        }

	records, truncated, err := plugin.A(ctx, &ks, zone, state, nil, plugin.Options{})
	if err != nil {
		log.Debug("Error", err)
		return dns.RcodeServerFailure, err
	}
	if len(records) == 0 {
		log.Debug("Error", "No records found")
		return dns.RcodeNameError, errors.New("No records found")
	}

	log.Debug("Records", records)

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
