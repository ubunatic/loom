// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"strings"
	"testing"
)

func TestRichTextEditF1HelpListsBindingsAndKeepsFormattingSeparate(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "format"}}}}})
	if result := edit.ConsumeKey(KeyEvent{Key: "f1"}); !result.Consumed || edit.helpPopup == nil {
		t.Fatalf("F1 result/help popup = %+v/%v, want open help", result, edit.helpPopup)
	}
	view, ok := edit.helpPopup.Inner.(*richTextEditHelp)
	if !ok {
		t.Fatalf("help popup inner = %T, want scrollable View", edit.helpPopup.Inner)
	}
	content := strings.Join(view.plainLines(72), "\n")
	for _, binding := range []string{"⌥F", "⌃S", "⌃⌥S", "⌃B", "⌃I", "⌃U", "⌃Space", "⌃C", "⌃Insert", "⌃X", "⇧Delete", "⌃V", "⇧Insert", "⌃Z", "⌃Y", "⌃R", "⇧⌃Z", "⌃Left", "⌃Right", "⇧Home", "⇧End", "⌃D", "F7", "F1", "Tab", "⇧Tab", "Esc"} {
		if !strings.Contains(content, binding) {
			t.Errorf("help omits binding %q", binding)
		}
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "esc"}); !result.Consumed || edit.helpPopup != nil {
		t.Fatal("Escape did not close F1 help")
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-space"}); !result.Consumed || !edit.ShowPopover || (!edit.HasSelection && !edit.popoverAtCursor) || edit.helpPopup != nil {
		t.Fatalf("Ctrl+Space did not open its formatting popover separately: %+v", result)
	}
}

func TestRichTextEditF1HelpFitsSmallBoundsAndScrolls(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{}}})
	edit.ConsumeKey(KeyEvent{Key: "f1"})
	canvas := NewCanvas(16, 5)
	edit.Draw(canvas, canvas.Bounds())
	if edit.helpPopup.innerRect.W <= 0 || edit.helpPopup.innerRect.H <= 0 {
		t.Fatalf("small help popup content bounds = %+v", edit.helpPopup.innerRect)
	}
	for row := 0; row < canvas.Rows(); row++ {
		if got := StringWidth(canvas.Row(row)); got > canvas.Cols() {
			t.Errorf("row %d width = %d, exceeds %d: %q", row, got, canvas.Cols(), canvas.Row(row))
		}
	}
	view := edit.helpPopup.Inner.(*richTextEditHelp)
	edit.ConsumeKey(KeyEvent{Key: "down"})
	if view.scroll == 0 {
		t.Fatal("Down did not scroll the help contents")
	}
}

func TestKeyHelpSectionsAlignsKeysAndFitsWidth(t *testing.T) {
	sections := []HelpSection{
		{Title: "One", Entries: []HelpEntry{{Keys: []string{"ctrl-alt-s"}, Label: "Save as"}, {Keys: []string{"f1"}, Label: "界面 a long description that must wrap"}}},
		{Title: "Two", Entries: []HelpEntry{{Cap: "Drag", Label: "Select"}}},
	}
	for _, width := range []int{12, 30, 60} {
		rows := KeyHelpSections(sections, width)
		keyCol := -1
		for i, row := range rows {
			if got := StringWidth(row.Key) + 2 + StringWidth(row.Text); !row.Header && got > width {
				t.Errorf("width %d row %d is %d wide: %+v", width, i, got, row)
			}
			if got := StringWidth(row.Text); row.Header && got > width {
				t.Errorf("width %d header %d is %d wide", width, i, got)
			}
			if !row.Header && row.Text != "" {
				if keyCol >= 0 && StringWidth(row.Key) != keyCol {
					t.Errorf("width %d: key column not aligned at row %d", width, i)
				}
				keyCol = StringWidth(row.Key)
			}
		}
	}
	rows := KeyHelpSections(sections, 60)
	want := []KeyHelpRow{
		{Header: true, Text: "One"},
		{Key: "⌃⌥S ", Text: "Save as"},
		{Key: "F1  ", Text: "界面 a long description that must wrap"},
		{},
		{Header: true, Text: "Two"},
		{Key: "Drag", Text: "Select"},
	}
	if len(rows) != len(want) {
		t.Fatalf("rows = %+v", rows)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("row %d = %+v, want %+v", i, rows[i], want[i])
		}
	}
}

func TestRichTextEditHelpIsSectionedAndTakesHostSections(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{}}})
	edit.AddHelpSection(HelpSection{Title: "Host", Entries: []HelpEntry{{Cap: "F3/⌃F", Label: "Search"}}})
	edit.ConsumeKey(KeyEvent{Key: "f1"})
	view := edit.helpPopup.Inner.(*richTextEditHelp)
	text := strings.Join(view.plainLines(68), "\n")
	for _, want := range []string{"File\n", "Navigation\n", "Selection\n", "Editing\n", "Formatting\n", "Host\n", "⌃S  ", "Save as", "F3/⌃F"} {
		if !strings.Contains(text, want) {
			t.Errorf("help omits %q:\n%s", want, text)
		}
	}
	if strings.Index(text, "Host") < strings.Index(text, "Help") {
		t.Error("host section must follow built-in sections")
	}
	canvas := NewCanvas(80, 24)
	edit.Draw(canvas, canvas.Bounds())
	for row := 0; row < canvas.Rows(); row++ {
		if StringWidth(canvas.Row(row)) > canvas.Cols() {
			t.Errorf("row %d overflows: %q", row, canvas.Row(row))
		}
	}
}

func TestSpeccedHelpRefsResolve(t *testing.T) {
	if b, l, ok := SpeccedDefaults.RichTextEdit.helpRef("HotkeySave"); !ok || b != "ctrl-s" || l != "Save" {
		t.Fatalf("HotkeySave ref = %q %q %v", b, l, ok)
	}
	if _, _, ok := SpeccedDefaults.RichTextEdit.helpRef("Nope"); ok {
		t.Fatal("unknown ref resolved")
	}
}

func TestRichTextEditHotkeyHintClipsAtWidthBoundaries(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{}}})
	for _, width := range []int{8, 20, 21, 22, 38, 39, 58, 59, 60, 80} {
		hint := edit.HotkeyHint(width)
		if got := StringWidth(hint); got > width {
			t.Errorf("hint width at %d columns = %d: %q", width, got, hint)
		}
		if width >= 10 && !strings.Contains(hint, "F1 Help") {
			t.Errorf("hint at %d columns omits F1 Help: %q", width, hint)
		}
		if width == 80 && (!strings.Contains(hint, "F7 View") || strings.Contains(hint, "Save as")) {
			t.Errorf("80-column hint omits complete shortcuts: %q", hint)
		}
	}
}
