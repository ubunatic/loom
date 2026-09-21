// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── M1: Canvas.DrawBorder basic functionality ──────────────────────────────────

func TestDrawBorderSharpStyle(t *testing.T) {
	canvas := loom.NewCanvas(10, 5)
	r := loom.Rect{X: 1, Y: 1, W: 5, H: 3}
	canvas.DrawBorder(r, loom.BoxBorderStyleSharp, loom.Reset)

	// Check corners
	if canvas.Get(1, 1).Text != "┌" {
		t.Errorf("top-left corner: got %q, want ┌", canvas.Get(1, 1).Text)
	}
	if canvas.Get(5, 1).Text != "┐" {
		t.Errorf("top-right corner: got %q, want ┐", canvas.Get(5, 1).Text)
	}
	if canvas.Get(1, 3).Text != "└" {
		t.Errorf("bottom-left corner: got %q, want └", canvas.Get(1, 3).Text)
	}
	if canvas.Get(5, 3).Text != "┘" {
		t.Errorf("bottom-right corner: got %q, want ┘", canvas.Get(5, 3).Text)
	}

	// Check edges
	if canvas.Get(2, 1).Text != "─" {
		t.Errorf("top edge: got %q, want ─", canvas.Get(2, 1).Text)
	}
	if canvas.Get(1, 2).Text != "│" {
		t.Errorf("left edge: got %q, want │", canvas.Get(1, 2).Text)
	}
}

func TestDrawBorderRoundedStyle(t *testing.T) {
	canvas := loom.NewCanvas(10, 5)
	r := loom.Rect{X: 1, Y: 1, W: 5, H: 3}
	canvas.DrawBorder(r, loom.BoxBorderStyleRounded, loom.Reset)

	// Check corners
	if canvas.Get(1, 1).Text != "╭" {
		t.Errorf("top-left corner: got %q, want ╭", canvas.Get(1, 1).Text)
	}
	if canvas.Get(5, 1).Text != "╮" {
		t.Errorf("top-right corner: got %q, want ╮", canvas.Get(5, 1).Text)
	}
	if canvas.Get(1, 3).Text != "╰" {
		t.Errorf("bottom-left corner: got %q, want ╰", canvas.Get(1, 3).Text)
	}
	if canvas.Get(5, 3).Text != "╯" {
		t.Errorf("bottom-right corner: got %q, want ╯", canvas.Get(5, 3).Text)
	}
}

func TestDrawBorderDoubleStyle(t *testing.T) {
	canvas := loom.NewCanvas(10, 5)
	r := loom.Rect{X: 1, Y: 1, W: 5, H: 3}
	canvas.DrawBorder(r, loom.BoxBorderStyleDouble, loom.Reset)

	// Check corners
	if canvas.Get(1, 1).Text != "╔" {
		t.Errorf("top-left corner: got %q, want ╔", canvas.Get(1, 1).Text)
	}
	if canvas.Get(1, 1).Text != "╔" {
		t.Errorf("top-left corner: got %q, want ╔", canvas.Get(1, 1).Text)
	}
	// Check horizontal edge (should be ═)
	if canvas.Get(2, 1).Text != "═" {
		t.Errorf("top edge: got %q, want ═", canvas.Get(2, 1).Text)
	}
	// Check vertical edge (should be ║)
	if canvas.Get(1, 2).Text != "║" {
		t.Errorf("left edge: got %q, want ║", canvas.Get(1, 2).Text)
	}
}

func TestDrawBorderASCIIStyle(t *testing.T) {
	canvas := loom.NewCanvas(10, 5)
	r := loom.Rect{X: 1, Y: 1, W: 5, H: 3}
	canvas.DrawBorder(r, loom.BoxBorderStyleASCII, loom.Reset)

	// Check corners
	if canvas.Get(1, 1).Text != "+" {
		t.Errorf("top-left corner: got %q, want +", canvas.Get(1, 1).Text)
	}
	// Check edges
	if canvas.Get(2, 1).Text != "-" {
		t.Errorf("top edge: got %q, want -", canvas.Get(2, 1).Text)
	}
	if canvas.Get(1, 2).Text != "|" {
		t.Errorf("left edge: got %q, want |", canvas.Get(1, 2).Text)
	}
}

func TestDrawBorderDegenerateSizes(t *testing.T) {
	tests := []struct {
		name string
		w, h int
	}{
		{"zero width", 0, 3},
		{"zero height", 3, 0},
		{"1x1", 1, 1},
		{"1x3", 1, 3},
		{"3x1", 3, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := loom.NewCanvas(10, 5)
			r := loom.Rect{X: 1, Y: 1, W: tt.w, H: tt.h}
			canvas.DrawBorder(r, loom.BoxBorderStyleSharp, loom.Reset)
			// Should not panic and leave canvas mostly blank
			if canvas.Get(1, 1).Text == "┌" && tt.w < 2 {
				t.Errorf("border drawn for %dx%d, should not draw", tt.w, tt.h)
			}
		})
	}
}

func TestDrawBorderClipping(t *testing.T) {
	canvas := loom.NewCanvas(5, 5)
	// Request a border that extends beyond canvas bounds
	r := loom.Rect{X: 2, Y: 2, W: 5, H: 5}
	canvas.DrawBorder(r, loom.BoxBorderStyleSharp, loom.Reset)

	// Top-left should be at (2, 2)
	if canvas.Get(2, 2).Text != "┌" {
		t.Errorf("top-left corner clipped: got %q, want ┌", canvas.Get(2, 2).Text)
	}

	// Out-of-bounds writes should be silently dropped
	// Canvas is 5×5, so position (7, 7) would be out of bounds
	// This is just verifying that Set handles out-of-bounds gracefully
}

// ── M1: Canvas.DrawBox with titles ────────────────────────────────────────────

func TestDrawBoxWithSimpleTitle(t *testing.T) {
	canvas := loom.NewCanvas(20, 5)
	r := loom.Rect{X: 1, Y: 1, W: 10, H: 3}
	canvas.DrawBox(r, loom.BoxBorderStyleSharp, "Test", loom.Reset)

	// Check that border is drawn
	if canvas.Get(1, 1).Text != "┌" {
		t.Errorf("border not drawn")
	}

	// Check title is written (should start at x=2, y=1)
	// "Test" becomes " Test " in DrawBox
	title := " Test "
	for i, ch := range title {
		cell := canvas.Get(2+i, 1)
		if cell.Text != string(ch) {
			t.Errorf("title at position %d: got %q, want %q", i, cell.Text, string(ch))
		}
	}
}

func TestDrawBoxWithLongTitle(t *testing.T) {
	canvas := loom.NewCanvas(10, 5)
	r := loom.Rect{X: 1, Y: 1, W: 8, H: 3}
	canvas.DrawBox(r, loom.BoxBorderStyleSharp, "VeryLongTitle", loom.Reset)

	// Title should be truncated with ellipsis
	// Available width is 8 - 2 = 6 cells (excluding corners)
	// " VeryLongTitle " is too long, so it should be truncated
	// The border should be drawn regardless
	if canvas.Get(1, 1).Text != "┌" {
		t.Errorf("border not drawn for long title")
	}
}

func TestDrawBoxWithEmptyTitle(t *testing.T) {
	canvas := loom.NewCanvas(10, 5)
	r := loom.Rect{X: 1, Y: 1, W: 5, H: 3}
	canvas.DrawBox(r, loom.BoxBorderStyleSharp, "", loom.Reset)

	// Border should be drawn without title
	if canvas.Get(1, 1).Text != "┌" {
		t.Errorf("border not drawn for empty title")
	}
}

func TestDrawBoxMultiByteTitle(t *testing.T) {
	canvas := loom.NewCanvas(30, 5)
	r := loom.Rect{X: 1, Y: 1, W: 15, H: 3}
	// Test with multi-byte characters (CJK)
	canvas.DrawBox(r, loom.BoxBorderStyleSharp, "你好", loom.Reset)

	// Border should be drawn
	if canvas.Get(1, 1).Text != "┌" {
		t.Errorf("border not drawn for multi-byte title")
	}
	// Title should be written (handling wide characters correctly)
}

func TestDrawBoxDegenerateSizes(t *testing.T) {
	tests := []struct {
		name string
		w, h int
	}{
		{"0x0", 0, 0},
		{"1x1", 1, 1},
		{"1x3", 1, 3},
		{"3x1", 3, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := loom.NewCanvas(10, 5)
			r := loom.Rect{X: 1, Y: 1, W: tt.w, H: tt.h}
			canvas.DrawBox(r, loom.BoxBorderStyleSharp, "Test", loom.Reset)
			// Should not panic
		})
	}
}

// ── M3: Popup.Draw output equivalence with Canvas.DrawBox ────────────────────

// TestPopupDrawEqualsCanvasDrawBox verifies that Popup.Draw produces identical
// output to Canvas.DrawBox(..., BoxBorderStyleSharp, ...) cell-by-cell.
// This confirms the migrated Popup implementation is correct.
func TestPopupDrawEqualsCanvasDrawBox(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{"simple title", "Test"},
		{"short title", "OK"},
		{"truncated title", "This is a very long title that should be truncated"},
		{"CJK title", "你好"},
		{"mixed CJK truncated", "你好世界大同小异"},
		{"empty title", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, rows := 40, 10

			// Draw using Popup.Draw
			popupCanvas := loom.NewCanvas(cols, rows)
			popup := loom.NewPopup(tt.title, &loom.Grid{})
			popup.Width = 30
			popup.Height = 6
			popup.Style = loom.Reset
			popup.Draw(popupCanvas, loom.Rect{X: 0, Y: 0, W: cols, H: rows})

			// Draw using Canvas.DrawBox directly
			boxCanvas := loom.NewCanvas(cols, rows)
			// Popup centers itself, calculate centered position
			x := (cols - popup.Width) / 2
			y := (rows - popup.Height) / 2
			boxCanvas.DrawBox(loom.Rect{X: x, Y: y, W: popup.Width, H: popup.Height},
				loom.BoxBorderStyleSharp, tt.title, loom.Reset)

			// Compare cell by cell
			for cy := 0; cy < rows; cy++ {
				for cx := 0; cx < cols; cx++ {
					popupCell := popupCanvas.Get(cx, cy)
					boxCell := boxCanvas.Get(cx, cy)

					if popupCell.Text != boxCell.Text {
						t.Errorf("cell (%d,%d) text mismatch: Popup.Draw got %q, DrawBox got %q",
							cx, cy, popupCell.Text, boxCell.Text)
					}
					// Note: Style comparison skipped as Popup may set different default styles
				}
			}
		})
	}
}
