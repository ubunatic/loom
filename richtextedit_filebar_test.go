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
	if canvas.Get(1, 4).Text != "F" || canvas.Get(2, 4).Text != "i" {
		t.Fatalf("bottom row has no File menu title: %q%q", canvas.Get(1, 4).Text, canvas.Get(2, 4).Text)
	}
}

func TestRichTextEditFileBarShowsTitleStatusAndHints(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.ShowFileBar = true
	canvas := NewCanvas(100, 5)
	edit.Draw(canvas, canvas.Bounds())
	row := hintBarRow(canvas, 4)
	if !strings.Contains(row, "Untitled") || !strings.Contains(row, "Unsaved") || strings.Contains(row, "Alt+F") || strings.Contains(row, "Ctrl+S") || strings.Contains(row, "Ctrl+Alt+S") {
		t.Fatalf("untitled file bar row = %q", row)
	}
	edit.ConsumeKey(KeyEvent{Text: "!"})
	edit.Draw(canvas, canvas.Bounds())
	if row := hintBarRow(canvas, 4); !strings.Contains(row, "Modified") {
		t.Fatalf("modified file bar row = %q", row)
	}
	path := filepath.Join(t.TempDir(), "notes.rtf")
	if err := edit.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	edit.Draw(canvas, canvas.Bounds())
	if row := hintBarRow(canvas, 4); !strings.Contains(row, "notes.rtf") || !strings.Contains(row, "Saved") {
		t.Fatalf("saved file bar row = %q", row)
	}
	if err := edit.SaveAs(filepath.Join(t.TempDir(), "missing", "notes.rtf")); err == nil {
		t.Fatal("SaveAs to missing directory succeeded")
	}
	edit.Draw(canvas, canvas.Bounds())
	if row := hintBarRow(canvas, 4); !strings.Contains(row, "Error") {
		t.Fatalf("error file bar row = %q", row)
	}
}

func TestRichTextEditFileBarHintsAtNarrowAndNormalWidths(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.ShowFileBar = true
	canvas := NewCanvas(100, 4)
	assertBar := func(status string) {
		t.Helper()
		edit.Draw(canvas, canvas.Bounds())
		row := hintBarRow(canvas, 3)
		if !strings.Contains(row, status) {
			t.Fatalf("100-column file bar row omits %q: %q", status, row)
		}
		if strings.Contains(row, "F7") {
			t.Fatalf("100-column file bar row shows an F7 hint: %q", row)
		}
	}
	assertBar("Untitled · Unsaved")
	edit.ConsumeKey(KeyEvent{Text: "!"})
	assertBar("Untitled · Modified")
	path := filepath.Join(t.TempDir(), "notes.rtf")
	if err := edit.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	assertBar("notes.rtf · Saved")
	if err := edit.SaveAs(filepath.Join(t.TempDir(), "missing", "notes.rtf")); err == nil {
		t.Fatal("SaveAs to missing directory succeeded")
	}
	assertBar("notes.rtf · Error")
	canvas = NewCanvas(80, 4)
	assertBar("notes.rtf · Error")
	canvas = NewCanvas(40, 4)
	edit.Draw(canvas, canvas.Bounds())
	row := hintBarRow(canvas, 3)
	if !strings.Contains(row, "· Error") || strings.Contains(row, "F7") {
		t.Fatalf("40-column file bar row = %q, want status and no F7 hint", row)
	}
	if strings.Contains(row, "F10") {
		t.Fatalf("narrow hints advertise F10 for the File menu: %q", row)
	}
}

// The file bar never shows an F7 hint: a right-aligned hint was clipped by
// host layouts (issue 269 M7). F7 lives in HotkeyHint and the F1 help.
func TestRichTextEditFileBarNeverShowsF7Hint(t *testing.T) {
	for _, mode := range []string{"edit", "view", "box"} {
		edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
		edit.ShowFileBar = true
		edit.ViewMode = mode == "view"
		edit.BoxMode = mode == "box"
		for width := 1; width <= 120; width++ {
			canvas := NewCanvas(width, 2)
			edit.Draw(canvas, canvas.Bounds())
			if row := hintBarRow(canvas, 1); strings.Contains(row, "F7") {
				t.Fatalf("%s mode, width %d: file bar shows F7: %q", mode, width, row)
			}
		}
	}
}

func TestRichTextEditFileBarBoxGuidanceIsAtomic(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.ShowFileBar = true
	edit.ConsumeKey(KeyEvent{Text: "!"})
	edit.BoxMode = true
	whole := false
	for width := 1; width <= 80; width++ {
		canvas := NewCanvas(width, 2)
		edit.Draw(canvas, canvas.Bounds())
		row := hintBarRow(canvas, 1)
		has := strings.Contains(row, "[Box] Esc exits")
		if !has && strings.Contains(row, "[Box") {
			t.Fatalf("width %d renders partial Box guidance: %q", width, row)
		}
		if whole && !has {
			t.Fatalf("width %d drops Box guidance shown at a narrower width: %q", width, row)
		}
		whole = whole || has
	}
	if !whole {
		t.Fatal("Box guidance never shown up to 80 columns")
	}
}

func TestRichTextEditHotkeyHintNamesF7DestinationMode(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "hello"}}}}})
	edit.ShowFileBar = true
	if hint := edit.HotkeyHint(80); !strings.Contains(hint, "F7 View") {
		t.Fatalf("edit-mode hotkey hint = %q, want F7 View", hint)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "f7"}); !result.Consumed || !edit.ViewMode {
		t.Fatalf("F7 view transition = %+v, view mode=%v", result, edit.ViewMode)
	}
	if hint := edit.HotkeyHint(80); !strings.Contains(hint, "F7 Edit") {
		t.Fatalf("view-mode hotkey hint = %q, want F7 Edit", hint)
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
	if data, err := os.ReadFile(path); err != nil || string(data) != edit.Document.ToANSI()+"\n" {
		t.Fatalf("Ctrl+S bytes/error = %q/%v", data, err)
	}

	edit.FilePath = ""
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-s"}); !result.Consumed || edit.savePicker == nil {
		t.Fatalf("pathless Ctrl+S result/picker = %+v/%v", result, edit.savePicker)
	}
	if edit.savePicker.nameFocus {
		t.Fatal("pathless Save picker stole focus from directory search")
	}
	edit.ConsumeKey(KeyEvent{Key: "esc"})
	if edit.FilePath != "" || edit.savePicker != nil {
		t.Fatalf("cancelled pathless Save changed path/picker = %q/%v", edit.FilePath, edit.savePicker)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-alt-s"}); !result.Consumed || edit.savePicker == nil {
		t.Fatalf("Ctrl+Alt+S result/picker = %+v/%v", result, edit.savePicker)
	}
	if edit.savePicker.nameFocus {
		t.Fatal("untitled Save as picker stole focus from directory search")
	}
	edit.ConsumeKey(KeyEvent{Key: "esc"})
	if edit.FilePath != "" || edit.savePicker != nil {
		t.Fatalf("cancelled Save as changed path/picker = %q/%v", edit.FilePath, edit.savePicker)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-s"}); !result.Consumed || edit.savePicker == nil {
		t.Fatalf("pathless Ctrl+S result/picker = %+v/%v", result, edit.savePicker)
	}
	edit.ConsumeKey(KeyEvent{Text: "created"})
	if got := edit.savePicker.List().Query(); got != "created" {
		t.Fatalf("Save directory query = %q, want created", got)
	}
	edit.ConsumeKey(KeyEvent{Key: "tab"})
	newPath := filepath.Join(root, "created.ansi")
	for _, r := range "created.ansi" {
		edit.ConsumeKey(KeyEvent{Text: string(r)})
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "enter"}); !result.Consumed || edit.savePicker != nil || edit.savePopup != nil {
		t.Fatalf("confirm pathless Save result/picker = %+v/%v", result, edit.savePicker)
	}
	if edit.FilePath != newPath {
		t.Fatalf("picker save associated path = %q, want %q", edit.FilePath, newPath)
	}
	initial, err := os.ReadFile(newPath)
	if err != nil || string(initial) != edit.Document.ToANSI()+"\n" {
		t.Fatalf("picker save bytes/error = %q/%v, want current document %q", initial, err, edit.Document.ToANSI())
	}
	if result := edit.ConsumeKey(KeyEvent{Text: "!"}); !result.Consumed {
		t.Fatalf("edit after picker save = %+v", result)
	}
	canvas := NewCanvas(100, 4)
	edit.Draw(canvas, canvas.Bounds())
	if row := hintBarRow(canvas, 3); !strings.Contains(row, "· Modified") {
		t.Fatalf("status after edit = %q, want Modified", row)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-s"}); !result.Consumed {
		t.Fatalf("second Ctrl+S result = %+v", result)
	}
	updated, err := os.ReadFile(newPath)
	if err != nil || string(updated) != edit.Document.ToANSI()+"\n" {
		t.Fatalf("second Ctrl+S bytes/error = %q/%v, want current document %q", updated, err, edit.Document.ToANSI())
	}
	edit.Draw(canvas, canvas.Bounds())
	if row := hintBarRow(canvas, 3); !strings.Contains(row, "· Saved") {
		t.Fatalf("status after successful second save = %q, want Saved", row)
	}
	if err := os.Remove(newPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(newPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-s"}); !result.Consumed {
		t.Fatalf("Ctrl+S to unwritable destination result = %+v", result)
	}
	edit.Draw(canvas, canvas.Bounds())
	if row := hintBarRow(canvas, 3); !strings.Contains(row, "· Error") || strings.Contains(row, "· Saved") {
		t.Fatalf("status after failed Ctrl+S = %q, want Error only", row)
	}
}

func TestRichTextEditSaveAsPopupDrawsBoundedBorderAndRoutesEvents(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "document"}}}}})
	if err := edit.openSavePicker(); err != nil {
		t.Fatal(err)
	}
	canvas := NewCanvas(100, 30)
	edit.Draw(canvas, canvas.Bounds())
	p := edit.savePopup
	if p == nil || p.innerRect.W <= 0 || p.innerRect.H <= 0 {
		t.Fatalf("popup inner rect = %+v", p)
	}
	if p.Width >= edit.lastRect.W || p.Height >= edit.lastRect.H {
		t.Fatalf("popup %dx%d is not smaller than editor %dx%d", p.Width, p.Height, edit.lastRect.W, edit.lastRect.H)
	}
	if !strings.Contains(hintBarRow(canvas, p.innerRect.Y-1), "Save as") || canvas.Get(p.innerRect.X-1, p.innerRect.Y).Text != "│" {
		t.Fatalf("popup border/title missing around picker: top=%q left=%q", hintBarRow(canvas, p.innerRect.Y-1), canvas.Get(p.innerRect.X-1, p.innerRect.Y).Text)
	}
	if result := edit.ConsumeKey(KeyEvent{Text: "notes"}); !result.Consumed || edit.savePicker.List().Query() != "notes" {
		t.Fatalf("popup did not route search input: result=%+v query=%q", result, edit.savePicker.List().Query())
	}
	filenameY := p.innerRect.Y + p.innerRect.H - 1
	edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: p.innerRect.X + 1, Y: filenameY})
	if !edit.savePicker.nameFocus {
		t.Fatal("popup filename click did not focus the filename field")
	}
	edit.ConsumeKey(KeyEvent{Text: ".ansi"})
	if got := edit.savePicker.FileName(); got != ".ansi" {
		t.Fatalf("filename after popup click = %q, want .ansi", got)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "esc"}); !result.Consumed || edit.savePicker != nil || edit.savePopup != nil {
		t.Fatalf("Escape did not cancel popup: result=%+v picker=%v popup=%v", result, edit.savePicker, edit.savePopup)
	}
}

func TestRichTextEditSaveAsPopupClosesOnOutsideClick(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{}}})
	if err := edit.openSavePicker(); err != nil {
		t.Fatal(err)
	}
	canvas := NewCanvas(80, 24)
	edit.Draw(canvas, canvas.Bounds())
	if result := edit.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 0, Y: 0}); !result.Consumed || edit.savePicker != nil || edit.savePopup != nil {
		t.Fatalf("outside click result/picker/popup = %+v/%v/%v, want consumed and dismissed", result, edit.savePicker, edit.savePopup)
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

func TestRichTextEditFileBarAltFOpensFileMenuAndF10StaysGlobal(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{}}})
	edit.ShowFileBar = true
	if result := edit.ConsumeKey(KeyEvent{Key: "f10"}); result.Consumed || edit.fileBar != nil && edit.fileBar.menu.Open {
		t.Fatalf("F10 was routed to File menu: result=%+v fileBar=%+v", result, edit.fileBar)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "alt-f"}); !result.Consumed || !edit.fileBar.menu.Open {
		t.Fatalf("Alt+F did not open File menu: result=%+v open=%v", result, edit.fileBar.menu.Open)
	}
	if result := edit.ConsumeKey(KeyEvent{Key: "esc"}); !result.Consumed || edit.fileBar.menu.Open {
		t.Fatalf("Escape did not close File menu: result=%+v open=%v", result, edit.fileBar.menu.Open)
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
	item := edit.fileBar.menu.itemRects[2] // Open and Close precede Save
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
	if len(items) != 4 || items[0].Label != "Open" || items[1].Label != "Close" || items[2].Label != "Save" || !strings.HasPrefix(items[3].Label, "Save as") {
		t.Fatalf("File menu items = %+v, want only Open, Close, Save and Save as", items)
	}
}

func TestRichTextEditFileMenuIsSpecDriven(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "text"}}}}})
	edit.ShowFileBar = true
	edit.ensureFileBar()
	defs := SpeccedDefaults.RichTextEdit
	menu := edit.fileBar.menu.Menus[0]
	if menu.Title != defs.FileMenuTitle || string(menu.Mnemonic) != defs.FileMenuMnemonic {
		t.Fatalf("menu title/mnemonic = %q/%q, want spec %q/%q", menu.Title, string(menu.Mnemonic), defs.FileMenuTitle, defs.FileMenuMnemonic)
	}
	want := []struct{ label, binding string }{
		{defs.HotkeyOpenLabel, defs.HotkeyOpenBinding}, {defs.HotkeyCloseLabel, defs.HotkeyCloseBinding},
		{defs.HotkeySaveLabel, defs.HotkeySaveBinding}, {defs.HotkeySaveAsLabel, defs.HotkeySaveAsBinding},
	}
	for i, w := range want {
		if !strings.HasPrefix(menu.Items[i].Label, w.label) || menu.Items[i].Shortcut != KeyCap(w.binding) {
			t.Fatalf("item %d = %q/%q, want spec %q/%q", i, menu.Items[i].Label, menu.Items[i].Shortcut, w.label, KeyCap(w.binding))
		}
	}
}

func TestRichTextEditOpenCloseActionsAndMenuItems(t *testing.T) {
	for _, fileBar := range []bool{true, false} {
		edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "text"}}}}})
		edit.ShowFileBar = fileBar
		edit.FilePath = "x.ansi"
		opens, closes := 0, 0
		edit.OnOpenRequest = func() { opens++ }
		edit.OnCloseRequest = func() { closes++ }
		if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-o"}); !result.Consumed || opens != 1 {
			t.Fatalf("fileBar=%v ^O consumed/opens = %v/%d", fileBar, result.Consumed, opens)
		}
		if result := edit.ConsumeKey(KeyEvent{Key: "ctrl-w"}); !result.Consumed || closes != 1 {
			t.Fatalf("fileBar=%v ^W consumed/closes = %v/%d", fileBar, result.Consumed, closes)
		}
		if edit.FilePath == "" || len(edit.Document.Lines[0].Spans) == 0 {
			t.Fatal("OnCloseRequest must leave closing to the host")
		}
	}

	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "text"}}}}})
	edit.ShowFileBar = true
	var labels []string
	for _, item := range edit.ensureFileBar().menu.Menus[0].Items {
		labels = append(labels, item.Label+" "+item.Shortcut)
	}
	got := strings.Join(labels, "|")
	for _, want := range []string{"Open ⌃O", "Close ⌃W"} {
		if !strings.Contains(got, want) {
			t.Fatalf("File menu %q missing %q", got, want)
		}
	}
}

func TestRichTextEditCloseWithoutHostResetsToUntitled(t *testing.T) {
	edit := NewRichTextEdit(&RichDocument{Lines: []RichLine{{Spans: []RichSpan{{Text: "text"}}}}})
	edit.ShowFileBar = true
	edit.FilePath = "x.ansi"
	edit.ConsumeKey(KeyEvent{Key: "ctrl-w"})
	if edit.FilePath != "" || edit.IsModified() || edit.DocState() != DocStateUntitled || len(edit.Document.Lines) != 1 || len(edit.Document.Lines[0].Spans) != 0 {
		t.Fatalf("closed state = %q/%v/%v", edit.FilePath, edit.IsModified(), edit.DocState())
	}
}
