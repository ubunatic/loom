// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"ubunatic.com/loom"
)

var (
	hotStyle  = loom.Style{FG: loom.ColorIndex(1), Bold: true}
	textStyle = loom.Style{FG: loom.ColorIndex(7)}
	metaStyle = loom.Style{FG: loom.ColorIndex(8), Dim: true}
)

// drawBox paints b into a fresh canvas of the given size and returns it.
func drawBox(b *loom.Box, w, h int) *loom.Canvas {
	c := loom.NewCanvas(w, h)
	b.Draw(c, loom.Rect{W: w, H: h})
	return c
}

// titleCells returns the title cells of row 0 from x=1 up to the right border
// (exclusive), so tests address columns by display position, not byte offset.
func titleCells(c *loom.Canvas, w int) []loom.Cell {
	var cells []loom.Cell
	for x := 1; x < w-1; x++ {
		cells = append(cells, c.Get(x, 0))
	}
	return cells
}

func TestBoxTitleSegmentsDrawInOrderWithOwnStyles(t *testing.T) {
	b := &loom.Box{
		Title:  "ignored",
		Border: loom.BoxBorder{TitlePrefix: " ", TitleSuffix: " "},
		TitleSegments: []loom.TitleSegment{
			{Text: "ab", Style: hotStyle},
			{Text: "c", Style: textStyle},
			{Text: "d", Style: metaStyle},
		},
	}
	c := drawBox(b, 20, 3)

	// Prefix space at x=1 uses the title style (zero here), then a, b, c, d.
	want := []struct {
		x    int
		text string
		st   loom.Style
	}{
		{1, " ", loom.Style{}},
		{2, "a", hotStyle},
		{3, "b", hotStyle},
		{4, "c", textStyle},
		{5, "d", metaStyle},
		{6, " ", loom.Style{}},
	}
	for _, w := range want {
		got := c.Get(w.x, 0)
		if got.Text != w.text || got.Style != w.st {
			t.Errorf("x=%d: got %q %+v, want %q %+v", w.x, got.Text, got.Style, w.text, w.st)
		}
	}
	if c.Get(7, 0).Text != "─" {
		t.Errorf("title must end at x=6; x=7 is %q, want border ─", c.Get(7, 0).Text)
	}
}

func TestBoxTitleSegmentsReplaceTitle(t *testing.T) {
	b := &loom.Box{
		Title:         "XYZ",
		TitleSegments: []loom.TitleSegment{{Text: "ok", Style: hotStyle}},
	}
	c := drawBox(b, 10, 3)
	for x := 1; x < 9; x++ {
		if got := c.Get(x, 0).Text; got == "X" || got == "Y" || got == "Z" {
			t.Fatalf("x=%d shows Title %q although TitleSegments is set", x, got)
		}
	}
	if c.Get(1, 0).Text != "o" || c.Get(2, 0).Text != "k" {
		t.Errorf("segments not drawn: %q %q", c.Get(1, 0).Text, c.Get(2, 0).Text)
	}
}

func TestBoxTitleSegmentsClipWideRunesAcrossSegments(t *testing.T) {
	// Interior is w-2 = 6 columns. "日本" takes 4, "ab" fits exactly, "c" is cut.
	b := &loom.Box{TitleSegments: []loom.TitleSegment{
		{Text: "日本", Style: hotStyle},
		{Text: "abc", Style: textStyle},
	}}
	const w = 8
	c := drawBox(b, w, 3)

	if got := c.Get(1, 0); got.Text != "日" || got.Style != hotStyle {
		t.Errorf("x=1: got %q %+v, want 日 hot", got.Text, got.Style)
	}
	if got := c.Get(3, 0); got.Text != "本" || got.Style != hotStyle {
		t.Errorf("x=3: got %q %+v, want 本 hot", got.Text, got.Style)
	}
	if got := c.Get(5, 0); got.Text != "a" || got.Style != textStyle {
		t.Errorf("x=5: got %q %+v, want a text", got.Text, got.Style)
	}
	if got := c.Get(6, 0); got.Text != "b" || got.Style != textStyle {
		t.Errorf("x=6: got %q %+v, want b text", got.Text, got.Style)
	}
	if got := c.Get(7, 0).Text; got != "┐" {
		t.Errorf("x=7 must be the top-right corner, got %q", got)
	}
}

func TestBoxTitleSegmentsWideRuneThatDoesNotFitIsDropped(t *testing.T) {
	// "abcde" fills columns 1..5. "日" needs two columns at 6..7 but only 6 is
	// inside the interior, so it must be dropped whole, not split.
	b := &loom.Box{TitleSegments: []loom.TitleSegment{
		{Text: "abcde", Style: textStyle},
		{Text: "日", Style: hotStyle},
	}}
	c := drawBox(b, 8, 3)
	if got := c.Get(6, 0); got.Text != "─" || got.Style == hotStyle {
		t.Errorf("x=6: got %q %+v, want untouched border ─", got.Text, got.Style)
	}
	if got := c.Get(7, 0).Text; got != "┐" {
		t.Errorf("x=7 must be the top-right corner, got %q", got)
	}
}

func TestBoxTitleSegmentsMeasureSumsSegmentWidths(t *testing.T) {
	b := &loom.Box{
		Border: loom.BoxBorder{TitlePrefix: " ", TitleSuffix: " "},
		TitleSegments: []loom.TitleSegment{
			{Text: "m", Style: hotStyle},
			{Text: "日", Style: textStyle},
			{Text: "[L]", Style: hotStyle},
		},
	}
	// Title width: 1 + 2 + 3 plus prefix and suffix = 8, plus 2 border columns.
	if got := b.Measure(0).Width; got != 10 {
		t.Errorf("Measure(0).Width = %d, want 10", got)
	}
}

func TestHotkeyTitleStylesFirstWordLetter(t *testing.T) {
	got := loom.HotkeyTitle("menu", 'm', hotStyle, textStyle)
	want := []loom.TitleSegment{{Text: "m", Style: hotStyle}, {Text: "enu", Style: textStyle}}
	assertSegments(t, got, want)
}

func TestHotkeyTitleMatchesCaseInsensitively(t *testing.T) {
	got := loom.HotkeyTitle("Load", 'l', hotStyle, textStyle)
	want := []loom.TitleSegment{{Text: "L", Style: hotStyle}, {Text: "oad", Style: textStyle}}
	assertSegments(t, got, want)
}

func TestHotkeyTitleFallsBackToKeyBadgeWhenLetterNotInFirstWord(t *testing.T) {
	// "d" appears only in the second word, so it is not a title-word match.
	got := loom.HotkeyTitle("Fetch diagnostics", 'd', hotStyle, textStyle)
	want := []loom.TitleSegment{{Text: "[d]", Style: hotStyle}, {Text: " Fetch diagnostics", Style: textStyle}}
	assertSegments(t, got, want)
}

func TestHotkeyTitleFallbackForAbsentLetter(t *testing.T) {
	got := loom.HotkeyTitle("Load", 'x', hotStyle, textStyle)
	want := []loom.TitleSegment{{Text: "[x]", Style: hotStyle}, {Text: " Load", Style: textStyle}}
	assertSegments(t, got, want)
}

func TestHotkeyTitleBadgeDrawsInBox(t *testing.T) {
	b := &loom.Box{TitleSegments: loom.HotkeyTitle("Fetch", 'd', hotStyle, textStyle)}
	c := drawBox(b, 20, 3)
	cells := titleCells(c, 20)
	if cells[0].Text != "[" || cells[0].Style != hotStyle {
		t.Errorf("badge start: got %q %+v, want [ hot", cells[0].Text, cells[0].Style)
	}
	if cells[2].Text != "]" || cells[2].Style != hotStyle {
		t.Errorf("badge end: got %q %+v, want ] hot", cells[2].Text, cells[2].Style)
	}
	if cells[4].Text != "F" || cells[4].Style != textStyle {
		t.Errorf("title after badge: got %q %+v, want F text", cells[4].Text, cells[4].Style)
	}
}

func TestHotkeyTitleKeepsTitleTextForWidthAccounting(t *testing.T) {
	for _, tc := range []struct {
		title string
		key   rune
	}{{"menu", 'm'}, {"Load", 'x'}, {"Fetch diagnostics", 'd'}, {"", 'q'}} {
		var joined string
		for _, s := range loom.HotkeyTitle(tc.title, tc.key, hotStyle, textStyle) {
			joined += s.Text
		}
		if tc.title == "" {
			continue
		}
		if joined != tc.title && joined != "["+string(tc.key)+"] "+tc.title {
			t.Errorf("HotkeyTitle(%q,%q) text = %q", tc.title, tc.key, joined)
		}
	}
}

func assertSegments(t *testing.T, got, want []loom.TitleSegment) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d segments %+v, want %d %+v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("segment %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
