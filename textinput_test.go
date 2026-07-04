// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

func typeKeys(t *loom.TextInput, texts ...string) {
	for _, s := range texts {
		t.HandleKey(loom.KeyEvent{Text: s})
	}
}

func TestTextInputInsertMidString(t *testing.T) {
	in := loom.NewTextInput("ac")
	if in.Caret() != 2 {
		t.Fatalf("caret after seed = %d, want 2", in.Caret())
	}
	in.HandleKey(loom.KeyEvent{Key: "left"}) // caret between a and c
	typeKeys(in, "b")
	if in.Value() != "abc" {
		t.Errorf("mid-string insert = %q, want abc", in.Value())
	}
	if in.Caret() != 2 {
		t.Errorf("caret after insert = %d, want 2", in.Caret())
	}
}

func TestTextInputBackspaceAtCaret(t *testing.T) {
	in := loom.NewTextInput("abc")
	in.HandleKey(loom.KeyEvent{Key: "left"}) // caret before c
	in.HandleKey(loom.KeyEvent{Key: "backspace"})
	if in.Value() != "ac" {
		t.Errorf("backspace at caret = %q, want ac", in.Value())
	}
	if in.Caret() != 1 {
		t.Errorf("caret = %d, want 1", in.Caret())
	}
	// Backspace at the start is a no-op.
	in.HandleKey(loom.KeyEvent{Key: "home"})
	in.HandleKey(loom.KeyEvent{Key: "backspace"})
	if in.Value() != "ac" || in.Caret() != 0 {
		t.Errorf("backspace at start changed state: %q caret=%d", in.Value(), in.Caret())
	}
}

func TestTextInputDelete(t *testing.T) {
	in := loom.NewTextInput("abc")
	in.HandleKey(loom.KeyEvent{Key: "home"})
	in.HandleKey(loom.KeyEvent{Key: "delete"}) // remove 'a'
	if in.Value() != "bc" || in.Caret() != 0 {
		t.Errorf("delete = %q caret=%d, want bc caret=0", in.Value(), in.Caret())
	}
	in.HandleKey(loom.KeyEvent{Key: "end"})
	in.HandleKey(loom.KeyEvent{Key: "delete"}) // no-op at end
	if in.Value() != "bc" {
		t.Errorf("delete at end mutated value: %q", in.Value())
	}
}

func TestTextInputHomeEnd(t *testing.T) {
	in := loom.NewTextInput("hello")
	in.HandleKey(loom.KeyEvent{Key: "home"})
	if in.Caret() != 0 {
		t.Errorf("home caret = %d, want 0", in.Caret())
	}
	typeKeys(in, "X") // insert at start
	if in.Value() != "Xhello" {
		t.Errorf("insert after home = %q, want Xhello", in.Value())
	}
	in.HandleKey(loom.KeyEvent{Key: "end"})
	typeKeys(in, "!")
	if in.Value() != "Xhello!" {
		t.Errorf("insert after end = %q, want Xhello!", in.Value())
	}
}

func TestTextInputConsumedReporting(t *testing.T) {
	in := loom.NewTextInput("")
	// Editing keys are consumed; navigation/commit keys are not.
	if !in.HandleKey(loom.KeyEvent{Text: "a"}) {
		t.Error("text key should be consumed")
	}
	if !in.HandleKey(loom.KeyEvent{Key: "left"}) {
		t.Error("left should be consumed")
	}
	if in.HandleKey(loom.KeyEvent{Key: "enter"}) {
		t.Error("enter should NOT be consumed (host owns it)")
	}
	if in.HandleKey(loom.KeyEvent{Key: "esc"}) {
		t.Error("esc should NOT be consumed (host owns it)")
	}
}

func TestTextInputDrawPlaceholderAndCaret(t *testing.T) {
	in := loom.NewTextInput("")
	in.Prompt = "> "
	in.Placeholder = "type here"
	c := loom.NewCanvas(20, 1)
	in.Draw(c, c.Bounds(), true)
	// Placeholder shows after the prompt; caret sits right after the prompt.
	if c.CursorX != 2 || c.CursorY != 0 {
		t.Errorf("empty caret at (%d,%d), want (2,0)", c.CursorX, c.CursorY)
	}

	in.SetValue("ab")
	c.Clear()
	in.Draw(c, c.Bounds(), true)
	if c.CursorX != 4 { // prompt(2) + "ab"(2)
		t.Errorf("caret after value = %d, want 4", c.CursorX)
	}

	// Unfocused: no cursor is placed.
	c.Clear()
	in.Draw(c, c.Bounds(), false)
	if c.CursorX != -1 || c.CursorY != -1 {
		t.Errorf("unfocused field set cursor to (%d,%d), want hidden", c.CursorX, c.CursorY)
	}
}
