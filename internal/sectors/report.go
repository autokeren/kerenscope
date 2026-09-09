package sectors

import (
	"context"
	"strings"
	"time"
)

type CompanyReport struct {
	Symbol      string `json:"symbol"`
	CompanyName string `json:"company_name"`
	Overview    struct {
		ListingBoard   string  `json:"listing_board"`
		Industry       string  `json:"industry"`
		SubIndustry    string  `json:"sub_industry"`
		Sector         string  `json:"sector"`
		SubSector      string  `json:"sub_sector"`
		MarketCap      int64   `json:"market_cap"`
		MarketCapRank  int     `json:"market_cap_rank"`
		EmployeeNum    float64 `json:"employee_num"`
		ListingDate    string  `json:"listing_date"`
		Website        string  `json:"website"`
		MainShareholder string `json:"main_shareholder"`
	} `json:"overview"`
	Valuation struct {
		LastClosePrice  int     `json:"last_close_price"`
		LatestCloseDate string  `json:"latest_close_date"`
		DailyCloseChange float64 `json:"daily_close_change"`
		ForwardPE       float64 `json:"forward_pe"`
		IntrinsicValue  int     `json:"intrinsic_value"`
		HistoricalValuation []ValuationYear `json:"historical_valuation"`
	} `json:"valuation"`
	Future struct {
		CompanyValueForecasts  []ValueForecast  `json:"company_value_forecasts"`
		CompanyGrowthForecasts []GrowthForecast  `json:"company_growth_forecasts"`
		AnalystRating         RatingBreakdown   `json:"analyst_rating_breakdown"`
	} `json:"future"`
}

type ValuationYear struct {
	PB      float64 `json:"pb"`
	PE      float64 `json:"pe"`
	PS      float64 `json:"ps"`
	PCF     float64 `json:"pcf"`
	PEG     float64 `json:"peg"`
	Year    int     `json:"year"`
	PBPeer  float64 `json:"pb_peer_avg"`
	PEPeer  float64 `json:"pe_peer_avg"`
	PSPeer  float64 `json:"ps_peer_avg"`
}

type ValueForecast struct {
	EPSEstimate     float64 `json:"eps_estimate"`
	EstimateYear    int     `json:"estimate_year"`
	RevenueEstimate float64 `json:"revenue_estimate"`
}

type GrowthForecast struct {
	BaseYear       int     `json:"base_year"`
	EstimateYear   int     `json:"estimate_year"`
	EPSGrowth      float64 `json:"eps_growth"`
	RevenueGrowth  float64 `json:"revenue_growth"`
}

type RatingBreakdown struct {
	Buy      int `json:"buy"`
	Hold     int `json:"hold"`
	Sell     int `json:"sell"`
	NAnalyst int `json:"n_analyst"`
}

func (r *CompanyReport) LatestValuation() *ValuationYear {
	if len(r.Valuation.HistoricalValuation) == 0 {
		return nil
	}
	return &r.Valuation.HistoricalValuation[len(r.Valuation.HistoricalValuation)-1]
}

func (c *Client) CompanyReport(ctx context.Context, symbol string) (*CompanyReport, error) {
	symbol = strings.ToUpper(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(symbol), ".JK")))
	if symbol == "" {
		return nil, errEmptySymbol
	}
	var report CompanyReport
	if err := c.GetJSON(ctx, "/v2/company/report/"+symbol+"/", nil, 24*time.Hour, &report); err != nil {
		return nil, err
	}
	return &report, nil
}
