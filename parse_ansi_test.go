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

// ── M4: Edge cases and advanced SGR codes ──────────────────────────────────────

func TestParseANSIEmptySGR(t *testing.T) {
	// Bare ESC[m (empty parameter list) should reset all
	cells := loom.ParseANSI("\x1b[1;31mBOLD RED\x1b[mNOT BOLD")
	if len(cells) < 12 {
		t.Fatalf("got %d cells, want at least 12", len(cells))
	}
	// First 8 cells should be bold+red
	for i := 0; i < 8; i++ {
		if !cells[i].Style.Bold {
			t.Errorf("cell %d (before reset): should be bold", i)
		}
	}
	// Last 8 cells should be reset
	for i := 8; i < 16; i++ {
		if i < len(cells) && cells[i].Style != loom.Reset {
			t.Errorf("cell %d (after reset): should be reset, got %#v", i, cells[i].Style)
		}
	}
}

func TestParseANSIEmptyParts(t *testing.T) {
	// ESC[;1m with empty first part should treat empty as 0
	cells := loom.ParseANSI("\x1b[;1mtext")
	if len(cells) != 4 {
		t.Fatalf("got %d cells, want 4", len(cells))
	}
	for _, cell := range cells {
		if !cell.Style.Bold {
			t.Errorf("expected bold, got %#v", cell.Style)
		}
	}
}

func TestParseANSIBoldOff(t *testing.T) {
	// SGR code 22 turns off bold and dim
	cells := loom.ParseANSI("\x1b[1mBOLD\x1b[22mNOT")
	if len(cells) < 7 {
		t.Fatalf("got %d cells, want at least 7", len(cells))
	}
	// First 4 should be bold
	for i := 0; i < 4; i++ {
		if !cells[i].Style.Bold {
			t.Errorf("cell %d: should be bold", i)
		}
	}
	// Last 3 should not be bold
	for i := 4; i < 7; i++ {
		if cells[i].Style.Bold {
			t.Errorf("cell %d: should not be bold after code 22", i)
		}
	}
}

func TestParseANSIUnderlineOff(t *testing.T) {
	// SGR code 24 turns off underline
	cells := loom.ParseANSI("\x1b[4mUNDER\x1b[24mNOT")
	if len(cells) < 8 {
		t.Fatalf("got %d cells, want at least 8", len(cells))
	}
	// First 5 should be underlined
	for i := 0; i < 5; i++ {
		if !cells[i].Style.Underline {
			t.Errorf("cell %d: should be underlined", i)
		}
	}
	// Last 3 should not be underlined
	for i := 5; i < 8; i++ {
		if cells[i].Style.Underline {
			t.Errorf("cell %d: should not be underlined after code 24", i)
		}
	}
}

func TestParseANSI256ColorOutOfRange(t *testing.T) {
	// Value > 255 in 256-color should be ignored, not wrapped
	cells := loom.ParseANSI("\x1b[38;5;256ma\x1b[38;5;100mb")
	if len(cells) < 2 {
		t.Fatalf("got %d cells, want at least 2", len(cells))
	}
	// First cell should have default (reset) color, not wrapped
	if cells[0].Style.FG != loom.ColorReset() {
		t.Errorf("cell 0: color 256 should be ignored, got %#v", cells[0].Style.FG)
	}
	// Second cell should have color 100
	if cells[1].Style.FG != loom.ColorIndex(100) {
		t.Errorf("cell 1: expected color 100, got %#v", cells[1].Style.FG)
	}
}

func TestParseANSIRGBOutOfRange(t *testing.T) {
	// RGB value > 255 should be ignored
	cells := loom.ParseANSI("\x1b[38;2;300;0;0ma\x1b[38;2;100;100;100mb")
	if len(cells) < 2 {
		t.Fatalf("got %d cells, want at least 2", len(cells))
	}
	// First cell should have default (reset) color
	if cells[0].Style.FG != loom.ColorReset() {
		t.Errorf("cell 0: RGB(300,0,0) should be ignored, got %#v", cells[0].Style.FG)
	}
	// Second cell should have color (100,100,100)
	r, g, b, ok := cells[1].Style.FG.RGB()
	if !ok || r != 100 || g != 100 || b != 100 {
		t.Errorf("cell 1: expected RGB(100,100,100), got (%d,%d,%d)", r, g, b)
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

// ── Fuzz/Property test ──────────────────────────────────────────────────────────

func TestParseANSIPropertyValidCells(t *testing.T) {
	// Property test: ParseANSI never returns invalid cell sequences
	testCases := []string{
		"",
		"plain text",
		"\x1b[31mred\x1b[0m",
		"\x1b[38;5;196mcolor196\x1b[0m",
		"\x1b[38;2;255;100;50mrgb\x1b[0m",
		"👍emoji",
		"\x1b[1;4;31mbold underline red\x1b[0m",
		"\x1b[1m\x1b[22m\x1b[4m\x1b[24mbold off underline off",
		"\x1b[;1m;empty parts;",
		"\x1b[mbare reset",
	}

	for _, input := range testCases {
		cells := loom.ParseANSI(input)
		for i := 0; i < len(cells); i++ {
			cell := cells[i]
			// Continuation cells should appear after wide cells
			if cell.Continuation {
				if i == 0 {
					t.Errorf("input %q: continuation at index 0", input)
				}
				prevCell := cells[i-1]
				if loom.StringWidth(prevCell.Text) != 2 {
					t.Errorf("input %q: continuation at %d not after wide cell", input, i)
				}
			}
			// All cells should have either text or be continuation
			if cell.Text == "" && !cell.Continuation {
				t.Errorf("input %q: cell %d is blank non-continuation", input, i)
			}
		}
	}
}

func TestParseANSINeverPanics(t *testing.T) {
	// Property test: ParseANSI never panics on any input
	testInputs := []string{
		"\x1b",           // incomplete escape
		"\x1b[",          // incomplete CSI
		"\x1b[1;2;3;4;5", // incomplete
		"\x1b[999m",      // high code
		"\x1b[38;5;999m", // high color
		"\x1b[38;2;999;999;999m", // high RGB
		"\x1b[38;999m",   // malformed
		string([]byte{255, 254, 253}), // invalid UTF-8
		"\x1b[;;;;;m",    // many empty parts
	}

	for _, input := range testInputs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ParseANSI panicked on input %q: %v", input, r)
				}
			}()
			loom.ParseANSI(input)
		}()
	}
}
