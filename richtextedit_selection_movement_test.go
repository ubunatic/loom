// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

func TestRichTextEditCommonSelectionMovements(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "one two three"}}}, {Spans: []RichSpan{{Text: "four five"}}}, {Spans: []RichSpan{{Text: "six seven"}}}}}
	for _, tc := range []struct {
		name, key, reverse string
		start, end         RichPosition
	}{
		{"word right", "ctrl-shift-right", "ctrl-shift-left", RichPosition{Offset: 4}, RichPosition{Offset: 8}},
		{"word left", "ctrl-shift-left", "ctrl-shift-right", RichPosition{Offset: 8}, RichPosition{Offset: 4}},
		{"document start", "ctrl-shift-home", "ctrl-shift-end", RichPosition{Line: 2, Offset: 9}, RichPosition{}},
		{"document end", "ctrl-shift-end", "ctrl-shift-home", RichPosition{}, RichPosition{Line: 2, Offset: 9}},
		{"page down", "shift-pgdown", "shift-pgup", RichPosition{}, RichPosition{Line: 2}},
		{"page up", "shift-pgup", "shift-pgdown", RichPosition{Line: 2}, RichPosition{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewRichTextEdit(doc)
			e.lastRect = Rect{H: 2}
			e.Cursor = tc.start
			e.ConsumeKey(KeyEvent{Key: tc.key})
			if e.Cursor != tc.end || !e.HasSelection {
				t.Fatalf("extend cursor=%+v selection=%v", e.Cursor, e.HasSelection)
			}
			e.ConsumeKey(KeyEvent{Key: tc.reverse})
			if e.HasSelection || e.Cursor != tc.start {
				t.Fatalf("shrink to anchor cursor=%+v selection=%v", e.Cursor, e.HasSelection)
			}
		})
	}
}

func TestRichTextEditSelectAllBindingsAndViewPageNavigation(t *testing.T) {
	for _, key := range []string{"ctrl-alt-a", "alt-a", "super-a"} {
		e := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "one"}}}, {Spans: []RichSpan{{Text: "two"}}}}})
		e.Cursor = RichPosition{Line: 1, Offset: 1}
		e.ConsumeKey(KeyEvent{Key: key})
		if !e.HasSelection || e.SelectionFrom != (RichPosition{}) || e.SelectionTo != (RichPosition{Line: 1, Offset: 3}) {
			t.Fatalf("%s selection=%+v..%+v active=%v", key, e.SelectionFrom, e.SelectionTo, e.HasSelection)
		}
		e.ConsumeKey(KeyEvent{Key: "ctrl-shift-left"})
		if e.Cursor != (RichPosition{Line: 1, Offset: 0}) || !e.HasSelection {
			t.Fatalf("%s selection did not shrink across anchor: %+v, %v", key, e.Cursor, e.HasSelection)
		}
	}
	e := NewRichTextView(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "a"}}}, {}, {Spans: []RichSpan{{Text: "c"}}}}})
	e.lastRect = Rect{H: 1}
	if got := e.ConsumeKey(KeyEvent{Key: "pgdown"}); got != Handled() || e.Cursor.Line != 1 {
		t.Fatalf("view pgdown = %+v cursor=%+v", got, e.Cursor)
	}
	if got := e.ConsumeKey(KeyEvent{Key: "shift-pgdown"}); got != Handled() || !e.HasSelection || e.Cursor.Line != 2 {
		t.Fatalf("view shift-pgdown = %+v cursor=%+v selection=%v", got, e.Cursor, e.HasSelection)
	}
	e.Cursor = RichPosition{}
	e.ClearSelection()
	if got := e.ConsumeKey(KeyEvent{Key: "ctrl-a"}); got != Handled() || e.Cursor != (RichPosition{}) {
		t.Fatalf("Ctrl+A should remain line start: %+v at %+v", got, e.Cursor)
	}
}
