package kubeslice

import (
	"context"
	"time"

	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/request"

	"github.com/coredns/coredns/plugin"
	dnsCache "github.com/kubeslice/dns/plugin/kubeslice/cache"
	"github.com/miekg/dns"
)

// implements plugin.servicebackend interface
type Kubeslice struct {
	Next           plugin.Handler
	EndpointsCache dnsCache.EndpointsCache
}

func (ks *Kubeslice) Services(ctx context.Context, state request.Request, exact bool, opt plugin.Options) ([]msg.Service, error) {
	startTime := time.Now()
	queryName := state.Name()
	
	log.Info("Services method called", "name", queryName, "exact", exact, "timestamp", startTime)

	var svcs []msg.Service

	// kubeslice only support A records for now, so return empty list if request is not A
	if state.QType() != dns.TypeA {
		log.Debug("received invalid request type, only A is supported now", "type", state.QType())
		return svcs, nil
	}

	// Time cache access
	cacheStart := time.Now()
	eps := ks.EndpointsCache.GetAll()
	cacheDuration := time.Since(cacheStart)
	
	// Get cache statistics
	cacheStats := ks.EndpointsCache.GetStats()
	
	log.Info("Cache access completed", 
		"name", queryName,
		"cache_duration_ms", cacheDuration.Milliseconds(),
		"total_endpoints", len(eps),
		"cache_hits", cacheStats.Hits,
		"cache_misses", cacheStats.Misses,
		"cache_puts", cacheStats.Puts,
		"cache_deletes", cacheStats.Deletes,
		"total_keys", cacheStats.TotalKeys,
		"last_access", cacheStats.LastAccess)

	// Time the matching logic
	matchStart := time.Now()
	name := queryName
	if len(name) > 0 && name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}
	
	log.Debug("Looking for matches", "search_name", name, "total_endpoints", len(eps))

	for _, ep := range eps {
		if ep.Host == name {
			svc := msg.Service{
				Host: ep.IP,
				TTL:  60,
			}
			svcs = append(svcs, svc)
			log.Debug("Found matching endpoint", "host", ep.Host, "ip", ep.IP)
		}
	}
	
	matchDuration := time.Since(matchStart)
	totalDuration := time.Since(startTime)
	
	log.Info("Services lookup completed", 
		"name", queryName,
		"total_duration_ms", totalDuration.Milliseconds(),
		"cache_duration_ms", cacheDuration.Milliseconds(),
		"match_duration_ms", matchDuration.Milliseconds(),
		"matches_found", len(svcs),
		"total_endpoints_searched", len(eps))

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
