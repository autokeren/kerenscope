package tools

import (
	"context"
	"encoding/json"

	"github.com/autokeren/kerenscope/internal/sectors"
)

func object(props map[string]any, required ...string) map[string]any {
	return map[string]any{"type": "object", "properties": props, "required": toAny(required)}
}

func toAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func strProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func intProp(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

type CompanyReportTool struct{ Client *sectors.Client }

func (t CompanyReportTool) Definition() Definition {
	return Definition{
		Name: "company_report",
		Description: "Full company report from Sectors: business overview, market cap and rank, valuation multiples (P/B, P/E, PEG vs peer averages), forward P/E, analyst consensus and growth forecasts. Use this first when investigating a specific ticker.",
		Parameters: object(map[string]any{
			"symbol": strProp("IDX ticker, e.g. BBCA"),
		}, "symbol"),
	}
}

func (t CompanyReportTool) Run(ctx context.Context, args map[string]any) Result {
	report, err := t.Client.CompanyReport(ctx, argString(args, "symbol"))
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: report}
}

type QuarterlyFinancialsTool struct{ Client *sectors.Client }

func (t QuarterlyFinancialsTool) Definition() Definition {
	return Definition{
		Name: "quarterly_financials",
		Description: "Quarterly financial statements for a company: revenue, earnings, and sector-specific metrics (banks: NII, gross loans, deposits). Use for trend analysis across recent quarters.",
		Parameters: object(map[string]any{
			"symbol":     strProp("IDX ticker, e.g. BBRI"),
			"n_quarters": intProp("Number of recent quarters to fetch (default 8)"),
		}, "symbol"),
	}
}

func (t QuarterlyFinancialsTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.QuarterlyFinancials(ctx, argString(args, "symbol"), argInt(args, "n_quarters", 8))
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type ScreenCompaniesTool struct{ Client *sectors.Client }

func (t ScreenCompaniesTool) Definition() Definition {
	return Definition{
		Name: "screen_companies",
		Description: "Screen/filter IDX-listed companies with a SQL-like `where` clause. ONLY documented field names are valid. Common fields — identification: symbol, company_name, sector, sub_sector, industry, listing_board, indices, tags; size/price: market_cap, market_cap_rank, last_close_price, all_time_high_price, ytd_high_price; valuation: pe, pe_ttm, forward_pe, pb, ps, pcf, peg, eps, eps_growth, dividend_yield_avg, intrinsic_value, pe_peer_avg, pb_peer_avg; profitability: roe, roe_ttm, roa, roa_ttm, gross_profit_margin, net_profit_margin, operating_profit_margin, cost_to_income_ratio; yearly financials (bracket notation): revenue[2024], earnings[2024], total_assets[2024], total_debt[2024], operating_cash_flow[2024]; growth: yoy_quarter_earnings_growth, yoy_quarter_revenue_growth, forecast_eps_growth, yearly_mcap_change; bank-specific: net_interest_margin, capital_adequacy_ratio, non_performing_loan, loan_to_deposit_ratio, casa_ratio, gross_loan, total_deposit; balance/liquidity: debt_to_equity_ratio, current_ratio, free_cash_flow. Operators: = != > >= < <= like in, combined with and/or; string values in single quotes. Examples: \"sector = 'Financials' and pe_ttm < 15\", \"roe[2024] > 0.15 and roe[2023] > 0.15\", \"indices in ['LQ45']\". `order_by` sorts (- prefix = descending), e.g. -market_cap or -(earnings[2024]/earnings[2023]). Alternatively pass a natural language `q` (LLM-translated server-side). Limit results (max 200, default 15).",
		Parameters: object(map[string]any{
			"where":    strProp("SQL-like filter conditions, e.g. \"sector = 'banks' and pe[2025] < 15\""),
			"q":        strProp("Natural language query alternative, e.g. \"top 10 tech companies by revenue in 2024\""),
			"order_by": strProp("Field to sort by, - prefix for descending, e.g. -revenue_growth[2024]"),
			"limit":    intProp("Maximum number of results (default 15)"),
		}),
	}
}

func (t ScreenCompaniesTool) Run(ctx context.Context, args map[string]any) Result {
	where := argString(args, "where")
	if where != "" {
		if err := ValidateScreenerWhere(where); err != nil {
			return Result{OK: false, Error: err.Error()}
		}
	}
	orderBy := argString(args, "order_by")
	if orderBy != "" {
		if err := ValidateScreenerOrderBy(orderBy); err != nil {
			return Result{OK: false, Error: err.Error()}
		}
	}
	data, err := t.Client.ScreenCompanies(ctx, sectors.ScreenOptions{
		Where:   where,
		Q:       argString(args, "q"),
		OrderBy: orderBy,
		Limit:   argInt(args, "limit", 15),
	})
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type SubsectorReportTool struct{ Client *sectors.Client }

func (t SubsectorReportTool) Definition() Definition {
	return Definition{
		Name: "subsector_report",
		Description: "Comprehensive report for an IDX subsector: aggregate financials, top companies by market cap, valuation and performance comparisons. Use kebab-case slugs like banks, food-beverage, telecommunication. Use this for peer/sector context.",
		Parameters: object(map[string]any{
			"subsector": strProp("Subsector slug, e.g. banks"),
		}, "subsector"),
	}
}

func (t SubsectorReportTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.SubsectorReport(ctx, argString(args, "subsector"))
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type SubsectorsTool struct{ Client *sectors.Client }

func (t SubsectorsTool) Definition() Definition {
	return Definition{
		Name: "list_subsectors",
		Description: "List all sector/subsector slug pairs (kebab-case). Call this when you need the correct subsector slug before calling subsector_report or screening by sector.",
		Parameters: object(map[string]any{}),
	}
}

func (t SubsectorsTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.Subsectors(ctx)
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type PriceHistoryTool struct{ Client *sectors.Client }

func (t PriceHistoryTool) Definition() Definition {
	return Definition{
		Name: "price_history",
		Description: "Daily close price, volume and market cap history for a ticker over up to 90 days. Use for momentum, drawdown and recent performance analysis.",
		Parameters: object(map[string]any{
			"symbol": strProp("IDX ticker, e.g. BBCA"),
			"days":   intProp("Lookback window in days, max 90 (default 30)"),
		}, "symbol"),
	}
}

func (t PriceHistoryTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.DailyPrices(ctx, argString(args, "symbol"), argInt(args, "days", 30))
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type ForeignFlowTool struct{ Client *sectors.Client }

func (t ForeignFlowTool) Definition() Definition {
	return Definition{
		Name: "foreign_flow",
		Description: "Daily net foreign-broker inflow (IDR) for a ticker: positive means foreign brokers net bought, negative means net sold. Use to gauge foreign sentiment and capital flow.",
		Parameters: object(map[string]any{
			"symbol": strProp("IDX ticker, e.g. BBCA"),
			"days":   intProp("Lookback window in days, max 90 (default 30)"),
		}, "symbol"),
	}
}

func (t ForeignFlowTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.ForeignFlow(ctx, argString(args, "symbol"), argInt(args, "days", 30))
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type BrokerTopTool struct{ Client *sectors.Client }

func (t BrokerTopTool) Definition() Definition {
	return Definition{
		Name: "broker_summary",
		Description: "Brokers most actively accumulating (net buying) and distributing (net selling) a ticker. Signals institutional interest. Optional origin filter: foreign or domestic.",
		Parameters: object(map[string]any{
			"symbol":    strProp("IDX ticker, e.g. BBCA"),
			"days":      intProp("Lookback window in days, max 90 (default 30)"),
			"n_brokers": intProp("Number of brokers per side (default 5)"),
			"origin":    strProp("Optional: foreign or domestic"),
		}, "symbol"),
	}
}

func (t BrokerTopTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.BrokerTop(ctx, argString(args, "symbol"), sectors.BrokerOptions{
		Days:     argInt(args, "days", 30),
		NBrokers: argInt(args, "n_brokers", 5),
		Origin:   argString(args, "origin"),
	})
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type InsiderFilingsTool struct{ Client *sectors.Client }

func (t InsiderFilingsTool) Definition() Definition {
	return Definition{
		Name: "insider_filings",
		Description: "IDX insider trading filings: buy/sell transactions by insiders and major shareholders for a ticker or the whole market. Use to spot insider confidence or selling pressure.",
		Parameters: object(map[string]any{
			"symbol":           strProp("IDX ticker, e.g. BBCA"),
			"transaction_type": strProp("Optional filter: buy or sell"),
			"days":             intProp("Lookback window in days (default 90)"),
			"limit":            intProp("Maximum number of filings (default 20)"),
		}, "symbol"),
	}
}

func (t InsiderFilingsTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.Filings(ctx, sectors.FilingsOptions{
		Symbol:          argString(args, "symbol"),
		TransactionType: argString(args, "transaction_type"),
		Days:            argInt(args, "days", 90),
		Limit:           argInt(args, "limit", 20),
	})
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type NewsTool struct{ Client *sectors.Client }

func (t NewsTool) Definition() Definition {
	return Definition{
		Name: "news",
		Description: "IDX news articles, optionally filtered by ticker(s) or keyword. Use for recent events and catalysts affecting a company or market.",
		Parameters: object(map[string]any{
			"symbol":  strProp("IDX ticker, e.g. BBCA"),
			"keyword": strProp("Optional keyword filter"),
			"days":    intProp("Lookback window in days (default 30)"),
			"limit":   intProp("Maximum number of articles (default 10)"),
		}, "symbol"),
	}
}

func (t NewsTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.News(ctx, sectors.NewsOptions{
		Symbol:  argString(args, "symbol"),
		Keyword: argString(args, "keyword"),
		Days:    argInt(args, "days", 30),
		Limit:   argInt(args, "limit", 10),
	})
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type TopMoversTool struct{ Client *sectors.Client }

func (t TopMoversTool) Definition() Definition {
	return Definition{
		Name: "top_movers",
		Description: "Top gainers or losers across periods. classification: top_gainers or top_losers. period: 1d, 7d, 14d, 30d, or 365d.",
		Parameters: object(map[string]any{
			"classification": strProp("top_gainers or top_losers"),
			"period":         strProp("1d, 7d, 14d, 30d, or 365d"),
			"n_stock":        intProp("Number of results (default 10)"),
		}),
	}
}

func (t TopMoversTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.TopMovers(ctx, sectors.MoversOptions{
		Classification: argString(args, "classification"),
		Period:         argString(args, "period"),
		NStock:         argInt(args, "n_stock", 10),
	})
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

type IndexDailyTool struct{ Client *sectors.Client }

func (t IndexDailyTool) Definition() Definition {
	return Definition{
		Name: "index_history",
		Description: "Daily history of an IDX index (default IHSG) over up to 90 days. Use for market context.",
		Parameters: object(map[string]any{
			"index": strProp("Index code, default IHSG"),
			"days": intProp("Lookback window in days, max 90 (default 30)"),
		}),
	}
}

func (t IndexDailyTool) Run(ctx context.Context, args map[string]any) Result {
	data, err := t.Client.IndexDaily(ctx, argString(args, "index"), argInt(args, "days", 30))
	if err != nil {
		return Result{OK: false, Error: err.Error()}
	}
	return Result{OK: true, Data: json.RawMessage(data)}
}

func DefaultRegistry(client *sectors.Client) *Registry {
	return NewRegistry().
		Register(CompanyReportTool{Client: client}).
		Register(QuarterlyFinancialsTool{Client: client}).
		Register(ScreenCompaniesTool{Client: client}).
		Register(SubsectorReportTool{Client: client}).
		Register(SubsectorsTool{Client: client}).
		Register(PriceHistoryTool{Client: client}).
		Register(ForeignFlowTool{Client: client}).
		Register(BrokerTopTool{Client: client}).
		Register(InsiderFilingsTool{Client: client}).
		Register(NewsTool{Client: client}).
		Register(TopMoversTool{Client: client}).
		Register(IndexDailyTool{Client: client}).
		Register(SectorsAPITool{Client: client})
}
