// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestRichTextEditDrawsStyledSpansAndCursor(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{
		{Text: "normal "},
		{Text: "bold", Style: Style{Bold: true, FG: ColorIndex(3)}},
	}}}}
	edit := NewRichTextEdit(doc)
	edit.Cursor.Offset = 11
	canvas := NewCanvas(12, 2)
	edit.Draw(canvas, Rect{X: 2, Y: 1, W: 8, H: 1})
	if got := canvas.Get(2, 1).Text; got != "a" {
		t.Fatalf("first visible cell = %q, want a", got)
	}
	if cell := canvas.Get(5, 1); cell.Text != "b" || !cell.Style.Bold || cell.Style.FG != ColorIndex(3) {
		t.Fatalf("styled cell = %#v", cell)
	}
	if canvas.CursorX != 9 || canvas.CursorY != 1 {
		t.Fatalf("cursor = (%d,%d), want (9,1)", canvas.CursorX, canvas.CursorY)
	}
}

func TestRichTextEditPaintsSelectionAcrossSpans(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{
		{Text: "abc"},
		{Text: "def", Style: Style{Italic: true, FG: ColorIndex(2)}},
	}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{Line: 0, Offset: 1}, RichPosition{Line: 0, Offset: 5})
	canvas := NewCanvas(8, 1)
	edit.Draw(canvas, Rect{W: 8, H: 1})
	if got := canvas.Get(0, 0).Style.BG; got != ColorReset() {
		t.Fatalf("unselected background = %v, want reset", got)
	}
	for _, x := range []int{1, 2, 3, 4} {
		if got := canvas.Get(x, 0).Style.BG; got != ColorIndex(24) {
			t.Errorf("selected cell %d background = %v", x, got)
		}
	}
	if !canvas.Get(3, 0).Style.Italic || canvas.Get(3, 0).Style.FG != ColorIndex(2) {
		t.Fatalf("selection erased span style: %#v", canvas.Get(3, 0).Style)
	}
}

func TestRichTextEditScrollsCursorIntoViewport(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{
		{Spans: []RichSpan{{Text: "0123456789"}}},
		{Spans: []RichSpan{{Text: "second"}}},
		{Spans: []RichSpan{{Text: "third"}}},
	}}
	edit := NewRichTextEdit(doc)
	edit.Cursor = RichPosition{Line: 2, Offset: 5}
	canvas := NewCanvas(5, 2)
	edit.Draw(canvas, Rect{W: 5, H: 2})
	if edit.ScrollY != 1 || edit.ScrollX != 1 {
		t.Fatalf("scroll = (%d,%d), want (1,1)", edit.ScrollX, edit.ScrollY)
	}
	if canvas.Get(0, 0).Text != "e" || canvas.Get(0, 1).Text != "h" {
		t.Fatalf("visible cells = %q, %q", canvas.Get(0, 0).Text, canvas.Get(0, 1).Text)
	}
	if canvas.CursorX != 4 || canvas.CursorY != 1 {
		t.Fatalf("cursor = (%d,%d), want (4,1)", canvas.CursorX, canvas.CursorY)
	}
}

func TestRichTextEditClipsToAssignedRectAndHidesUnfocusedCursor(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "abcdef"}}}}})
	edit.Cursor.Offset = 2
	edit.SetFocus(false)
	canvas := NewCanvas(10, 1)
	canvas.Set(6, 0, Cell{Text: "x"})
	edit.Draw(canvas, Rect{X: 1, W: 3, H: 1})
	if canvas.Get(4, 0).Text != " " || canvas.Get(6, 0).Text != "x" {
		t.Fatalf("draw crossed widget bounds: right cell=%q outside cell=%q", canvas.Get(4, 0).Text, canvas.Get(6, 0).Text)
	}
	if canvas.CursorX != -1 || canvas.CursorY != -1 {
		t.Fatalf("unfocused draw set cursor to (%d,%d)", canvas.CursorX, canvas.CursorY)
	}
}
