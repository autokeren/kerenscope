package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/autokeren/kerenscope/internal/sectors"
	"github.com/autokeren/kerenscope/internal/tools"
)

func resultSummary(name string, args map[string]any, result tools.Result) string {
	if !result.OK {
		msg := result.Error
		if len(msg) > 90 {
			msg = msg[:90] + "…"
		}
		return msg
	}
	switch name {
	case "company_report":
		if r, ok := result.Data.(*sectors.CompanyReport); ok {
			return fmt.Sprintf("%s · %s · market cap %s (#%d IDX)", r.Symbol, r.CompanyName, compactIDR(float64(r.Overview.MarketCap)), r.Overview.MarketCapRank)
		}
	case "quarterly_financials":
		if s, ok := peekQuarters(result.Data); ok {
			return fmt.Sprintf("%s · %d quarters, latest %s", argSymbol(args), s.n, s.latest)
		}
	case "price_history":
		if p, ok := peekPrices(result.Data); ok {
			return fmt.Sprintf("%s · %s → %s close over %d days (%s)", argSymbol(args), compactIDR(p.first), compactIDR(p.last), p.n, pct(p.change))
		}
	case "foreign_flow":
		if f, ok := peekFlow(result.Data); ok {
			return fmt.Sprintf("%s · net foreign flow %s over %d days", argSymbol(args), signedIDR(f.net), f.n)
		}
	case "broker_summary":
		if b, ok := peekBroker(result.Data); ok {
			return fmt.Sprintf("%s · %d top buyers / %d top sellers, %d days", argSymbol(args), b.buyers, b.sellers, b.n)
		}
	case "news":
		if n, ok := peekResults(result.Data); ok {
			return fmt.Sprintf("%d articles (latest: %s)", n.n, n.first)
		}
	case "insider_filings":
		if n, ok := peekResults(result.Data); ok {
			if n.n == 0 {
				return "no filings in window"
			}
			return fmt.Sprintf("%d filings (latest: %s)", n.n, n.first)
		}
	case "screen_companies":
		if n, ok := peekArrayLen(result.Data); ok {
			return fmt.Sprintf("%d companies matched", n)
		}
	case "subsector_report":
		if s, ok := peekSubsector(result.Data); ok {
			return fmt.Sprintf("%s · %d companies", s.sub, s.total)
		}
	case "sectors_api":
		return argString(args, "path")
	case "top_movers":
		if n, ok := peekArrayLen(result.Data); ok {
			return fmt.Sprintf("%d movers", n)
		}
	}
	s := mustJSON(result.Data)
	if len(s) > 70 {
		s = s[:70] + "…"
	}
	return s
}

func argSymbol(args map[string]any) string {
	return argString(args, "symbol")
}

func argString(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return strings.ToUpper(strings.TrimSpace(v))
	}
	return "?"
}

type quartersPeek struct {
	n      int
	latest string
}

func peekQuarters(data any) (quartersPeek, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return quartersPeek{}, false
	}
	var rows []struct {
		Date string `json:"date"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) == 0 {
		return quartersPeek{}, false
	}
	return quartersPeek{n: len(rows), latest: rows[len(rows)-1].Date}, true
}

type pricesPeek struct {
	n      int
	first  float64
	last   float64
	change float64
}

func peekPrices(data any) (pricesPeek, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return pricesPeek{}, false
	}
	var rows []struct {
		Close *float64 `json:"close"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) < 2 {
		return pricesPeek{}, false
	}
	var closes []float64
	for _, r := range rows {
		if r.Close != nil {
			closes = append(closes, *r.Close)
		}
	}
	if len(closes) < 2 {
		return pricesPeek{}, false
	}
	first, last := closes[0], closes[len(closes)-1]
	change := 0.0
	if first != 0 {
		change = (last - first) / first
	}
	return pricesPeek{n: len(closes), first: first, last: last, change: change}, true
}

type flowPeek struct {
	n   int
	net float64
}

func peekFlow(data any) (flowPeek, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return flowPeek{}, false
	}
	var payload struct {
		Data []struct {
			Net *float64 `json:"net_foreign_inflow"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return flowPeek{}, false
	}
	peek := flowPeek{}
	for _, d := range payload.Data {
		peek.n++
		if d.Net != nil {
			peek.net += *d.Net
		}
	}
	return peek, peek.n > 0
}

type brokerPeek struct {
	buyers  int
	sellers int
	n       int
}

func peekBroker(data any) (brokerPeek, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return brokerPeek{}, false
	}
	var payload struct {
		Start      string `json:"start"`
		End        string `json:"end"`
		TopBuyers  []any  `json:"top_buyers"`
		TopSellers []any  `json:"top_sellers"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return brokerPeek{}, false
	}
	buyers, sellers := len(payload.TopBuyers), len(payload.TopSellers)
	if buyers == 0 && sellers == 0 {
		return brokerPeek{}, false
	}
	days := daysBetween(payload.Start, payload.End)
	return brokerPeek{buyers: buyers, sellers: sellers, n: days}, true
}

type resultsPeek struct {
	n     int
	first string
}

func peekResults(data any) (resultsPeek, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return resultsPeek{}, false
	}
	var payload struct {
		Results []struct {
			Title string `json:"title"`
			Name  string `json:"name"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return resultsPeek{}, false
	}
	peek := resultsPeek{n: len(payload.Results)}
	if len(payload.Results) > 0 {
		peek.first = payload.Results[0].Title
		if peek.first == "" {
			peek.first = payload.Results[0].Name
		}
		if len(peek.first) > 48 {
			peek.first = peek.first[:48] + "…"
		}
	}
	return peek, true
}

func peekArrayLen(data any) (int, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return 0, false
	}
	var arr []any
	if err := json.Unmarshal(raw, &arr); err != nil {
		return 0, false
	}
	return len(arr), true
}

type subsectorPeek struct {
	sub   string
	total int
}

func peekSubsector(data any) (subsectorPeek, bool) {
	raw, err := json.Marshal(data)
	if err != nil {
		return subsectorPeek{}, false
	}
	var payload struct {
		SubSector   string `json:"sub_sector"`
		Statistics  struct {
			TotalCompanies int `json:"total_companies"`
		} `json:"statistics"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return subsectorPeek{}, false
	}
	return subsectorPeek{sub: payload.SubSector, total: payload.Statistics.TotalCompanies}, payload.SubSector != ""
}

func compactIDR(v float64) string {
	abs := v
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs >= 1e12:
		return fmt.Sprintf("Rp%.1fT", v/1e12)
	case abs >= 1e9:
		return fmt.Sprintf("Rp%.1fB", v/1e9)
	case abs >= 1e6:
		return fmt.Sprintf("Rp%.1fM", v/1e6)
	default:
		return fmt.Sprintf("Rp%.0f", v)
	}
}

func signedIDR(v float64) string {
	if v >= 0 {
		return "+" + compactIDR(v)
	}
	return "-" + compactIDR(-v)
}

func pct(v float64) string {
	return fmt.Sprintf("%+.1f%%", v*100)
}

func daysBetween(start, end string) int {
	t1, err1 := time.Parse("2006-01-02", strings.TrimSpace(start))
	t2, err2 := time.Parse("2006-01-02", strings.TrimSpace(end))
	if err1 != nil || err2 != nil {
		return 30
	}
	d := int(t2.Sub(t1).Hours() / 24)
	if d <= 0 {
		return 30
	}
	return d
}
