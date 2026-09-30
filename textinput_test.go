// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom_test

import (
	"testing"

	"codeberg.org/ubunatic/loom"
)

func typeKeys(t *loom.TextInput, texts ...string) {
	for _, s := range texts {
		t.ConsumeKey(loom.KeyEvent{Text: s})
	}
}

func TestTextInputInsertMidString(t *testing.T) {
	in := loom.NewTextInput("ac")
	if in.Caret() != 2 {
		t.Fatalf("caret after seed = %d, want 2", in.Caret())
	}
	in.ConsumeKey(loom.KeyEvent{Key: "left"}) // caret between a and c
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
	in.ConsumeKey(loom.KeyEvent{Key: "left"}) // caret before c
	in.ConsumeKey(loom.KeyEvent{Key: "backspace"})
	if in.Value() != "ac" {
		t.Errorf("backspace at caret = %q, want ac", in.Value())
	}
	if in.Caret() != 1 {
		t.Errorf("caret = %d, want 1", in.Caret())
	}
	// Backspace at the start is a no-op.
	in.ConsumeKey(loom.KeyEvent{Key: "home"})
	in.ConsumeKey(loom.KeyEvent{Key: "backspace"})
	if in.Value() != "ac" || in.Caret() != 0 {
		t.Errorf("backspace at start changed state: %q caret=%d", in.Value(), in.Caret())
	}
}

func TestTextInputDelete(t *testing.T) {
	in := loom.NewTextInput("abc")
	in.ConsumeKey(loom.KeyEvent{Key: "home"})
	in.ConsumeKey(loom.KeyEvent{Key: "delete"}) // remove 'a'
	if in.Value() != "bc" || in.Caret() != 0 {
		t.Errorf("delete = %q caret=%d, want bc caret=0", in.Value(), in.Caret())
	}
	in.ConsumeKey(loom.KeyEvent{Key: "end"})
	in.ConsumeKey(loom.KeyEvent{Key: "delete"}) // no-op at end
	if in.Value() != "bc" {
		t.Errorf("delete at end mutated value: %q", in.Value())
	}
}

func TestTextInputHomeEnd(t *testing.T) {
	in := loom.NewTextInput("hello")
	in.ConsumeKey(loom.KeyEvent{Key: "home"})
	if in.Caret() != 0 {
		t.Errorf("home caret = %d, want 0", in.Caret())
	}
	typeKeys(in, "X") // insert at start
	if in.Value() != "Xhello" {
		t.Errorf("insert after home = %q, want Xhello", in.Value())
	}
	in.ConsumeKey(loom.KeyEvent{Key: "end"})
	typeKeys(in, "!")
	if in.Value() != "Xhello!" {
		t.Errorf("insert after end = %q, want Xhello!", in.Value())
	}
}

func TestTextInputConsumedReporting(t *testing.T) {
	in := loom.NewTextInput("")
	// Editing keys are consumed; navigation/commit keys are not.
	if !in.ConsumeKey(loom.KeyEvent{Text: "a"}).Consumed {
		t.Error("text key should be consumed")
	}
	if !in.ConsumeKey(loom.KeyEvent{Key: "left"}).Consumed {
		t.Error("left should be consumed")
	}
	if in.ConsumeKey(loom.KeyEvent{Key: "enter"}).Consumed {
		t.Error("enter should NOT be consumed (host owns it)")
	}
	if in.ConsumeKey(loom.KeyEvent{Key: "esc"}).Consumed {
		t.Error("esc should NOT be consumed (host owns it)")
	}
}

func TestTextInputPasteReplacesNewlinesWithSpaces(t *testing.T) {
	in := loom.NewTextInput("ac")
	in.ConsumeKey(loom.KeyEvent{Key: "left"})
	if !loom.DispatchPasteEvent(in, loom.PasteEvent{Text: "b\nc\r\nd"}).Consumed {
		t.Fatal("paste was not consumed")
	}
	if got := in.Value(); got != "ab c dc" {
		t.Fatalf("paste value = %q, want %q", got, "ab c dc")
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

func TestTextInputMaskedDrawKeepsValueAndMasksPerRune(t *testing.T) {
	in := loom.NewTextInput("界🙂")
	in.Prompt = "Password: "
	in.Mask = '•'
	in.ConsumeKey(loom.KeyEvent{Key: "left"})

	c := loom.NewCanvas(30, 1)
	c.ColorProfile = loom.ColorProfileNone
	in.Draw(c, c.Bounds(), true)

	if got := in.Value(); got != "界🙂" {
		t.Fatalf("masked value = %q, want original runes", got)
	}
	if c.CursorX != loom.StringWidth("Password: •") {
		t.Errorf("masked caret x = %d, want %d", c.CursorX, loom.StringWidth("Password: •"))
	}
	if got := c.Get(10, 0).Text; got != "•" {
		t.Errorf("first masked cell = %q, want bullet", got)
	}
	if got := c.Get(11, 0).Text; got != "•" {
		t.Errorf("second masked cell = %q, want bullet", got)
	}
	for x := 0; x < 30; x++ {
		if got := c.Get(x, 0).Text; got == "界" || got == "🙂" {
			t.Errorf("masked render exposed secret at x=%d: %q", x, got)
		}
	}
}

func TestTextInputMaskedWideMaskCaretAndDeletion(t *testing.T) {
	in := loom.NewTextInput("ab界")
	in.Mask = '界'
	in.ConsumeKey(loom.KeyEvent{Key: "home"})
	in.ConsumeKey(loom.KeyEvent{Key: "right"})
	in.ConsumeKey(loom.KeyEvent{Key: "right"})
	c := loom.NewCanvas(20, 1)
	c.ColorProfile = loom.ColorProfileNone
	in.Draw(c, c.Bounds(), true)
	if c.CursorX != loom.StringWidth("界界") {
		t.Errorf("wide mask caret x = %d, want %d", c.CursorX, loom.StringWidth("界界"))
	}
	in.ConsumeKey(loom.KeyEvent{Key: "delete"})
	if got := in.Value(); got != "ab" || in.Caret() != 2 {
		t.Errorf("delete in masked input = %q caret=%d, want %q caret=2", got, in.Caret(), "ab")
	}
}

func TestTextInputScrollsWideValueWithCaret(t *testing.T) {
	in := loom.NewTextInput("ab界🙂cdef")
	c := loom.NewCanvas(5, 1)
	c.ColorProfile = loom.ColorProfileNone
	in.Draw(c, c.Bounds(), true)
	if c.CursorX < 0 || c.CursorX >= 5 {
		t.Fatalf("end caret x = %d, want visible within width 5", c.CursorX)
	}
	if got := c.Get(0, 0).Text; got != "‹" {
		t.Errorf("left clipped marker = %q, want ‹", got)
	}
	if got := c.Get(4, 0).Text; got == "›" {
		t.Errorf("end caret showed a right marker despite no hidden suffix")
	}
	// Move inside the value with hidden content on both sides.
	in.ConsumeKey(loom.KeyEvent{Key: "home"})
	for range 4 {
		in.ConsumeKey(loom.KeyEvent{Key: "right"})
	}
	c.Clear()
	in.Draw(c, c.Bounds(), true)
	if got := c.Get(0, 0).Text; got != "‹" {
		t.Errorf("interior left marker = %q, want ‹", got)
	}
	if got := c.Get(4, 0).Text; got != "›" {
		t.Errorf("interior right marker = %q, want ›", got)
	}
	in.ConsumeKey(loom.KeyEvent{Key: "home"})
	c.Clear()
	in.Draw(c, c.Bounds(), true)
	if c.CursorX != 0 {
		t.Errorf("home caret x = %d, want 0", c.CursorX)
	}
	if got := c.Get(0, 0).Text; got != "a" {
		t.Errorf("home window starts with %q, want a", got)
	}
	in.ConsumeKey(loom.KeyEvent{Key: "end"})
	in.ConsumeKey(loom.KeyEvent{Key: "backspace"})
	c.Clear()
	in.Draw(c, c.Bounds(), true)
	if c.CursorX < 0 || c.CursorX >= 5 {
		t.Errorf("caret after end backspace x = %d, want visible", c.CursorX)
	}
	if got := in.Value(); got != "ab界🙂cde" {
		t.Errorf("value after end backspace = %q, want ab界🙂cde", got)
	}
}
