package verify

import (
	"strings"
	"testing"
)

func TestExtractIndonesianNumbers(t *testing.T) {
	draft := "BBRI diperdagangkan pada P/E 8,23x dengan market cap Rp 513,1 triliun. ROE 19,1% dan kredit Rp 1.592 T. Analyst 46 buy."
	nums := ExtractNumbers(draft)
	found := map[string]DraftNumber{}
	for _, n := range nums {
		found[n.Raw] = n
	}
	cases := []struct {
		raw string
		value float64
		kind NumberKind
		scale float64
	}{
		{"8,23", 8.23, KindRatio, 1},
		{"513,1", 513.1, KindIDR, 1e12},
		{"19,1", 19.1, KindPercent, 1},
		{"1.592", 1592, KindIDR, 1e12},
		{"46", 46, KindPlain, 1},
	}
	for _, c := range cases {
		n, ok := found[c.raw]
		if !ok {
			t.Fatalf("expected %q extracted, got %v", c.raw, nums)
		}
		if n.Value != c.value {
			t.Fatalf("%q: value want %v got %v", c.raw, c.value, n.Value)
		}
		if n.Kind != c.kind {
			t.Fatalf("%q: kind want %v got %v", c.raw, c.kind, n.Kind)
		}
		if n.Scale != c.scale {
			t.Fatalf("%q: scale want %v got %v", c.raw, c.scale, n.Scale)
		}
	}
}

func TestCheckNumbersMatchesRealFacts(t *testing.T) {
	evidence := `{"symbol":"BBRI.JK","overview":{"market_cap":513100000000000},"valuation":{"historical_valuation":[{"pe":8.23,"pb":1.59,"pe_peer_avg":10.21}]},"computed_roe":0.191,"analyst":{"buy":46}}`
	facts := CollectFacts(evidence)
	draft := "BBRI: market cap Rp 513,1 triliun, P/E 8,23x (peer 10,21), ROE 19,1%, konsensus 46 buy, laba Rp 15,7 triliun."
	result := CheckNumbers(draft, facts)
	matched := map[string]bool{}
	for _, m := range result.Matched {
		matched[m.Number] = true
	}
	for _, want := range []string{"513,1", "8,23", "10,21", "19,1", "46"} {
		if !matched[want] {
			t.Fatalf("expected %q matched; matched=%v unmatched=%v", want, result.Matched, result.Unmatched)
		}
	}
	stillUnmatched := []string{}
	for _, u := range result.Unmatched {
		if strings.Contains(u, "15,7") {
			stillUnmatched = append(stillUnmatched, u)
		}
	}
	if len(stillUnmatched) != 1 {
		t.Fatalf("15,7 should be the only unmatched number, got %v", result.Unmatched)
	}
}

func TestSkipTrivialAndYears(t *testing.T) {
	draft := "Pada 2026, kuartal Q2 menunjukkan tren naik selama 3 kuartal terakhir. Laba Rp 15,7 triliun."
	nums := ExtractNumbers(draft)
	for _, n := range nums {
		if n.Raw == "2026" || n.Raw == "2" || n.Raw == "3" {
			t.Fatalf("expected %q to be skipped", n.Raw)
		}
	}
	if len(nums) != 1 || nums[0].Raw != "15,7" {
		t.Fatalf("expected only 15,7 extracted, got %v", nums)
	}
}
