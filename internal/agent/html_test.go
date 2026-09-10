package agent

import (
	"strings"
	"testing"
	"time"

	"github.com/autokeren/kerenscope/internal/verify"
)

func TestRenderHTMLProducesStandalonePage(t *testing.T) {
	r := &Report{
		Question: "test question?",
		Plan:     Plan{Objective: "objective", Steps: []PlanStep{{ID: 1, Intent: "intent", Tool: "company_report", Args: map[string]any{}}}},
		Draft:    "## Analysis\n\nP/E 8,23x dengan market cap Rp 513,1 triliun.\n\n| Metric | Value |\n|---|---|\n| ROE | 21% |\n",
		Verification: Verification{
			Claims:     []Claim{{Claim: "ROE tinggi", Supported: true, Evidence: "computed quarterly metrics"}},
			Confidence: 80,
			Numeric:    verify.NumericCheck{Matched: []verify.MatchedNumber{{Number: "8,23", Source: "computed"}}, Unmatched: []string{}},
		},
		StartedAt: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
	}
	html, err := r.RenderHTML()
	if err != nil {
		t.Fatal(err)
	}
	page := string(html)
	for _, want := range []string{"<table", "513,1", "<style>", "not investment advice", "80/100", "8,23"} {
		if !strings.Contains(page, want) {
			t.Fatalf("expected %q in HTML output", want)
		}
	}
	if !strings.Contains(page, "</html>") {
		t.Fatal("expected complete HTML document")
	}
}
