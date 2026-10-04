// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
	"time"
)

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
	if len(edit.popoverButtons) != 8 {
		t.Fatalf("popover buttons = %d, want 8", len(edit.popoverButtons))
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

func TestRichTextEditWrapSelectionInBoxAndUndo(t *testing.T) {
	tests := []struct {
		name string
		doc  *RichDocument
		from RichPosition
		to   RichPosition
		want string
	}{
		{
			name: "single line",
			doc:  &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "before text after"}}}}},
			from: RichPosition{Offset: 7}, to: RichPosition{Offset: 11},
			want: "before \n┌────┐\n│text│\n└────┘\n after",
		},
		{
			name: "multiple lines and wide runes",
			doc: &RichDocument{Lines: []RichLine{
				{Spans: []RichSpan{{Text: "界", Style: Style{Bold: true}}, {Text: "x"}}},
				{Spans: []RichSpan{{Text: "yz"}}},
			}},
			from: RichPosition{}, to: RichPosition{Line: 1, Offset: 2},
			want: "┌───┐\n│界x│\n│yz │\n└───┘",
		},
		{
			name: "selection within styled span",
			doc:  &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "abCDxy", Style: Style{Italic: true}}}}}},
			from: RichPosition{Offset: 2}, to: RichPosition{Offset: 4},
			want: "ab\n┌──┐\n│CD│\n└──┘\nxy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := tt.doc.ToPlainText()
			edit := NewRichTextEdit(tt.doc)
			edit.SetSelection(tt.from, tt.to)
			edit.wrapSelectionInBox()
			if got := tt.doc.ToPlainText(); got != tt.want {
				t.Fatalf("boxed text = %q, want %q", got, tt.want)
			}
			if tt.name == "multiple lines and wide runes" && !tt.doc.Lines[1].Spans[0].Style.Bold {
				t.Fatalf("selected styled span lost style: %#v", tt.doc.Lines[1].Spans)
			}
			if tt.name == "selection within styled span" && !tt.doc.Lines[2].Spans[0].Style.Italic {
				t.Fatalf("styled selection lost style: %#v", tt.doc.Lines[2].Spans)
			}
			edit.undo()
			if got := tt.doc.ToPlainText(); got != before {
				t.Fatalf("undo text = %q, want %q", got, before)
			}
		})
	}
}

func TestRichTextEditBoxActionAvailableInPopover(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}}
	edit := NewRichTextEdit(doc)
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 5})
	canvas := NewCanvas(60, 4)
	edit.Draw(canvas, Rect{W: 60, H: 4})
	for _, button := range edit.popoverButtons {
		if button.label == "Box" {
			edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: button.rect.X + 1, Y: button.rect.Y})
			if got := doc.ToPlainText(); got != "┌─────┐\n│hello│\n└─────┘" {
				t.Fatalf("Box action text = %q", got)
			}
			return
		}
	}
	t.Fatal("Box action missing from popover")
}

func TestRichTextEditBoxModeDrawsTurnsAndUndoesOneStroke(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{}}}
	edit := NewRichTextEdit(doc)
	if got := edit.ConsumeKey(KeyEvent{Key: "ctrl-shift-b"}); !got.Consumed || !edit.BoxMode {
		t.Fatalf("box mode toggle result/mode = %+v/%v", got, edit.BoxMode)
	}
	for _, key := range []string{"right", "right", "down", "left"} {
		if got := edit.ConsumeKey(KeyEvent{Key: key}); !got.Consumed {
			t.Fatalf("%s was not consumed in box mode", key)
		}
	}
	if got, want := doc.ToPlainText(), "╶─┐\n ╶┘"; got != want {
		t.Fatalf("drawn box path = %q, want %q", got, want)
	}
	if got := edit.ConsumeKey(KeyEvent{Key: "esc"}); !got.Consumed || edit.BoxMode {
		t.Fatalf("escape result/mode = %+v/%v", got, edit.BoxMode)
	}
	edit.ConsumeKey(KeyEvent{Key: "ctrl-z"})
	if got := doc.ToPlainText(); got != "" {
		t.Fatalf("one undo did not restore pre-stroke document: %q", got)
	}
}

func TestRichTextEditBoxModePadsPastLineEndAndBelowDocument(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "abc"}}}}}
	edit := NewRichTextEdit(doc)
	edit.Cursor.Offset = 3
	edit.ConsumeKey(KeyEvent{Key: "f5"})
	edit.ConsumeKey(KeyEvent{Key: "right"})
	edit.ConsumeKey(KeyEvent{Key: "down"})
	if got, want := doc.ToPlainText(), "abc╶┐\n    ╵"; got != want {
		t.Fatalf("padded box path = %q, want %q", got, want)
	}
	if edit.Cursor != (RichPosition{Line: 1, Offset: 4}) {
		t.Fatalf("cursor after downward stroke = %+v, want line 1 offset 4", edit.Cursor)
	}
	edit.ConsumeKey(KeyEvent{Key: "f5"})
	if edit.BoxMode {
		t.Fatal("F5 did not exit box mode")
	}
}

func TestRichTextEditBoxModeHandlesCSIuToggle(t *testing.T) {
	key := DecodeKey([]byte("\x1b[98;6u"))
	if !key.Is("ctrl-shift-b") {
		t.Fatalf("CSI-u Ctrl-Shift-B decoded as %+v", key)
	}
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{}}})
	edit.ConsumeKey(key)
	if !edit.BoxMode {
		t.Fatal("CSI-u Ctrl-Shift-B did not enter box mode")
	}
	if key := DecodeKey([]byte{2}); !key.Is("ctrl-b") {
		t.Fatalf("legacy Ctrl-B changed from bold key: %+v", key)
	}
}

func TestRichTextEditBoxModeCompletesCrossingAndTJunction(t *testing.T) {
	t.Run("crossing", func(t *testing.T) {
		doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: " "}}}, {Spans: []RichSpan{{Text: "───"}}}, {}}}
		edit := NewRichTextEdit(doc)
		edit.Cursor.Offset = 1
		edit.ConsumeKey(KeyEvent{Key: "f5"})
		edit.ConsumeKey(KeyEvent{Key: "down"})
		edit.ConsumeKey(KeyEvent{Key: "down"})
		if got := richLineCell(doc.Lines[1], 1); got != "┼" {
			t.Fatalf("crossing glyph = %q, want ┼", got)
		}
	})
	t.Run("stroke ends on existing side", func(t *testing.T) {
		doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: " "}}}, {Spans: []RichSpan{{Text: "───"}}}}}
		edit := NewRichTextEdit(doc)
		edit.Cursor.Offset = 1
		edit.ConsumeKey(KeyEvent{Key: "f5"})
		edit.ConsumeKey(KeyEvent{Key: "down"})
		if got := richLineCell(doc.Lines[1], 1); got != "┴" {
			t.Fatalf("T-junction glyph = %q, want ┴", got)
		}
	})
}

func TestRichTextEditBoxModeDoesNotModifyAdjacentBox(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "│ "}}}}}
	edit := NewRichTextEdit(doc)
	edit.Cursor.Offset = 1
	edit.ConsumeKey(KeyEvent{Key: "f5"})
	edit.ConsumeKey(KeyEvent{Key: "right"})
	if got := richLineCell(doc.Lines[0], 1); got != "╶" {
		t.Fatalf("drawn cell = %q, want ╶", got)
	}
	if got := richLineCell(doc.Lines[0], 0); got != "│" {
		t.Fatalf("adjacent existing glyph changed to %q", got)
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

func richTestEditor(spans ...RichSpan) *RichTextEdit {
	return NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: spans}}})
}

func richTestText(e *RichTextEdit) string {
	var out []string
	for _, line := range e.Document.Lines {
		var text string
		for _, span := range line.Spans {
			text += span.Text
		}
		out = append(out, text)
	}
	return strings.Join(out, "\n")
}

func richTestKey(e *RichTextEdit, keys ...string) {
	for _, key := range keys {
		e.ConsumeKey(KeyEvent{Key: key})
	}
}

func richTestType(e *RichTextEdit, text string) {
	for _, r := range text {
		e.ConsumeKey(KeyEvent{Text: string(r)})
	}
}

func TestRichTextEditWordScopedStyling(t *testing.T) {
	for key, get := range map[string]func(Style) bool{
		"ctrl-b": func(s Style) bool { return s.Bold },
		"ctrl-i": func(s Style) bool { return s.Italic },
		"ctrl-u": func(s Style) bool { return s.Underline },
	} {
		e := richTestEditor(RichSpan{Text: "one two three"})
		e.Cursor.Offset = 5
		richTestKey(e, key)
		spans := e.Document.Lines[0].Spans
		if len(spans) != 3 || spans[1].Text != "two" || !get(spans[1].Style) || get(spans[0].Style) || get(spans[2].Style) {
			t.Fatalf("%s spans = %#v", key, spans)
		}
		richTestKey(e, key)
		if len(e.Document.Lines[0].Spans) != 1 {
			t.Fatalf("%s toggle off spans = %#v", key, e.Document.Lines[0].Spans)
		}
	}
}

func TestRichTextEditStyleWithoutWordKeepsDocument(t *testing.T) {
	for _, offset := range []int{3, 7} { // whitespace, line end
		e := richTestEditor(RichSpan{Text: "one two"})
		e.Cursor.Offset = offset
		richTestKey(e, "ctrl-b")
		if len(e.Document.Lines[0].Spans) != 1 || e.Document.Lines[0].Spans[0].Style.Bold {
			t.Fatalf("offset %d changed document: %#v", offset, e.Document.Lines[0].Spans)
		}
		if len(e.undoStack) != 0 {
			t.Fatalf("offset %d recorded undo", offset)
		}
	}
}

func TestRichTextEditStyleSelectionSpanningStyledSpans(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "ab"}, RichSpan{Text: "cd", Style: Style{Bold: true}}, RichSpan{Text: "ef"})
	e.SetSelection(RichPosition{Offset: 1}, RichPosition{Offset: 5})
	richTestKey(e, "ctrl-b")
	spans := e.Document.Lines[0].Spans
	if len(spans) != 3 || spans[0].Style.Bold || !spans[1].Style.Bold || spans[1].Text != "bcde" || spans[2].Style.Bold {
		t.Fatalf("spans = %#v", spans)
	}
}

func TestRichTextEditCtrlSpaceSelectsWordAndOpensPopover(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "one two"})
	e.ShowPopover = false
	e.Cursor.Offset = 5
	richTestKey(e, "ctrl-space")
	from, to := e.selectionBounds()
	if !e.HasSelection || from.Offset != 4 || to.Offset != 7 || !e.ShowPopover {
		t.Fatalf("selection %v-%v has=%v popover=%v", from, to, e.HasSelection, e.ShowPopover)
	}
	e = richTestEditor(RichSpan{Text: "one two"})
	e.ShowPopover = false
	e.Cursor.Offset = 3
	richTestKey(e, "ctrl-space")
	if e.HasSelection || e.ShowPopover {
		t.Fatal("ctrl-space on whitespace must be a no-op")
	}
}

func TestRichTextEditCtrlSpaceSelectsSharpAndRoundedBoxRectangle(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
	}{
		{name: "sharp styled", lines: []string{"┌───┐", "│abc│", "└───┘"}},
		{name: "rounded", lines: []string{"╭───╮", "│abc│", "╰───╯"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := richEditorLines(tc.lines...)
			e.Document.Lines[1].Spans = []RichSpan{{Text: "│"}, {Text: "abc", Style: Style{Bold: true}}, {Text: "│"}}
			e.Cursor = RichPosition{Line: 0, Offset: 0}
			if got := e.ConsumeKey(KeyEvent{Key: "ctrl-space"}); got != Handled() {
				t.Fatalf("ctrl-space = %+v, want Handled", got)
			}
			if e.boxSelection == nil || e.boxSelection.top != 0 || e.boxSelection.left != 0 || e.boxSelection.bottom != 2 || e.boxSelection.right != 4 || !e.ShowPopover {
				t.Fatalf("box selection = %+v, popover=%v", e.boxSelection, e.ShowPopover)
			}
			canvas := NewCanvas(5, 3)
			e.Draw(canvas, Rect{W: 5, H: 3})
			for _, pos := range [][2]int{{0, 0}, {4, 0}, {0, 1}, {4, 1}, {0, 2}, {4, 2}} {
				if got := canvas.Get(pos[0], pos[1]).Style.BG; got != ColorIndex(uint8(SpeccedDefaults.RichTextEdit.SelectionBG)) {
					t.Errorf("perimeter cell (%d,%d) not selected: bg=%v", pos[0], pos[1], got)
				}
			}
			if !canvas.Get(1, 1).Style.Bold {
				t.Fatal("box selection erased interior span styling")
			}
		})
	}
}

func TestRichTextEditCtrlSpaceBoxGeometryAndDeterministicTies(t *testing.T) {
	t.Run("crossing arms", func(t *testing.T) {
		e := richEditorLines("┌─┬─┐", "├─┼─┤", "└─┴─┘")
		e.Cursor = RichPosition{Line: 1, Offset: 2}
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		if e.boxSelection == nil || e.boxSelection.top != 0 || e.boxSelection.left != 0 || e.boxSelection.bottom != 1 || e.boxSelection.right != 2 {
			t.Fatalf("crossing box selection = %+v", e.boxSelection)
		}
	})
	t.Run("nested smallest rectangle", func(t *testing.T) {
		e := richEditorLines("┌──────┐", "│┌──┐  │", "││xy│  │", "│└──┘  │", "└──────┘")
		e.Cursor = RichPosition{Line: 1, Offset: 1}
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		if got := e.boxSelection; got == nil || got.top != 1 || got.left != 1 || got.bottom != 3 || got.right != 4 {
			t.Fatalf("nested selection = %+v, want inner box", got)
		}
	})
	t.Run("shared edge ties choose left box", func(t *testing.T) {
		e := richEditorLines("┌─┬─┐", "│a│b│", "└─┴─┘")
		e.Cursor = RichPosition{Line: 0, Offset: 2}
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		if got := e.boxSelection; got == nil || got.left != 0 || got.right != 2 {
			t.Fatalf("shared-edge selection = %+v, want left box", got)
		}
	})
}

func TestRichTextEditCtrlSpaceBoxFallbackAndRectangleOperations(t *testing.T) {
	t.Run("plain word", func(t *testing.T) {
		e := richTestEditor(RichSpan{Text: "plain text"})
		e.Cursor.Offset = 2
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		from, to := e.selectionBounds()
		if e.boxSelection != nil || from.Offset != 0 || to.Offset != 5 {
			t.Fatalf("plain fallback = %v-%v, box=%+v", from, to, e.boxSelection)
		}
	})
	t.Run("word beside box", func(t *testing.T) {
		e := richEditorLines("┌─┐word", "│x│", "└─┘")
		e.Cursor = RichPosition{Offset: 3}
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		from, to := e.selectionBounds()
		if e.boxSelection != nil || from.Offset != 3 || to.Offset != 7 {
			t.Fatalf("adjacent word fallback = %v-%v, box=%+v", from, to, e.boxSelection)
		}
	})
	t.Run("broken perimeter falls back to word", func(t *testing.T) {
		e := richEditorLines("┌──┐", "Xabc│", "└──┘")
		e.Cursor = RichPosition{Line: 1, Offset: 1}
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		from, to := e.selectionBounds()
		if e.boxSelection != nil || from.Line != 1 || from.Offset != 0 || to.Offset != 4 {
			t.Fatalf("broken fallback = %v-%v, box=%+v", from, to, e.boxSelection)
		}
	})
	t.Run("partial box without word stays unselected", func(t *testing.T) {
		e := richEditorLines("┌──┐", "│ab│", "└──x")
		e.Cursor = RichPosition{Offset: 0}
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		if e.HasSelection || e.boxSelection != nil {
			t.Fatalf("partial perimeter selected: %+v", e.boxSelection)
		}
	})
	t.Run("wide and combining text; preserve outside text on cut and undo", func(t *testing.T) {
		e := richEditorLines("x┌───┐z", "x│界é│z", "x└───┘z")
		e.Cursor = RichPosition{Offset: 1}
		e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
		if e.boxSelection == nil || e.boxSelection.left != 1 || e.boxSelection.right != 5 {
			t.Fatalf("wide content box = %+v", e.boxSelection)
		}
		if !e.copySelectionOrWord() || richLinesText(e.clipboard) != "┌───┐\n│界é│\n└───┘" {
			t.Fatalf("copied rectangle = %q", richLinesText(e.clipboard))
		}
		e.mutate(false, false, func() { e.deleteSelection() })
		if got := richLinesText(e.Document.Lines); got != "xz\nxz\nxz" {
			t.Fatalf("document after rectangular cut = %q", got)
		}
		e.undo()
		if got := richLinesText(e.Document.Lines); got != "x┌───┐z\nx│界é│z\nx└───┘z" {
			t.Fatalf("document after undo = %q", got)
		}
		if e.boxSelection == nil {
			t.Fatal("undo did not restore the validated box selection")
		}
		e.redo()
		if e.boxSelection != nil || richLinesText(e.Document.Lines) != "xz\nxz\nxz" {
			t.Fatalf("redo left stale box selection or wrong text: box=%+v text=%q", e.boxSelection, richLinesText(e.Document.Lines))
		}
	})
}

func TestRichTextEditMovingCursorClearsBoxSelection(t *testing.T) {
	e := richEditorLines("┌──┐", "│ab│", "└──┘")
	e.Cursor = RichPosition{Offset: 0}
	e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	if e.boxSelection == nil {
		t.Fatal("ctrl-space did not select the box")
	}
	e.ConsumeKey(KeyEvent{Key: "right"})
	if e.boxSelection != nil || e.HasSelection {
		t.Fatalf("cursor movement retained box selection: box=%+v has=%v", e.boxSelection, e.HasSelection)
	}
}

func TestRichTextEditBoxRectangleFormattingLeavesSideTextOutside(t *testing.T) {
	e := richEditorLines("x┌─┐z", "x│a│z", "x└─┘z")
	e.Cursor = RichPosition{Offset: 1}
	e.ConsumeKey(KeyEvent{Key: "ctrl-space"})
	e.toggleAttribute(func(s *Style, on bool) { s.Bold = on }, func(s Style) bool { return s.Bold })
	if e.boxSelection == nil {
		t.Fatal("styling failed to revalidate the box selection")
	}
	if !e.Document.Lines[0].Spans[1].Style.Bold || !e.Document.Lines[1].Spans[1].Style.Bold {
		t.Fatal("box perimeter and interior were not styled")
	}
	if e.Document.Lines[0].Spans[0].Style.Bold || e.Document.Lines[0].Spans[len(e.Document.Lines[0].Spans)-1].Style.Bold {
		t.Fatal("formatting crossed the rectangle into side text")
	}
}

func richEditorLines(lines ...string) *RichTextEdit {
	doc := &RichDocument{Lines: make([]RichLine, len(lines))}
	for i, line := range lines {
		doc.Lines[i] = RichLine{Spans: []RichSpan{{Text: line}}}
	}
	return NewRichTextEdit(doc)
}

func richLinesText(lines []RichLine) string {
	parts := make([]string, len(lines))
	for i, line := range lines {
		parts[i] = richTestText(NewRichTextEdit(&RichDocument{Lines: []RichLine{line}}))
	}
	return strings.Join(parts, "\n")
}

func TestRichTextEditCopyPasteKeepsStyles(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "ab "}, RichSpan{Text: "cd", Style: Style{Bold: true}})
	e.SetSelection(RichPosition{Offset: 1}, RichPosition{Offset: 5})
	richTestKey(e, "ctrl-c")
	e.ClearSelection()
	e.Cursor.Offset = 5
	richTestKey(e, "ctrl-v")
	if got := richTestText(e); got != "ab cdb cd" {
		t.Fatalf("text = %q", got)
	}
	spans := e.Document.Lines[0].Spans
	if len(spans) != 4 || !spans[3].Style.Bold || spans[3].Text != "cd" || spans[2].Text != "b " {
		t.Fatalf("spans = %#v", spans)
	}
}

func TestRichTextEditCopyWordAndAliases(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "one two"})
	e.Cursor.Offset = 1
	richTestKey(e, "ctrl-c")
	e.Cursor.Offset = 7
	e.ClearSelection()
	richTestKey(e, "shift-insert")
	if got := richTestText(e); got != "one twoone" {
		t.Fatalf("word copy/paste = %q", got)
	}
	e.ClearSelection()
	e.Cursor.Offset = 5
	richTestKey(e, "ctrl-insert")
	e.Cursor = RichPosition{}
	e.ClearSelection()
	richTestKey(e, "shift-insert")
	if got := richTestText(e); got != "twooneone twoone" {
		t.Fatalf("aliases = %q", got)
	}
}

func TestRichTextEditCut(t *testing.T) {
	for _, key := range []string{"ctrl-x", "shift-delete"} {
		e := richTestEditor(RichSpan{Text: "one two"})
		e.SetSelection(RichPosition{Offset: 3}, RichPosition{Offset: 7})
		richTestKey(e, key)
		if got := richTestText(e); got != "one" || e.HasSelection {
			t.Fatalf("%s text = %q", key, got)
		}
		richTestKey(e, "ctrl-v")
		if got := richTestText(e); got != "one two" {
			t.Fatalf("%s paste = %q", key, got)
		}
	}
}

func TestRichTextEditCutWordWithoutSelection(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "one two"})
	e.Cursor.Offset = 1
	richTestKey(e, "ctrl-x")
	if got := richTestText(e); got != " two" {
		t.Fatalf("text = %q", got)
	}
}

func TestRichTextEditPasteIntoStyledSpan(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "abcd", Style: Style{Italic: true}})
	e.clipboard = []RichLine{{Spans: []RichSpan{{Text: "XY", Style: Style{Bold: true}}}}}
	e.Cursor.Offset = 2
	richTestKey(e, "ctrl-v")
	spans := e.Document.Lines[0].Spans
	if len(spans) != 3 || spans[0].Text != "ab" || !spans[0].Style.Italic || spans[1].Text != "XY" || !spans[1].Style.Bold || spans[1].Style.Italic || spans[2].Text != "cd" || !spans[2].Style.Italic {
		t.Fatalf("spans = %#v", spans)
	}
	if e.Cursor.Offset != 4 {
		t.Fatalf("cursor = %d", e.Cursor.Offset)
	}
}

func TestRichTextEditMultiLineCopyPaste(t *testing.T) {
	e := NewRichTextEdit(&RichDocument{Lines: []RichLine{
		{Spans: []RichSpan{{Text: "abc"}}}, {Spans: []RichSpan{{Text: "def", Style: Style{Bold: true}}}}, {Spans: []RichSpan{{Text: "end"}}},
	}})
	e.SetSelection(RichPosition{Offset: 1}, RichPosition{Line: 1, Offset: 2})
	richTestKey(e, "ctrl-c")
	e.Cursor = RichPosition{Line: 2, Offset: 1}
	e.ClearSelection()
	richTestKey(e, "ctrl-v")
	if got := richTestText(e); got != "abc\ndef\nebc\ndend" {
		t.Fatalf("text = %q", got)
	}
	if e.Cursor != (RichPosition{Line: 3, Offset: 2}) {
		t.Fatalf("cursor = %v", e.Cursor)
	}
}

func TestRichTextEditEmptyClipboardPasteIsNoop(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "abc"})
	richTestKey(e, "ctrl-v", "shift-insert")
	if richTestText(e) != "abc" || len(e.undoStack) != 0 {
		t.Fatalf("text = %q undo=%d", richTestText(e), len(e.undoStack))
	}
}

func TestRichTextEditCopyWithoutWordIsIgnored(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "ab cd"})
	e.Cursor.Offset = 2
	if res := e.ConsumeKey(KeyEvent{Key: "ctrl-c"}); res.Consumed {
		t.Fatal("copy on whitespace should be ignored")
	}
}

func TestRichTextEditUndoRedoTypingRuns(t *testing.T) {
	e := richTestEditor()
	richTestType(e, "hello world")
	for _, want := range []string{"hello ", ""} {
		richTestKey(e, "ctrl-z")
		if got := richTestText(e); got != want {
			t.Fatalf("undo = %q, want %q", got, want)
		}
	}
	richTestKey(e, "ctrl-z") // empty stack
	richTestKey(e, "ctrl-r")
	if got := richTestText(e); got != "hello " {
		t.Fatalf("redo = %q", got)
	}
	richTestKey(e, "ctrl-shift-y")
	if got := richTestText(e); got != "hello world" {
		t.Fatalf("redo = %q", got)
	}
	richTestKey(e, "ctrl-y", "ctrl-shift-z")
	if got := richTestText(e); got != "hello world" {
		t.Fatalf("undo/redo roundtrip = %q", got)
	}
}

func TestRichTextEditUndoStyleCutPasteSteps(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "one two"})
	e.Cursor.Offset = 1
	richTestKey(e, "ctrl-b")
	e.Cursor.Offset = 5
	richTestKey(e, "ctrl-x", "ctrl-v")
	if got := richTestText(e); got != "one two" || len(e.undoStack) != 3 {
		t.Fatalf("text = %q steps = %d", got, len(e.undoStack))
	}
	richTestKey(e, "ctrl-z")
	if got := richTestText(e); got != "one " {
		t.Fatalf("undo paste = %q", got)
	}
	richTestKey(e, "ctrl-z")
	if got := richTestText(e); got != "one two" {
		t.Fatalf("undo cut = %q", got)
	}
	richTestKey(e, "ctrl-z")
	if spans := e.Document.Lines[0].Spans; len(spans) != 1 || spans[0].Style.Bold {
		t.Fatalf("undo style = %#v", spans)
	}
}

func TestRichTextEditNewEditClearsRedoBranch(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "ab"})
	e.Cursor.Offset = 2
	richTestType(e, "c")
	richTestKey(e, "ctrl-z")
	richTestKey(e, "left")
	richTestType(e, "X")
	richTestKey(e, "ctrl-r")
	if got := richTestText(e); got != "aXb" {
		t.Fatalf("text = %q", got)
	}
	richTestKey(e, "ctrl-z")
	if got := richTestText(e); got != "ab" {
		t.Fatalf("undo after redo branch = %q", got)
	}
}

func TestRichTextEditUndoLimit(t *testing.T) {
	e := richTestEditor()
	for range richUndoLimit + 20 {
		richTestType(e, "a ")
	}
	if len(e.undoStack) != richUndoLimit {
		t.Fatalf("undo stack = %d", len(e.undoStack))
	}
}

func TestRichTextEditDoubleAndTripleClick(t *testing.T) {
	e := NewRichTextEdit(&RichDocument{Lines: []RichLine{
		{Spans: []RichSpan{{Text: "one two"}}}, {Spans: []RichSpan{{Text: "last line"}}},
	}})
	e.Draw(NewCanvas(20, 5), Rect{W: 20, H: 5})
	clock := time.Unix(0, 0)
	e.now = func() time.Time { return clock }
	press := func(x, y int) {
		e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: x, Y: y})
		e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: x, Y: y})
		clock = clock.Add(100 * time.Millisecond)
	}
	press(5, 0)
	press(5, 0)
	if from, to := e.selectionBounds(); from.Offset != 4 || to.Offset != 7 {
		t.Fatalf("double click selection %v-%v", from, to)
	}
	press(5, 0)
	if from, to := e.selectionBounds(); from.Offset != 0 || to.Offset != 7 {
		t.Fatalf("triple click selection %v-%v", from, to)
	}
	clock = clock.Add(time.Second)
	press(2, 1)
	press(2, 1)
	press(2, 1)
	from, to := e.selectionBounds()
	if from != (RichPosition{Line: 1}) || to != (RichPosition{Line: 1, Offset: 9}) {
		t.Fatalf("last line triple click %v-%v", from, to)
	}
	clock = clock.Add(time.Second)
	press(2, 1)
	press(2, 1)
	if from, to := e.selectionBounds(); from.Offset != 0 || to.Offset != 4 {
		t.Fatalf("double click on last line %v-%v", from, to)
	}
}

func TestRichTextEditSlowClicksDoNotSelectWord(t *testing.T) {
	e := richTestEditor(RichSpan{Text: "one two"})
	e.Draw(NewCanvas(20, 2), Rect{W: 20, H: 2})
	clock := time.Unix(0, 0)
	e.now = func() time.Time { return clock }
	for range 2 {
		e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 5, Y: 0})
		e.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft, X: 5, Y: 0})
		clock = clock.Add(time.Second)
	}
	if e.HasSelection {
		t.Fatal("slow clicks selected text")
	}
}

func richViewTestEditor() *RichTextEdit {
	e := NewRichTextView(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{
		{Text: "see "}, {Text: "loom", Link: "https://ubunatic.com/loom"}, {Text: " now"},
	}}}})
	return e
}

func TestRichTextViewRejectsEveryEditPath(t *testing.T) {
	e := richViewTestEditor()
	e.clipboard = []RichLine{{Spans: []RichSpan{{Text: "zz"}}}}
	e.SetSelection(RichPosition{Offset: 0}, RichPosition{Offset: 3})
	e.Cursor.Offset = 3
	before := richTestText(e)
	for _, key := range []KeyEvent{
		{Text: "x"}, {Key: "enter"}, {Key: "backspace"}, {Key: "delete"}, {Key: "ctrl-b"}, {Key: "ctrl-i"},
		{Key: "ctrl-u"}, {Key: "ctrl-v"}, {Key: "shift-insert"}, {Key: "ctrl-x"}, {Key: "shift-delete"},
		{Key: "ctrl-z"}, {Key: "ctrl-y"}, {Key: "ctrl-r"}, {Key: "ctrl-space"},
	} {
		if got := e.ConsumeKey(key); got != Ignored() {
			t.Fatalf("%+v result = %v, want Ignored", key, got)
		}
	}
	if got := richTestText(e); got != before || len(e.undoStack) != 0 {
		t.Fatalf("text = %q undo=%d, want unchanged", got, len(e.undoStack))
	}
	if !e.HasSelection || e.ShowPopover {
		t.Fatal("selection lost or popover enabled")
	}
	c := NewCanvas(40, 6)
	e.ShowPopover = true // even if forced on, view mode draws no popover
	e.Draw(c, Rect{W: 40, H: 6})
	if len(e.popoverButtons) != 0 {
		t.Fatal("view mode drew a popover")
	}
	e.ConsumeMouse(MouseEvent{X: 1, Y: 0, Action: MousePress, Button: MouseLeft})
	if richTestText(e) != before {
		t.Fatal("mouse changed text")
	}
}

func TestRichTextViewDoesNotPlaceEditCursor(t *testing.T) {
	e := richViewTestEditor()
	c := NewCanvas(40, 2)
	e.Draw(c, Rect{W: 40, H: 2})
	if c.CursorX != -1 || c.CursorY != -1 {
		t.Fatalf("view draw placed cursor at (%d,%d)", c.CursorX, c.CursorY)
	}
}

func TestRichTextViewAllowsSelectionAndCopy(t *testing.T) {
	e := richViewTestEditor()
	if got := e.ConsumeKey(KeyEvent{Key: "shift-right"}); got != Handled() {
		t.Fatalf("shift-right = %v", got)
	}
	richTestKey(e, "shift-right", "shift-right", "shift-right")
	if !e.HasSelection {
		t.Fatal("no selection")
	}
	if got := e.ConsumeKey(KeyEvent{Key: "ctrl-c"}); got != Handled() {
		t.Fatalf("ctrl-c = %v", got)
	}
	if len(e.clipboard) != 1 || e.clipboard[0].Spans[0].Text != "see " {
		t.Fatalf("clipboard = %#v", e.clipboard)
	}
	e.ConsumeMouse(MouseEvent{X: 0, Y: 0, Action: MousePress, Button: MouseLeft})
	e.ConsumeMouse(MouseEvent{X: 5, Y: 0, Action: MouseRelease, Button: MouseLeft})
	if !e.HasSelection {
		t.Fatal("mouse drag selection failed")
	}
}

func TestRichTextEditLinkSpanStyleAndCopyPaste(t *testing.T) {
	for _, focused := range []bool{true, false} {
		e := richViewTestEditor()
		e.SetFocus(focused)
		c := NewCanvas(20, 2)
		e.Draw(c, Rect{W: 20, H: 2})
		link, plain := c.Get(4, 0).Style, c.Get(0, 0).Style
		if link.FG != ColorIndex(39) || !link.Underline || plain.Underline || plain.FG == link.FG {
			t.Fatalf("focused=%v link=%+v plain=%+v", focused, link, plain)
		}
	}
	e := richViewTestEditor()
	e.ViewMode = false
	e.SetSelection(RichPosition{Offset: 4}, RichPosition{Offset: 8})
	richTestKey(e, "ctrl-c")
	e.ClearSelection()
	e.Cursor.Offset = 12
	richTestKey(e, "ctrl-v")
	if got := richTestText(e); got != "see loom nowloom" {
		t.Fatalf("text = %q", got)
	}
	spans := e.Document.Lines[0].Spans
	last := spans[len(spans)-1]
	if last.Text != "loom" || last.Link != "https://ubunatic.com/loom" {
		t.Fatalf("pasted span = %#v", last)
	}
}

func TestRichTextEditLineStartEndKeys(t *testing.T) {
	long := strings.Repeat("abcdef ", 20)
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}, {}, {Spans: []RichSpan{{Text: long}}}}}
	edit := NewRichTextEdit(doc)
	edit.Cursor = RichPosition{Line: 0, Offset: 2}
	edit.ClearSelection()
	edit.ConsumeKey(KeyEvent{Key: "ctrl-e"})
	if edit.Cursor != (RichPosition{Line: 0, Offset: 5}) || edit.HasSelection {
		t.Fatalf("ctrl-e cursor = %+v sel=%v", edit.Cursor, edit.HasSelection)
	}
	edit.ConsumeKey(KeyEvent{Key: "ctrl-a"})
	if edit.Cursor != (RichPosition{}) {
		t.Fatalf("ctrl-a cursor = %+v", edit.Cursor)
	}
	edit.Cursor = RichPosition{Line: 0, Offset: 2}
	edit.ClearSelection()
	edit.ConsumeKey(KeyEvent{Key: "ctrl-shift-e"})
	from, to := edit.selectionBounds()
	if !edit.HasSelection || from.Offset != 2 || to.Offset != 5 {
		t.Fatalf("ctrl-shift-e selection = %+v..%+v", from, to)
	}
	edit.ConsumeKey(KeyEvent{Key: "shift-home"})
	from, to = edit.selectionBounds()
	if !edit.HasSelection || from.Offset != 0 || to.Offset != 2 {
		t.Fatalf("shift-home selection = %+v..%+v", from, to)
	}
	edit.Cursor = RichPosition{Line: 0, Offset: 3}
	edit.ClearSelection()
	edit.ConsumeKey(KeyEvent{Key: "ctrl-shift-a"})
	from, to = edit.selectionBounds()
	if from.Offset != 0 || to.Offset != 3 {
		t.Fatalf("ctrl-shift-a selection = %+v..%+v", from, to)
	}
	edit.ConsumeKey(KeyEvent{Key: "shift-end"})
	from, to = edit.selectionBounds()
	if from.Offset != 3 || to.Offset != 5 {
		t.Fatalf("shift-end selection = %+v..%+v", from, to)
	}
	edit.Cursor = RichPosition{Line: 1}
	edit.ClearSelection()
	for _, k := range []string{"ctrl-a", "ctrl-e", "ctrl-shift-a", "ctrl-shift-e"} {
		edit.ConsumeKey(KeyEvent{Key: k})
		if edit.Cursor != (RichPosition{Line: 1}) {
			t.Fatalf("%s on empty line cursor = %+v", k, edit.Cursor)
		}
	}
	edit.Cursor = RichPosition{Line: 2, Offset: 30}
	edit.ClearSelection()
	edit.ConsumeKey(KeyEvent{Key: "ctrl-e"})
	if edit.Cursor != (RichPosition{Line: 2, Offset: len([]rune(long))}) {
		t.Fatalf("ctrl-e on long line = %+v", edit.Cursor)
	}
	edit.ConsumeKey(KeyEvent{Key: "ctrl-a"})
	if edit.Cursor != (RichPosition{Line: 2}) {
		t.Fatalf("ctrl-a on long line = %+v", edit.Cursor)
	}
}

func TestRichTextViewLineStartEndKeys(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}}
	view := NewRichTextView(doc)
	view.Cursor = RichPosition{Offset: 2}
	view.ClearSelection()
	view.ConsumeKey(KeyEvent{Key: "ctrl-e"})
	if view.Cursor.Offset != 5 {
		t.Fatalf("view ctrl-e cursor = %+v", view.Cursor)
	}
	view.ConsumeKey(KeyEvent{Key: "ctrl-shift-a"})
	if !view.HasSelection || doc.Lines[0].Spans[0].Text != "hello" {
		t.Fatalf("view ctrl-shift-a sel=%v", view.HasSelection)
	}
}
