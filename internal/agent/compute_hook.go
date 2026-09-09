package agent

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/autokeren/kerenscope/internal/compute"
	"github.com/autokeren/kerenscope/internal/sectors"
	"github.com/autokeren/kerenscope/internal/tools"
	"github.com/autokeren/kerenscope/internal/verify"
)

type computeState struct {
	mu           sync.Mutex
	companies    map[string]*compute.CompanyMetrics
	facts        []verify.Fact
	lastCompared int
}

func newComputeState() *computeState {
	return &computeState{companies: map[string]*compute.CompanyMetrics{}}
}

func (s *computeState) company(symbol string) *compute.CompanyMetrics {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.companies[symbol]; ok {
		return c
	}
	c := &compute.CompanyMetrics{Symbol: symbol}
	s.companies[symbol] = c
	return c
}

func (s *computeState) registerFacts(prefix string, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	facts := verify.CollectFacts(string(data))
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range facts {
		s.facts = append(s.facts, verify.Fact{Value: f.Value, Path: prefix + f.Path})
	}
}

func (s *computeState) comparison() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var ms []compute.CompanyMetrics
	for _, c := range s.companies {
		if c.Quarterly != nil || c.Valuation != nil {
			ms = append(ms, *c)
		}
	}
	if len(ms) < 2 || len(ms) <= s.lastCompared {
		return ""
	}
	s.lastCompared = len(ms)
	cmp := compute.CompareCompanies(ms)
	s.registerFacts("comparison", cmp)
	out, err := json.MarshalIndent(cmp, "", " ")
	if err != nil {
		return ""
	}
	return string(out)
}

func symbolFromArgs(args map[string]any) string {
	if v, ok := args["symbol"].(string); ok {
		return strings.ToUpper(strings.TrimSpace(v))
	}
	return "?"
}

func (a *Agent) postProcess(name string, args map[string]any, result tools.Result, state *computeState) string {
	if !result.OK {
		return ""
	}
	var block strings.Builder
	switch name {
	case "company_report":
		if report, ok := result.Data.(*sectors.CompanyReport); ok {
			vm := compute.FromCompanyReport(report)
			c := state.company(report.Symbol)
			c.Name = report.CompanyName
			c.Valuation = &vm
			state.registerFacts("computed:"+c.Symbol+".valuation", vm)
			block.WriteString("\n\n[COMPUTED VALUATION METRICS — deterministic, computed by the KerenScope engine. These are verified values: use them verbatim in your analysis and never recompute arithmetic yourself]:\n")
			block.WriteString(mustJSONIndent(vm))
		}
	case "quarterly_financials":
		if raw, err := json.Marshal(result.Data); err == nil {
			if quarters, err := compute.ParseQuarters(raw); err == nil && len(quarters) > 0 {
				qm := compute.MetricsFromQuarters(symbolFromArgs(args), quarters)
				c := state.company(symbolFromArgs(args))
				c.Quarterly = &qm
				state.registerFacts("computed:"+c.Symbol+".quarterly", qm)
				block.WriteString("\n\n[COMPUTED QUARTERLY METRICS — deterministic, computed by the KerenScope engine. These are verified values: use them verbatim in your analysis and never recompute arithmetic yourself]:\n")
				block.WriteString(mustJSONIndent(qm))
			}
		}
	}
	if block.Len() > 0 {
		if cmp := state.comparison(); cmp != "" {
			block.WriteString("\n\n[COMPARISON — deterministic cross-company scoring by the KerenScope engine (see rubric)]:\n")
			block.WriteString(cmp)
		}
	}
	return block.String()
}

func mustJSONIndent(v any) string {
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
