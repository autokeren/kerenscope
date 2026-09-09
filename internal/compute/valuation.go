package compute

import (
	"sort"

	"github.com/autokeren/kerenscope/internal/sectors"
)

type ValuationMetrics struct {
	Symbol          string   `json:"symbol"`
	LastClose       *float64 `json:"last_close,omitempty"`
	CloseDate       string   `json:"close_date,omitempty"`
	MarketCap       *float64 `json:"market_cap,omitempty"`
	MarketCapRank   int      `json:"market_cap_rank,omitempty"`
	PE              *float64 `json:"pe,omitempty"`
	PEPeer          *float64 `json:"pe_peer_avg,omitempty"`
	PEPremium       *float64 `json:"pe_premium_vs_peer,omitempty"`
	PB              *float64 `json:"pb,omitempty"`
	PBPeer          *float64 `json:"pb_peer_avg,omitempty"`
	PBPremium       *float64 `json:"pb_premium_vs_peer,omitempty"`
	PS              *float64 `json:"ps,omitempty"`
	PSPeer          *float64 `json:"ps_peer_avg,omitempty"`
	PSPremium       *float64 `json:"ps_premium_vs_peer,omitempty"`
	PEG             *float64 `json:"peg,omitempty"`
	ForwardPE       *float64 `json:"forward_pe,omitempty"`
	PEHistoryMean   *float64 `json:"pe_history_mean,omitempty"`
	PEVsHistory     *float64 `json:"pe_vs_history_mean,omitempty"`
	PEHistoryYears  []int    `json:"pe_history_years,omitempty"`
	GrowthForecast  *float64 `json:"eps_growth_forecast,omitempty"`
	AnalystBuy      int      `json:"analyst_buy,omitempty"`
	AnalystHold     int      `json:"analyst_hold,omitempty"`
	AnalystSell     int      `json:"analyst_sell,omitempty"`
	AnalystTotal    int      `json:"analyst_total,omitempty"`
}

func FromCompanyReport(r *sectors.CompanyReport) ValuationMetrics {
	m := ValuationMetrics{
		Symbol:        r.Symbol,
		CloseDate:     r.Valuation.LatestCloseDate,
		MarketCap:     nz(float64(r.Overview.MarketCap)),
		MarketCapRank: r.Overview.MarketCapRank,
		ForwardPE:     nz(r.Valuation.ForwardPE),
	}
	if r.Valuation.LastClosePrice > 0 {
		m.LastClose = nz(float64(r.Valuation.LastClosePrice))
	}
	if v := r.LatestValuation(); v != nil {
		m.PE = nz(v.PE)
		m.PEPeer = nz(v.PEPeer)
		m.PB = nz(v.PB)
		m.PBPeer = nz(v.PBPeer)
		m.PS = nz(v.PS)
		m.PSPeer = nz(v.PSPeer)
		m.PEG = nz(v.PEG)
		if len(r.Valuation.HistoricalValuation) > 0 {
			years := make([]int, 0, len(r.Valuation.HistoricalValuation))
			for _, hv := range r.Valuation.HistoricalValuation {
				years = append(years, hv.Year)
			}
			m.PEHistoryYears = years
			if v.PE > 0 {
				sum, n := 0.0, 0
				for _, hv := range r.Valuation.HistoricalValuation {
					if hv.PE > 0 {
						sum += hv.PE
						n++
					}
				}
				if n > 0 {
					mean := sum / float64(n)
					m.PEHistoryMean = nz(mean)
					rel := (v.PE - mean) / mean
					m.PEVsHistory = nz(rel)
				}
			}
		}
	}
	if m.PE != nil && m.PEPeer != nil && *m.PEPeer > 0 {
		rel := (*m.PE - *m.PEPeer) / *m.PEPeer
		m.PEPremium = nz(rel)
	}
	if m.PB != nil && m.PBPeer != nil && *m.PBPeer > 0 {
		rel := (*m.PB - *m.PBPeer) / *m.PBPeer
		m.PBPremium = nz(rel)
	}
	if m.PS != nil && m.PSPeer != nil && *m.PSPeer > 0 {
		rel := (*m.PS - *m.PSPeer) / *m.PSPeer
		m.PSPremium = nz(rel)
	}
	maxEstimateYear := 0
	for _, f := range r.Future.CompanyGrowthForecasts {
		if f.EstimateYear > maxEstimateYear {
			maxEstimateYear = f.EstimateYear
		}
	}
	for _, f := range r.Future.CompanyGrowthForecasts {
		if f.EstimateYear == maxEstimateYear {
			m.GrowthForecast = nz(f.EPSGrowth)
		}
	}
	m.AnalystBuy = r.Future.AnalystRating.Buy
	m.AnalystHold = r.Future.AnalystRating.Hold
	m.AnalystSell = r.Future.AnalystRating.Sell
	m.AnalystTotal = r.Future.AnalystRating.NAnalyst
	return m
}

type CompanyMetrics struct {
	Symbol    string           `json:"symbol"`
	Name      string           `json:"name,omitempty"`
	Quarterly *QuarterMetrics  `json:"quarterly,omitempty"`
	Valuation *ValuationMetrics `json:"valuation,omitempty"`
}

type Pillar struct {
	Name   string             `json:"name"`
	Weight float64            `json:"weight"`
	Scores map[string]float64 `json:"scores"`
}

type Comparison struct {
	Companies []string `json:"companies"`
	Pillars   []Pillar `json:"pillars"`
	Overall   map[string]float64 `json:"overall_score"`
	Ranking   []string `json:"ranking"`
	Rubric    string   `json:"rubric"`
}

func CompareCompanies(metrics []CompanyMetrics) Comparison {
	pick := func(m CompanyMetrics, fn func(QuarterMetrics) *float64) (float64, bool) {
		if m.Quarterly == nil {
			return 0, false
		}
		v := fn(*m.Quarterly)
		if v == nil {
			return 0, false
		}
		return *v, true
	}
	pickV := func(m CompanyMetrics, fn func(ValuationMetrics) *float64) (float64, bool) {
		if m.Valuation == nil {
			return 0, false
		}
		v := fn(*m.Valuation)
		if v == nil {
			return 0, false
		}
		return *v, true
	}

	type metric struct {
		name   string
		weight float64
		get    func(CompanyMetrics) (float64, bool)
		lower  bool
	}
	mets := []metric{
		{"roe", 1.0, func(m CompanyMetrics) (float64, bool) { return pick(m, func(q QuarterMetrics) *float64 { return q.ROE }) }, false},
		{"roa", 0.7, func(m CompanyMetrics) (float64, bool) { return pick(m, func(q QuarterMetrics) *float64 { return q.ROA }) }, false},
		{"nim", 0.7, func(m CompanyMetrics) (float64, bool) { return pick(m, func(q QuarterMetrics) *float64 { return q.NIM }) }, false},
		{"casa", 0.5, func(m CompanyMetrics) (float64, bool) { return pick(m, func(q QuarterMetrics) *float64 { return q.CASA }) }, false},
		{"cost_to_income", 0.5, func(m CompanyMetrics) (float64, bool) { return pick(m, func(q QuarterMetrics) *float64 { return q.CostToIncome }) }, true},
		{"provision_ratio", 0.5, func(m CompanyMetrics) (float64, bool) { return pick(m, func(q QuarterMetrics) *float64 { return q.ProvisionRatio }) }, true},
		{"earnings_growth", 0.8, func(m CompanyMetrics) (float64, bool) {
			return pick(m, func(q QuarterMetrics) *float64 {
				if q.EarningsYoY != nil {
					return q.EarningsYoY
				}
				return q.EarningsQoQ
			})
		}, false},
		{"loan_growth", 0.4, func(m CompanyMetrics) (float64, bool) { return pick(m, func(q QuarterMetrics) *float64 { return q.LoanGrowth }) }, false},
		{"pe_discount", 0.8, func(m CompanyMetrics) (float64, bool) { return pickV(m, func(v ValuationMetrics) *float64 { return v.PEPremium }) }, true},
		{"pb_discount", 0.5, func(m CompanyMetrics) (float64, bool) { return pickV(m, func(v ValuationMetrics) *float64 { return v.PBPremium }) }, true},
	}

	companies := make([]string, len(metrics))
	scores := map[string]float64{}
	contributed := map[string]int{}
	for i, m := range metrics {
		companies[i] = m.Symbol
	}
	for _, met := range mets {
		vals := map[string]float64{}
		present := make([]string, 0, len(metrics))
		for _, m := range metrics {
			if v, ok := met.get(m); ok {
				vals[m.Symbol] = v
				present = append(present, m.Symbol)
			}
		}
		if len(present) < 2 {
			continue
		}
		for _, sym := range present {
			s := normalize(vals[sym], vals, met.lower)
			scores[sym] += s * met.weight
			contributed[sym]++
		}
	}
	overall := map[string]float64{}
	for sym, s := range scores {
		w := float64(contributed[sym])
		if w > 0 {
			overall[sym] = round2(s / w)
		}
	}
	ranked := append([]string(nil), companies...)
	sort.Slice(ranked, func(i, j int) bool { return overall[ranked[i]] > overall[ranked[j]] })
	return Comparison{
		Companies: companies,
		Overall:   overall,
		Ranking:   ranked,
		Rubric: "Each metric is min-max normalized to 0-100 across the compared companies, then averaged per company with relative weights (profitability: roe 1.0, roa 0.7, nim 0.7, casa 0.5, cost_to_income 0.5 lower-better, provision_ratio 0.5 lower-better; growth: earnings_growth 0.8, loan_growth 0.4; valuation: pe_discount 0.8 lower-better, pb_discount 0.5 lower-better). Descriptive ranking of computed data only — not investment advice.",
	}
}

func normalize(v float64, vals map[string]float64, lower bool) float64 {
	min, max := v, v
	for _, x := range vals {
		if x < min {
			min = x
		}
		if x > max {
			max = x
		}
	}
	if max == min {
		return 50
	}
	s := (v - min) / (max - min) * 100
	if lower {
		s = 100 - s
	}
	return s
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
