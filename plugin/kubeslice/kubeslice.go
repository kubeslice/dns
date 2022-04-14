package kubeslice

import (
	"context"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("kubeslice")

type Kubeslice struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (ks Kubeslice) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	log.Debug("Received response")

	// Wrap.
	pw := NewResponsePrinter(w)

	// Call next plugin (if any).
	return plugin.NextOrFailure(ks.Name(), ks.Next, ctx, pw, r)
}

// Name implements the Handler interface.
func (e Kubeslice) Name() string { return "kubeslice" }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("got a request for kubeslice")
	return r.ResponseWriter.WriteMsg(res)
}
