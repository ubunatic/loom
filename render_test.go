// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── A4: Canvas.Flush — absolute row positioning + cursor placement ────────────

func TestCanvasFlushRowPositioning(t *testing.T) {
	c := loom.NewCanvas(3, 2)
	c.Write(0, 0, "ab", loom.Reset)
	var b strings.Builder
	c.Flush(&b, 5) // canvas top at terminal row 5
	out := b.String()

	if !strings.Contains(out, "\x1b[5;1H") {
		t.Errorf("Flush missing row-1 move \\x1b[5;1H: %q", out)
	}
	if !strings.Contains(out, "\x1b[6;1H") {
		t.Errorf("Flush missing row-2 move \\x1b[6;1H: %q", out)
	}
}

func TestCanvasFlushCursorMove(t *testing.T) {
	c := loom.NewCanvas(4, 2)
	c.CursorX = 1
	c.CursorY = 0
	var b strings.Builder
	c.Flush(&b, 5)
	// Cursor at (startRow+CursorY, CursorX+1) = (5, 2), then made visible.
	if !strings.HasSuffix(b.String(), "\x1b[5;2H\x1b[?25h") {
		t.Errorf("Flush should position then show the cursor (\\x1b[5;2H\\x1b[?25h): %q", b.String())
	}
}

func TestCanvasFlushNoCursorWhenHidden(t *testing.T) {
	c := loom.NewCanvas(4, 1) // default CursorX/Y are -1 (hidden)
	var b strings.Builder
	c.Flush(&b, 3)
	// No final cursor move should be emitted; output ends with the row reset.
	if strings.HasSuffix(b.String(), "\x1b[3;1H") {
		t.Error("hidden cursor should not emit a trailing absolute-position move")
	}
}

// ── A5: View scroll-indicator visual math ─────────────────────────────────────

// hasIndicator reports whether the rightmost column holds the specced thumb.
func hasIndicator(c *loom.Canvas) (row int, ok bool) {
	x := c.Cols() - 1
	for y := 0; y < c.Rows(); y++ {
		if c.Get(x, y).Text == loom.SpeccedDefaults.Scrollbar.ForegroundChar {
			return y, true
		}
	}
	return -1, false
}

func TestViewNoIndicatorWhenFits(t *testing.T) {
	v := loom.NewView([]string{"a", "b"})
	c := loom.NewCanvas(6, 3) // 3 rows ≥ 2 lines → not scrollable
	v.Draw(c, c.Bounds())
	if _, ok := hasIndicator(c); ok {
		t.Error("non-scrollable view should not draw a scroll indicator")
	}
}

func TestViewIndicatorAtTopAndBottom(t *testing.T) {
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line"
	}
	v := loom.NewView(lines)
	c := loom.NewCanvas(6, 3) // scrollable: 10 lines, height 3, maxScroll = 7

	// Scroll=0 → indicator at the top row (0).
	v.Draw(c, c.Bounds())
	if row, ok := hasIndicator(c); !ok || row != 0 {
		t.Errorf("at top: indicator row = %d, ok=%v; want row 0", row, ok)
	}

	// Scroll to max → indicator at the bottom row (H-1 = 2).
	v.Scroll = 7
	c.Clear()
	v.Draw(c, c.Bounds())
	if row, ok := hasIndicator(c); !ok || row != c.Rows()-1 {
		t.Errorf("at bottom: indicator row = %d, ok=%v; want row %d", row, ok, c.Rows()-1)
	}
}

func TestViewContentTruncatedToReserveIndicatorColumn(t *testing.T) {
	lines := make([]string, 5)
	for i := range lines {
		lines[i] = strings.Repeat("X", 10) // wider than the canvas
	}
	v := loom.NewView(lines)
	c := loom.NewCanvas(6, 3) // scrollable → content width is 5 (W-1)
	v.Draw(c, c.Bounds())

	// Rightmost column on the indicator row holds the thumb, not content.
	if got := c.Get(5, 0).Text; got != loom.SpeccedDefaults.Scrollbar.ForegroundChar {
		t.Errorf("rightmost col on indicator row = %q, want %q", got, loom.SpeccedDefaults.Scrollbar.ForegroundChar)
	}
	// Column W-2 still holds content.
	if got := c.Get(4, 0).Text; got != "X" {
		t.Errorf("content column 4 = %q, want X", got)
	}
}
