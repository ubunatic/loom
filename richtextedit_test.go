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
		if got := canvas.Get(x, 0).Style.BG; got != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SelectionBG)) {
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

func TestRichTextEditTypingNewlinesAndLineJoin(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "abcd"}}}}}
	edit := NewRichTextEdit(doc)
	edit.Cursor.Offset = 2
	edit.ConsumeKey(KeyEvent{Text: "XY"})
	if got := doc.ToPlainText(); got != "abXYcd" || edit.Cursor.Offset != 4 {
		t.Fatalf("typed doc/cursor = %q/%+v", got, edit.Cursor)
	}
	edit.ConsumeKey(KeyEvent{Key: "enter"})
	edit.ConsumeKey(KeyEvent{Text: "z"})
	if got := doc.ToPlainText(); got != "abXY\nzcd" {
		t.Fatalf("newline insertion = %q", got)
	}
	edit.ConsumeKey(KeyEvent{Key: "backspace"})
	if got := doc.ToPlainText(); got != "abXY\ncd" {
		t.Fatalf("backspace = %q", got)
	}
	edit.Cursor = RichPosition{Line: 1}
	edit.ConsumeKey(KeyEvent{Key: "backspace"})
	if got := doc.ToPlainText(); got != "abXYcd" || edit.Cursor != (RichPosition{Line: 0, Offset: 4}) {
		t.Fatalf("line join = %q, cursor %+v", got, edit.Cursor)
	}
}

func TestRichTextEditDeletesPillsAtomically(t *testing.T) {
	newDoc := func() *RichDocument {
		return &RichDocument{Lines: []RichLine{{Spans: []RichSpan{
			{Text: "a"},
			{Text: "@team-core", PillData: &RichPill{Kind: "mention", ID: "team-core"}},
			{Text: "b"},
		}}}}
	}
	t.Run("backspace", func(t *testing.T) {
		doc := newDoc()
		edit := NewRichTextEdit(doc)
		edit.Cursor.Offset = 10
		edit.ConsumeKey(KeyEvent{Key: "backspace"})
		if got := doc.ToPlainText(); got != "ab" {
			t.Fatalf("backspace split pill: %q", got)
		}
	})
	t.Run("delete", func(t *testing.T) {
		doc := newDoc()
		edit := NewRichTextEdit(doc)
		edit.Cursor.Offset = 1
		edit.ConsumeKey(KeyEvent{Key: "delete"})
		if got := doc.ToPlainText(); got != "ab" {
			t.Fatalf("delete split pill: %q", got)
		}
	})
	t.Run("selection", func(t *testing.T) {
		doc := newDoc()
		edit := NewRichTextEdit(doc)
		edit.SetSelection(RichPosition{Offset: 3}, RichPosition{Offset: 4})
		edit.ConsumeKey(KeyEvent{Key: "delete"})
		if got := doc.ToPlainText(); got != "ab" {
			t.Fatalf("selection split pill: %q", got)
		}
	})
}

func TestRichTextEditNavigationAndSelection(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{
		{Spans: []RichSpan{{Text: "one two"}}},
		{Spans: []RichSpan{{Text: "next"}}},
	}}
	edit := NewRichTextEdit(doc)
	edit.ConsumeKey(KeyEvent{Key: "ctrl-right"})
	if edit.Cursor.Offset != 4 {
		t.Fatalf("ctrl-right cursor = %+v, want offset 4", edit.Cursor)
	}
	edit.ConsumeKey(KeyEvent{Key: "ctrl-left"})
	if edit.Cursor.Offset != 0 {
		t.Fatalf("ctrl-left cursor = %+v, want offset 0", edit.Cursor)
	}
	edit.ConsumeKey(KeyEvent{Key: "shift-right"})
	edit.ConsumeKey(KeyEvent{Key: "shift-right"})
	if !edit.HasSelection || edit.SelectionFrom.Offset != 0 || edit.SelectionTo.Offset != 2 {
		t.Fatalf("shift selection = %+v..%+v active=%v", edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
	edit.ConsumeKey(KeyEvent{Key: "shift-left"})
	if !edit.HasSelection || edit.Cursor.Offset != 1 {
		t.Fatalf("reversed selection = %+v active=%v", edit.Cursor, edit.HasSelection)
	}
	edit.ConsumeKey(KeyEvent{Key: "left"})
	if edit.HasSelection || edit.Cursor.Offset != 0 {
		t.Fatalf("left should collapse selection to start: %+v active=%v", edit.Cursor, edit.HasSelection)
	}
	edit.ConsumeKey(KeyEvent{Key: "end"})
	edit.ConsumeKey(KeyEvent{Key: "down"})
	if edit.Cursor != (RichPosition{Line: 1, Offset: 4}) {
		t.Fatalf("vertical navigation cursor = %+v", edit.Cursor)
	}
}

func TestRichTextEditFormattingShortcuts(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello world"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 5})
	for _, key := range []string{"ctrl-b", "ctrl-i", "ctrl-u"} {
		edit.ConsumeKey(KeyEvent{Key: key})
	}
	if len(doc.Lines[0].Spans) != 2 {
		t.Fatalf("formatting should split spans: %#v", doc.Lines[0].Spans)
	}
	styled := doc.Lines[0].Spans[0].Style
	if !styled.Bold || !styled.Italic || !styled.Underline || doc.Lines[0].Spans[1].Style != (Style{}) {
		t.Fatalf("formatting range = %#v", doc.Lines[0].Spans)
	}
	edit.ClearSelection()
	edit.Cursor.Offset = 11
	for _, key := range []string{"ctrl-b", "ctrl-i", "ctrl-u"} {
		edit.ConsumeKey(KeyEvent{Key: key})
	}
	edit.ConsumeKey(KeyEvent{Text: "!"})
	last := doc.Lines[0].Spans[len(doc.Lines[0].Spans)-1]
	if last.Text != "!" || !last.Style.Bold || !last.Style.Italic || !last.Style.Underline {
		t.Fatalf("active style insertion = %#v", last)
	}
}

func TestRichTextEditMouseClickAndDragSelection(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "abcdef"}}}, {Spans: []RichSpan{{Text: "ghij"}}}}}
	edit := NewRichTextEdit(doc)
	canvas := NewCanvas(14, 4)
	edit.Draw(canvas, Rect{X: 3, Y: 1, W: 8, H: 2})
	if got := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 0}); !got.Consumed || edit.Cursor != (RichPosition{Offset: 2}) {
		t.Fatalf("mouse click result=%+v cursor=%+v", got, edit.Cursor)
	}
	edit.ConsumeMouse(MouseEvent{Action: MouseDrag, Button: MouseLeft, X: 3, Y: 1})
	edit.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 3, Y: 1})
	if !edit.HasSelection || edit.SelectionFrom != (RichPosition{Offset: 2}) || edit.SelectionTo != (RichPosition{Line: 1, Offset: 3}) {
		t.Fatalf("drag selection = %+v..%+v active=%v", edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
	edit.Draw(canvas, Rect{X: 3, Y: 1, W: 8, H: 2})
	if canvas.Get(5, 1).Style.BG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SelectionBG)) || canvas.Get(3, 2).Style.BG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SelectionBG)) {
		t.Fatal("mouse selection was not painted across lines")
	}
}

func TestRichTextEditPopoverGeometryAndFormatActions(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{
		{Spans: []RichSpan{{Text: "first"}}},
		{Spans: []RichSpan{{Text: "second"}}},
		{Spans: []RichSpan{{Text: "some selected text"}}},
		{Spans: []RichSpan{{Text: "last"}}},
	}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{Line: 2, Offset: 5}, RichPosition{Line: 2, Offset: 13})
	canvas := NewCanvas(50, 6)
	edit.Draw(canvas, Rect{X: 3, Y: 1, W: 42, H: 5})
	if len(edit.popoverButtons) != 7 {
		t.Fatalf("popover buttons = %d, want 7", len(edit.popoverButtons))
	}
	for i, button := range edit.popoverButtons {
		if button.label != SpeccedDefaults.RichTextEdit.PopoverLabels[i] {
			t.Fatalf("popover label %d = %q, want %q", i, button.label, SpeccedDefaults.RichTextEdit.PopoverLabels[i])
		}
	}
	first := edit.popoverButtons[0].rect
	if first.Y != 0 || canvas.Get(3+first.X-1, 1+first.Y).Text != "[" {
		t.Fatalf("popover top-left = %+v, expected above selection in child-local coordinates", first)
	}
	toolbar := canvas.Get(3+first.X+1, 1+first.Y)
	if toolbar.Style.FG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.ToolbarFG)) || toolbar.Style.BG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.ToolbarBG)) {
		t.Fatalf("toolbar style = %+v", toolbar.Style)
	}
	separator := canvas.Get(3+first.X+first.W, 1+first.Y)
	if separator.Text != SpeccedDefaults.RichTextEdit.SeparatorGlyph || separator.Style.FG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SeparatorFG)) || separator.Style.BG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SeparatorBG)) {
		t.Fatalf("separator cell = %+v", separator)
	}
	downPointer := canvas.Get(3+first.X+4, 1+first.Y+1)
	if downPointer.Text != SpeccedDefaults.RichTextEdit.PointerDownGlyph || downPointer.Style.FG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.PointerFG)) {
		t.Fatal("popover did not draw a downward anchor between toolbar and selection")
	}
	selection := edit.SelectionFrom
	bold := edit.popoverButtons[0].rect
	if result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: bold.X + 1, Y: bold.Y}); !result.Consumed {
		t.Fatal("bold button click was not consumed")
	}
	if len(doc.Lines[2].Spans) < 2 || !doc.Lines[2].Spans[1].Style.Bold || edit.SelectionFrom != selection || !edit.HasSelection {
		t.Fatalf("bold action/selection = %#v, %+v..%+v active=%v", doc.Lines[2].Spans, edit.SelectionFrom, edit.SelectionTo, edit.HasSelection)
	}
	if result := edit.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: bold.X + 1, Y: bold.Y}); !result.Consumed {
		t.Fatal("popover button release was not consumed")
	}
	for _, label := range []string{"I", "U", "S"} {
		for _, button := range edit.popoverButtons {
			if button.label == label {
				edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: button.rect.X + 1, Y: button.rect.Y})
				break
			}
		}
	}
	if style := doc.Lines[2].Spans[1].Style; !style.Bold || !style.Italic || !style.Underline || !style.Strike {
		t.Fatalf("format action styles = %+v", style)
	}
}

func TestRichTextEditPopoverFallsBelowAndConsumesPlaceholders(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{Offset: 1}, RichPosition{Offset: 6})
	canvas := NewCanvas(48, 5)
	edit.Draw(canvas, Rect{W: 40, H: 5})
	first := edit.popoverButtons[0].rect
	upPointer := canvas.Get(first.X, 1)
	if first.Y != 2 || upPointer.Text != SpeccedDefaults.RichTextEdit.PointerUpGlyph || upPointer.Style.FG != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.PointerFG)) {
		t.Fatalf("popover failed below-selection placement: first=%+v pointer=%q", first, canvas.Get(first.X, 1).Text)
	}
	for _, label := range []string{"Link", "#FG", "#BG"} {
		for _, button := range edit.popoverButtons {
			if button.label == label {
				result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: button.rect.X + 1, Y: button.rect.Y})
				if !result.Consumed || !edit.HasSelection {
					t.Fatalf("placeholder %s result=%+v selection=%v", label, result, edit.HasSelection)
				}
				break
			}
		}
	}
	edit.SetSelection(RichPosition{}, RichPosition{})
	edit.Draw(canvas, Rect{W: 40, H: 5})
	if len(edit.popoverButtons) != 0 {
		t.Fatal("popover remained visible after selection collapsed")
	}
}

func TestRichTextEditPopoverFGAndBGPalettesStyleSelectedSpans(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{
		{Text: "ab", Style: Style{FG: ColorIndex(2), BG: ColorIndex(3)}},
		{Text: "cdef", Style: Style{FG: ColorIndex(4), BG: ColorIndex(5), Italic: true}},
		{Text: "ghij", Style: Style{FG: ColorIndex(6), BG: ColorIndex(7), Underline: true}},
	}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{Offset: 1}, RichPosition{Offset: 8})
	canvas := NewCanvas(50, 5)
	edit.Draw(canvas, Rect{W: 40, H: 5})

	clickButton := func(label string) {
		t.Helper()
		for _, button := range edit.popoverButtons {
			if button.label == label {
				if result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: button.rect.X + 1, Y: button.rect.Y}); !result.Consumed {
					t.Fatalf("%s click was not consumed", label)
				}
				return
			}
		}
		t.Fatalf("popover button %s not found", label)
	}
	clickSwatch := func(index int) {
		t.Helper()
		if len(edit.popoverSwatches) != 16 {
			t.Fatalf("palette swatches = %d, want 16", len(edit.popoverSwatches))
		}
		swatch := edit.popoverSwatches[index]
		if result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: swatch.rect.X, Y: swatch.rect.Y}); !result.Consumed {
			t.Fatalf("color %d click was not consumed", index)
		}
		if edit.popoverPalette != "" {
			t.Fatalf("palette remained open after choosing color %d", index)
		}
	}
	styleAt := func(offset int) Style {
		for _, span := range doc.Lines[0].Spans {
			length := len([]rune(span.Text))
			if offset < length {
				return span.Style
			}
			offset -= length
		}
		t.Fatalf("no span at offset")
		return Style{}
	}

	clickButton("#FG")
	edit.Draw(canvas, Rect{W: 40, H: 5})
	clickSwatch(10)
	if got := styleAt(0).FG; got != ColorIndex(2) {
		t.Fatalf("unselected prefix FG = %v, want 2", got)
	}
	for _, offset := range []int{1, 2, 5, 6, 7} {
		if got := styleAt(offset).FG; got != ColorIndex(10) {
			t.Errorf("selected offset %d FG = %v, want 10", offset, got)
		}
	}
	if got := styleAt(8).FG; got != ColorIndex(6) {
		t.Fatalf("unselected suffix FG = %v, want 6", got)
	}
	if !styleAt(3).Italic || !styleAt(7).Underline {
		t.Fatal("FG palette erased existing attributes")
	}

	edit.Draw(canvas, Rect{W: 40, H: 5})
	clickButton("#BG")
	edit.Draw(canvas, Rect{W: 40, H: 5})
	clickSwatch(9)
	if got := styleAt(0).BG; got != ColorIndex(3) {
		t.Fatalf("unselected prefix BG = %v, want 3", got)
	}
	for _, offset := range []int{1, 2, 5, 6, 7} {
		if got := styleAt(offset).BG; got != ColorIndex(9) {
			t.Errorf("selected offset %d BG = %v, want 9", offset, got)
		}
	}
	if got := styleAt(8).BG; got != ColorIndex(7) {
		t.Fatalf("unselected suffix BG = %v, want 7", got)
	}
}
