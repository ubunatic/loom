// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestStyledRowsDrawsANSIStylesAndResetsEachRow(t *testing.T) {
	rows := loom.NewStyledRows("\x1b[31mred", "plain")
	canvas := loom.NewCanvas(8, 2)
	rows.Draw(canvas, canvas.Bounds())

	if got := canvas.Get(0, 0).Style.FG; got != loom.ColorIndex(1) {
		t.Fatalf("styled row foreground = %#v, want red", got)
	}
	if got := canvas.Get(0, 1).Style; got != loom.Reset {
		t.Fatalf("style leaked into next row: got %#v, want reset", got)
	}
}

func TestStyledRowsRoundTripsANSIStyle(t *testing.T) {
	canvas := loom.NewCanvas(6, 1)
	loom.NewStyledRows("\x1b[31mred\x1b[0m").Draw(canvas, canvas.Bounds())
	rendered := loom.ParseANSI(canvas.Row(0))
	var text strings.Builder
	for _, cell := range rendered {
		if !cell.Continuation {
			text.WriteString(cell.Text)
		}
	}
	if got := strings.TrimRight(text.String(), " "); got != "red" {
		t.Fatalf("rendered text = %q, want %q", got, "red")
	}
	if got := canvas.Get(0, 0).Style.FG; got != loom.ColorIndex(1) {
		t.Fatalf("rendered foreground = %#v, want red", got)
	}
}

func TestStyledRowsDropsMalformedAndUnsupportedEscapes(t *testing.T) {
	for _, input := range []string{
		"\x1b[31mred",
		"ab\x1b[3",
		"\x1b[2Jab",
		"\x1b]0;title\x07ab",
	} {
		t.Run(input, func(t *testing.T) {
			canvas := loom.NewCanvas(8, 1)
			loom.NewStyledRows(input).Draw(canvas, canvas.Bounds())
			for x := 0; x < canvas.Cols(); x++ {
				cell := canvas.Get(x, 0)
				if strings.ContainsAny(cell.Text, "\x1b[]") {
					t.Fatalf("escape syntax rendered as text at x=%d: %#v", x, cell)
				}
				if x >= 3 && cell.Style.FG == loom.ColorIndex(1) {
					t.Fatalf("style leaked past visible text at x=%d: %#v", x, cell)
				}
			}
		})
	}
}

func TestStyledRowsRedrawClearsOldCells(t *testing.T) {
	canvas := loom.NewCanvas(6, 2)
	rows := loom.NewStyledRows("\x1b[31mlong", "second")
	rows.Draw(canvas, canvas.Bounds())
	rows.Lines = []string{"x"}
	rows.Draw(canvas, canvas.Bounds())
	for y := 0; y < canvas.Rows(); y++ {
		for x := 0; x < canvas.Cols(); x++ {
			cell := canvas.Get(x, y)
			if (y == 0 && x == 0 && cell.Text != "x") ||
				((y != 0 || x != 0) && (cell.Text != " " || cell.Style != loom.Reset)) {
				t.Fatalf("stale cell at (%d,%d): %#v", x, y, cell)
			}
		}
	}
}

func TestStyledRowsClipsToRectByDisplayWidth(t *testing.T) {
	rows := loom.NewStyledRows("ab👍z")
	canvas := loom.NewCanvas(8, 1)
	rows.Draw(canvas, loom.Rect{X: 2, Y: 0, W: 3, H: 1})

	if canvas.Get(2, 0).Text != "a" || canvas.Get(3, 0).Text != "b" {
		t.Fatalf("visible prefix = %q%q, want ab", canvas.Get(2, 0).Text, canvas.Get(3, 0).Text)
	}
	if canvas.Get(4, 0).Text != " " {
		t.Fatalf("wide cluster was partially drawn: cell 4 = %#v", canvas.Get(4, 0))
	}
	if canvas.Get(5, 0).Text != " " {
		t.Fatalf("draw escaped its rect: cell 5 = %#v", canvas.Get(5, 0))
	}
}

func TestStyledRowsClipsHeight(t *testing.T) {
	rows := loom.NewStyledRows("first", "second")
	canvas := loom.NewCanvas(8, 2)
	rows.Draw(canvas, loom.Rect{X: 0, Y: 1, W: 8, H: 1})
	if canvas.Get(0, 0).Text != " " || canvas.Get(0, 1).Text != "f" {
		t.Fatalf("rows were not clipped to rect: top=%q bottom=%q", canvas.Get(0, 0).Text, canvas.Get(0, 1).Text)
	}
}
