package compute

import (
	"encoding/json"
	"math"
	"testing"
)

const bbriFixture = `[
  {"symbol":"BBRI.JK","date":"2025-06-30",
   "earnings":15724782000000,"revenue":53348647000000,"total_assets":2352380589000000,
   "total_equity":328674878000000,"operating_expense":20251399000000,"provision":12061556000000,
   "non_interest_income":12604952000000,
   "financials_sector_metrics":{
     "net_interest_income":40375410000000,"interest_income":55085430000000,"interest_expense":14710020000000,
     "gross_loan":1580422630000000,"net_loan":1498037923000000,"total_deposit":1580681886000000,
     "current_account":449565770000000,"savings_account":619099972000000,"time_deposit":512016144000000}},
  {"symbol":"BBRI.JK","date":"2025-03-31",
   "earnings":15770304000000,"revenue":52000000000000,"total_assets":2250000000000000,
   "total_equity":320000000000000,"operating_expense":20000000000000,"provision":11000000000000,
   "non_interest_income":12000000000000,
   "financials_sector_metrics":{
     "net_interest_income":39000000000000,"interest_income":54000000000000,"interest_expense":15000000000000,
     "gross_loan":1460000000000000,"net_loan":1400000000000000,"total_deposit":1500000000000000,
     "current_account":440000000000000,"savings_account":610000000000000,"time_deposit":450000000000000}}
]`

func TestMetricsFromQuartersBBRI(t *testing.T) {
	quarters, err := ParseQuarters(json.RawMessage(bbriFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(quarters) != 2 {
		t.Fatalf("expected 2 quarters, got %d", len(quarters))
	}
	m := MetricsFromQuarters("BBRI", quarters)
	if m.AsOf != "2025-06-30" {
		t.Fatalf("expected latest date 2025-06-30, got %s", m.AsOf)
	}
	want := 15724782000000.0 * 4 / 328674878000000
	if m.ROE == nil || math.Abs(*m.ROE-want) > 1e-9 {
		t.Fatalf("ROE mismatch: want %v got %v", want, m.ROE)
	}
	wantCASA := (449565770000000.0 + 619099972000000.0) / 1580681886000000.0
	if m.CASA == nil || math.Abs(*m.CASA-wantCASA) > 1e-9 {
		t.Fatalf("CASA mismatch: want %v got %v", wantCASA, m.CASA)
	}
	wantLDR := 1498037923000000.0 / 1580681886000000.0
	if m.LDR == nil || math.Abs(*m.LDR-wantLDR) > 1e-9 {
		t.Fatalf("LDR mismatch: want %v got %v", wantLDR, m.LDR)
	}
	wantQoQ := (15724782000000.0 - 15770304000000.0) / 15770304000000.0
	if m.EarningsQoQ == nil || math.Abs(*m.EarningsQoQ-wantQoQ) > 1e-9 {
		t.Fatalf("EarningsQoQ mismatch: want %v got %v", wantQoQ, m.EarningsQoQ)
	}
	wantCTI := 20251399000000.0 / 53348647000000.0
	if m.CostToIncome == nil || math.Abs(*m.CostToIncome-wantCTI) > 1e-9 {
		t.Fatalf("CostToIncome mismatch: want %v got %v", wantCTI, m.CostToIncome)
	}
}

func TestParseQuartersNonBank(t *testing.T) {
	raw := `[{"date":"2025-06-30","earnings":1000,"revenue":5000,"total_assets":100000,"total_equity":50000}]`
	quarters, err := ParseQuarters(json.RawMessage(raw))
	if err != nil {
		t.Fatal(err)
	}
	m := MetricsFromQuarters("TLKM", quarters)
	if m.ROE == nil || math.Abs(*m.ROE-0.08) > 1e-9 {
		t.Fatalf("ROE for non-bank: got %v", m.ROE)
	}
	if m.NIM != nil {
		t.Fatal("NIM must be nil for non-bank")
	}
	if len(m.Notes) == 0 {
		t.Fatal("expected a note about short history")
	}
}

func TestCompareCompanies(t *testing.T) {
	roe := func(v float64) *float64 { return &v }
	a := CompanyMetrics{Symbol: "AAA", Quarterly: &QuarterMetrics{ROE: roe(0.20), CostToIncome: roe(0.5)}}
	b := CompanyMetrics{Symbol: "BBB", Quarterly: &QuarterMetrics{ROE: roe(0.10), CostToIncome: roe(0.6)}}
	c := CompareCompanies([]CompanyMetrics{a, b})
	if c.Ranking[0] != "AAA" {
		t.Fatalf("expected AAA first, got %v", c.Ranking)
	}
	if c.Overall["AAA"] <= c.Overall["BBB"] {
		t.Fatalf("AAA must outscore BBB: %v", c.Overall)
	}
	if len(c.Rubric) < 50 {
		t.Fatal("rubric must explain the scoring")
	}
}
