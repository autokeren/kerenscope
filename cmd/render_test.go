package cmd

import (
	"strings"
	"testing"
)

func TestBannerRowsHaveEqualDisplayWidth(t *testing.T) {
	b := buildBanner(true)
	lines := strings.Split(b, "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines, got %d", len(lines))
	}
	widths := make([]int, len(lines))
	for i, line := range lines {
		widths[i] = displayLen(strings.TrimRight(line, "\x1b[0m"))
	}
	for i, w := range widths {
		if i == 0 || i == len(widths)-1 {
			continue
		}
		if w != widths[0] {
			t.Fatalf("line %d width %d != border width %d", i, w, widths[0])
		}
	}
}

func TestWrapIndentAlignment(t *testing.T) {
	out := wrapIndent("kata sangat panjang yang harus membungkus ke beberapa baris agar rapi", 30, "    ")
	for _, line := range strings.Split(out, "\n") {
		if displayLen(line) > 36 {
			t.Fatalf("line too wide: %q (%d)", line, displayLen(line))
		}
	}
}
