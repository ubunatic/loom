// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"ubunatic.com/loom"
)

func TestCanvasSubCanvasUsesLocalOriginAndCopiesCells(t *testing.T) {
	parent := loom.NewCanvas(6, 3)
	style := loom.Style{FG: loom.ColorIndex(2), BG: loom.ColorIndex(4), Bold: true}
	parent.Set(3, 1, loom.Cell{Text: "A", Style: style})

	sub := parent.SubCanvas(loom.Rect{X: 2, Y: 1, W: 3, H: 2})
	if sub.Cols() != 3 || sub.Rows() != 2 {
		t.Fatalf("SubCanvas size = %dx%d, want 3x2", sub.Cols(), sub.Rows())
	}
	if got := sub.Get(1, 0); got.Text != "A" || got.Style != style {
		t.Fatalf("SubCanvas local cell = %#v, want styled A", got)
	}
	sub.Set(0, 1, loom.Cell{Text: "B"})
	if got := parent.Get(2, 2); got.Text != " " {
		t.Fatalf("mutating subcanvas changed parent: %#v", got)
	}
}

func TestCanvasBlitPreservesStyledWideCells(t *testing.T) {
	src := loom.NewCanvas(4, 1)
	style := loom.Style{FG: loom.ColorIndex(5), BG: loom.ColorIndex(1), Underline: true}
	src.Set(0, 0, loom.Cell{Text: "界", Style: style})
	src.Set(2, 0, loom.Cell{Text: "x", Style: style})
	dst := loom.NewCanvas(5, 1)

	dst.Blit(src, 1, 0)
	lead, continuation := dst.Get(1, 0), dst.Get(2, 0)
	if lead.Text != "界" || lead.Style != style {
		t.Fatalf("blitted lead = %#v, want styled wide rune", lead)
	}
	if !continuation.Continuation || continuation.Style != style {
		t.Fatalf("blitted continuation = %#v, want styled continuation", continuation)
	}
	if got := dst.Get(3, 0); got.Text != "x" || got.Style != style {
		t.Fatalf("blitted trailing cell = %#v, want styled x", got)
	}
}

func TestCanvasBlitClipsWithoutOrphaningWideCells(t *testing.T) {
	src := loom.NewCanvas(3, 1)
	src.Set(0, 0, loom.Cell{Text: "界"})
	dst := loom.NewCanvas(2, 1)
	dst.Blit(src, -1, 0)
	for x := 0; x < dst.Cols(); x++ {
		if got := dst.Get(x, 0); got.Continuation || got.Text != " " {
			t.Fatalf("clipped wide glyph left malformed cell at %d: %#v", x, got)
		}
	}
	dst.Blit(src, 1, 0)
	if got := dst.Get(1, 0); got.Continuation || got.Text != " " {
		t.Fatalf("right-clipped wide glyph wrote partial lead: %#v", got)
	}
}

func TestCanvasSubCanvasAndBlitCursor(t *testing.T) {
	parent := loom.NewCanvas(10, 5)
	parent.CursorX, parent.CursorY = 4, 2

	sub := parent.SubCanvas(loom.Rect{X: 2, Y: 1, W: 5, H: 3})
	if sub.CursorX != 2 || sub.CursorY != 1 {
		t.Fatalf("SubCanvas cursor = (%d,%d), want (2,1)", sub.CursorX, sub.CursorY)
	}

	outside := parent.SubCanvas(loom.Rect{X: 5, Y: 0, W: 3, H: 2})
	if outside.CursorX != -1 || outside.CursorY != -1 {
		t.Fatalf("SubCanvas outside cursor = (%d,%d), want (-1,-1)", outside.CursorX, outside.CursorY)
	}

	dst := loom.NewCanvas(12, 6)
	dst.Blit(sub, 3, 1)
	if dst.CursorX != 5 || dst.CursorY != 2 {
		t.Fatalf("Blit cursor = (%d,%d), want (5,2)", dst.CursorX, dst.CursorY)
	}
}
