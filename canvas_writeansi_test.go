// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── M2: Canvas.WriteANSI basic functionality ──────────────────────────────────

func TestCanvasWriteANSIBasic(t *testing.T) {
	canvas := loom.NewCanvas(10, 3)
	n := canvas.WriteANSI(0, 0, "hello")
	if n != 5 {
		t.Errorf("WriteANSI returned %d, want 5", n)
	}

	// Verify the cells were written
	for i, ch := range "hello" {
		cell := canvas.Get(i, 0)
		if cell.Text != string(ch) {
			t.Errorf("cell %d: got %q, want %q", i, cell.Text, string(ch))
		}
	}
}

func TestCanvasWriteANSIWithColor(t *testing.T) {
	canvas := loom.NewCanvas(20, 3)
	input := "\x1b[31mRED\x1b[0m \x1b[32mGREEN\x1b[0m"
	n := canvas.WriteANSI(0, 0, input)

	// "RED GREEN" is 9 characters
	if n != 9 {
		t.Errorf("WriteANSI returned %d, want 9", n)
	}

	// Check first 3 cells are RED
	for i := 0; i < 3; i++ {
		cell := canvas.Get(i, 0)
		if cell.Text == "" {
			continue
		}
		if cell.Style.FG != loom.ColorIndex(1) {
			t.Errorf("cell %d: FG should be red (index 1), got %#v", i, cell.Style.FG)
		}
	}

	// Check cells after space are GREEN
	for i := 4; i < 9; i++ {
		cell := canvas.Get(i, 0)
		if cell.Text == "" {
			continue
		}
		if cell.Style.FG != loom.ColorIndex(2) {
			t.Errorf("cell %d: FG should be green (index 2), got %#v", i, cell.Style.FG)
		}
	}
}

func TestCanvasWriteANSIPositioning(t *testing.T) {
	canvas := loom.NewCanvas(10, 3)
	// Write at position (3, 1)
	n := canvas.WriteANSI(3, 1, "test")
	if n != 4 {
		t.Errorf("WriteANSI returned %d, want 4", n)
	}

	// Verify positions
	for i, ch := range "test" {
		cell := canvas.Get(3+i, 1)
		if cell.Text != string(ch) {
			t.Errorf("cell at (3+%d, 1): got %q, want %q", i, cell.Text, string(ch))
		}
	}

	// Verify cells outside this region are blank
	if canvas.Get(0, 1).Text != " " && canvas.Get(0, 1).Text != "" {
		t.Error("cell at (0, 1) should be blank")
	}
}

func TestCanvasWriteANSIClipping(t *testing.T) {
	canvas := loom.NewCanvas(5, 3)
	// Write a string longer than canvas width
	n := canvas.WriteANSI(0, 0, "toolong")
	if n != 5 {
		t.Errorf("WriteANSI returned %d, want 5 (clipped to canvas width)", n)
	}

	// Verify only 5 cells were written
	for i := 0; i < 5; i++ {
		if i < len("toolo") {
			ch := "toolo"[i]
			cell := canvas.Get(i, 0)
			if cell.Text != string(ch) {
				t.Errorf("cell %d: got %q, want %q", i, cell.Text, string(ch))
			}
		}
	}
}

func TestCanvasWriteANSIClippingAtLeft(t *testing.T) {
	canvas := loom.NewCanvas(10, 3)
	// Start writing at x=8 (only 2 columns left)
	n := canvas.WriteANSI(8, 0, "test")
	if n != 2 {
		t.Errorf("WriteANSI returned %d, want 2 (clipped to bounds)", n)
	}

	// Verify only 2 cells were written
	if canvas.Get(8, 0).Text != "t" {
		t.Error("cell at (8, 0) should be 't'")
	}
	if canvas.Get(9, 0).Text != "e" {
		t.Error("cell at (9, 0) should be 'e'")
	}
}

func TestCanvasWriteANSIWideCharacters(t *testing.T) {
	canvas := loom.NewCanvas(10, 3)
	// Emoji "👍" has width 2
	n := canvas.WriteANSI(0, 0, "👍test")
	// 👍 = 2 cells, + 4 characters = 6 cells total
	if n != 6 {
		t.Errorf("WriteANSI returned %d, want 6", n)
	}

	// First cell should contain the emoji
	cell := canvas.Get(0, 0)
	if cell.Text != "👍" {
		t.Errorf("cell at (0, 0): got %q, want %q", cell.Text, "👍")
	}

	// Second cell should be continuation
	cell = canvas.Get(1, 0)
	if !cell.Continuation {
		t.Errorf("cell at (1, 0): should be continuation")
	}

	// Next cells should be the text
	for i, ch := range "test" {
		cell := canvas.Get(2+i, 0)
		if cell.Text != string(ch) {
			t.Errorf("cell at (2+%d, 0): got %q, want %q", i, cell.Text, string(ch))
		}
	}
}

func TestCanvasWriteANSIWideCharacterClipping(t *testing.T) {
	canvas := loom.NewCanvas(5, 3)
	// Wide char at position 4 (would overflow to position 5)
	n := canvas.WriteANSI(3, 0, "ab👍cd")
	// "ab" = 2 cells, "👍" = 2 cells but would overflow, so stop at "ab" = 2
	// So we get 4 cells: a, b, at positions 3, 4, then clipping
	// Actually, WriteANSI should write "ab" and then try "👍" but can't fit 2 cells,
	// so it stops. Return is 2
	if n != 2 {
		t.Errorf("WriteANSI returned %d, want 2", n)
	}

	if canvas.Get(3, 0).Text != "a" {
		t.Error("cell at (3, 0) should be 'a'")
	}
	if canvas.Get(4, 0).Text != "b" {
		t.Error("cell at (4, 0) should be 'b'")
	}
}

func TestCanvasWriteANSIStyleReset(t *testing.T) {
	canvas := loom.NewCanvas(20, 3)
	// Write bold, then reset, then normal
	input := "\x1b[1mbold\x1b[0mnormal"
	n := canvas.WriteANSI(0, 0, input)
	if n != 10 { // 4 + 1 + 6 - 1 space = "bold" + "normal" = 10 chars
		t.Errorf("WriteANSI returned %d, want 10", n)
	}

	// Check first 4 cells are bold
	for i := 0; i < 4; i++ {
		cell := canvas.Get(i, 0)
		if !cell.Style.Bold {
			t.Errorf("cell %d: should be bold", i)
		}
	}

	// Check last 6 cells are not bold
	for i := 4; i < 10; i++ {
		cell := canvas.Get(i, 0)
		if cell.Style.Bold {
			t.Errorf("cell %d: should not be bold", i)
		}
	}
}

func TestCanvasWriteANSIOutOfBounds(t *testing.T) {
	canvas := loom.NewCanvas(10, 3)
	// Write completely outside the canvas
	n := canvas.WriteANSI(20, 20, "test")
	if n != 0 {
		t.Errorf("WriteANSI returned %d, want 0 for out-of-bounds", n)
	}

	// Write with negative position
	n = canvas.WriteANSI(-5, 0, "test")
	if n != 0 {
		t.Errorf("WriteANSI returned %d, want 0 for negative position", n)
	}
}

func TestCanvasWriteANSIRoundTrip(t *testing.T) {
	// Write ANSI string, then render back to ANSI and compare cells
	canvas := loom.NewCanvas(30, 3)
	input := "\x1b[31mRED\x1b[32mGREEN\x1b[0mNORMAL"
	_ = canvas.WriteANSI(0, 0, input)

	// Parse the rendered row
	rendered := canvas.Row(0)
	cells := loom.ParseANSI(rendered)

	// Should have text for RED, GREEN, NORMAL (11 characters)
	textCells := 0
	for _, cell := range cells {
		if !cell.Continuation && cell.Text != " " && cell.Text != "" {
			textCells++
		}
	}
	if textCells < 11 {
		t.Errorf("parsed cells has %d text cells, want at least 11", textCells)
	}
}

func TestCanvasWriteANSI256Color(t *testing.T) {
	canvas := loom.NewCanvas(30, 3)
	input := "\x1b[38;5;196mRED"
	n := canvas.WriteANSI(0, 0, input)
	if n != 3 {
		t.Errorf("WriteANSI returned %d, want 3", n)
	}

	// Check that color 196 is set
	for i := 0; i < 3; i++ {
		cell := canvas.Get(i, 0)
		if cell.Style.FG != loom.ColorIndex(196) {
			t.Errorf("cell %d: expected color index 196, got %#v", i, cell.Style.FG)
		}
	}
}

func TestCanvasWriteANSIRGBColor(t *testing.T) {
	canvas := loom.NewCanvas(30, 3)
	input := "\x1b[38;2;255;0;128mtext"
	n := canvas.WriteANSI(0, 0, input)
	if n != 4 {
		t.Errorf("WriteANSI returned %d, want 4", n)
	}

	// Check RGB color
	for i := 0; i < 4; i++ {
		cell := canvas.Get(i, 0)
		r, g, b, ok := cell.Style.FG.RGB()
		if !ok || r != 255 || g != 0 || b != 128 {
			t.Errorf("cell %d: expected RGB(255,0,128), got (%d,%d,%d)", i, r, g, b)
		}
	}
}

// ── Benchmark ───────────────────────────────────────────────────────────────

func BenchmarkCanvasWriteANSI(b *testing.B) {
	canvas := loom.NewCanvas(80, 24)
	input := "\x1b[31mERROR\x1b[0m: " +
		"\x1b[38;5;196mFailed\x1b[0m - " +
		"\x1b[1;34mDetails\x1b[0m: " +
		"Some error message with lots of words that goes on for a bit"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = canvas.WriteANSI(0, 0, input)
	}
}

// ── Parity tests between old and new Canvas WriteANSI / set paths ───────────

func TestCanvasWriteANSIOldNewParity(t *testing.T) {
	inputs := []string{
		"hello world",
		"\x1b[31mRED\x1b[0m \x1b[32mGREEN\x1b[0m",
		"\x1b[1mbold\x1b[0mnormal",
		"👍test emoji",
		"\x1b[38;2;255;0;128mRGB test\x1b[0m",
	}

	for i, input := range inputs {
		cOld := loom.NewCanvas(30, 5)
		cNew := loom.NewCanvas(30, 5)

		t.Setenv("LOOM_FAST_ANSI", "0")
		_ = cOld.WriteANSI(0, 0, input)

		t.Setenv("LOOM_FAST_ANSI", "1")
		_ = cNew.WriteANSI(0, 0, input)

		rowOld := cOld.Row(0)
		rowNew := cNew.Row(0)

		if rowOld != rowNew {
			t.Errorf("case %d (%q): rendered row mismatch:\n old: %q\n new: %q", i, input, rowOld, rowNew)
		}
	}
}

func BenchmarkCanvasWriteANSI_Old(b *testing.B) {
	b.Setenv("LOOM_FAST_ANSI", "0")
	canvas := loom.NewCanvas(80, 24)
	input := "\x1b[31mERROR\x1b[0m: " +
		"\x1b[38;5;196mFailed\x1b[0m - " +
		"\x1b[1;34mDetails\x1b[0m: " +
		"Some error message with lots of words that goes on for a bit"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = canvas.WriteANSI(0, 0, input)
	}
}

func BenchmarkCanvasWriteANSI_New(b *testing.B) {
	b.Setenv("LOOM_FAST_ANSI", "1")
	canvas := loom.NewCanvas(80, 24)
	input := "\x1b[31mERROR\x1b[0m: " +
		"\x1b[38;5;196mFailed\x1b[0m - " +
		"\x1b[1;34mDetails\x1b[0m: " +
		"Some error message with lots of words that goes on for a bit"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = canvas.WriteANSI(0, 0, input)
	}
}
