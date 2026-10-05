// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRichTextEditFileBarIsOptInAndReservesOneRow(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	canvas := NewCanvas(30, 5)
	edit.Draw(canvas, canvas.Bounds())
	if edit.lastRect.H != 5 {
		t.Fatalf("default editor height = %d, want unchanged 5", edit.lastRect.H)
	}
	edit.ShowFileBar = true
	edit.Draw(canvas, canvas.Bounds())
	if edit.lastRect.H != 4 {
		t.Fatalf("file bar body height = %d, want 4", edit.lastRect.H)
	}
	if !strings.Contains(canvas.Row(4), "File") {
		t.Fatalf("bottom row has no File menu: %q", canvas.Row(4))
	}
}

func TestRichTextEditFileBarShowsTitleStatusAndHints(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.ShowFileBar = true
	canvas := NewCanvas(100, 5)
	edit.Draw(canvas, canvas.Bounds())
	row := canvas.Row(4)
	if !strings.Contains(row, "Untitled") || !strings.Contains(row, "Unsaved") || !strings.Contains(row, "Ctrl+S") || !strings.Contains(row, "Ctrl+Shift+S") {
		t.Fatalf("untitled file bar row = %q", row)
	}
	edit.ConsumeKey(KeyEvent{Text: "!"})
	edit.Draw(canvas, canvas.Bounds())
	if row := canvas.Row(4); !strings.Contains(row, "Modified") {
		t.Fatalf("modified file bar row = %q", row)
	}
	path := filepath.Join(t.TempDir(), "notes.rtf")
	if err := edit.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	edit.Draw(canvas, canvas.Bounds())
	if row := canvas.Row(4); !strings.Contains(row, "notes.rtf") || !strings.Contains(row, "Saved") {
		t.Fatalf("saved file bar row = %q", row)
	}
	if err := edit.SaveAs(filepath.Join(t.TempDir(), "missing", "notes.rtf")); err == nil {
		t.Fatal("SaveAs to missing directory succeeded")
	}
	edit.Draw(canvas, canvas.Bounds())
	if row := canvas.Row(4); !strings.Contains(row, "Error") {
		t.Fatalf("error file bar row = %q", row)
	}
}

func TestRichTextEditFileBarF7HintNamesDestinationMode(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.ShowFileBar = true
	canvas := NewCanvas(100, 2)

	edit.Draw(canvas, canvas.Bounds())
	if row := canvas.Row(1); !strings.Contains(row, "[F7] View") {
		t.Fatalf("edit-mode F7 hint = %q, want destination View", row)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "f7"}); !result.Consumed || !edit.ViewMode {
		t.Fatalf("F7 view transition = %+v, view mode=%v", result, edit.ViewMode)
	}
	edit.Draw(canvas, canvas.Bounds())
	if row := canvas.Row(1); !strings.Contains(row, "[F7] Edit") {
		t.Fatalf("view-mode F7 hint = %q, want destination Edit", row)
	}
}

func TestRichTextEditFileBarClipsNarrowBounds(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "long file bar"}}}}})
	edit.ShowFileBar = true
	canvas := NewCanvas(9, 2)
	edit.Draw(canvas, canvas.Bounds())
	if canvas.Cols() != 9 || canvas.Get(8, 1).Text == "" {
		t.Fatalf("narrow file bar was not clipped to canvas bounds: cols=%d last cell=%+v", canvas.Cols(), canvas.Get(8, 1))
	}
	edit.Draw(NewCanvas(4, 1), Rect{W: 4, H: 1})
}

func TestRichTextEditFileBarKeyboardSaveAndPathlessSaveAs(t *testing.T) {
	root := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(originalDir) }()
	path := filepath.Join(root, "existing.ansi")
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "saved"}}}}})
	edit.ShowFileBar = true
	edit.FilePath = path
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-s"}); !result.Consumed {
		t.Fatalf("Ctrl+S result = %+v", result)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != edit.Document.ToANSI() {
		t.Fatalf("Ctrl+S bytes/error = %q/%v", data, err)
	}

	edit.FilePath = ""
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-s"}); !result.Consumed || edit.savePicker == nil {
		t.Fatalf("pathless Ctrl+S result/picker = %+v/%v", result, edit.savePicker)
	}
	if !edit.savePicker.nameFocus {
		t.Fatal("pathless Save picker did not focus its filename field")
	}
	edit.ConsumeKey(KeyEvent{Key: "esc"})
	if edit.FilePath != "" || edit.savePicker != nil {
		t.Fatalf("cancelled pathless Save changed path/picker = %q/%v", edit.FilePath, edit.savePicker)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-shift-s"}); !result.Consumed || edit.savePicker == nil {
		t.Fatalf("Ctrl+Shift+S result/picker = %+v/%v", result, edit.savePicker)
	}
	if !edit.savePicker.nameFocus {
		t.Fatal("untitled Save as picker did not focus its filename field")
	}
	newPath := filepath.Join(root, "created.ansi")
	for _, r := range "created.ansi" {
		edit.ConsumeKey(KeyEvent{Text: string(r)})
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "enter"}); !result.Consumed || edit.savePicker != nil {
		t.Fatalf("confirm pathless Save as result/picker = %+v/%v", result, edit.savePicker)
	}
	if edit.FilePath != newPath {
		t.Fatalf("Save as associated path = %q, want %q", edit.FilePath, newPath)
	}
}

func TestRichTextEditFileBarMenuPreservesSelectionAndPopoverRouting(t *testing.T) {
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected text"}}}}}
	edit := NewRichTextEdit(doc)
	edit.ShowFileBar = true
	edit.Cursor = RichPosition{Offset: 4}
	edit.SetSelection(RichPosition{Offset: 1}, RichPosition{Offset: 8})
	canvas := NewCanvas(80, 6)
	edit.Draw(canvas, canvas.Bounds())
	if len(edit.popoverButtons) == 0 {
		t.Fatal("selection popover was not shown")
	}
	selectionFrom, selectionTo, cursor := edit.SelectionFrom, edit.SelectionTo, edit.Cursor
	text := doc.ToPlainText()

	for _, key := range []KeyEvent{{Key: "alt-f"}, {Key: "down"}, {Key: "down"}} {
		if result := edit.ConsumeKey(key); !result.Consumed {
			t.Fatalf("File menu key %+v was not consumed: %+v", key, result)
		}
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "esc"}); !result.Consumed {
		t.Fatalf("Escape did not dismiss File menu: %+v", result)
	}
	if edit.Cursor != cursor || edit.SelectionFrom != selectionFrom || edit.SelectionTo != selectionTo || doc.ToPlainText() != text || !edit.HasSelection {
		t.Fatalf("menu navigation changed editor state: cursor=%+v selection=%+v..%+v text=%q", edit.Cursor, edit.SelectionFrom, edit.SelectionTo, doc.ToPlainText())
	}
	if !edit.popoverSuppressed && len(edit.popoverButtons) == 0 {
		t.Fatal("dismissing File menu lost the selection popover")
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "tab"}); !result.Consumed || !edit.popoverFocusSet {
		t.Fatalf("popover did not receive Tab after menu dismissal: result=%+v focus=%v", result, edit.popoverFocusSet)
	}
	if doc.ToPlainText() != text || !edit.HasSelection {
		t.Fatal("popover navigation formatted text or cleared selection")
	}
}

func TestRichTextEditFileBarMouseMenuActionKeepsSelection(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "mouse.ansi")
	doc := &RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}}
	edit := NewRichTextEdit(doc)
	edit.ShowFileBar = true
	edit.FilePath = path
	edit.SetSelection(RichPosition{}, RichPosition{Offset: 8})
	canvas := NewCanvas(80, 6)
	edit.Draw(canvas, canvas.Bounds())
	from, to := edit.SelectionFrom, edit.SelectionTo
	fileTitle := edit.fileBar.menu.titleRects[0]
	edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: fileTitle.X + 1, Y: fileTitle.Y})
	edit.Draw(canvas, canvas.Bounds())
	item := edit.fileBar.menu.itemRects[0]
	if result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: item.X, Y: item.Y}); !result.Consumed {
		t.Fatalf("clicking Save menu item was not consumed: %+v", result)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("mouse Save did not write document: %v", err)
	}
	if !edit.HasSelection || edit.SelectionFrom != from || edit.SelectionTo != to {
		t.Fatalf("mouse menu action changed selection: %v %+v..%+v", edit.HasSelection, edit.SelectionFrom, edit.SelectionTo)
	}
}

func TestRichTextEditFileBarHasNoFormattingActions(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "selected"}}}}})
	edit.ShowFileBar = true
	edit.ensureFileBar()
	items := edit.fileBar.menu.Menus[0].Items
	if len(items) != 2 || items[0].Label != "Save" || !strings.HasPrefix(items[1].Label, "Save as") {
		t.Fatalf("File menu items = %+v, want only Save and Save as", items)
	}
}
