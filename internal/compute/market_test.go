package compute

import (
	"encoding/json"
	"math"
	"testing"
)

func TestDigestPrice(t *testing.T) {
	raw := json.RawMessage(`[
		{"symbol":"BBCA.JK","date":"2026-06-11","close":100,"volume":1000},
		{"symbol":"BBCA.JK","date":"2026-06-12","close":110,"volume":2000},
		{"symbol":"BBCA.JK","date":"2026-06-15","close":90,"volume":3000},
		{"symbol":"BBCA.JK","date":"2026-06-16","close":105,"volume":4000}
	]`)
	series, err := ParsePriceSeries(raw)
	if err != nil {
		t.Fatal(err)
	}
	d := DigestPrice(series)
	if d.FirstClose != 100 || d.LastClose != 105 {
		t.Fatalf("closes: %v -> %v", d.FirstClose, d.LastClose)
	}
	if d.ChangePct == nil || math.Abs(*d.ChangePct-0.05) > 1e-9 {
		t.Fatalf("change: %v", d.ChangePct)
	}
	if d.HighClose != 110 || d.LowClose != 90 {
		t.Fatalf("high/low: %v/%v", d.HighClose, d.LowClose)
	}
	if d.MaxDrawdownPct == nil || math.Abs(*d.MaxDrawdownPct-(20.0/110.0)) > 1e-9 {
		t.Fatalf("max drawdown: %v (want %v)", d.MaxDrawdownPct, 20.0/110.0)
	}
	if d.AvgVolume != 2500 || d.LastVolume != 4000 {
		t.Fatalf("volumes: avg=%v last=%v", d.AvgVolume, d.LastVolume)
	}
}

func TestDigestFlow(t *testing.T) {
	raw := json.RawMessage(`{"symbol":"BBCA.JK","start":"2026-06-11","end":"2026-06-13","data":[
		{"date":"2026-06-11","net_foreign_inflow":100},
		{"date":"2026-06-12","net_foreign_inflow":-300},
		{"date":"2026-06-13","net_foreign_inflow":50}
	]}`)
	d, err := DigestFlow(raw)
	if err != nil {
		t.Fatal(err)
	}
	if d.NetTotal != -150 {
		t.Fatalf("net total: %v", d.NetTotal)
	}
	if d.BuyDays != 2 || d.SellDays != 1 {
		t.Fatalf("days: buy=%d sell=%d", d.BuyDays, d.SellDays)
	}
	if d.NetLast10d == nil || *d.NetLast10d != -150 {
		t.Fatalf("last 10d: %v", d.NetLast10d)
	}
	if d.MaxInflowDay == nil || *d.MaxInflowDay.NetFlow != 100 {
		t.Fatalf("max inflow: %v", d.MaxInflowDay)
	}
	if d.MaxOutflowDay == nil || *d.MaxOutflowDay.NetFlow != -300 {
		t.Fatalf("max outflow: %v", d.MaxOutflowDay)
	}
}
