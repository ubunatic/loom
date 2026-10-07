// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
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

	// Set pending open path and trigger quit prompt
	view.pendingOpenPath = path2
	res := view.handleQuit()
	if res.Quit {
		t.Fatal("handleQuit on modified document should open dialog, not quit immediately")
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
	if view.shouldQuit {
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

	// 5. Toggle Regex mode with Tab
	view.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if !view.regexMode {
		t.Fatal("Tab in search mode should enable regexMode")
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

	// Press Ctrl-Q inside search mode
	res := view.ConsumeKey(loom.KeyEvent{Key: "ctrl-q"})
	if res.Quit || view.shouldQuit {
		t.Fatalf("ctrl-q in search mode on modified doc quit immediately, want unsaved dialog")
	}
	if view.unsavedDialog == nil || !view.unsavedDialog.Open {
		t.Fatal("ctrl-q in search mode should open unsavedDialog")
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
		"theme: mc-dark",
		"mouse: on",
		"altscreen: on",
		"F2 Files",
		"F3/^F Search",
		"F10 Quit",
		"└",
	} {
		if !strings.Contains(screenText, want) {
			t.Errorf("rendered canvas missing %q: %q", want, screenText)
		}
	}
}

func TestGenerateAnsiDesignScreenshots(t *testing.T) {
	designDir := t.TempDir()
	if os.Getenv("UPDATE_GOLDEN") == "1" || os.Getenv("UPDATE_SNAPSHOTS") == "1" || os.Getenv("GENERATE_DESIGN_SCREENSHOTS") == "1" {
		designDir = filepath.Join("..", "..", "docs", "design")
	}

	readmePath := filepath.Join(t.TempDir(), "README.md")
	readmeContent := "# Loom\n\nTerminal widgets for Go.\n\n## Getting started\n\ngo get ubunatic.com/loom\n\nBuild terminal interfaces with Loom.\nCompose widgets, panes and layouts.\n\n## Editing\n\nloom edit README.md\n\nSave changes with ^S.\n"
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
