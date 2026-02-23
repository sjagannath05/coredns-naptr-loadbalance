package naptrloadbalance

import (
	"testing"

	"github.com/miekg/dns"
)

func TestShuffleNaptrNoRecords(t *testing.T) {
	in := []dns.RR{}
	out := shuffleNaptr(in, false)
	if len(out) != 0 {
		t.Errorf("expected empty, got %d records", len(out))
	}
}

func TestShuffleNaptrSingleRecord(t *testing.T) {
	in := []dns.RR{
		makeNAPTR("example.com.", 20, 100, "node01.example.com."),
	}
	out := shuffleNaptr(in, false)
	if len(out) != 1 {
		t.Errorf("expected 1 record, got %d", len(out))
	}
}

func TestShuffleNaptrPreservesNonNaptr(t *testing.T) {
	a := &dns.A{Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 120}}
	naptr1 := makeNAPTR("example.com.", 20, 100, "node01.example.com.")
	naptr2 := makeNAPTR("example.com.", 20, 100, "node03.example.com.")

	in := []dns.RR{a, naptr1, naptr2}
	out := shuffleNaptr(in, false)

	if len(out) != 3 {
		t.Fatalf("expected 3 records, got %d", len(out))
	}

	// A record should be first (non-NAPTR records come before shuffled NAPTRs)
	if out[0].Header().Rrtype != dns.TypeA {
		t.Errorf("expected A record first, got type %d", out[0].Header().Rrtype)
	}

	// Both NAPTRs should still be present
	naptrCount := 0
	for _, r := range out {
		if r.Header().Rrtype == dns.TypeNAPTR {
			naptrCount++
		}
	}
	if naptrCount != 2 {
		t.Errorf("expected 2 NAPTR records, got %d", naptrCount)
	}
}

func TestShuffleNaptrDistribution(t *testing.T) {
	node01First := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		naptr1 := makeNAPTR("example.com.", 20, 100, "node01.example.com.")
		naptr2 := makeNAPTR("example.com.", 20, 100, "node03.example.com.")
		in := []dns.RR{naptr1, naptr2}

		out := shuffleNaptr(in, false)
		if out[0].(*dns.NAPTR).Replacement == "node01.example.com." {
			node01First++
		}
	}

	ratio := float64(node01First) / float64(iterations)
	if ratio < 0.35 || ratio > 0.65 {
		t.Errorf("uneven distribution: node01 was first %d/%d times (%.1f%%)", node01First, iterations, ratio*100)
	}
}

func TestShuffleNaptrSingleMode(t *testing.T) {
	naptr1 := makeNAPTR("example.com.", 20, 100, "node01.example.com.")
	naptr2 := makeNAPTR("example.com.", 20, 100, "node03.example.com.")

	in := []dns.RR{naptr1, naptr2}
	out := shuffleNaptr(in, true)

	if len(out) != 1 {
		t.Fatalf("single mode: expected 1 record, got %d", len(out))
	}

	if out[0].Header().Rrtype != dns.TypeNAPTR {
		t.Errorf("expected NAPTR record, got type %d", out[0].Header().Rrtype)
	}
}

func TestShuffleNaptrSingleModeDistribution(t *testing.T) {
	node01Count := 0
	iterations := 1000

	for i := 0; i < iterations; i++ {
		naptr1 := makeNAPTR("example.com.", 20, 100, "node01.example.com.")
		naptr2 := makeNAPTR("example.com.", 20, 100, "node03.example.com.")
		in := []dns.RR{naptr1, naptr2}

		out := shuffleNaptr(in, true)
		if len(out) != 1 {
			t.Fatalf("single mode: expected 1 record, got %d", len(out))
		}
		if out[0].(*dns.NAPTR).Replacement == "node01.example.com." {
			node01Count++
		}
	}

	ratio := float64(node01Count) / float64(iterations)
	if ratio < 0.35 || ratio > 0.65 {
		t.Errorf("single mode uneven distribution: node01 returned %d/%d times (%.1f%%)", node01Count, iterations, ratio*100)
	}
}

func TestShuffleNaptrSingleModePreservesNonNaptr(t *testing.T) {
	a := &dns.A{Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 120}}
	naptr1 := makeNAPTR("example.com.", 20, 100, "node01.example.com.")
	naptr2 := makeNAPTR("example.com.", 20, 100, "node03.example.com.")

	in := []dns.RR{a, naptr1, naptr2}
	out := shuffleNaptr(in, true)

	if len(out) != 2 {
		t.Fatalf("single mode with A: expected 2 records, got %d", len(out))
	}

	if out[0].Header().Rrtype != dns.TypeA {
		t.Errorf("expected A record first, got type %d", out[0].Header().Rrtype)
	}
	if out[1].Header().Rrtype != dns.TypeNAPTR {
		t.Errorf("expected NAPTR record second, got type %d", out[1].Header().Rrtype)
	}
}

func TestShuffleNaptrSingleModeOneRecord(t *testing.T) {
	in := []dns.RR{
		makeNAPTR("example.com.", 20, 100, "node01.example.com."),
	}
	out := shuffleNaptr(in, true)
	if len(out) != 1 {
		t.Errorf("single mode with 1 NAPTR: expected 1 record, got %d", len(out))
	}
}

func makeNAPTR(name string, order, preference uint16, replacement string) *dns.NAPTR {
	return &dns.NAPTR{
		Hdr:         dns.RR_Header{Name: name, Rrtype: dns.TypeNAPTR, Class: dns.ClassINET, Ttl: 120},
		Order:       order,
		Preference:  preference,
		Flags:       "a",
		Service:     "x-3gpp-pgw:x-s5-gtp:x-s8-gtp:x-gn:x-gp",
		Regexp:      "",
		Replacement: replacement,
	}
}
