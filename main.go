package main

import (
	_ "bitbucket.org/realtimeai/kubeslice-dns/plugin/kubeslice"
	_ "github.com/coredns/coredns/core/plugin" // Plug in CoreDNS.

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/coremain"
)

func init() {
	// add kubeslice to the list of plugins for coredns
	dnsserver.Directives = append([]string{"kubeslice"}, dnsserver.Directives...)
}

func main() {
	coremain.Run()
}
