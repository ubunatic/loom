// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"reflect"
	"slices"
	"testing"
)

func TestTextEntryBoldOptInAndMarkdown(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	typeEntryText(e, "a*b\\c")
	if r := entryKey(e, "ctrl-b"); r != Ignored() {
		t.Fatalf("disabled Ctrl+B = %+v, want ignored", r)
	}
	if got := e.MarkdownValue(); got != "a\\*b\\\\c" {
		t.Fatalf("plain MarkdownValue() = %q", got)
	}
	e.EnableBold()
	if r := entryKey(e, "ctrl-b"); !r.Consumed {
		t.Fatalf("typing Ctrl+B = %+v", r)
	}
	typeEntryText(e, "xy")
	if got := e.Value(); got != "a*b\\cxy" {
		t.Fatalf("Value() = %q", got)
	}
	if got := e.MarkdownValue(); got != "a\\*b\\\\c**xy**" {
		t.Fatalf("MarkdownValue() = %q", got)
	}
	if r := entryKey(e, "ctrl-z"); !r.Consumed || e.MarkdownValue() != "a\\*b\\\\c" {
		t.Fatalf("undo typing bold: result=%+v markdown=%q", r, e.MarkdownValue())
	}
	if r := entryKey(e, "ctrl-y"); !r.Consumed || e.MarkdownValue() != "a\\*b\\\\c**xy**" {
		t.Fatalf("redo typing bold: result=%+v markdown=%q", r, e.MarkdownValue())
	}
}

func TestTextEntryBoldTypingModeUndoRedo(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.EnableBold()
	entryKey(e, "ctrl-b")
	if r := entryKey(e, "ctrl-z"); !r.Consumed || e.typingBold {
		t.Fatalf("undo typing style: result=%+v typingBold=%v", r, e.typingBold)
	}
	if r := entryKey(e, "ctrl-y"); !r.Consumed || !e.typingBold {
		t.Fatalf("redo typing style: result=%+v typingBold=%v", r, e.typingBold)
	}
	typeEntryText(e, "x")
	if got := e.MarkdownValue(); got != "**x**" {
		t.Fatalf("typed after redo = %q, want bold x", got)
	}
}

func TestTextEntryBoldSelectionUndoAndPlainNotifications(t *testing.T) {
	changes := 0
	e := NewTextEntry(TextEntryOptions{OnChange: func(string) { changes++ }})
	e.EnableBold()
	e.SetValue("ab界")
	e.SetCursorOffset(0)
	entryKey(e, "shift-right")
	entryKey(e, "shift-right")
	before := changes
	if r := entryKey(e, "ctrl-b"); !r.Consumed || e.MarkdownValue() != "**ab**界" || changes != before {
		t.Fatalf("toggle selection: result=%+v markdown=%q changes=%d", r, e.MarkdownValue(), changes)
	}
	if r := entryKey(e, "ctrl-z"); !r.Consumed || e.MarkdownValue() != "ab界" || changes != before {
		t.Fatalf("undo format-only edit: result=%+v markdown=%q changes=%d", r, e.MarkdownValue(), changes)
	}
	if r := entryKey(e, "ctrl-y"); !r.Consumed || e.MarkdownValue() != "**ab**界" || changes != before {
		t.Fatalf("redo format-only edit: result=%+v markdown=%q changes=%d", r, e.MarkdownValue(), changes)
	}
}

func TestTextEntryBoldSelectionToggleAndMutationAlignment(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.EnableBold()
	e.SetValue("a界e\u0301z")
	entryKey(e, "ctrl-a")
	entryKey(e, "ctrl-b")
	if got, want := e.MarkdownValue(), "**a界e\u0301z**"; got != want {
		t.Fatalf("bold selection = %q, want %q", got, want)
	}
	entryKey(e, "ctrl-b")
	if got := e.MarkdownValue(); got != "a界e\u0301z" {
		t.Fatalf("unbold selection = %q", got)
	}
	e.SetCursorOffset(0)
	entryKey(e, "shift-right")
	entryKey(e, "ctrl-b") // make only the first grapheme bold
	e.SetCursorOffset(0)
	entryKey(e, "ctrl-a")
	entryKey(e, "ctrl-b") // a mixed selection turns fully bold
	if got := e.MarkdownValue(); got != "**a界e\u0301z**" {
		t.Fatalf("mixed selection toggle = %q", got)
	}
	e.SetCursorOffset(len("a界"))
	e.ReplaceRange(len("a界"), len("a界"), "q")
	if got := e.MarkdownValue(); got != "**a界**q**e\u0301z**" {
		t.Fatalf("ReplaceRange did not inherit typing style: %q", got)
	}
	if len(e.units) != len(e.bold) {
		t.Fatalf("unit/style lengths = %d/%d", len(e.units), len(e.bold))
	}
	e.ConsumePaste(PasteEvent{Text: "!"})
	if len(e.units) != len(e.bold) || e.units[3] != "!" || e.bold[3] {
		t.Fatalf("paste did not preserve aligned plain style: units=%d bold=%v", len(e.units), e.bold)
	}
}

func TestTextEntryBoldNewlineCtrlUPlaceholderAndClipboard(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{Placeholder: "hint"})
	e.EnableBold()
	entryKey(e, "ctrl-b")
	typeEntryText(e, "AB")
	entryKey(e, "shift-enter")
	typeEntryText(e, "C")
	if got, want := e.MarkdownValue(), "**AB**\n**C**"; got != want {
		t.Fatalf("multiline typing style = %q, want %q", got, want)
	}
	e.SetCursorOffset(len(e.Value()))
	entryKey(e, "ctrl-u")
	if got, want := e.MarkdownValue(), "**AB**\n"; got != want {
		t.Fatalf("Ctrl+U result = %q, want %q", got, want)
	}
	entryKey(e, "ctrl-z")
	if got, want := e.MarkdownValue(), "**AB**\n**C**"; got != want {
		t.Fatalf("Ctrl+U undo = %q, want %q", got, want)
	}
	e.SetCursorOffset(0)
	entryKey(e, "shift-right")
	entryKey(e, "ctrl-c")
	if e.clipboard != "A" {
		t.Fatalf("clipboard copied %q, want plain text A", e.clipboard)
	}

	p := NewTextEntry(TextEntryOptions{Placeholder: "hint"})
	p.EnableBold()
	entryKey(p, "ctrl-b")
	if !p.AcceptPlaceholder() || p.MarkdownValue() != "**hint**" {
		t.Fatalf("placeholder did not inherit typing style: %q", p.MarkdownValue())
	}
}

func TestTextEntryBoldExactMutationPaths(t *testing.T) {
	t.Run("backspace-and-delete", func(t *testing.T) {
		e := NewTextEntry(TextEntryOptions{})
		e.EnableBold()
		e.SetValue("abcd")
		entryKey(e, "ctrl-a")
		entryKey(e, "ctrl-b")
		e.SetCursorOffset(2)
		entryKey(e, "backspace")
		if e.Value() != "acd" || e.MarkdownValue() != "**acd**" || e.CursorOffset() != 1 {
			t.Fatalf("Backspace result: value=%q markdown=%q caret=%d", e.Value(), e.MarkdownValue(), e.CursorOffset())
		}
		entryKey(e, "delete")
		if e.Value() != "ad" || e.MarkdownValue() != "**ad**" || e.CursorOffset() != 1 {
			t.Fatalf("Delete result: value=%q markdown=%q caret=%d", e.Value(), e.MarkdownValue(), e.CursorOffset())
		}
		entryKey(e, "ctrl-z")
		if e.Value() != "acd" || e.MarkdownValue() != "**acd**" || e.CursorOffset() != 1 {
			t.Fatalf("Delete undo: value=%q markdown=%q caret=%d", e.Value(), e.MarkdownValue(), e.CursorOffset())
		}
	})

	t.Run("cut-plain-and-undo", func(t *testing.T) {
		e := NewTextEntry(TextEntryOptions{})
		e.EnableBold()
		e.SetValue("left MID right")
		e.SetCursorOffset(5)
		for range "MID" {
			entryKey(e, "shift-right")
		}
		entryKey(e, "ctrl-b")
		entryKey(e, "ctrl-x")
		if e.Value() != "left  right" || e.MarkdownValue() != "left  right" || e.clipboard != "MID" || e.CursorOffset() != 5 {
			t.Fatalf("cut result: value=%q markdown=%q clipboard=%q caret=%d", e.Value(), e.MarkdownValue(), e.clipboard, e.CursorOffset())
		}
		entryKey(e, "ctrl-z")
		if e.Value() != "left MID right" || e.MarkdownValue() != "left **MID** right" || e.clipboard != "MID" || e.CursorOffset() != 8 {
			t.Fatalf("cut undo: value=%q markdown=%q clipboard=%q caret=%d", e.Value(), e.MarkdownValue(), e.clipboard, e.CursorOffset())
		}
	})

	t.Run("internal-and-bracketed-paste", func(t *testing.T) {
		e := NewTextEntry(TextEntryOptions{})
		e.EnableBold()
		e.clipboard = "x*y"
		entryKey(e, "ctrl-b")
		entryKey(e, "ctrl-v")
		if e.Value() != "x*y" || e.MarkdownValue() != "**x\\*y**" || e.CursorOffset() != len("x*y") {
			t.Fatalf("internal paste: value=%q markdown=%q caret=%d", e.Value(), e.MarkdownValue(), e.CursorOffset())
		}
		e.ConsumePaste(PasteEvent{Text: "界\\"})
		if e.Value() != "x*y界\\" || e.MarkdownValue() != "**x\\*y界\\\\**" || e.CursorOffset() != len("x*y界\\") {
			t.Fatalf("bracketed paste: value=%q markdown=%q caret=%d", e.Value(), e.MarkdownValue(), e.CursorOffset())
		}
		entryKey(e, "ctrl-z")
		if e.Value() != "x*y" || e.MarkdownValue() != "**x\\*y**" || e.CursorOffset() != len("x*y") {
			t.Fatalf("bracketed paste undo: value=%q markdown=%q caret=%d", e.Value(), e.MarkdownValue(), e.CursorOffset())
		}
	})

	t.Run("selection-replacement-and-setvalue-reset", func(t *testing.T) {
		e := NewTextEntry(TextEntryOptions{})
		e.EnableBold()
		e.SetValue("one TWO three")
		e.SetCursorOffset(len(e.Value()))
		entryKey(e, "ctrl-b")
		e.SetCursorOffset(4)
		for range "TWO" {
			entryKey(e, "shift-right")
		}
		typeEntryText(e, "new")
		if e.Value() != "one new three" || e.MarkdownValue() != "one **new** three" || e.CursorOffset() != 7 {
			t.Fatalf("selection replacement: value=%q markdown=%q caret=%d", e.Value(), e.MarkdownValue(), e.CursorOffset())
		}
		e.SetValue("replacement")
		if e.Value() != "replacement" || e.MarkdownValue() != "replacement" || e.typingBold || slices.Contains(e.bold, true) || e.CursorOffset() != len("replacement") {
			t.Fatalf("SetValue did not reset format: value=%q markdown=%q typingBold=%v bold=%v caret=%d", e.Value(), e.MarkdownValue(), e.typingBold, e.bold, e.CursorOffset())
		}
	})

	t.Run("setvalue-clears-format-redo", func(t *testing.T) {
		e := NewTextEntry(TextEntryOptions{})
		e.EnableBold()
		e.SetValue("xy")
		entryKey(e, "ctrl-a")
		entryKey(e, "ctrl-b")
		entryKey(e, "ctrl-z")
		if len(e.redo) == 0 || e.MarkdownValue() != "xy" {
			t.Fatalf("setup did not create formatting redo: markdown=%q redo=%d", e.MarkdownValue(), len(e.redo))
		}
		e.SetValue("xy")
		if len(e.redo) != 0 || entryKey(e, "ctrl-y") != Ignored() || e.MarkdownValue() != "xy" {
			t.Fatalf("SetValue retained stale formatting redo: markdown=%q redo=%d", e.MarkdownValue(), len(e.redo))
		}
	})
}

func TestTextEntryBoldMarkdownExactWhitespaceAdjacentAndPure(t *testing.T) {
	changes := 0
	e := NewTextEntry(TextEntryOptions{OnChange: func(string) { changes++ }})
	e.EnableBold()
	e.SetValue("  a#b!c_(x)[y]\\z  plain界 c  \n    \nΩ")
	plain := e.Value()
	e.SetCursorOffset(0)
	for range "  a#b!c_(x)[y]\\z  " {
		entryKey(e, "shift-right")
	}
	entryKey(e, "ctrl-b")
	// Add a second bold run after unbolded text, and bold only the final Unicode unit.
	e.SetCursorOffset(len("  a#b!c_(x)[y]\\z  plain界 "))
	entryKey(e, "shift-right")
	entryKey(e, "ctrl-b")
	want := "  **a\\#b\\!c\\_\\(x\\)\\[y\\]\\\\z**  plain界 **c**  \n    \nΩ"
	if got := e.MarkdownValue(); got != want {
		t.Fatalf("MarkdownValue() = %q, want %q", got, want)
	}
	before := e.snapshot()
	undo, redo := slices.Clone(e.undo), slices.Clone(e.redo)
	beforeChanges := changes
	if got := e.MarkdownValue(); got != want || e.Value() != plain || !reflect.DeepEqual(e.snapshot(), before) || !reflect.DeepEqual(e.undo, undo) || !reflect.DeepEqual(e.redo, redo) || changes != beforeChanges {
		t.Fatalf("serialization mutated state: markdown=%q value=%q changes=%d", got, e.Value(), changes)
	}
}

func TestTextEntryBoldWhitespaceOnlySelectionIsHarmless(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.EnableBold()
	e.SetValue(" \n  ")
	entryKey(e, "ctrl-a")
	undoLen := len(e.undo)
	entryKey(e, "ctrl-b")
	if e.MarkdownValue() != " \n  " || len(e.undo) != undoLen+1 || !slices.Contains(e.bold, true) {
		t.Fatalf("whitespace-only formatting result: markdown=%q undo=%d bold=%v", e.MarkdownValue(), len(e.undo), e.bold)
	}
	entryKey(e, "ctrl-z")
	e.SetCursorOffset(1)
	entryKey(e, "shift-right") // select the newline only
	undoLen = len(e.undo)
	entryKey(e, "ctrl-b")
	if e.MarkdownValue() != " \n  " || len(e.undo) != undoLen || slices.Contains(e.bold, true) {
		t.Fatalf("newline-only formatting changed state: markdown=%q undo=%d bold=%v", e.MarkdownValue(), len(e.undo), e.bold)
	}
}

func TestTextEntryMarkdownWhitespaceNewlinesAndEscapes(t *testing.T) {
	e := NewTextEntry(TextEntryOptions{})
	e.EnableBold()
	e.SetValue("  *hi*\n\n界  ")
	e.SetCursorOffset(2)
	for range "*hi*" {
		entryKey(e, "shift-right")
	}
	entryKey(e, "ctrl-b")
	if got, want := e.MarkdownValue(), "  **\\*hi\\***\n\n界  "; got != want {
		t.Fatalf("MarkdownValue() = %q, want %q", got, want)
	}
}
