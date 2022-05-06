package main

import (
	_ "github.com/coredns/coredns/core/plugin" // Plug in CoreDNS.
	_ "github.com/kubeslice/dns/plugin/kubeslice"

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
