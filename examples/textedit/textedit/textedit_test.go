// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package textedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/ubunatic/loom"
)

func TestNewAppAndRender(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatalf("NewApp returned nil")
	}

	frames := loom.Render(app, 80, 24)
	if len(frames) != 24 {
		t.Errorf("rendered %d lines, want 24", len(frames))
	}
	firstLine := frames[0]
	if !strings.Contains(firstLine, "Loom TextEdit") {
		t.Errorf("expected Loom TextEdit title in frame line 0: %q", firstLine)
	}
	joined := strings.Join(frames, "\n")
	for _, want := range []string{"Loom TextEdit", "demo.go", "Focus: ", "Explorer", "1 │", "package", "💻 Terminal", "❯"} {
		if !strings.Contains(joined, want) {
			t.Errorf("render missing %q", want)
		}
	}
}

func TestNewWidgetArgs(t *testing.T) {
	w, err := NewWidget(nil)
	if err != nil || w == nil {
		t.Fatalf("NewWidget(nil) failed: %v", err)
	}
	_, err = NewWidget([]string{"extra"})
	if err == nil {
		t.Errorf("NewWidget with extra args should return error")
	}
}

func TestFileSaveAndLoad(t *testing.T) {
	app := NewApp()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	app.editor.SetValue("line 1\nline 2")
	app.SaveFile(testFile)

	if app.modified {
		t.Errorf("expected modified=false after SaveFile")
	}
	if app.activePath != testFile {
		t.Errorf("activePath = %q, want %q", app.activePath, testFile)
	}
	if len(app.mruList) == 0 || app.mruList[0] != testFile {
		t.Errorf("expected %q at top of MRU list: %v", testFile, app.mruList)
	}

	// Verify content on disk
	b, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(b) != "line 1\nline 2" {
		t.Errorf("file content = %q, want %q", string(b), "line 1\nline 2")
	}

	// Test OpenFile
	app2 := NewApp()
	app2.OpenFile(testFile)
	if app2.editor.Value() != "line 1\nline 2" {
		t.Errorf("OpenFile content = %q, want %q", app2.editor.Value(), "line 1\nline 2")
	}
}

func TestKeybindingsClipboardAndSidebar(t *testing.T) {
	app := NewApp()
	app.editor.SetValue("first line\nsecond line")

	// C-c (Copy line)
	app.HandleKey(loom.KeyEvent{Key: "ctrl-c"})
	if app.clipboard != "second line" { // caret on last line by default
		t.Errorf("clipboard = %q, want %q", app.clipboard, "second line")
	}

	// C-v (Paste)
	app.HandleKey(loom.KeyEvent{Key: "ctrl-v"})
	if !strings.Contains(app.editor.Value(), "second line") {
		t.Errorf("editor should contain pasted content")
	}

	// C-x (Cut line)
	app.editor.SetValue("line A\nline B")
	app.HandleKey(loom.KeyEvent{Key: "ctrl-x"})
	if app.clipboard != "line B" {
		t.Errorf("cut clipboard = %q, want %q", app.clipboard, "line B")
	}

	// C-b (Toggle sidebar)
	initialShow := app.showBrowser
	app.HandleKey(loom.KeyEvent{Key: "ctrl-b"})
	if app.showBrowser == initialShow {
		t.Errorf("C-b did not toggle sidebar showBrowser")
	}

	// Tab (Cycle focus)
	f0 := app.activeFocus
	app.HandleKey(loom.KeyEvent{Key: "tab"})
	if app.activeFocus == f0 {
		t.Errorf("Tab did not change focus area")
	}
}

func TestInputDoesNotTreatConsumedEventsAsQuit(t *testing.T) {
	app := NewApp()
	original := app.editor.Value()
	app.activeFocus = focusBrowser

	if app.HandleKey(loom.KeyEvent{Text: "x"}) {
		t.Fatal("ordinary key was treated as a quit request")
	}
	if app.editor.Value() != original {
		t.Fatal("key for the file browser was dispatched a second time to the editor")
	}

	loom.Render(app, 80, 24)
	if app.HandleMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 5, Y: 5}) {
		t.Fatal("mouse click was treated as a quit request")
	}

	// When browser is focused, 'q' is unhandled and does not quit
	if app.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("unhandled 'q' in browser quit unexpectedly")
	}
}

func TestF10AndCtrlQAlwaysQuitWhenEditorFocused(t *testing.T) {
	app := NewApp()
	app.activeFocus = focusEditor
	valBefore := app.editor.Value()

	// Typing 'q' into editor should insert 'q' without quitting
	if app.HandleKey(loom.KeyEvent{Text: "q"}) {
		t.Fatal("typing 'q' in editor caused app to quit")
	}
	if app.editor.Value() == valBefore || !strings.Contains(app.editor.Value(), "q") {
		t.Fatal("typing 'q' in editor did not update buffer")
	}

	// F10 and Ctrl-Q must quit even with focus in editor
	if !app.HandleKey(loom.KeyEvent{Key: "f10"}) {
		t.Fatal("F10 did not quit when editor was focused")
	}
	if !app.HandleKey(loom.KeyEvent{Key: "ctrl-q"}) {
		t.Fatal("Ctrl-Q did not quit when editor was focused")
	}
}

func TestHotkeysStayVisibleAfterSave(t *testing.T) {
	app := NewApp()
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	app.SaveFile(tmpFile)

	frames := loom.Render(app, 120, 24)
	statusLine := frames[len(frames)-1]
	if !strings.Contains(statusLine, "Saved") {
		t.Fatalf("status line did not report Saved: %q", statusLine)
	}
	if !strings.Contains(statusLine, "C-s Save") || !strings.Contains(statusLine, "F10 Quit") {
		t.Fatalf("status line did not keep hotkeys visible after save: %q", statusLine)
	}
}

func TestEditorGutterAndExplorerSelection(t *testing.T) {
	app := NewApp()
	frames := loom.Render(app, 100, 24)
	all := strings.Join(frames, "\n")
	if !strings.Contains(all, " 1 │") || !strings.Contains(all, " 2 │") {
		t.Fatalf("editor line number gutter missing: %q", frames[1])
	}
	if !strings.Contains(all, "▸ ") {
		t.Fatalf("explorer directory/file indicator missing")
	}
}

func TestTerminalCommands(t *testing.T) {
	app := NewApp()
	tw := app.terminal

	tw.execCommand("help")
	tw.execCommand("echo hello world")
	tw.execCommand("date")
	tw.execCommand("clear")

	if len(tw.logs) != 0 {
		t.Errorf("expected clear command to empty logs, got %d lines", len(tw.logs))
	}
}

func TestFileBrowserNavigation(t *testing.T) {
	app := NewApp()
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "f1.txt"), []byte("file 1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "f2.txt"), []byte("file 2"), 0644)

	fb := newFileBrowser(tmpDir, app)
	if len(fb.files) != 2 {
		t.Fatalf("expected 2 files in temp dir, got %d", len(fb.files))
	}

	fb.HandleKey(loom.KeyEvent{Key: "down"})
	if fb.selected != 1 {
		t.Errorf("selected = %d, want 1", fb.selected)
	}

	fb.HandleKey(loom.KeyEvent{Key: "up"})
	if fb.selected != 0 {
		t.Errorf("selected = %d, want 0", fb.selected)
	}

	fb.HandleKey(loom.KeyEvent{Key: "enter"})
	if app.activePath != filepath.Join(tmpDir, "f1.txt") {
		t.Errorf("activePath = %q, want %q", app.activePath, filepath.Join(tmpDir, "f1.txt"))
	}

	// Backspace navigates to parent directory without quitting
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	fb2 := newFileBrowser(subDir, app)
	if fb2.dir != subDir {
		t.Fatalf("fb2.dir = %q, want %q", fb2.dir, subDir)
	}
	fb2.HandleKey(loom.KeyEvent{Key: "backspace"})
	if fb2.dir != tmpDir {
		t.Fatalf("fb2.dir after backspace = %q, want %q", fb2.dir, tmpDir)
	}
}
