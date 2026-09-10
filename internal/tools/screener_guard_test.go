package tools

import (
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
