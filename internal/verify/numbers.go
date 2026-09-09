package verify

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

type Fact struct {
	Value float64
	Path  string
}

type NumberKind int

const (
	KindPlain NumberKind = iota
	KindPercent
	KindRatio
	KindIDR
)

type DraftNumber struct {
	Raw     string     `json:"raw"`
	Value   float64    `json:"value"`
	Kind    NumberKind `json:"kind"`
	Scale   float64    `json:"scale"`
	Context string     `json:"context"`
}

type MatchedNumber struct {
	Number string `json:"number"`
	Source string `json:"source"`
}

type NumericCheck struct {
	Matched   []MatchedNumber `json:"matched"`
	Unmatched []string       `json:"unmatched"`
	Skipped   int            `json:"skipped"`
}

var numberRe = regexp.MustCompile(`[-+]?(?:\d{1,3}(?:\.\d{3})+(?:,\d+)?|\d+(?:[.,]\d+)?)`)
var yearRe = regexp.MustCompile(`^(19|20)\d{2}$`)

func CollectFacts(payloads ...string) []Fact {
	var facts []Fact
	for _, p := range payloads {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" || trimmed[0] != '{' && trimmed[0] != '[' {
			continue
		}
		var v any
		if err := json.Unmarshal([]byte(trimmed), &v); err != nil {
			continue
		}
		walk(v, "", &facts)
	}
	return facts
}

func walk(v any, path string, facts *[]Fact) {
	switch t := v.(type) {
	case float64:
		*facts = append(*facts, Fact{Value: t, Path: path})
	case int:
		*facts = append(*facts, Fact{Value: float64(t), Path: path})
	case int64:
		*facts = append(*facts, Fact{Value: float64(t), Path: path})
	case map[string]any:
		for k, child := range t {
			walk(child, path+"."+k, facts)
		}
	case []any:
		for i, child := range t {
			walk(child, fmt.Sprintf("%s[%d]", path, i), facts)
		}
	}
}

func ExtractNumbers(text string) []DraftNumber {
	text = strings.NewReplacer("–", "-", "—", "-", "\u00a0", " ").Replace(text)
	locs := numberRe.FindAllStringIndex(text, -1)
	var out []DraftNumber
	for _, loc := range locs {
		raw := text[loc[0]:loc[1]]
		value, ok := parseLocaleNumber(raw)
		if !ok {
			continue
		}
		before := text[max(0, loc[0]-24):loc[0]]
		after := text[loc[1]:min(len(text), loc[1]+24)]
		num := DraftNumber{Raw: raw, Value: value, Kind: KindPlain, Scale: 1, Context: strings.TrimSpace(before + " ⟦" + raw + "⟧ " + after)}
		classify(&num, before, after)
		if shouldSkip(&num) {
			continue
		}
		out = append(out, num)
	}
	return out
}

func classify(n *DraftNumber, before, after string) {
	afterLower := strings.ToLower(after)
	beforeLower := strings.ToLower(before)
	switch {
	case strings.HasPrefix(strings.TrimSpace(after), "%"):
		n.Kind = KindPercent
	case regexp.MustCompile(`^\s*(x|×)`).MatchString(strings.ToLower(after)):
		n.Kind = KindRatio
	case containsAnyWord(afterLower, "triliun", "trillion") || strings.HasPrefix(strings.TrimSpace(after), "T") || endsWithRupiah(beforeLower) && strings.HasPrefix(strings.TrimSpace(after), "t "):
		n.Kind = KindIDR
		n.Scale = 1e12
	case containsAnyWord(afterLower, "miliar", "milyar", "billion"):
		n.Kind = KindIDR
		n.Scale = 1e9
	case containsAnyWord(afterLower, "juta", "million"):
		n.Kind = KindIDR
		n.Scale = 1e6
	case endsWithRupiah(beforeLower):
		n.Kind = KindIDR
	}
}

func endsWithRupiah(beforeLower string) bool {
	fields := strings.Fields(strings.TrimSpace(beforeLower))
	return len(fields) > 0 && fields[len(fields)-1] == "rp"
}

func containsAnyWord(s string, words ...string) bool {
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

func shouldSkip(n *DraftNumber) bool {
	raw := strings.TrimLeft(n.Raw, "+-")
	if n.Kind == KindPlain {
		if yearRe.MatchString(raw) {
			return true
		}
		if v := math.Abs(n.Value); v < 10 && v == math.Trunc(v) {
			return true
		}
	}
	ctx := strings.ToLower(n.Context)
	if regexp.MustCompile(`q\s*⟦[1-4]⟧|step \d|\[\d+/\d+\]`).MatchString(ctx) {
		return true
	}
	return false
}

func parseLocaleNumber(raw string) (float64, bool) {
	neg := false
	if strings.HasPrefix(raw, "-") || strings.HasPrefix(raw, "+") {
		neg = strings.HasPrefix(raw, "-")
		raw = raw[1:]
	}
	hasDot := strings.Contains(raw, ".")
	hasComma := strings.Contains(raw, ",")
	switch {
	case hasDot && hasComma:
		if strings.LastIndex(raw, ",") > strings.LastIndex(raw, ".") {
			raw = strings.ReplaceAll(raw, ".", "")
			raw = strings.ReplaceAll(raw, ",", ".")
		} else {
			raw = strings.ReplaceAll(raw, ",", "")
		}
	case hasComma:
		parts := strings.Split(raw, ",")
		if len(parts) == 2 && len(parts[1]) == 3 && !strings.Contains(parts[0], ".") {
			raw = strings.ReplaceAll(raw, ",", "")
		} else {
			raw = strings.ReplaceAll(raw, ",", ".")
		}
	case hasDot:
		parts := strings.Split(raw, ".")
		if len(parts) == 2 && len(parts[1]) == 3 {
			raw = strings.ReplaceAll(raw, ".", "")
		}
	}
	var parsed float64
	if _, err := fmt.Sscanf(raw, "%g", &parsed); err != nil {
		return 0, false
	}
	if neg {
		parsed = -parsed
	}
	return parsed, true
}

func CheckNumbers(draft string, facts []Fact) NumericCheck {
	numbers := ExtractNumbers(draft)
	result := NumericCheck{Matched: []MatchedNumber{}, Unmatched: []string{}}
	seen := map[string]bool{}
	for _, n := range numbers {
		if seen[n.Raw] {
			result.Skipped++
			continue
		}
		if f, ok := match(n, facts); ok {
			seen[n.Raw] = true
			result.Matched = append(result.Matched, MatchedNumber{Number: n.Raw, Source: f.Path})
		}
	}
	for _, n := range numbers {
		if !seen[n.Raw] {
			seen[n.Raw] = true
			result.Unmatched = append(result.Unmatched, fmt.Sprintf("%s — context: %s", n.Raw, n.Context))
		}
	}
	return result
}

func match(n DraftNumber, facts []Fact) (Fact, bool) {
	tol := func(ref float64) float64 {
		return math.Max(math.Abs(ref)*0.006, 0.015)
	}
	close := func(a, b float64) bool {
		return math.Abs(a-b) <= tol(math.Max(math.Abs(a), math.Abs(b)))
	}
	for _, f := range facts {
		switch n.Kind {
		case KindPercent:
			if close(n.Value, f.Value*100) || close(n.Value, f.Value) {
				return f, true
			}
		case KindRatio:
			if close(n.Value, f.Value) {
				return f, true
			}
		case KindIDR:
			if n.Scale > 1 {
				if close(n.Value, f.Value/n.Scale) || close(n.Value*n.Scale, f.Value) {
					return f, true
				}
			} else {
				if close(n.Value, f.Value) {
					return f, true
				}
			}
		default:
			if close(n.Value, f.Value) || close(n.Value, f.Value/1e12) || close(n.Value, f.Value/1e9) {
				return f, true
			}
		}
	}
	return Fact{}, false
}
