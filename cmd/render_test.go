package cmd

import (
	"strings"
	"testing"
	"time"
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

func TestTermWidthFallbackAndHelpers(t *testing.T) {
	if w := termWidth(); w < 40 {
		t.Fatalf("termWidth must never go below 40, got %d", w)
	}
	if c := contentWidth(); c != termWidth()-4 {
		t.Fatalf("contentWidth must be termWidth-4, got %d", c)
	}
}

func TestRevealPrintInstantOnNonTTY(t *testing.T) {
	start := time.Now()
	revealPrint("line1\nline2\nline3", true)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("revealPrint must be instant on non-TTY, took %v", elapsed)
	}
}
