// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func taType(t *loom.TextArea, texts ...string) {
	for _, s := range texts {
		t.HandleKey(loom.KeyEvent{Text: s})
	}
}

func TestTextAreaValueRoundTrip(t *testing.T) {
	ta := loom.NewTextArea("line one\nline two")
	if ta.Value() != "line one\nline two" {
		t.Fatalf("round trip = %q", ta.Value())
	}
	if ta.LineCount() != 2 {
		t.Errorf("LineCount = %d, want 2", ta.LineCount())
	}
	r, col := ta.Caret()
	if r != 1 || col != len("line two") {
		t.Errorf("seed caret = (%d,%d), want (1,8)", r, col)
	}
}

func TestTextAreaEnterSplitsLine(t *testing.T) {
	ta := loom.NewTextArea("abcd")
	ta.HandleKey(loom.KeyEvent{Key: "home"})
	ta.HandleKey(loom.KeyEvent{Key: "right"})
	ta.HandleKey(loom.KeyEvent{Key: "right"}) // caret after "ab"
	ta.HandleKey(loom.KeyEvent{Key: "enter"})
	if ta.Value() != "ab\ncd" {
		t.Errorf("enter split = %q, want \"ab\\ncd\"", ta.Value())
	}
	r, col := ta.Caret()
	if r != 1 || col != 0 {
		t.Errorf("caret after split = (%d,%d), want (1,0)", r, col)
	}
}

func TestTextAreaBackspaceJoinsLines(t *testing.T) {
	ta := loom.NewTextArea("ab\ncd")
	ta.HandleKey(loom.KeyEvent{Key: "home"}) // start of "cd" (caret seeded on last line)
	ta.HandleKey(loom.KeyEvent{Key: "backspace"})
	if ta.Value() != "abcd" {
		t.Errorf("backspace join = %q, want abcd", ta.Value())
	}
	r, col := ta.Caret()
	if r != 0 || col != 2 {
		t.Errorf("caret after join = (%d,%d), want (0,2)", r, col)
	}
}

func TestTextAreaDeleteMergesNextLine(t *testing.T) {
	ta := loom.NewTextArea("ab\ncd")
	ta.HandleKey(loom.KeyEvent{Key: "up"})  // to line 0
	ta.HandleKey(loom.KeyEvent{Key: "end"}) // end of "ab"
	ta.HandleKey(loom.KeyEvent{Key: "delete"})
	if ta.Value() != "abcd" {
		t.Errorf("delete merge = %q, want abcd", ta.Value())
	}
}

func TestTextAreaVerticalCaretClamp(t *testing.T) {
	ta := loom.NewTextArea("longline\nx")
	// Caret seeded at (1,1) on "x"; moving up must clamp col to len("longline")? No —
	// up keeps col then clamps to the shorter target line. Here target is longer.
	ta.HandleKey(loom.KeyEvent{Key: "up"})
	_, col := ta.Caret()
	if col != 1 {
		t.Errorf("col after up = %d, want 1 (preserved)", col)
	}
	// Move to end of the long line, then down onto the short line clamps col.
	ta.HandleKey(loom.KeyEvent{Key: "end"})
	ta.HandleKey(loom.KeyEvent{Key: "down"})
	r, col := ta.Caret()
	if r != 1 || col != 1 {
		t.Errorf("caret after down-clamp = (%d,%d), want (1,1)", r, col)
	}
}

func TestTextAreaScrollKeepsCaretVisible(t *testing.T) {
	ta := loom.NewTextArea("l0\nl1\nl2\nl3\nl4") // 5 lines; caret on l4
	c := loom.NewCanvas(10, 3)                   // only 3 rows visible
	ta.Draw(c, c.Bounds(), true)
	// Caret is on the last line; it must be visible (cursor set within bounds).
	if c.CursorY < 0 || c.CursorY >= 3 {
		t.Errorf("caret not scrolled into view: CursorY=%d", c.CursorY)
	}
}

func TestTextAreaPlaceholder(t *testing.T) {
	ta := loom.NewTextArea("")
	ta.Placeholder = "body here"
	c := loom.NewCanvas(20, 3)
	ta.Draw(c, c.Bounds(), true)
	if got := strings.TrimRight(rowText(c, 0), " "); got != "body here" {
		t.Errorf("placeholder not drawn, row0 = %q", got)
	}
	// Typing replaces the placeholder.
	taType(ta, "x")
	c.Clear()
	ta.Draw(c, c.Bounds(), true)
	if got := strings.TrimRight(rowText(c, 0), " "); got != "x" {
		t.Errorf("after typing, row0 = %q, want \"x\"", got)
	}
}
