// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── M1: ParseANSI basic functionality ────────────────────────────────────────

func TestParseANSIEmpty(t *testing.T) {
	cells := loom.ParseANSI("")
	if len(cells) != 0 {
		t.Errorf("ParseANSI(\"\") = %d cells, want 0", len(cells))
	}
}

func TestParseANSIPlainText(t *testing.T) {
	cells := loom.ParseANSI("hello")
	if len(cells) != 5 {
		t.Fatalf("ParseANSI(\"hello\") = %d cells, want 5", len(cells))
	}
	for i, ch := range "hello" {
		if cells[i].Text != string(ch) {
			t.Errorf("cell %d: got %q, want %q", i, cells[i].Text, string(ch))
		}
		if cells[i].Style != loom.Reset {
			t.Errorf("cell %d: got style %#v, want Reset", i, cells[i].Style)
		}
	}
}

// ── Reset (code 0) ──────────────────────────────────────────────────────────

func TestParseANSIReset(t *testing.T) {
	// \x1b[0m followed by text
	cells := loom.ParseANSI("\x1b[0mtext")
	if len(cells) != 4 {
		t.Fatalf("got %d cells, want 4", len(cells))
	}
	for i, ch := range "text" {
		if cells[i].Text != string(ch) {
			t.Errorf("cell %d: got %q, want %q", i, cells[i].Text, string(ch))
		}
		if cells[i].Style != loom.Reset {
			t.Errorf("cell %d style not reset", i)
		}
	}
}

// ── Bold (code 1) ──────────────────────────────────────────────────────────

func TestParseANSIBold(t *testing.T) {
	cells := loom.ParseANSI("\x1b[1mbold")
	if len(cells) != 4 {
		t.Fatalf("got %d cells, want 4", len(cells))
	}
	for _, cell := range cells {
		if !cell.Style.Bold {
			t.Errorf("expected Bold=true, got %#v", cell.Style)
		}
	}
}

// ── Dim (code 2) ────────────────────────────────────────────────────────────

func TestParseANSIDim(t *testing.T) {
	cells := loom.ParseANSI("\x1b[2mdim")
	if len(cells) != 3 {
		t.Fatalf("got %d cells, want 3", len(cells))
	}
	for _, cell := range cells {
		if !cell.Style.Dim {
			t.Errorf("expected Dim=true, got %#v", cell.Style)
		}
	}
}

// ── Underline (code 4) ──────────────────────────────────────────────────────

func TestParseANSIUnderline(t *testing.T) {
	cells := loom.ParseANSI("\x1b[4munder")
	if len(cells) != 5 {
		t.Fatalf("got %d cells, want 5", len(cells))
	}
	for _, cell := range cells {
		if !cell.Style.Underline {
			t.Errorf("expected Underline=true, got %#v", cell.Style)
		}
	}
}

// ── Italic (code 3) – if supported ──────────────────────────────────────────
// Currently not supported in loom.Style; unknown codes should be dropped

func TestParseANSIIgnoresItalic(t *testing.T) {
	// Code 3 (italic) is not in loom.Style, so it should be ignored
	cells := loom.ParseANSI("\x1b[3mitalic")
	if len(cells) != 6 {
		t.Fatalf("got %d cells, want 6", len(cells))
	}
	// The text should be rendered, but without italic (since loom doesn't support it)
	for _, cell := range cells {
		// Just verify the cells exist and have the text
		if cell.Text == "" {
			t.Error("cell should have text, not be empty")
		}
	}
}

// ── 16-color foreground (codes 30-37, 90-97) ──────────────────────────────

func TestParseANSI16ColorFG(t *testing.T) {
	tests := []struct {
		code  string
		index uint8
	}{
		{"\x1b[30m", 0},  // black
		{"\x1b[31m", 1},  // red
		{"\x1b[32m", 2},  // green
		{"\x1b[33m", 3},  // yellow
		{"\x1b[34m", 4},  // blue
		{"\x1b[35m", 5},  // magenta
		{"\x1b[36m", 6},  // cyan
		{"\x1b[37m", 7},  // white
		{"\x1b[90m", 8},  // bright black
		{"\x1b[91m", 9},  // bright red
		{"\x1b[92m", 10}, // bright green
		{"\x1b[93m", 11}, // bright yellow
		{"\x1b[94m", 12}, // bright blue
		{"\x1b[95m", 13}, // bright magenta
		{"\x1b[96m", 14}, // bright cyan
		{"\x1b[97m", 15}, // bright white
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			cells := loom.ParseANSI(tt.code + "a")
			if len(cells) != 1 {
				t.Fatalf("got %d cells, want 1", len(cells))
			}
			r, g, b, ok := cells[0].Style.FG.RGB()
			if !ok {
				t.Fatalf("FG is not RGB-convertible (code %q)", tt.code)
			}
			// Verify it's a valid 16-color by checking against xterm colors
			wantR, wantG, wantB, _ := loom.ColorIndex(tt.index).RGB()
			if r != wantR || g != wantG || b != wantB {
				t.Errorf("got RGB(%d,%d,%d), want (%d,%d,%d)", r, g, b, wantR, wantG, wantB)
			}
		})
	}
}

// ── 16-color background (codes 40-47, 100-107) ──────────────────────────────

func TestParseANSI16ColorBG(t *testing.T) {
	cells := loom.ParseANSI("\x1b[41ma")
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	wantR, wantG, wantB, _ := loom.ColorIndex(1).RGB()
	r, g, b, ok := cells[0].Style.BG.RGB()
	if !ok {
		t.Fatalf("BG is not RGB-convertible")
	}
	if r != wantR || g != wantG || b != wantB {
		t.Errorf("got RGB(%d,%d,%d), want (%d,%d,%d)", r, g, b, wantR, wantG, wantB)
	}
}

// ── 256-color foreground (38;5;n) ──────────────────────────────────────────

func TestParseANSI256ColorFG(t *testing.T) {
	cells := loom.ParseANSI("\x1b[38;5;196ma")
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	r, g, b, ok := cells[0].Style.FG.RGB()
	if !ok {
		t.Fatalf("FG is not RGB-convertible")
	}
	// Color 196 in xterm 256 palette is pure red
	wantR, wantG, wantB, _ := loom.ColorIndex(196).RGB()
	if r != wantR || g != wantG || b != wantB {
		t.Errorf("got RGB(%d,%d,%d), want (%d,%d,%d)", r, g, b, wantR, wantG, wantB)
	}
}

// ── 256-color background (48;5;n) ──────────────────────────────────────────

func TestParseANSI256ColorBG(t *testing.T) {
	cells := loom.ParseANSI("\x1b[48;5;21ma")
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	r, g, b, ok := cells[0].Style.BG.RGB()
	if !ok {
		t.Fatalf("BG is not RGB-convertible")
	}
	wantR, wantG, wantB, _ := loom.ColorIndex(21).RGB()
	if r != wantR || g != wantG || b != wantB {
		t.Errorf("got RGB(%d,%d,%d), want (%d,%d,%d)", r, g, b, wantR, wantG, wantB)
	}
}

// ── 24-bit RGB foreground (38;2;r;g;b) ──────────────────────────────────────

func TestParseANSIRGBFG(t *testing.T) {
	cells := loom.ParseANSI("\x1b[38;2;255;0;128ma")
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	r, g, b, ok := cells[0].Style.FG.RGB()
	if !ok {
		t.Fatalf("FG is not RGB-convertible")
	}
	if r != 255 || g != 0 || b != 128 {
		t.Errorf("got RGB(%d,%d,%d), want (255,0,128)", r, g, b)
	}
}

// ── 24-bit RGB background (48;2;r;g;b) ──────────────────────────────────────

func TestParseANSIRGBBG(t *testing.T) {
	cells := loom.ParseANSI("\x1b[48;2;10;20;30ma")
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	r, g, b, ok := cells[0].Style.BG.RGB()
	if !ok {
		t.Fatalf("BG is not RGB-convertible")
	}
	if r != 10 || g != 20 || b != 30 {
		t.Errorf("got RGB(%d,%d,%d), want (10,20,30)", r, g, b)
	}
}

// ── Reset FG (code 39) ──────────────────────────────────────────────────────

func TestParseANSIResetFG(t *testing.T) {
	cells := loom.ParseANSI("\x1b[31m\x1b[39mtext")
	if len(cells) != 4 {
		t.Fatalf("got %d cells, want 4", len(cells))
	}
	for _, cell := range cells {
		if cell.Style.FG != loom.ColorReset() {
			t.Errorf("FG not reset: %#v", cell.Style)
		}
	}
}

// ── Reset BG (code 49) ──────────────────────────────────────────────────────

func TestParseANSIResetBG(t *testing.T) {
	cells := loom.ParseANSI("\x1b[41m\x1b[49mtext")
	if len(cells) != 4 {
		t.Fatalf("got %d cells, want 4", len(cells))
	}
	for _, cell := range cells {
		if cell.Style.BG != loom.ColorReset() {
			t.Errorf("BG not reset: %#v", cell.Style)
		}
	}
}

// ── Multiple attributes in one sequence ──────────────────────────────────────

func TestParseANSIMultipleAttributes(t *testing.T) {
	// \x1b[1;4;31m means bold, underline, red
	cells := loom.ParseANSI("\x1b[1;4;31mtext")
	if len(cells) != 4 {
		t.Fatalf("got %d cells, want 4", len(cells))
	}
	for _, cell := range cells {
		if !cell.Style.Bold {
			t.Error("Bold not set")
		}
		if !cell.Style.Underline {
			t.Error("Underline not set")
		}
		r, _, _, ok := cell.Style.FG.RGB()
		if !ok || r != 205 { // Red in xterm
			t.Errorf("FG not red, got %#v", cell.Style.FG)
		}
	}
}

// ── Wide characters (emoji) ─────────────────────────────────────────────────

func TestParseANSIWideCharacter(t *testing.T) {
	// Emoji "👍" has width 2
	cells := loom.ParseANSI("👍")
	if len(cells) != 2 {
		t.Fatalf("got %d cells, want 2 (wide char + continuation)", len(cells))
	}
	if cells[0].Text != "👍" {
		t.Errorf("first cell text: got %q, want %q", cells[0].Text, "👍")
	}
	if !cells[1].Continuation {
		t.Errorf("second cell should be continuation, got %#v", cells[1])
	}
}

// ── Combining marks ──────────────────────────────────────────────────────────

func TestParseANSICombiningMarks(t *testing.T) {
	// "e" + combining acute accent
	str := "é" // é as e + combining accent
	cells := loom.ParseANSI(str)
	if len(cells) < 1 {
		t.Fatalf("got %d cells, want at least 1", len(cells))
	}
	// The first cell should contain the combined cluster
	if !strings.Contains(cells[0].Text, "e") {
		t.Errorf("expected base 'e' in first cell, got %q", cells[0].Text)
	}
}

// ── Unknown escape sequences are dropped ──────────────────────────────────────

func TestParseANSIUnknownSequences(t *testing.T) {
	// CSI sequences that are not SGR (m) should be dropped
	// For example, cursor movement should not produce cells
	cells := loom.ParseANSI("\x1b[Atext") // ESC [ A = cursor up
	if len(cells) != 4 {
		t.Fatalf("got %d cells, want 4 (unknown escape dropped, text rendered)", len(cells))
	}
	for i, ch := range "text" {
		if cells[i].Text != string(ch) {
			t.Errorf("cell %d: got %q, want %q", i, cells[i].Text, string(ch))
		}
	}
}

// ── Mixed content: text, escapes, text ────────────────────────────────────────

func TestParseANSIMixedContent(t *testing.T) {
	input := "hello \x1b[1mworld\x1b[0m end"
	cells := loom.ParseANSI(input)

	// Expected: h e l l o (space) w o r l d (space) e n d
	// That's 15 characters total
	if len(cells) != 15 {
		t.Fatalf("got %d cells, want 15", len(cells))
	}

	// First 6 cells are "hello " with default style
	for i := 0; i < 6; i++ {
		if cells[i].Style != loom.Reset {
			t.Errorf("cell %d: expected Reset style, got %#v", i, cells[i].Style)
		}
	}

	// Next 5 cells are "world" with bold
	for i := 6; i < 11; i++ {
		if !cells[i].Style.Bold {
			t.Errorf("cell %d: expected Bold, got %#v", i, cells[i].Style)
		}
	}

	// Last 4 cells are " end" with reset style
	for i := 11; i < 15; i++ {
		if cells[i].Style != loom.Reset {
			t.Errorf("cell %d: expected Reset style, got %#v", i, cells[i].Style)
		}
	}
}

// ── Incomplete sequences are handled gracefully ───────────────────────────────

func TestParseANSIIncompleteSequence(t *testing.T) {
	// Incomplete sequence at end
	cells := loom.ParseANSI("text\x1b[")
	if len(cells) < 4 {
		t.Fatalf("got %d cells, want at least 4", len(cells))
	}
	// The "text" part should be rendered
	for i, ch := range "text" {
		if cells[i].Text != string(ch) {
			t.Errorf("cell %d: got %q, want %q", i, cells[i].Text, string(ch))
		}
	}
}

// ── Benchmark ────────────────────────────────────────────────────────────────

func BenchmarkParseANSI(b *testing.B) {
	// A typical colored output line with multiple color changes
	input := "\x1b[31mERROR\x1b[0m: " +
		"\x1b[38;5;196mFailed\x1b[0m - " +
		"\x1b[1;34mDetails\x1b[0m: " +
		"Some error message with lots of words that goes on for a bit"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = loom.ParseANSI(input)
	}
}
