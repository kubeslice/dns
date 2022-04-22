package kubeslice

import (
	"context"

	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/request"

	dnsCache "bitbucket.org/realtimeai/kubeslice-dns/plugin/kubeslice/cache"
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// implements plugin.servicebackend interface
type Kubeslice struct {
	Next           plugin.Handler
	EndpointsCache dnsCache.EndpointsCache
}

func (ks *Kubeslice) Services(ctx context.Context, state request.Request, exact bool, opt plugin.Options) ([]msg.Service, error) {

	var svcs []msg.Service

	// kubeslice only support A records for now, so return empty list if request is not A
	if state.QType() != dns.TypeA {
		log.Debug("received invalid request type, only A is supported now")
		return svcs, nil
	}

	log.Info("fetching kubeslice services")

	name := state.Name()
	name = name[:len(name)-1]

	eps := ks.EndpointsCache.GetAll()

	for _, ep := range eps {
		if ep.Host == name {
			svc := msg.Service{
				Host: ep.IP,
			}

			svcs = append(svcs, svc)
		}
	}

	return svcs, nil

}

// TODO fill later
func (ks *Kubeslice) Reverse(ctx context.Context, state request.Request, exact bool, opt plugin.Options) ([]msg.Service, error) {
	var svcs []msg.Service
	log.Debug("kubeslice reverse lookup")
	return svcs, nil
}

// TODO fill later
func (ks *Kubeslice) Lookup(ctx context.Context, state request.Request, name string, typ uint16) (*dns.Msg, error) {
	log.Debug("kubeslice lookup")
	msg := &dns.Msg{}
	return msg, nil
}

// TODO fill later
func (ks *Kubeslice) Records(ctx context.Context, state request.Request, exact bool) ([]msg.Service, error) {
	var svcs []msg.Service
	log.Debug("kubeslice records")
	return svcs, nil
}

// TODO fill later
func (ks *Kubeslice) MinTTL(state request.Request) uint32 {
	log.Debug("kubeslice ttl")
	return 60
}

// TODO fill later
func (ks *Kubeslice) Serial(state request.Request) uint32 {
	log.Debug("kubeslice soa")
	return 1
}

// TODO fill later
func (ks *Kubeslice) IsNameError(err error) bool {
	log.Debug("kubeslice isnameerror")
	return false
}
