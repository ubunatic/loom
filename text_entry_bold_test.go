// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import "testing"

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
