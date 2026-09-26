// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
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
