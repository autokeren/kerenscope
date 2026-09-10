package tools

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestValidateScreenerWhere(t *testing.T) {
	valid := []string{
		"sector = 'Financials' and pe_ttm < 15",
		"roe[2024] > 0.15 and roe[2023] > 0.15",
		"indices in ['LQ45', 'IDX30']",
		"company_name like '%energi%'",
		"revenue[2023] > earnings[2023] * 5",
		"market_cap > 500000000000000 and not symbol = 'BBRI'",
	}
	for _, w := range valid {
		if err := ValidateScreenerWhere(w); err != nil {
			t.Fatalf("expected %q valid, got: %v", w, err)
		}
	}
	invalid := []string{
		"revenue_growth[2024] > 0.1",
		"roe_ttm > 0.1 and pe_ratio < 15",
		"growth_percent > 5",
	}
	for _, w := range invalid {
		err := ValidateScreenerWhere(w)
		if err == nil {
			t.Fatalf("expected %q rejected", w)
		}
		if !strings.Contains(err.Error(), "invalid field") || !strings.Contains(err.Error(), "common ones") {
			t.Fatalf("expected helpful error for %q, got: %v", w, err)
		}
	}
}

func TestValidateScreenerOrderBy(t *testing.T) {
	for _, f := range []string{"market_cap", "-market_cap", "-(earnings[2024]/earnings[2023])", "pe_ttm"} {
		if err := ValidateScreenerOrderBy(f); err != nil {
			t.Fatalf("expected %q valid, got %v", f, err)
		}
	}
	if err := ValidateScreenerOrderBy("price_earnings"); err == nil {
		t.Fatal("expected unknown order_by field rejected")
	}
}

func TestAutoBracketFix(t *testing.T) {
	apiErr := `sectors: HTTP 400: Field 'debt_to_equity_ratio' requires bracket notation with a year. Example: debt_to_equity_ratio[2024]`
	fixed, ok := AutoBracketFix("debt_to_equity_ratio < 1 and pe_ttm < 15", apiErr)
	if !ok {
		t.Fatal("expected repair")
	}
	if fixed != "debt_to_equity_ratio[2024] < 1 and pe_ttm < 15" {
		t.Fatalf("unexpected repair result: %s", fixed)
	}
	_, ok = AutoBracketFix("pe_ttm < 15", apiErr)
	if ok {
		t.Fatal("no repair expected when the field is absent")
	}
}

func TestScreenCompaniesToolSelfHealsBracketErrors(t *testing.T) {
	calls := 0
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		where := r.URL.Query().Get("where")
		if !strings.Contains(where, "[") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"INVALID_WHERE_CLAUSE","message":"Field 'forecast_eps_growth' requires bracket notation with a year. Example: forecast_eps_growth[2025]"}`))
			return
		}
		w.Write([]byte(`{"results":[{"symbol":"BRIS.JK"}]}`))
	})
	tool := ScreenCompaniesTool{Client: client}
	res := tool.Run(context.Background(), map[string]any{
		"where": "forecast_eps_growth > 0.1 and pe_ttm < 15",
	})
	if !res.OK {
		t.Fatalf("expected self-healed success, got: %s", res.Error)
	}
	if calls != 2 {
		t.Fatalf("expected exactly 2 calls (1 rejection + 1 repaired), got %d", calls)
	}
}
