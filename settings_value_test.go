// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

// ── A6: Settings value rendering ──────────────────────────────────────────────
// KindString supports inline editing: Enter/→ enters edit mode, printable keys
// append, Backspace deletes, Enter commits, Esc cancels (see issue 014).

func rowText(c *loom.Canvas, y int) string {
	var b strings.Builder
	for x := 0; x < c.Cols(); x++ {
		t := c.Get(x, y).Text
		if t == "" {
			continue // wide-char continuation cell
		}
		b.WriteString(t)
	}
	return b.String()
}

func TestSettingsStringValueRenders(t *testing.T) {
	str := "hello"
	s := loom.NewSettings([]loom.Setting{
		{Label: "name", Kind: loom.KindString, Str: &str},
	})
	c := loom.NewCanvas(40, 1)
	s.Draw(c, c.Bounds())
	row := rowText(c, 0)
	if !strings.Contains(row, "name") || !strings.Contains(row, "hello") {
		t.Errorf("string setting row = %q, want label + value", row)
	}
}

func TestSettingsStringInlineEdit(t *testing.T) {
	str := "abc"
	s := loom.NewSettings([]loom.Setting{
		{Label: "name", Kind: loom.KindString, Str: &str},
	})
	// Outside edit mode, typing must not mutate the value.
	s.HandleKey(loom.KeyEvent{Text: "z"})
	if str != "abc" || s.Editing() {
		t.Fatalf("value/edit-state changed before entering edit mode: %q editing=%v", str, s.Editing())
	}
	// Enter edit mode, type, delete, and commit.
	s.HandleKey(loom.KeyEvent{Key: "enter"})
	if !s.Editing() {
		t.Fatal("Enter on a KindString row did not enter edit mode")
	}
	s.HandleKey(loom.KeyEvent{Text: "d"})
	s.HandleKey(loom.KeyEvent{Text: "e"})
	s.HandleKey(loom.KeyEvent{Key: "backspace"})
	q := s.HandleKey(loom.KeyEvent{Key: "enter"})
	if q || s.Editing() {
		t.Errorf("Enter should commit without quitting: quit=%v editing=%v", q, s.Editing())
	}
	if str != "abcd" {
		t.Errorf("after edit, value = %q, want %q", str, "abcd")
	}
}

func TestSettingsStringEditCancel(t *testing.T) {
	str := "abc"
	s := loom.NewSettings([]loom.Setting{
		{Label: "name", Kind: loom.KindString, Str: &str},
	})
	s.HandleKey(loom.KeyEvent{Key: "enter"}) // begin edit
	s.HandleKey(loom.KeyEvent{Text: "x"})
	if str != "abcx" {
		t.Fatalf("typing in edit mode did not append: %q", str)
	}
	q := s.HandleKey(loom.KeyEvent{Key: "esc"}) // cancel restores prior value
	if q {
		t.Error("Esc during edit should cancel the edit, not quit the widget")
	}
	if s.Editing() {
		t.Error("Esc did not leave edit mode")
	}
	if str != "abc" {
		t.Errorf("Esc did not cancel edit, value = %q, want abc", str)
	}
	// After the edit is cancelled, Esc quits the widget as usual.
	if !s.HandleKey(loom.KeyEvent{Key: "esc"}) {
		t.Error("Esc outside edit mode should quit the widget")
	}
}

func TestSettingsChoiceValueRendersMarkers(t *testing.T) {
	idx := 1
	s := loom.NewSettings([]loom.Setting{
		{Label: "theme", Kind: loom.KindChoice, Options: []string{"dark", "light"}, Index: &idx},
	})
	c := loom.NewCanvas(40, 1)
	s.Draw(c, c.Bounds())
	row := rowText(c, 0)
	if !strings.Contains(row, "●") || !strings.Contains(row, "light") {
		t.Errorf("choice row = %q, want selected ● on 'light'", row)
	}
	if !strings.Contains(row, "○") || !strings.Contains(row, "dark") {
		t.Errorf("choice row = %q, want unselected ○ on 'dark'", row)
	}
}
