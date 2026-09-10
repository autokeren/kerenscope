package compute

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

type PriceSeries struct {
	Symbol  string    `json:"symbol"`
	Dates   []string  `json:"dates"`
	Closes  []float64 `json:"closes"`
	Volumes []float64 `json:"volumes"`
}

type PriceDigest struct {
	Symbol          string   `json:"symbol"`
	Start           string   `json:"start"`
	End             string   `json:"end"`
	NDays           int      `json:"n_days"`
	FirstClose      float64  `json:"first_close"`
	LastClose       float64  `json:"last_close"`
	ChangePct       *float64 `json:"change_pct"`
	HighClose       float64  `json:"high_close"`
	LowClose        float64  `json:"low_close"`
	MaxDrawdownPct  *float64 `json:"max_drawdown_pct"`
	VolatilityDaily *float64 `json:"volatility_daily_pct"`
	AvgVolume       float64  `json:"avg_volume"`
	LastVolume      float64  `json:"last_volume"`
}

func ParsePriceSeries(raw json.RawMessage) (PriceSeries, error) {
	var rows []struct {
		Symbol  string   `json:"symbol"`
		Date    string   `json:"date"`
		Close   *float64 `json:"close"`
		Volume  *float64 `json:"volume"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return PriceSeries{}, fmt.Errorf("compute: unexpected price payload: %w", err)
	}
	s := PriceSeries{}
	for _, r := range rows {
		if r.Close == nil {
			continue
		}
		if s.Symbol == "" {
			s.Symbol = r.Symbol
		}
		s.Dates = append(s.Dates, r.Date)
		s.Closes = append(s.Closes, *r.Close)
		if r.Volume != nil {
			s.Volumes = append(s.Volumes, *r.Volume)
		}
	}
	if len(s.Closes) < 2 {
		return s, fmt.Errorf("compute: need at least 2 closes, got %d", len(s.Closes))
	}
	return s, nil
}

func DigestPrice(series PriceSeries) PriceDigest {
	d := PriceDigest{
		Symbol:     series.Symbol,
		Start:      series.Dates[0],
		End:        series.Dates[len(series.Dates)-1],
		NDays:      len(series.Dates),
		FirstClose: series.Closes[0],
		LastClose:  series.Closes[len(series.Closes)-1],
	}
	d.HighClose = series.Closes[0]
	d.LowClose = series.Closes[0]
	for _, c := range series.Closes {
		if c > d.HighClose {
			d.HighClose = c
		}
		if c < d.LowClose {
			d.LowClose = c
		}
	}
	if d.FirstClose != 0 {
		v := (d.LastClose - d.FirstClose) / d.FirstClose
		d.ChangePct = &v
	}
	peak := series.Closes[0]
	maxDD := 0.0
	for _, c := range series.Closes {
		if c > peak {
			peak = c
		}
		if peak > 0 {
			dd := (peak - c) / peak
			if dd > maxDD {
				maxDD = dd
			}
		}
	}
	if maxDD > 0 {
		d.MaxDrawdownPct = &maxDD
	}
	if stdev := returnsStdev(series.Closes); stdev > 0 {
		d.VolatilityDaily = &stdev
	}
	if len(series.Volumes) > 0 {
		sum := 0.0
		for _, v := range series.Volumes {
			sum += v
		}
		d.AvgVolume = sum / float64(len(series.Volumes))
		d.LastVolume = series.Volumes[len(series.Volumes)-1]
	}
	return d
}

func returnsStdev(closes []float64) float64 {
	if len(closes) < 3 {
		return 0
	}
	returns := make([]float64, 0, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		if closes[i-1] != 0 {
			returns = append(returns, (closes[i]-closes[i-1])/closes[i-1])
		}
	}
	if len(returns) < 2 {
		return 0
	}
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))
	variance := 0.0
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	variance /= float64(len(returns) - 1)
	return math.Sqrt(variance)
}

type FlowDay struct {
	Date    string   `json:"date"`
	NetFlow *float64 `json:"net_foreign_inflow"`
}

type FlowDigest struct {
	Symbol        string   `json:"symbol"`
	Start         string   `json:"start"`
	End           string   `json:"end"`
	NDays         int      `json:"n_days"`
	NetTotal      float64  `json:"net_total_idr"`
	NetLast10d    *float64 `json:"net_last_10d_idr"`
	BuyDays       int      `json:"net_buy_days"`
	SellDays      int      `json:"net_sell_days"`
	MaxInflowDay  *FlowDay `json:"max_inflow_day,omitempty"`
	MaxOutflowDay *FlowDay `json:"max_outflow_day,omitempty"`
}

func DigestFlow(raw json.RawMessage) (FlowDigest, error) {
	var payload struct {
		Symbol string    `json:"symbol"`
		Start  string    `json:"start"`
		End    string    `json:"end"`
		Data   []FlowDay `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return FlowDigest{}, fmt.Errorf("compute: unexpected flow payload: %w", err)
	}
	if len(payload.Data) == 0 {
		return FlowDigest{}, fmt.Errorf("compute: flow payload has no data rows")
	}
	days := make([]FlowDay, 0, len(payload.Data))
	for _, day := range payload.Data {
		if day.NetFlow != nil {
			days = append(days, day)
		}
	}
	if len(days) == 0 {
		return FlowDigest{}, fmt.Errorf("compute: flow payload has no usable rows")
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Date < days[j].Date })
	d := FlowDigest{
		Symbol: payload.Symbol,
		Start:  days[0].Date,
		End:    days[len(days)-1].Date,
		NDays:  len(days),
	}
	var maxIn, maxOut FlowDay
	var haveIn, haveOut bool
	for _, day := range days {
		flow := *day.NetFlow
		d.NetTotal += flow
		if flow > 0 {
			d.BuyDays++
			if !haveIn || flow > *maxIn.NetFlow {
				maxIn, haveIn = day, true
			}
		} else if flow < 0 {
			d.SellDays++
			if !haveOut || flow < *maxOut.NetFlow {
				maxOut, haveOut = day, true
			}
		}
	}
	tail := 10
	if len(days) < tail {
		tail = len(days)
	}
	tailSum := 0.0
	for _, day := range days[len(days)-tail:] {
		tailSum += *day.NetFlow
	}
	d.NetLast10d = &tailSum
	if haveIn {
		d.MaxInflowDay = &maxIn
	}
	if haveOut {
		d.MaxOutflowDay = &maxOut
	}
	return d, nil
}
