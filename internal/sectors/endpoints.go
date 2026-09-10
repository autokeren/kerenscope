package sectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *Client) QuarterlyFinancials(ctx context.Context, symbol string, nQuarters int) (json.RawMessage, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return nil, errEmptySymbol
	}
	params := url.Values{}
	if nQuarters > 0 {
		params.Set("n_quarters", strconv.Itoa(nQuarters))
	}
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/financials/quarterly/"+symbol+"/", params, 24*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SubsectorReport(ctx context.Context, subSector string) (json.RawMessage, error) {
	subSector = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(subSector, " ", "-")))
	if subSector == "" {
		return nil, fmt.Errorf("subsector is empty")
	}
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/subsector/report/"+subSector+"/", nil, 24*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Subsectors(ctx context.Context) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/subsectors/", nil, 7*24*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type ScreenOptions struct {
	Where   string
	Q       string
	OrderBy string
	Desc    bool
	Limit   int
}

func (c *Client) ScreenCompanies(ctx context.Context, opts ScreenOptions) (json.RawMessage, error) {
	params := url.Values{}
	if opts.Where != "" {
		params.Set("where", opts.Where)
	}
	if opts.Q != "" {
		params.Set("q", opts.Q)
	}
	if opts.OrderBy != "" {
		if opts.Desc && !strings.HasPrefix(opts.OrderBy, "-") {
			opts.OrderBy = "-" + opts.OrderBy
		}
		params.Set("order_by", opts.OrderBy)
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/companies/", params, 12*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DailyPrices(ctx context.Context, symbol string, days int) (json.RawMessage, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return nil, errEmptySymbol
	}
	start, end := dateRange(days)
	params := url.Values{}
	params.Set("start", start)
	params.Set("end", end)
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/daily/"+symbol+"/", params, 12*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ForeignFlow(ctx context.Context, symbol string, days int) (json.RawMessage, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return nil, errEmptySymbol
	}
	start, end := dateRange(days)
	params := url.Values{}
	params.Set("start", start)
	params.Set("end", end)
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/foreign-flow/"+symbol+"/", params, 12*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type BrokerOptions struct {
	Days      int
	NBrokers  int
	Origin    string
	Cohort    string
}

func (c *Client) BrokerTop(ctx context.Context, symbol string, opts BrokerOptions) (json.RawMessage, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return nil, errEmptySymbol
	}
	start, end := dateRange(opts.Days)
	params := url.Values{}
	params.Set("start", start)
	params.Set("end", end)
	if opts.NBrokers > 0 {
		params.Set("n_brokers", strconv.Itoa(opts.NBrokers))
	}
	if opts.Origin != "" {
		params.Set("origin", opts.Origin)
	}
	if opts.Cohort != "" {
		params.Set("cohort", opts.Cohort)
	}
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/broker-summary/"+symbol+"/top/", params, 12*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type FilingsOptions struct {
	Symbol          string
	TransactionType string
	Days            int
	Limit           int
}

func (c *Client) Filings(ctx context.Context, opts FilingsOptions) (json.RawMessage, error) {
	params := url.Values{}
	if opts.Symbol != "" {
		params.Set("symbol", normalizeSymbol(opts.Symbol))
	}
	if opts.TransactionType != "" {
		params.Set("transaction_type", opts.TransactionType)
	}
	if opts.Days > 0 {
		start, end := dateRange(opts.Days)
		params.Set("start", start)
		params.Set("end", end)
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/filings/", params, 6*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type NewsOptions struct {
	Symbol  string
	Keyword string
	Days    int
	Limit   int
}

func (c *Client) News(ctx context.Context, opts NewsOptions) (json.RawMessage, error) {
	params := url.Values{}
	if opts.Symbol != "" {
		params.Set("symbols", normalizeSymbol(opts.Symbol))
	}
	if opts.Keyword != "" {
		params.Set("keyword", opts.Keyword)
	}
	if opts.Days > 0 {
		start, end := dateRange(opts.Days)
		params.Set("start", start)
		params.Set("end", end)
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/news/", params, 3*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type MoversOptions struct {
	Classification string
	Period         string
	NStock         int
	MinMCapBillion float64
}

func (c *Client) TopMovers(ctx context.Context, opts MoversOptions) (json.RawMessage, error) {
	params := url.Values{}
	if opts.Classification != "" {
		params.Set("classifications", opts.Classification)
	}
	if opts.Period != "" {
		params.Set("periods", opts.Period)
	}
	if opts.NStock > 0 {
		params.Set("n_stock", strconv.Itoa(opts.NStock))
	}
	if opts.MinMCapBillion > 0 {
		params.Set("min_mcap_billion", strconv.FormatFloat(opts.MinMCapBillion, 'f', -1, 64))
	}
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/companies/top-changes/", params, 6*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) IndexDaily(ctx context.Context, indexCode string, days int) (json.RawMessage, error) {
	indexCode = strings.ToLower(strings.TrimSpace(indexCode))
	if indexCode == "" {
		indexCode = "IHSG"
	}
	start, end := dateRange(days)
	params := url.Values{}
	params.Set("start", start)
	params.Set("end", end)
	var out json.RawMessage
	if err := c.GetJSON(ctx, "/v2/index-daily/"+indexCode+"/", params, 12*time.Hour, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizeSymbol(symbol string) string {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	symbol = strings.TrimSuffix(symbol, ".JK")
	return symbol
}

func dateRange(days int) (string, string) {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -days)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}
