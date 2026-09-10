package tools

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	quotedRe    = regexp.MustCompile(`'[^']*'`)
	bracketRe   = regexp.MustCompile(`\[[^\]]*\]`)
	tokenRe     = regexp.MustCompile(`[a-z_]+(\[[0-9]{4}\])?`)
	screenerKw  = map[string]bool{"and": true, "or": true, "not": true, "like": true, "in": true, "is": true, "null": true, "true": true, "false": true, "between": true, "all": true}
)

func screenerFieldBase(token string) string {
	if idx := strings.Index(token, "["); idx > 0 {
		return token[:idx]
	}
	return token
}

func ValidateScreenerWhere(where string) error {
	stripped := bracketRe.ReplaceAllString(quotedRe.ReplaceAllString(where, " "), " ")
	var unknown []string
	for _, token := range tokenRe.FindAllString(strings.ToLower(stripped), -1) {
		base := screenerFieldBase(token)
		if screenerKw[base] {
			continue
		}
		if screenerFields[base] {
			continue
		}
		unknown = append(unknown, token)
	}
	if len(unknown) > 0 {
		return fmt.Errorf("invalid field(s) in where clause: %s — use documented screener fields (common ones: symbol, sector, market_cap, pe_ttm, pb, roe_ttm, roa_ttm, revenue[YYYY], earnings[YYYY], eps_growth, dividend_yield_avg, forward_pe, yoy_quarter_earnings_growth, net_interest_margin, non_performing_loan, indices, tags)", strings.Join(unknown, ", "))
	}
	return nil
}

func ValidateScreenerOrderBy(field string) error {
	name := strings.TrimPrefix(strings.TrimSpace(strings.ToLower(field)), "-")
	name = strings.TrimSuffix(name, ")")
	if name == "" {
		return nil
	}
	// allow simple arithmetic like earnings[2024]/earnings[2023]
	for _, token := range tokenRe.FindAllString(name, -1) {
		base := screenerFieldBase(token)
		if screenerKw[base] {
			continue
		}
		if !screenerFields[base] {
			return fmt.Errorf("invalid order_by field %q — use documented screener fields", name)
		}
	}
	return nil
}

var bracketFixRe = regexp.MustCompile(`Field '([a-z_]+)' requires bracket notation with a year\. Example: [a-z_]+\[(\d{4})\]`)

func AutoBracketFix(clause string, apiError string) (string, bool) {
	m := bracketFixRe.FindStringSubmatch(apiError)
	if m == nil {
		return clause, false
	}
	field, year := m[1], m[2]
	fixed := tokenRe.ReplaceAllStringFunc(clause, func(tok string) string {
		base := screenerFieldBase(tok)
		if base == field && !strings.Contains(tok, "[") {
			return field + "[" + year + "]"
		}
		return tok
	})
	if fixed == clause {
		return clause, false
	}
	return fixed, true
}
