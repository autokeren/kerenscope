package agent

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func (r *Report) Render() string {
	var b strings.Builder
	b.WriteString("# KerenScope Research Report\n\n")
	b.WriteString(fmt.Sprintf("**Question:** %s\n\n", r.Question))
	b.WriteString(fmt.Sprintf("**Objective:** %s\n\n", r.Plan.Objective))
	b.WriteString(fmt.Sprintf("*Generated %s · Tools used: %s*\n\n", r.StartedAt.Format("2006-01-02 15:04 MST"), strings.Join(unique(r.ToolsUsed), ", ")))
	b.WriteString("## Analysis\n\n")
	b.WriteString(r.Draft)
	b.WriteString("\n\n")
	if len(r.Verification.Claims) > 0 {
		b.WriteString("## Claim verification\n\n")
		supported := 0
		for _, c := range r.Verification.Claims {
			icon := "✗"
			if c.Supported {
				icon = "✓"
				supported++
			}
			b.WriteString(fmt.Sprintf("- %s **%s** — %s\n", icon, c.Claim, c.Evidence))
		}
		b.WriteString(fmt.Sprintf("\n**%d/%d claims supported by the evidence.**\n\n", supported, len(r.Verification.Claims)))
	}
	if r.Verification.Confidence > 0 {
		b.WriteString(fmt.Sprintf("## Confidence\n\n**%d/100**\n\n", r.Verification.Confidence))
	}
	if len(r.Verification.Limitations) > 0 {
		b.WriteString("## Limitations\n\n")
		for _, l := range r.Verification.Limitations {
			b.WriteString(fmt.Sprintf("- %s\n", l))
		}
		b.WriteString("\n")
	}
	b.WriteString("---\n\n*Disclaimer: KerenScope is an information and analysis tool, not investment advice. Data: Sectors API.*\n")
	return b.String()
}

func (r *Report) Slug() string {
	s := strings.ToLower(r.Question)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	if s == "" {
		s = "research"
	}
	return s + "-" + time.Now().Format("20060102-1504")
}

func unique(ss []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
