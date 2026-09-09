package compute

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Quarter struct {
	Date                string  `json:"date"`
	Earnings            float64 `json:"earnings"`
	Revenue             float64 `json:"revenue"`
	TotalAssets         float64 `json:"total_assets"`
	TotalEquity         float64 `json:"total_equity"`
	OperatingExpense    float64 `json:"operating_expense"`
	Provision           float64 `json:"provision"`
	NonInterestIncome   float64 `json:"non_interest_income"`
	NII                 float64 `json:"net_interest_income"`
	InterestIncome      float64 `json:"interest_income"`
	InterestExpense     float64 `json:"interest_expense"`
	GrossLoan           float64 `json:"gross_loan"`
	NetLoan             float64 `json:"net_loan"`
	TotalDeposit        float64 `json:"total_deposit"`
	CurrentAccount      float64 `json:"current_account"`
	SavingsAccount      float64 `json:"savings_account"`
	TimeDeposit         float64 `json:"time_deposit"`
	IsBank              bool    `json:"-"`
}

type QuarterMetrics struct {
	Symbol           string            `json:"symbol"`
	AsOf             string            `json:"as_of"`
	NQuarters        int               `json:"n_quarters"`
	ROE              *float64          `json:"roe,omitempty"`
	ROA              *float64          `json:"roa,omitempty"`
	NIM              *float64          `json:"nim,omitempty"`
	CASA             *float64          `json:"casa,omitempty"`
	LDR              *float64          `json:"ldr,omitempty"`
	CostToIncome     *float64          `json:"cost_to_income,omitempty"`
	ProvisionRatio   *float64          `json:"provision_ratio,omitempty"`
	NetMargin        *float64          `json:"net_margin,omitempty"`
	Earnings         *float64          `json:"earnings_latest,omitempty"`
	Revenue          *float64          `json:"revenue_latest,omitempty"`
	EarningsQoQ      *float64          `json:"earnings_qoq,omitempty"`
	EarningsYoY      *float64          `json:"earnings_yoy,omitempty"`
	LoanGrowth       *float64          `json:"loan_growth,omitempty"`
	DepositGrowth    *float64          `json:"deposit_growth,omitempty"`
	Notes            []string          `json:"notes,omitempty"`
}

func ParseQuarters(raw json.RawMessage) ([]Quarter, error) {
	var records []struct {
		Date                      string     `json:"date"`
		Earnings                  *float64   `json:"earnings"`
		Revenue                   *float64   `json:"revenue"`
		TotalAssets               *float64   `json:"total_assets"`
		TotalEquity               *float64   `json:"total_equity"`
		OperatingExpense          *float64   `json:"operating_expense"`
		Provision                 *float64   `json:"provision"`
		NonInterestIncome         *float64   `json:"non_interest_income"`
		FinancialsSectorMetrics   *struct {
			NetInterestIncome  *float64 `json:"net_interest_income"`
			InterestIncome     *float64 `json:"interest_income"`
			InterestExpense    *float64 `json:"interest_expense"`
			GrossLoan          *float64 `json:"gross_loan"`
			NetLoan            *float64 `json:"net_loan"`
			TotalDeposit       *float64 `json:"total_deposit"`
			CurrentAccount     *float64 `json:"current_account"`
			SavingsAccount     *float64 `json:"savings_account"`
			TimeDeposit        *float64 `json:"time_deposit"`
		} `json:"financials_sector_metrics"`
	}
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, fmt.Errorf("compute: unexpected quarterly payload: %w", err)
	}
	quarters := make([]Quarter, 0, len(records))
	for _, r := range records {
		q := Quarter{Date: r.Date}
		set := func(dst *float64, src *float64) {
			if src != nil {
				*dst = *src
			}
		}
		set(&q.Earnings, r.Earnings)
		set(&q.Revenue, r.Revenue)
		set(&q.TotalAssets, r.TotalAssets)
		set(&q.TotalEquity, r.TotalEquity)
		set(&q.OperatingExpense, r.OperatingExpense)
		set(&q.Provision, r.Provision)
		set(&q.NonInterestIncome, r.NonInterestIncome)
		if r.FinancialsSectorMetrics != nil {
			q.IsBank = true
			set(&q.NII, r.FinancialsSectorMetrics.NetInterestIncome)
			set(&q.InterestIncome, r.FinancialsSectorMetrics.InterestIncome)
			set(&q.InterestExpense, r.FinancialsSectorMetrics.InterestExpense)
			set(&q.GrossLoan, r.FinancialsSectorMetrics.GrossLoan)
			set(&q.NetLoan, r.FinancialsSectorMetrics.NetLoan)
			set(&q.TotalDeposit, r.FinancialsSectorMetrics.TotalDeposit)
			set(&q.CurrentAccount, r.FinancialsSectorMetrics.CurrentAccount)
			set(&q.SavingsAccount, r.FinancialsSectorMetrics.SavingsAccount)
			set(&q.TimeDeposit, r.FinancialsSectorMetrics.TimeDeposit)
		}
		quarters = append(quarters, q)
	}
	sort.Slice(quarters, func(i, j int) bool { return quarters[i].Date < quarters[j].Date })
	return quarters, nil
}

func MetricsFromQuarters(symbol string, quarters []Quarter) QuarterMetrics {
	m := QuarterMetrics{Symbol: symbol, NQuarters: len(quarters)}
	if len(quarters) == 0 {
		return m
	}
	latest := quarters[len(quarters)-1]
	m.AsOf = latest.Date
	if val, ok := ratio(latest.Earnings*4, latest.TotalEquity); ok {
		m.ROE = val
	}
	if val, ok := ratio(latest.Earnings*4, latest.TotalAssets); ok {
		m.ROA = val
	}
	if val, ok := ratio(latest.NII*4, latest.GrossLoan); ok {
		m.NIM = val
	}
	if val, ok := ratio(latest.CurrentAccount+latest.SavingsAccount, latest.TotalDeposit); ok {
		m.CASA = val
	}
	if val, ok := ratio(latest.NetLoan, latest.TotalDeposit); ok {
		m.LDR = val
	}
	if val, ok := ratio(latest.OperatingExpense, latest.Revenue); ok {
		m.CostToIncome = val
	}
	if val, ok := ratio(latest.Provision, latest.Revenue); ok {
		m.ProvisionRatio = val
	}
	if val, ok := ratio(latest.Earnings, latest.Revenue); ok {
		m.NetMargin = val
	}
	m.Earnings = nz(latest.Earnings)
	m.Revenue = nz(latest.Revenue)
	if len(quarters) >= 2 {
		prev := quarters[len(quarters)-2]
		if val, ok := growth(latest.Earnings, prev.Earnings); ok {
			m.EarningsQoQ = val
		}
		if val, ok := growth(latest.NetLoan, prev.NetLoan); ok {
			m.LoanGrowth = val
		}
		if val, ok := growth(latest.TotalDeposit, prev.TotalDeposit); ok {
			m.DepositGrowth = val
		}
	}
	if len(quarters) >= 5 {
		yearAgo := quarters[len(quarters)-5]
		if val, ok := growth(latest.Earnings, yearAgo.Earnings); ok {
			m.EarningsYoY = val
		}
	}
	if len(quarters) < 4 {
		m.Notes = append(m.Notes, fmt.Sprintf("only %d quarters available; annualized metrics use latest quarter × 4", len(quarters)))
	}
	return m
}

func ratio(num, den float64) (*float64, bool) {
	if den == 0 || num == 0 {
		return nil, false
	}
	v := num / den
	return &v, true
}

func growth(now, before float64) (*float64, bool) {
	if before == 0 {
		return nil, false
	}
	v := (now - before) / before
	return &v, true
}

func nz(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}

func QuarterDate(date string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	if err != nil {
		t, err = time.Parse("2006-01", strings.TrimSpace(date))
		if err != nil {
			return time.Time{}, false
		}
	}
	return t, true
}
