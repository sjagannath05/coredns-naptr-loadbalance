// Package naptrloadbalance is a CoreDNS plugin that shuffles NAPTR records
// in DNS responses for round-robin load balancing.
package naptrloadbalance

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

var plog = log.NewWithPlugin("naptr_loadbalance")

// NaptrLoadBalance is a plugin that shuffles NAPTR records in responses.
type NaptrLoadBalance struct {
	Next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (n NaptrLoadBalance) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	rw := &naptrResponseWriter{ResponseWriter: w}
	return plugin.NextOrFailure(n.Name(), n.Next, ctx, rw, r)
}

// Name implements the Handler interface.
func (n NaptrLoadBalance) Name() string { return "naptr_loadbalance" }

// naptrResponseWriter is a response writer that shuffles NAPTR records.
type naptrResponseWriter struct {
	dns.ResponseWriter
}

// WriteMsg implements the dns.ResponseWriter interface.
func (w *naptrResponseWriter) WriteMsg(res *dns.Msg) error {
	if res.Rcode != dns.RcodeSuccess {
		return w.ResponseWriter.WriteMsg(res)
	}

	if res.Question[0].Qtype == dns.TypeAXFR || res.Question[0].Qtype == dns.TypeIXFR {
		return w.ResponseWriter.WriteMsg(res)
	}

	res.Answer = shuffleNaptr(res.Answer)
	res.Ns = shuffleNaptr(res.Ns)
	res.Extra = shuffleNaptr(res.Extra)

	return w.ResponseWriter.WriteMsg(res)
}

// Write implements the dns.ResponseWriter interface.
func (w *naptrResponseWriter) Write(buf []byte) (int, error) {
	plog.Warning("NaptrLoadBalance called with Write: not shuffling records")
	return w.ResponseWriter.Write(buf)
}

// shuffleNaptr separates NAPTR records, shuffles them, and reassembles.
func shuffleNaptr(in []dns.RR) []dns.RR {
	naptr := []dns.RR{}
	rest := []dns.RR{}
	for _, r := range in {
		if r.Header().Rrtype == dns.TypeNAPTR {
			naptr = append(naptr, r)
		} else {
			rest = append(rest, r)
		}
	}

	if len(naptr) < 2 {
		return in
	}

	shuffleRecords(naptr)
	return append(rest, naptr...)
}

// shuffleRecords randomizes the order of DNS records using the same
// algorithm as CoreDNS loadbalance plugin.
func shuffleRecords(records []dns.RR) {
	switch l := len(records); l {
	case 0, 1:
		break
	case 2:
		if dns.Id()%2 == 0 {
			records[0], records[1] = records[1], records[0]
		}
	default:
		for j := range l {
			p := j + (int(dns.Id()) % (l - j))
			if j == p {
				continue
			}
			records[j], records[p] = records[p], records[j]
		}
	}
}
