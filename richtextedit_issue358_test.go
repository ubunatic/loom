// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func issue358Editor(t *testing.T) *RichTextEdit {
	t.Helper()
	e := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello world"}}}}})
	e.SetSelection(RichPosition{}, RichPosition{Offset: 5})
	e.Cursor = RichPosition{Offset: 5}
	e.Draw(NewCanvas(60, 6), Rect{W: 60, H: 6})
	return e
}

func Test358PopoverClosePathsClearSelection(t *testing.T) {
	for _, tc := range []struct {
		name  string
		close func(*RichTextEdit)
	}{
		{"Escape", func(e *RichTextEdit) { e.ConsumeKey(KeyEvent{Key: "esc"}) }},
		{"CtrlSpace toggle", func(e *RichTextEdit) { e.ConsumeKey(KeyEvent{Key: "ctrl-space"}) }},
		{"outside click", func(e *RichTextEdit) { e.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 59, Y: 5}) }},
		{"B action", func(e *RichTextEdit) { e.ConsumeKey(KeyEvent{Key: "down"}); e.ConsumeKey(KeyEvent{Key: "enter"}) }},
		{"B hotkey", func(e *RichTextEdit) { e.ConsumeKey(KeyEvent{Key: "down"}); e.ConsumeKey(KeyEvent{Text: "b"}) }},
		{"Frame Plain", func(e *RichTextEdit) {
			e.ConsumeKey(KeyEvent{Key: "down"})
			for i := 0; i < popoverActionIndex("Box"); i++ {
				e.ConsumeKey(KeyEvent{Key: "right"})
			}
			e.ConsumeKey(KeyEvent{Key: "enter"})
			e.ConsumeKey(KeyEvent{Key: "enter"})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := issue358Editor(t)
			cursor := e.Cursor
			text := richLinesText(e.Document.Lines)
			tc.close(e)
			checkCursor := tc.name == "Escape" || tc.name == "CtrlSpace toggle" || tc.name == "outside click"
			if e.HasSelection || (checkCursor && e.Cursor != cursor) {
				t.Fatalf("close left selection=%v cursor=%+v want cursor=%+v", e.HasSelection, e.Cursor, cursor)
			}
			if tc.name == "Escape" || tc.name == "CtrlSpace toggle" || tc.name == "outside click" {
				if got := richLinesText(e.Document.Lines); got != text {
					t.Fatalf("cancel changed text: %q", got)
				}
			}
		})
	}
}

func Test358ColorPickerPickClosesAndClearsSelection(t *testing.T) {
	e := issue358Editor(t)
	popoverFocusAction(t, e, "ColorPicker")
	e.ConsumeKey(KeyEvent{Key: "space"})
	e.ConsumeKey(KeyEvent{Key: "right"})
	e.ConsumeKey(KeyEvent{Key: "space"})
	if e.HasSelection || e.popoverColorPicker != nil {
		t.Fatalf("ColorPicker pick left selection=%v picker=%v", e.HasSelection, e.popoverColorPicker != nil)
	}
}

func Test358SelectAllActionKeepsNewSelection(t *testing.T) {
	e := issue358Editor(t)
	e.applyHamburgerItem(0)
	if !e.HasSelection || e.SelectionFrom != (RichPosition{}) || e.SelectionTo.Offset != 11 {
		t.Fatalf("Select all selection = %v..%v, selected=%v", e.SelectionFrom, e.SelectionTo, e.HasSelection)
	}
}
