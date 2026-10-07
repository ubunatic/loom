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
	content := strings.Join(view.lines, "\n")
	for _, binding := range []string{"⌥F", "⌃S", "⌃⌥S", "⌃B", "⌃I", "⌃U", "⌃Space", "⌃C", "⌃Insert", "⌃X", "⇧Delete", "⌃V", "⇧Insert", "⌃Z", "⌃Y", "⌃R", "⇧⌃Y", "⇧⌃Z", "⌃Left", "⌃Right", "⇧Home", "⇧End", "F5", "F7", "F1", "Tab", "⇧Tab", "Escape"} {
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

func TestRichTextEditHelpWrappingPreservesBindingsWithinWidth(t *testing.T) {
	text := "Ctrl+Alt+S Save as · 界面 filename"
	rows := wrapRichTextHelpLine(text, 12)
	if got, want := strings.Join(strings.Fields(strings.Join(rows, " ")), " "), strings.Join(strings.Fields(text), " "); got != want {
		t.Fatalf("wrapped help words = %q, want %q", got, want)
	}
	for i, row := range rows {
		if got := StringWidth(row); got > 12 {
			t.Errorf("wrapped row %d width = %d, exceeds 12: %q", i, got, row)
		}
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
		if width == 80 && (!strings.Contains(hint, "F7 View") || !strings.Contains(hint, "⌃⌥S Save as")) {
			t.Errorf("80-column hint omits complete shortcuts: %q", hint)
		}
	}
}
