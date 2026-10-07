// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestEditViewSidePanelF2ToggleAndTabFocus(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "doc1.txt")
	_ = os.WriteFile(path1, []byte("Document One\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path1)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path1, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	if view.showSidePanel {
		t.Fatal("side panel should initially be hidden")
	}

	// Press F2 to show side panel
	res := view.ConsumeKey(loom.KeyEvent{Key: "f2"})
	if !res.Consumed {
		t.Fatalf("F2 result = %+v, want consumed", res)
	}
	if !view.showSidePanel {
		t.Fatal("F2 did not show side panel")
	}
	if view.focused != focusBrowser {
		t.Fatalf("focus = %v, want focusBrowser", view.focused)
	}

	// Press Tab to switch focus to editor
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if view.focused != focusEditor {
		t.Fatalf("after Tab focus = %v, want focusEditor", view.focused)
	}

	// Press Tab again to switch focus back to browser
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if view.focused != focusBrowser {
		t.Fatalf("after second Tab focus = %v, want focusBrowser", view.focused)
	}

	// Press F2 to hide side panel
	view.ConsumeKey(loom.KeyEvent{Key: "f2"})
	if view.showSidePanel {
		t.Fatal("F2 did not hide side panel")
	}
	if view.focused != focusEditor {
		t.Fatalf("after hiding side panel focus = %v, want focusEditor", view.focused)
	}
}

func TestEditViewSidePanelOpenFile(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "doc1.txt")
	path2 := filepath.Join(dir, "doc2.txt")
	_ = os.WriteFile(path1, []byte("Document One\n"), 0600)
	_ = os.WriteFile(path2, []byte("Document Two\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path1)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path1, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	// Open side panel
	view.ConsumeKey(loom.KeyEvent{Key: "f2"})

	// Direct file open via view.openFile
	if err := view.openFile(path2); err != nil {
		t.Fatalf("openFile(%s): %v", path2, err)
	}

	if view.filePath != path2 {
		t.Fatalf("filePath = %q, want %q", view.filePath, path2)
	}

	canvas := loom.NewCanvas(80, 10)
	view.Draw(canvas, canvas.Bounds())
	if !strings.Contains(canvasScreenText(canvas), "doc2.txt") {
		t.Fatalf("rendered canvas missing doc2.txt: %q", canvasScreenText(canvas))
	}
}

func TestEditViewSidePanelUnsavedChangesProtection(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "doc1.txt")
	path2 := filepath.Join(dir, "doc2.txt")
	_ = os.WriteFile(path1, []byte("Document One\n"), 0600)
	_ = os.WriteFile(path2, []byte("Document Two\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path1)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path1, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	// Modify document
	view.edit.ConsumeKey(loom.KeyEvent{Text: " modified"})
	if !view.edit.IsModified() {
		t.Fatal("document expected to be modified")
	}

	// Set the pending action (pendingOpenPath became the generic pendingAction
	// when Close shared the guard) and trigger the prompt.
	view.pendingAction = func() { _ = view.openFile(path2) }
	quits := countQuits(view)
	res := view.showUnsavedDialog()
	if res.Quit {
		t.Fatal("showUnsavedDialog should open dialog, not quit immediately")
	}
	if view.unsavedDialog == nil || !view.unsavedDialog.Open {
		t.Fatal("unsavedDialog should be open")
	}

	// Select Discard in dialog
	view.ConsumeKey(loom.KeyEvent{Key: "right"}) // Discard button
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})

	if view.filePath != path2 {
		t.Fatalf("after Discard filePath = %q, want %q", view.filePath, path2)
	}
	if *quits != 0 {
		t.Fatal("opening new file after Discard should not quit editor")
	}
}

func TestEditViewSearchNormalAndRegex(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	content := "1 # Loom\n2 Terminal widgets for Go.\n3 Build TUI with Loom."
	_ = os.WriteFile(path, []byte(content), 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	// 1. Open search panel via F3
	res := view.ConsumeKey(loom.KeyEvent{Key: "f3"})
	if !res.Consumed || !view.showSearch || view.focused != focusSearch {
		t.Fatalf("F3 search open = %+v, showSearch=%v, focus=%v", res, view.showSearch, view.focused)
	}

	// 2. Type "Loom" into search
	for _, ch := range "Loom" {
		view.ConsumeKey(loom.KeyEvent{Text: string(ch)})
	}

	if view.searchQuery != "Loom" {
		t.Fatalf("searchQuery = %q, want Loom", view.searchQuery)
	}
	if len(view.searchMatches) != 2 {
		t.Fatalf("searchMatches count = %d, want 2 (line 0 and line 2)", len(view.searchMatches))
	}

	// 3. Test Enter for next match
	if view.searchIndex != 0 {
		t.Fatalf("initial searchIndex = %d, want 0", view.searchIndex)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if view.searchIndex != 1 {
		t.Fatalf("after Enter searchIndex = %d, want 1", view.searchIndex)
	}

	// 4. Test Shift-Enter for previous match
	view.ConsumeKey(loom.KeyEvent{Key: "shift-enter"})
	if view.searchIndex != 0 {
		t.Fatalf("after Shift-Enter searchIndex = %d, want 0", view.searchIndex)
	}

	// 5. Toggle Regex mode with Ctrl-R; Tab belongs to the container.
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-r"})
	if !view.regexMode {
		t.Fatal("Ctrl-R in search mode should enable regexMode")
	}

	// Type regex query: "Loom|widgets"
	view.searchBar.SetQuery("")
	view.searchQuery = ""
	for _, ch := range "Loom|widgets" {
		view.ConsumeKey(loom.KeyEvent{Text: string(ch)})
	}
	if len(view.searchMatches) != 3 {
		t.Fatalf("regex matches count = %d, want 3", len(view.searchMatches))
	}

	// 6. Test invalid regex syntax handling
	view.searchBar.SetQuery("")
	view.searchQuery = ""
	for _, ch := range "[unclosed" {
		view.ConsumeKey(loom.KeyEvent{Text: string(ch)})
	}
	if view.searchErr != "Invalid regex" {
		t.Fatalf("searchErr = %q, want Invalid regex", view.searchErr)
	}

	// 7. Test Esc closes search panel
	view.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if view.showSearch {
		t.Fatal("Esc should close search panel")
	}
	if view.focused != focusEditor {
		t.Fatalf("after search close focus = %v, want focusEditor", view.focused)
	}
}

func TestEditViewUnicodeSearchNoCrash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unicode.txt")
	content := "1. Übersicht der Grüße\n2. こんにちは世界\n3. übersicht test"
	_ = os.WriteFile(path, []byte(content), 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	view.ConsumeKey(loom.KeyEvent{Key: "f3"})

	// Search "übersicht" (case-insensitive Unicode match)
	for _, ch := range "übersicht" {
		view.ConsumeKey(loom.KeyEvent{Text: string(ch)})
	}

	if len(view.searchMatches) != 2 {
		t.Fatalf("searchMatches count = %d, want 2", len(view.searchMatches))
	}

	// Search CJK characters "こんにちは"
	view.searchBar.SetQuery("")
	view.searchQuery = ""
	for _, ch := range "こんにちは" {
		view.ConsumeKey(loom.KeyEvent{Text: string(ch)})
	}

	if len(view.searchMatches) != 1 {
		t.Fatalf("searchMatches count = %d, want 1", len(view.searchMatches))
	}
}

func TestEditViewSearchCtrlQPromptsUnsavedChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	_ = os.WriteFile(path, []byte("Content\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	// Modify document
	view.edit.ConsumeKey(loom.KeyEvent{Text: "X"})

	// Open search panel
	view.ConsumeKey(loom.KeyEvent{Key: "f3"})

	// Ctrl-Q inside search mode is left to the Pane (old expectation: the view
	// consumed it itself, which bypassed the Pane's close-request contract).
	quits := countQuits(view)
	res := view.ConsumeKey(loom.KeyEvent{Key: "ctrl-q"})
	if res.Consumed || res.Quit || *quits != 0 {
		t.Fatalf("ctrl-q in search mode = %+v, want ignored so the Pane handles it", res)
	}
	if view.closeRequest(loom.CloseReasonQuitKey) != loom.CloseVeto {
		t.Fatal("close request on modified doc must be vetoed")
	}
	if view.unsavedDialog == nil || !view.unsavedDialog.Open {
		t.Fatal("close request should open unsavedDialog")
	}
}

func TestEditViewSearchMouseClickFocus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	_ = os.WriteFile(path, []byte("Content Line 1\nContent Line 2\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	// Open search
	view.ConsumeKey(loom.KeyEvent{Key: "f3"})

	// Draw view so rects (lastRect, editorRect, searchRect) are computed
	canvas := loom.NewCanvas(100, 24)
	view.Draw(canvas, canvas.Bounds())

	if !view.searchRect.Contains(view.searchRect.X+5, view.searchRect.Y+1) {
		t.Fatal("searchRect expected to contain click target")
	}

	// Click inside search panel
	res := view.ConsumeMouse(loom.MouseEvent{
		Action: loom.MousePress,
		Button: loom.MouseLeft,
		X:      view.searchRect.X + 5,
		Y:      view.searchRect.Y + 1,
	})

	if !res.Consumed {
		t.Fatalf("click in search panel result = %+v, want consumed", res)
	}
	if view.focused != focusSearch {
		t.Fatalf("focus = %v, want focusSearch", view.focused)
	}

	// Verify typing goes to search panel, not editor
	view.ConsumeKey(loom.KeyEvent{Text: "X"})
	if view.searchQuery != "X" {
		t.Fatalf("searchQuery = %q, want X", view.searchQuery)
	}
}

func TestEditViewCtrlPScreenshot(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	_ = os.WriteFile(path, []byte("Content\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	// Press Ctrl-P
	res := view.ConsumeKey(loom.KeyEvent{Key: "ctrl-p"})
	if !res.Consumed {
		t.Fatalf("ctrl-p result = %+v, want consumed", res)
	}

	shotDir := filepath.Join(tempHome, "Pictures", "Screenshots")
	entries, err := os.ReadDir(shotDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("ReadDir(%s) err=%v, entries=%d", shotDir, err, len(entries))
	}
}

func TestEditViewNarrowSearchLayout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	_ = os.WriteFile(path, []byte("Content\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	view.ConsumeKey(loom.KeyEvent{Key: "f3"})

	// Draw on narrow canvas (width 20)
	canvas := loom.NewCanvas(20, 10)
	view.Draw(canvas, canvas.Bounds())

	if view.searchRect.X < 0 {
		t.Fatalf("searchRect.X = %d, want >= 0", view.searchRect.X)
	}
	if view.searchRect.W > 20 {
		t.Fatalf("searchRect.W = %d, want <= 20", view.searchRect.W)
	}
}

func TestEditViewLayoutHeaderAndStatusBars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "README.md")
	_ = os.WriteFile(path, []byte("# Loom\n"), 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}

	cfg := loom.EditorConfig{
		Theme:     "mc-dark",
		MouseGrab: true,
		AltScreen: true,
	}
	view, err := newEditView(edit, path, cfg)
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}

	canvas := loom.NewCanvas(100, 24)
	view.Draw(canvas, canvas.Bounds())
	screenText := canvasScreenText(canvas)

	for _, want := range []string{
		"┌─ loom edit",
		"File README.md",
		loom.SpeccedDefaults.Editor.StatusIcons.Theme,
		"mc-dark",
		"^O Files",
		"^F Search",
		"F10 Quit",
		"└",
	} {
		if !strings.Contains(screenText, want) {
			t.Errorf("rendered canvas missing %q: %q", want, screenText)
		}
	}
	if view.editorRect.Y != 1 || view.editorRect.H != 20 {
		t.Fatalf("editor rectangle = %+v, want Y=1 H=20", view.editorRect)
	}
	status := strings.Split(screenText, "\n")[21]
	for _, icon := range []string{loom.SpeccedDefaults.Editor.StatusIcons.Theme, loom.SpeccedDefaults.Editor.StatusIcons.MouseOn, loom.SpeccedDefaults.Editor.StatusIcons.AltOn} {
		if !strings.Contains(status, " "+icon+" ") {
			t.Errorf("bottom status %q missing %q", status, icon)
		}
	}
}

func TestGenerateAnsiDesignScreenshots(t *testing.T) {
	designDir := t.TempDir()

	readmePath := filepath.Join(t.TempDir(), "README.md")
	readmeContent := "# Loom\n\nTerminal widgets for Go.\n\n## Getting started\n\ngo get ubunatic.com/loom\n\nBuild terminal interfaces with Loom.\nCompose widgets, panes and layouts.\n\n## Editing\n\nloom edit README.md\n\nSave changes with ⌃S.\n"
	_ = os.WriteFile(readmePath, []byte(readmeContent), 0600)

	editorYamlPath := filepath.Join(t.TempDir(), "editor.yaml")
	editorYamlContent := "# ~/.config/loom/editor.yaml\n# yaml-language-server: $schema=./editor.schema.json\n\ntheme: mc-dark\nmousegrab: true\naltscreen: true\n\n# Boolean values: true / false\n# CLI flags override these preferences.\n\n# Example: loom edit --mousegrab README.md\n"
	_ = os.WriteFile(editorYamlPath, []byte(editorYamlContent), 0600)

	cfg := loom.EditorConfig{
		Theme:     "mc-dark",
		MouseGrab: true,
		AltScreen: true,
	}

	saveShot := func(fileName string, setup func(v *editView)) {
		edit, err := loom.NewRichTextEditFromFile(readmePath)
		if fileName == "loom-edit-05-settings.ansi" {
			edit, err = loom.NewRichTextEditFromFile(editorYamlPath)
		}
		if err != nil {
			t.Fatalf("NewRichTextEditFromFile: %v", err)
		}
		view, err := newEditView(edit, readmePath, cfg)
		if fileName == "loom-edit-05-settings.ansi" {
			view, err = newEditView(edit, "editor.yaml", cfg)
		}
		if err != nil {
			t.Fatalf("newEditView: %v", err)
		}

		setup(view)

		outPath := filepath.Join(designDir, fileName)
		rows := loom.Render(view, 100, 24)
		content := strings.Join(rows, "\n") + "\n"
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", outPath, err)
		}
	}

	// 1. Editor
	saveShot("loom-edit-01-editor.ansi", func(v *editView) {
		v.showSidePanel = false
		v.showSearch = false
	})

	// 2. File browser
	saveShot("loom-edit-02-file-browser.ansi", func(v *editView) {
		v.showSidePanel = true
		v.showSearch = false
	})

	// 3. Search normal
	saveShot("loom-edit-03-search-normal.ansi", func(v *editView) {
		v.showSidePanel = true
		v.showSearch = true
		v.regexMode = false
		v.searchBar.SetQuery("Loom")
		v.searchQuery = "Loom"
		v.updateSearchMatches()
	})

	// 4. Search regex
	saveShot("loom-edit-04-search-regex.ansi", func(v *editView) {
		v.showSidePanel = true
		v.showSearch = true
		v.regexMode = true
		v.searchBar.SetQuery("Loom|widgets")
		v.searchQuery = "Loom|widgets"
		v.updateSearchMatches()
	})

	// 5. Settings
	saveShot("loom-edit-05-settings.ansi", func(v *editView) {
		v.showSidePanel = false
		v.showSearch = false
	})
}

func integrationEditView(t *testing.T) *editView {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte("alpha beta alpha\n"), 0600); err != nil {
		t.Fatal(err)
	}
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatal(err)
	}
	return view
}

func TestScreenshotPreservesExistingCapturesAndIncludesDialog(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	view := integrationEditView(t)
	dir := filepath.Join(os.Getenv("HOME"), "Pictures", "Screenshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(dir, "02-loom-edit-doc.txt.ansi")
	if err := os.WriteFile(oldPath, []byte("keep me"), 0600); err != nil {
		t.Fatal(err)
	}
	view.edit.ConsumeKey(loom.KeyEvent{Text: "X"})
	quits := countQuits(view)
	view.showUnsavedDialog()
	loom.Render(view, 100, 24)
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-p"})
	old, err := os.ReadFile(oldPath)
	if err != nil || string(old) != "keep me" {
		t.Fatalf("existing capture changed: %q, %v", old, err)
	}
	shot, err := os.ReadFile(filepath.Join(dir, "03-loom-edit-doc.txt.ansi"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(shot), "Save changes?") {
		t.Fatal("screenshot omitted open dialog")
	}
	if !view.unsavedDialog.Open || *quits != 0 {
		t.Fatal("capture changed dialog lifecycle")
	}
	if !strings.Contains(view.statusMessage, "Saved screenshot:") {
		t.Fatal("missing capture feedback")
	}
}

func TestSearchHighlightDoesNotAlterSelectionOrEditingCursor(t *testing.T) {
	view := integrationEditView(t)
	from, to := loom.RichPosition{Offset: 6}, loom.RichPosition{Offset: 10}
	view.edit.SetSelection(from, to)
	view.ConsumeKey(loom.KeyEvent{Key: "f3"})
	view.ConsumeKey(loom.KeyEvent{Text: "alpha"})
	if view.edit.SelectionFrom != from || view.edit.SelectionTo != to || !view.edit.HasSelection {
		t.Fatal("search changed editing selection")
	}
	if len(view.edit.Highlights) != 1 {
		t.Fatal("search did not create display highlight")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if view.focused != focusEditor || view.regexMode {
		t.Fatal("Tab did not move focus independently of regex mode")
	}
	view.edit.Cursor = loom.RichPosition{Offset: 4}
	view.edit.ClearSelection()
	view.ConsumeKey(loom.KeyEvent{Text: "!"})
	if view.edit.Cursor.Offset != 5 || view.edit.HasSelection {
		t.Fatal("match refresh moved the editing cursor or selected text")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "f3"})
	if len(view.edit.Highlights) != 0 {
		t.Fatal("closing search retained highlights")
	}
}

func TestSearchBoundsOnSmallTerminals(t *testing.T) {
	view := integrationEditView(t)
	view.showSearch, view.showSidePanel = true, true
	for _, size := range [][2]int{{100, 24}, {20, 10}, {8, 8}, {4, 7}, {1, 1}} {
		loom.Render(view, size[0], size[1])
		r, editor := view.searchRect, view.editorRect
		if r.W > 0 && (r.X < editor.X || r.Y < editor.Y || r.X+r.W > editor.X+editor.W || r.Y+r.H > editor.Y+editor.H) {
			t.Fatalf("%v: search %v exceeds editor %v", size, r, editor)
		}
	}
}

func TestEditorReportsScreenshotAndFileOpenFailures(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	view := integrationEditView(t)
	if err := os.WriteFile(filepath.Join(os.Getenv("HOME"), "Pictures"), []byte("blocked"), 0600); err != nil {
		t.Fatal(err)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-p"})
	if view.statusMessage == "" {
		t.Fatal("capture failure was silent")
	}
	original := view.filePath
	view.statusMessage = ""
	// A listed file changes into a directory before activation: opening fails.
	if err := os.Remove(original); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(original, 0755); err != nil {
		t.Fatal(err)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "f2"})
	view.filePicker.List().SearchBar().SetQuery("doc.txt")
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if view.filePath != original || view.statusMessage == "" {
		t.Fatal("failed file open lost the document or hid the error")
	}
}

// countQuits replaces the pane's Quit with a counter.
func countQuits(v *editView) *int {
	n := new(int)
	v.quit = func() { *n++ }
	return n
}

func newDirtyEditView(t *testing.T) (*editView, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte("Initial\n"), 0600); err != nil {
		t.Fatal(err)
	}
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatal(err)
	}
	return view, path
}

func TestEditViewCloseRequestCleanAllows(t *testing.T) {
	view, _ := newDirtyEditView(t)
	if got := view.closeRequest(loom.CloseReasonInterrupt); got != loom.CloseAllow {
		t.Fatalf("clean close request = %v, want allow", got)
	}
	if view.unsavedDialog != nil {
		t.Fatal("clean close request opened a dialog")
	}
}

func TestEditViewEscNeverQuits(t *testing.T) {
	view, _ := newDirtyEditView(t)
	quits := countQuits(view)
	view.edit.ConsumeKey(loom.KeyEvent{Text: "X"})
	for _, ke := range []loom.KeyEvent{{Key: "esc"}, {Text: "q"}} {
		if res := view.ConsumeKey(ke); res.Quit {
			t.Fatalf("%+v returned quit", ke)
		}
	}
	if *quits != 0 || view.unsavedDialog != nil {
		t.Fatal("Esc/q must not quit or open the dialog")
	}
	if !view.edit.IsModified() {
		t.Fatal("text lost")
	}
}

func TestEditViewCloseDialogButtons(t *testing.T) {
	// Dirty buffer: veto, then Cancel keeps running, Discard quits, Save writes then quits.
	view, path := newDirtyEditView(t)
	quits := countQuits(view)
	view.edit.ConsumeKey(loom.KeyEvent{Text: "X"})

	if view.closeRequest(loom.CloseReasonQuitKey) != loom.CloseVeto {
		t.Fatal("dirty close request not vetoed")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "right"})
	view.ConsumeKey(loom.KeyEvent{Key: "right"}) // Cancel
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if *quits != 0 || view.unsavedDialog != nil {
		t.Fatalf("Cancel: quits=%d dialog=%v", *quits, view.unsavedDialog)
	}

	view.closeRequest(loom.CloseReasonQuitKey)
	view.ConsumeKey(loom.KeyEvent{Key: "right"}) // Discard
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if *quits != 1 {
		t.Fatalf("Discard quits = %d, want 1", *quits)
	}
	if data, _ := os.ReadFile(path); strings.Contains(string(data), "X") {
		t.Fatal("Discard wrote the file")
	}

	view.closeRequest(loom.CloseReasonInterrupt)
	view.ConsumeKey(loom.KeyEvent{Key: "enter"}) // Save
	if *quits != 2 {
		t.Fatalf("Save quits = %d, want 2", *quits)
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "X") {
		t.Fatal("Save did not write the file")
	}
}

func TestEditViewSaveAfterSpanDeletionCompletesClose(t *testing.T) {
	for _, mouse := range []bool{false, true} {
		t.Run(fmt.Sprintf("mouse=%v", mouse), func(t *testing.T) {
			view, path := newDirtyEditView(t)
			quits := countQuits(view)
			view.ConsumeKey(loom.KeyEvent{Key: "end"})
			view.ConsumeKey(loom.KeyEvent{Key: "enter"})
			view.ConsumeKey(loom.KeyEvent{Text: "x"})
			view.ConsumeKey(loom.KeyEvent{Key: "backspace"})
			if view.closeRequest(loom.CloseReasonQuitKey) != loom.CloseVeto {
				t.Fatal("new empty logical line should need saving")
			}
			canvas := loom.NewCanvas(100, 24)
			view.Draw(canvas, canvas.Bounds())
			if mouse {
				clicked := false
				for y := 0; y < canvas.Rows() && !clicked; y++ {
					for x := 0; x < canvas.Cols(); x++ {
						if canvas.Get(x, y).Text == "▶" {
							view.ConsumeMouse(loom.MouseEvent{X: x + 2, Y: y, Button: loom.MouseLeft, Action: loom.MousePress})
							clicked = true
							break
						}
					}
				}
				if !clicked {
					t.Fatal("Save button not found")
				}
			} else {
				view.ConsumeKey(loom.KeyEvent{Key: "enter"})
			}
			if *quits != 1 || view.edit.IsModified() || view.edit.DocState() != loom.DocStateSaved {
				t.Fatalf("Save did not finish close: quits=%d state=%v", *quits, view.edit.DocState())
			}
			data, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(data), "Initial\n") {
				t.Fatalf("saved content = %q, error=%v", data, err)
			}
		})
	}
}

func TestEditViewAttachWiresPane(t *testing.T) {
	view, _ := newDirtyEditView(t)
	pane := &loom.Pane{}
	view.attach(pane)
	if pane.OnCloseRequest == nil || view.quit == nil {
		t.Fatal("attach did not wire OnCloseRequest and quit")
	}
	view.edit.ConsumeKey(loom.KeyEvent{Text: "X"})
	if pane.OnCloseRequest(loom.CloseReasonQuitKey) != loom.CloseVeto {
		t.Fatal("pane close request on dirty view not vetoed")
	}
}

func editHeaderRow(view *editView) string {
	canvas := loom.NewCanvas(100, 24)
	view.Draw(canvas, loom.Rect{W: 100, H: 24})
	return strings.Split(canvasScreenText(canvas), "\n")[0]
}

// editScreenCount counts the screen rows of a 100x24 render containing text.
func editScreenCount(view *editView, text string) int {
	canvas := loom.NewCanvas(100, 24)
	view.Draw(canvas, loom.Rect{W: 100, H: 24})
	n := 0
	for _, row := range strings.Split(canvasScreenText(canvas), "\n") {
		if strings.Contains(row, text) {
			n++
		}
	}
	return n
}

func TestEditViewFileBarOwnsDocNameAndState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.txt")
	_ = os.WriteFile(path, []byte("base\n"), 0600)
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatal(err)
	}
	check := func(state string) {
		t.Helper()
		if row := editHeaderRow(view); strings.Contains(row, "theme:") || strings.Contains(row, "File") || !strings.Contains(row, "─") {
			t.Fatalf("top row = %q, want top box border without header info", row)
		}
		if n := editScreenCount(view, "doc.txt · "+state); n != 1 {
			t.Fatalf("file bar status %q shown on %d rows, want 1", "doc.txt · "+state, n)
		}
		if n := editScreenCount(view, " File "); n != 1 {
			t.Fatalf("File menu title shown on %d rows, want 1", n)
		}
	}
	check("Saved")
	view.ConsumeKey(loom.KeyEvent{Text: "z"})
	check("Modified")
	// Save must work with the file browser focused too.
	view.ConsumeKey(loom.KeyEvent{Key: "f2"})
	if view.focused != focusBrowser {
		t.Fatal("browser not focused")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-s"})
	check("Saved")
	if !strings.Contains(view.statusMessage, "Saved") {
		t.Fatalf("status message = %q, want save confirmation", view.statusMessage)
	}
}

func TestEditViewBrowserSingleClickSelectsRow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	_ = os.WriteFile(path, []byte("x\n"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0600)
	_ = os.WriteFile(filepath.Join(dir, "b.txt"), nil, 0600)

	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}
	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "f2"})
	canvas := loom.NewCanvas(100, 24)
	view.Draw(canvas, canvas.Bounds())

	// Entries: "..", a.txt, b.txt, doc.txt; list starts one row below the header.
	for i, want := range []string{"..", "a.txt", "b.txt", "doc.txt"} {
		view.filePicker.List().SelectIndex(3 - i)
		view.ConsumeMouse(loom.MouseEvent{
			Action: loom.MousePress,
			Button: loom.MouseLeft,
			X:      view.browserRect.X + view.browserRect.W - 2,
			Y:      view.browserRect.Y + 1 + i,
		})
		if got, ok := view.filePicker.Selected(); !ok || got.Name != want {
			t.Fatalf("click row %d selected %q, want %q", i, got.Name, want)
		}
	}
}

func newCloseTestView(t *testing.T) *editView {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte("base\n"), 0600); err != nil {
		t.Fatal(err)
	}
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatal(err)
	}
	return view
}

func TestEditViewCtrlOTogglesBrowserFromEveryFocus(t *testing.T) {
	for _, focus := range []editFocus{focusEditor, focusBrowser, focusSearch} {
		view := newCloseTestView(t)
		view.showSearch = focus == focusSearch
		view.focused = focus
		if focus == focusBrowser {
			view.showSidePanel = true
		}
		if res := view.ConsumeKey(loom.KeyEvent{Key: "ctrl-o"}); !res.Consumed {
			t.Fatalf("focus %v: ^O not consumed", focus)
		}
		wantOpen, wantFocus := true, focusBrowser
		if focus == focusBrowser {
			wantOpen, wantFocus = false, focusEditor
		}
		if view.showSidePanel != wantOpen || view.focused != wantFocus {
			t.Fatalf("focus %v: panel/focus = %v/%v, want %v/%v", focus, view.showSidePanel, view.focused, wantOpen, wantFocus)
		}
	}
}

func TestEditViewCtrlWCleanBufferResetsToUntitled(t *testing.T) {
	view := newCloseTestView(t)
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-w"})
	if view.unsavedDialog != nil {
		t.Fatal("clean close must not ask")
	}
	if view.edit.FilePath != "" || view.filePath != "" || view.edit.DocState() != loom.DocStateUntitled {
		t.Fatalf("after close path/state = %q/%v", view.edit.FilePath, view.edit.DocState())
	}
}

func TestEditViewCtrlWDirtyGuardsDiscardSaveCancel(t *testing.T) {
	// Cancel keeps the buffer.
	view := newCloseTestView(t)
	view.edit.ConsumeKey(loom.KeyEvent{Text: "x"})
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-w"})
	if view.unsavedDialog == nil || !view.unsavedDialog.Open {
		t.Fatal("dirty close must show the dialog")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if view.edit.FilePath == "" || !view.edit.IsModified() {
		t.Fatal("Cancel must keep the buffer")
	}

	// Discard resets to Untitled without writing.
	path := view.edit.FilePath
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-w"})
	view.ConsumeKey(loom.KeyEvent{Key: "right"})
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if view.edit.FilePath != "" || view.edit.DocState() != loom.DocStateUntitled {
		t.Fatalf("Discard path/state = %q/%v", view.edit.FilePath, view.edit.DocState())
	}
	if data, _ := os.ReadFile(path); string(data) != "base\n" {
		t.Fatalf("Discard wrote the file: %q", data)
	}

	// Save writes, then resets.
	view = newCloseTestView(t)
	path = view.edit.FilePath
	view.edit.ConsumeKey(loom.KeyEvent{Text: "x"})
	view.ConsumeKey(loom.KeyEvent{Key: "ctrl-w"})
	view.ConsumeKey(loom.KeyEvent{Key: "enter"}) // Save is the default
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "x") {
		t.Fatalf("Save did not write: %q", data)
	}
	if view.edit.FilePath != "" {
		t.Fatal("Save must then close the document")
	}
}
