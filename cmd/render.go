package cmd

import (
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/term"
)

const (
	cReset  = "\x1b[0m"
	cBold   = "\x1b[1m"
	cDim    = "\x1b[2m"
	cCyan   = "\x1b[36m"
	cGreen  = "\x1b[32m"
	cYellow = "\x1b[33m"
	cRed    = "\x1b[31m"
	cBlue   = "\x1b[94m"
)

var colorOn = os.Getenv("NO_COLOR") == "" && term.IsTerminal(int(os.Stdout.Fd()))

func stylize(code, s string) string {
	if !colorOn {
		return s
	}
	return code + s + cReset
}

func bold(s string) string    { return stylize(cBold, s) }
func dim(s string) string     { return stylize(cDim, s) }
func cyan(s string) string    { return stylize(cCyan, s) }
func green(s string) string   { return stylize(cGreen, s) }
func yellow(s string) string  { return stylize(cYellow, s) }
func red(s string) string     { return stylize(cRed, s) }
func blue(s string) string    { return stylize(cBlue, s) }

func wrap(s string, width int, indent string) string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		line := ""
		for _, w := range words {
			candidate := line
			if candidate != "" {
				candidate += " "
			}
			candidate += w
			if displayLen(candidate) > width && line != "" {
				out = append(out, line)
				line = w
			} else {
				line = candidate
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func displayLen(s string) int {
	s = ansiRe.ReplaceAllString(s, "")
	n := 0
	for _, r := range s {
		if unicode.IsPrint(r) {
			n++
		}
	}
	return n
}

func wrapIndent(s string, width int, indent string) string {
	wrapped := wrap(s, width-displayLen(indent), indent)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		if i == 0 {
			lines[i] = indent + line
		} else {
			lines[i] = strings.Repeat(" ", displayLen(indent)) + line
		}
	}
	return strings.Join(lines, "\n")
}

func banner() string {
	return buildBanner(colorOn)
}

func buildBanner(color bool) string {
	const width = 54
	rows := []string{
		"  " + "KerenScope" + " — Autonomous Financial Research",
		"  " + "Indonesian market intelligence · Powered by Sectors",
	}
	if !color {
		return "KerenScope — Autonomous Financial Research\n" + rows[1][2:] + "\n"
	}
	var b strings.Builder
	b.WriteString(cCyan + "╭" + strings.Repeat("─", width) + "╮\n")
	for i, row := range rows {
		if i == 0 {
			row = "  " + bold("KerenScope") + " — Autonomous Financial Research"
		} else {
			row = "  " + dim(row[2:])
		}
		pad := width - displayLen(row)
		if pad < 0 {
			pad = 0
		}
		b.WriteString(cCyan + "│" + cReset + row + strings.Repeat(" ", pad) + cCyan + "│" + cReset + "\n")
	}
	b.WriteString(cCyan + "╰" + strings.Repeat("─", width) + "╯" + cReset)
	return b.String()
}

func section(title string) string {
	return "\n  " + cyan("◆ ") + bold(title)
}

func statusIcon(ok bool) string {
	if ok {
		return green("✓")
	}
	return red("✗")
}
