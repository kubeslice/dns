package kubeslice_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net"

	"github.com/kubeslice/dns/plugin/kubeslice"
	dnsCache "github.com/kubeslice/dns/plugin/kubeslice/cache"
	"github.com/kubeslice/dns/plugin/kubeslice/slice"
	"github.com/miekg/dns"
)

var _ = Describe("Handler", func() {

	Context("ServeDNS", func() {

		It("should serve A records", func() {
			cache := dnsCache.NewEndpointsCache()
			ks := kubeslice.Kubeslice{
				EndpointsCache: cache,
			}

			r := &dns.Msg{
				Question: []dns.Question{{
					Name:  "nginx.default.slice.local",
					Qtype: dns.TypeA,
				}},
			}
			w := &mockResponse{}

			code, err := ks.ServeDNS(context.Background(), w, r)

			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(dns.RcodeSuccess))
			Expect(w.Msg.Answer).To(HaveLen(0))
		})

		It("should return correct A record", func() {
			cache := dnsCache.NewEndpointsCache()

			cache.Put("nginx", "green", "default", []slice.Endpoint{{
				Host: "nginx.default.slice.local",
				IP:   "10.0.0.1",
			}, {
				Host: "wrong-host",
				IP:   "10.0.1.1",
			}})

			ks := kubeslice.Kubeslice{
				EndpointsCache: cache,
			}

			r := &dns.Msg{
				Question: []dns.Question{{
					Name:  "nginx.default.slice.local.",
					Qtype: dns.TypeA,
				}},
			}
			w := &mockResponse{}

			code, err := ks.ServeDNS(context.Background(), w, r)

			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(dns.RcodeSuccess))
			Expect(w.Msg.Answer).To(HaveLen(1))
			Expect(w.Msg.Answer[0].String()).To(Equal("nginx.default.slice.local.	0	IN	A	10.0.0.1"))
		})

		It("should return multiple A records", func() {
			cache := dnsCache.NewEndpointsCache()

			cache.Put("nginx", "green", "default", []slice.Endpoint{{
				Host: "nginx.default.slice.local",
				IP:   "10.0.0.1",
			}, {
				Host: "nginx.default.slice.local",
				IP:   "10.0.1.1",
			}})

			ks := kubeslice.Kubeslice{
				EndpointsCache: cache,
			}

			r := &dns.Msg{
				Question: []dns.Question{{
					Name:  "nginx.default.slice.local.",
					Qtype: dns.TypeA,
				}},
			}
			w := &mockResponse{}

			code, err := ks.ServeDNS(context.Background(), w, r)

			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(dns.RcodeSuccess))
			Expect(w.Msg.Answer).To(HaveLen(2))
			Expect(w.Msg.Answer[0].String()).To(Equal("nginx.default.slice.local.	0	IN	A	10.0.0.1"))
			Expect(w.Msg.Answer[1].String()).To(Equal("nginx.default.slice.local.	0	IN	A	10.0.1.1"))
		})

		It("should return empty response for AAAA requests", func() {
			cache := dnsCache.NewEndpointsCache()

			cache.Put("nginx", "green", "default", []slice.Endpoint{{
				Host: "nginx.default.slice.local",
				IP:   "10.0.0.1",
			}, {
				Host: "wrong-host",
				IP:   "10.0.1.1",
			}})

			ks := kubeslice.Kubeslice{
				EndpointsCache: cache,
			}

			r := &dns.Msg{
				Question: []dns.Question{{
					Name:  "nginx.default.slice.local.",
					Qtype: dns.TypeAAAA,
				}},
			}
			w := &mockResponse{}

			code, err := ks.ServeDNS(context.Background(), w, r)

			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(dns.RcodeSuccess))
			Expect(w.Msg.Answer).To(HaveLen(0))
		})

	})

})

type mockResponse struct {
	Msg *dns.Msg
}

func (m *mockResponse) LocalAddr() net.Addr {
	return nil
}

func (m *mockResponse) RemoteAddr() net.Addr {
	return nil
}

func (m *mockResponse) WriteMsg(msg *dns.Msg) error {
	m.Msg = msg
	return nil
}

func (m *mockResponse) Write(b []byte) (int, error) {
	return 0, nil
}

func (m *mockResponse) Close() error {
	return nil
}

func (m *mockResponse) TsigStatus() error {
	return nil
}

func (m *mockResponse) TsigTimersOnly(b bool) {
}

func (m *mockResponse) Hijack() {
}
