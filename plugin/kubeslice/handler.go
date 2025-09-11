package kubeslice

import (
	"context"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"

	// "github.com/coredns/coredns/plugin/etcd/msg"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("kubeslice")

// ServeDNS implements the plugin.Handler interface.
func (ks Kubeslice) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	startTime := time.Now()
	queryName := r.Question[0].Name
	queryType := r.Question[0].Qtype
	
	log.Info("DNS Query received", "name", queryName, "type", queryType, "timestamp", startTime)

	state := request.Request{W: w, Req: r}
	zone := "slice.local"

	// Check if this is a slice.local query
	if !strings.HasSuffix(queryName, zone+".") {
		log.Debug("Query not for slice.local zone, passing to next plugin", "name", queryName)
		return plugin.NextOrFailure(ks.Name(), ks.Next, ctx, w, r)
	}

	// Time the A record lookup
	lookupStart := time.Now()
	records, truncated, err := plugin.A(ctx, &ks, zone, state, nil, plugin.Options{})
	lookupDuration := time.Since(lookupStart)
	
	log.Info("A record lookup completed", 
		"name", queryName, 
		"duration_ms", lookupDuration.Milliseconds(),
		"records_found", len(records),
		"truncated", truncated,
		"error", err)

	if err != nil {
		log.Error(err, "A record lookup failed", "name", queryName, "duration_ms", lookupDuration.Milliseconds())
		return dns.RcodeServerFailure, err
	}

	// Time the response construction
	responseStart := time.Now()
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Truncated = truncated
	m.Answer = records

	w.WriteMsg(m)
	responseDuration := time.Since(responseStart)
	totalDuration := time.Since(startTime)
	
	log.Info("DNS response sent", 
		"name", queryName,
		"total_duration_ms", totalDuration.Milliseconds(),
		"lookup_duration_ms", lookupDuration.Milliseconds(),
		"response_duration_ms", responseDuration.Milliseconds(),
		"records_count", len(records))
	
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (e Kubeslice) Name() string { return "kubeslice" }
