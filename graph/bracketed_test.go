// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package graph

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom/measure"
)

func TestBracketedBarDeterminate(t *testing.T) {
	opts := BracketedBarOptions{
		Width:   24,
		SubChar: true,
	}

	// 0% -> all empty
	bar0 := RenderBracketedBar(0, opts)
	if !strings.HasPrefix(bar0, "[") || !strings.HasSuffix(bar0, "]") {
		t.Fatalf("unexpected wrapper: %q", bar0)
	}
	if w := measure.StringWidth(bar0); w != 26 {
		t.Fatalf("width of bar0 = %d, want 26", w)
	}

	// 100% -> all filled
	bar100 := RenderBracketedBar(100, opts)
	want100 := "[" + strings.Repeat("⣿", 24) + "]"
	if bar100 != want100 {
		t.Fatalf("bar100 = %q, want %q", bar100, want100)
	}
	if w := measure.StringWidth(bar100); w != 26 {
		t.Fatalf("width of bar100 = %d, want 26", w)
	}

	// Partial fill matching target example: 4 full ⣿, 1 partial ⡇, 19 spaces
	// 4.5 / 24 = 18.75%
	barMid := RenderBracketedBar(18.75, opts)
	wantMid := "[⣿⣿⣿⣿⡇" + strings.Repeat(" ", 19) + "]"
	if barMid != wantMid {
		t.Fatalf("barMid = %q, want %q", barMid, wantMid)
	}
	if w := measure.StringWidth(barMid); w != 26 {
		t.Fatalf("width of barMid = %d, want 26", w)
	}
}

func TestBracketedBarPattern(t *testing.T) {
	opts := BracketedBarOptions{
		Width:   32,
		Pattern: ":",
	}

	got := RenderBracketedBar(100, opts)
	want := "[" + strings.Repeat(":", 32) + "]"
	if got != want {
		t.Fatalf("got pattern bar %q, want %q", got, want)
	}
	if w := measure.StringWidth(got); w != 34 {
		t.Fatalf("width = %d, want 34", w)
	}
}

func TestBracketedBarWidthHelper(t *testing.T) {
	opts := BracketedBarOptions{
		Width: 20,
		Left:  "<",
		Right: ">",
	}
	if w := BracketedBarWidth(opts); w != 22 {
		t.Fatalf("BracketedBarWidth = %d, want 22", w)
	}
}
