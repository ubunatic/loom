// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package textedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	app.ConsumeKey(loom.KeyEvent{Key: "ctrl-c"})
	if app.clipboard != "second line" { // caret on last line by default
		t.Errorf("clipboard = %q, want %q", app.clipboard, "second line")
	}

	// C-v (Paste)
	app.ConsumeKey(loom.KeyEvent{Key: "ctrl-v"})
	if !strings.Contains(app.editor.Value(), "second line") {
		t.Errorf("editor should contain pasted content")
	}

	// C-x (Cut line)
	app.editor.SetValue("line A\nline B")
	app.ConsumeKey(loom.KeyEvent{Key: "ctrl-x"})
	if app.clipboard != "line B" {
		t.Errorf("cut clipboard = %q, want %q", app.clipboard, "line B")
	}

	// C-b (Toggle sidebar)
	initialShow := app.showBrowser
	app.ConsumeKey(loom.KeyEvent{Key: "ctrl-b"})
	if app.showBrowser == initialShow {
		t.Errorf("C-b did not toggle sidebar showBrowser")
	}

	// Tab (Cycle focus)
	f0 := app.activeFocus
	app.ConsumeKey(loom.KeyEvent{Key: "tab"})
	if app.activeFocus == f0 {
		t.Errorf("Tab did not change focus area")
	}
}

func TestInputDoesNotTreatConsumedEventsAsQuit(t *testing.T) {
	app := NewApp()
	original := app.editor.Value()
	app.activeFocus = focusBrowser

	if app.ConsumeKey(loom.KeyEvent{Text: "x"}).Quit {
		t.Fatal("ordinary key was treated as a quit request")
	}
	if app.editor.Value() != original {
		t.Fatal("key for the file browser was dispatched a second time to the editor")
	}

	loom.Render(app, 80, 24)
	if app.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 5, Y: 5}).Quit {
		t.Fatal("mouse click was treated as a quit request")
	}

	// When browser is focused, 'q' is unhandled and does not quit
	if app.ConsumeKey(loom.KeyEvent{Text: "q"}).Quit {
		t.Fatal("unhandled 'q' in browser quit unexpectedly")
	}
}

func TestF10AndCtrlQAlwaysQuitWhenEditorFocused(t *testing.T) {
	app := NewApp()
	app.activeFocus = focusEditor
	valBefore := app.editor.Value()

	// Typing 'q' into editor should insert 'q' without quitting
	if app.ConsumeKey(loom.KeyEvent{Text: "q"}).Quit {
		t.Fatal("typing 'q' in editor caused app to quit")
	}
	if app.editor.Value() == valBefore || !strings.Contains(app.editor.Value(), "q") {
		t.Fatal("typing 'q' in editor did not update buffer")
	}

	// F10 and Ctrl-Q must quit even with focus in editor
	if !app.ConsumeKey(loom.KeyEvent{Key: "f10"}).Quit {
		t.Fatal("F10 did not quit when editor was focused")
	}
	if !app.ConsumeKey(loom.KeyEvent{Key: "ctrl-q"}).Quit {
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

	fb.ConsumeKey(loom.KeyEvent{Key: "down"})
	if fb.selected != 1 {
		t.Errorf("selected = %d, want 1", fb.selected)
	}

	fb.ConsumeKey(loom.KeyEvent{Key: "up"})
	if fb.selected != 0 {
		t.Errorf("selected = %d, want 0", fb.selected)
	}

	fb.ConsumeKey(loom.KeyEvent{Key: "enter"})
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
	fb2.ConsumeKey(loom.KeyEvent{Key: "backspace"})
	if fb2.dir != tmpDir {
		t.Fatalf("fb2.dir after backspace = %q, want %q", fb2.dir, tmpDir)
	}
}

func TestLanguageEngineSelection(t *testing.T) {
	app := NewApp()
	if app.editor.Highlighter() == nil {
		t.Fatalf("expected NewApp editor to have highlighter for demo.go")
	}
	if got := app.editor.Highlighter().Language(); got != "go" {
		t.Errorf("demo.go language = %q, want \"go\"", got)
	}

	tmpDir := t.TempDir()
	tests := []struct {
		name     string
		file     string
		content  string
		wantLang string
		hasHL    bool
	}{
		{"json file", "config.json", `{"key": "value"}`, "json", true},
		{"markdown file", "doc.md", "# Title\n`code`", "markdown", true},
		{"yaml file", "deploy.yaml", "app: test\nversion: 1", "yaml", true},
		{"yml file", "deploy.yml", "app: test\nversion: 1", "yaml", true},
		{"unknown file", "notes.txt", "just plain text", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.file)
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write %s: %v", filePath, err)
			}
			app.OpenFile(filePath)
			hl := app.editor.Highlighter()
			if !tt.hasHL {
				if hl != nil {
					t.Errorf("expected no highlighter for %s, got %v", tt.file, hl)
				}
				return
			}
			if hl == nil {
				t.Fatalf("expected highlighter for %s, got nil", tt.file)
			}
			if hl.Language() != tt.wantLang {
				t.Errorf("%s language = %q, want %q", tt.file, hl.Language(), tt.wantLang)
			}
		})
	}
}

func TestSinglePassSyntaxHighlightingInTextEdit(t *testing.T) {
	app := NewApp()
	frames := loom.Render(app, 80, 24)
	joined := strings.Join(frames, "\n")
	if !strings.Contains(joined, "package") || !strings.Contains(joined, "main") {
		t.Fatalf("expected rendered editor to contain keywords: %q", joined)
	}

	// Verify syntax highlighting renders via canvas cells
	c := loom.NewCanvas(80, 24)
	app.Draw(c, c.Bounds())

	// Caret/editor line 0 has "package main" inside textRect (x starts at 5 + borders)
	// Find cell containing 'p' of "package"
	found := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			cell := c.Get(x, y)
			if cell.Text == "p" && cell.Style.Bold && cell.Style.FG == loom.ColorIndex(5) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Errorf("expected to find syntax-highlighted 'p' cell with Bold and ColorIndex(5)")
	}
}

func TestTopBarBreadcrumbs(t *testing.T) {
	app := NewApp()
	// Default demo.go has caret on line 6 (inside func main)
	frames := loom.Render(app, 100, 24)
	topLine := frames[0]
	if !strings.Contains(topLine, "📁 demo.go › 🔧 main") {
		t.Errorf("top bar missing breadcrumb '📁 demo.go › 🔧 main', got: %q", topLine)
	}

	// Move caret to line 0 (package main)
	app.editor.SetCaret(0, 0)
	frames = loom.Render(app, 100, 24)
	topLine = frames[0]
	if strings.Contains(topLine, "› 🔧 main") {
		t.Errorf("top bar should not show main when caret at line 0, got: %q", topLine)
	}
	if !strings.Contains(topLine, "📁 demo.go") {
		t.Errorf("top bar should show 📁 demo.go, got: %q", topLine)
	}

	// Move caret back to line 5 (inside func main)
	app.editor.SetCaret(5, 0)
	frames = loom.Render(app, 100, 24)
	topLine = frames[0]
	if !strings.Contains(topLine, "📁 demo.go › 🔧 main") {
		t.Errorf("top bar missing restored breadcrumb '📁 demo.go › 🔧 main', got: %q", topLine)
	}
}

func TestSidebarOutlineAndJumpCaret(t *testing.T) {
	app := NewApp()
	if app.SidebarMode() != SidebarFiles {
		t.Errorf("default sidebar mode = %v, want SidebarFiles", app.SidebarMode())
	}

	// Toggle to Outline
	app.ToggleSidebarMode()
	if app.SidebarMode() != SidebarOutline {
		t.Errorf("sidebar mode after toggle = %v, want SidebarOutline", app.SidebarMode())
	}

	frames := loom.Render(app, 100, 24)
	joined := strings.Join(frames, "\n")
	if !strings.Contains(joined, "📋 Outline") || !strings.Contains(joined, "🔧 main") {
		t.Fatalf("outline rendering missing symbols: %q", joined)
	}

	// Set caret to top of document
	app.editor.SetCaret(0, 0)
	row, col := app.editor.Caret()
	if row != 0 || col != 0 {
		t.Fatalf("caret = (%d, %d), want (0, 0)", row, col)
	}

	// Select symbol in outline and press Enter
	app.activeFocus = focusBrowser
	app.ConsumeKey(loom.KeyEvent{Key: "enter"})

	row, _ = app.editor.Caret()
	if row != 5 { // func main() is at line index 5 (gutter line 6)
		t.Errorf("caret row after jumping to main = %d, want 5", row)
	}
	if app.activeFocus != focusEditor {
		t.Errorf("activeFocus after jump = %v, want %v", app.activeFocus, focusEditor)
	}
}

func TestOutlineDoubleClickJumpsToSelectedSymbol(t *testing.T) {
	app := NewApp()
	app.editor.SetCaret(0, 0)
	app.activeFocus = focusBrowser
	app.sidebar.outline.refresh()
	if len(app.sidebar.outline.symbols) == 0 {
		t.Fatal("outline has no symbols to select")
	}
	symbol := app.sidebar.outline.symbols[0]
	click := loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, Y: 0}
	app.sidebar.outline.ConsumeMouse(click)
	row, col := app.editor.Caret()
	if app.sidebar.outline.selected != 0 || row != 0 || col != 0 || app.activeFocus != focusBrowser {
		t.Fatalf("first click selected %d, caret (%d,%d), focus %v; want selection only", app.sidebar.outline.selected, row, col, app.activeFocus)
	}
	app.sidebar.outline.ConsumeMouse(loom.MouseEvent{Action: loom.MouseRelease, Button: loom.MouseLeft, Y: 0})
	time.Sleep(10 * time.Millisecond)
	app.sidebar.outline.ConsumeMouse(click)
	if row, col := app.editor.Caret(); row != symbol.Line || col != symbol.Column {
		t.Fatalf("double-click caret = (%d,%d), want symbol at (%d,%d)", row, col, symbol.Line, symbol.Column)
	}
	if app.activeFocus != focusEditor {
		t.Fatalf("focus after double-click = %v, want editor", app.activeFocus)
	}
}

func TestCodeFoldingInTextEdit(t *testing.T) {
	app := NewApp()
	originalText := app.editor.Value()

	// Initial render: gutter should show unfolded fold icon ▾ at line 6 (index 5)
	frames := loom.Render(app, 100, 24)
	joined := strings.Join(frames, "\n")
	if !strings.Contains(joined, " 6▾│ ") {
		t.Fatalf("expected 6▾│ fold indicator in gutter: %q", joined)
	}
	if !strings.Contains(joined, "Hello, Loom!") {
		t.Fatalf("expected Hello, Loom! to be visible before folding")
	}

	// Move caret to line 5 (func main() {)
	app.editor.SetCaret(5, 0)

	// Press F2 to fold
	app.ConsumeKey(loom.KeyEvent{Key: "f2"})
	if !app.editor.IsFolded(5) {
		t.Fatalf("expected line 5 to be folded after F2")
	}

	// Render after folding: gutter should show folded icon ▸ at line 6
	frames = loom.Render(app, 100, 24)
	joined = strings.Join(frames, "\n")
	if !strings.Contains(joined, " 6▸│ ") {
		t.Fatalf("expected 6▸│ fold indicator in gutter after folding: %q", joined)
	}
	if strings.Contains(joined, "Hello, Loom!") {
		t.Fatalf("expected Hello, Loom! to be collapsed inside fold")
	}

	// Buffer must remain intact!
	if app.editor.Value() != originalText {
		t.Fatalf("editor buffer was modified/corrupted by folding!")
	}

	// Press F2 again to unfold
	app.ConsumeKey(loom.KeyEvent{Key: "f2"})
	if app.editor.IsFolded(5) {
		t.Fatalf("expected line 5 to be unfolded after second F2")
	}

	frames = loom.Render(app, 100, 24)
	joined = strings.Join(frames, "\n")
	if !strings.Contains(joined, " 6▾│ ") {
		t.Fatalf("expected 6▾│ fold indicator in gutter after unfolding: %q", joined)
	}
	if !strings.Contains(joined, "Hello, Loom!") {
		t.Fatalf("expected Hello, Loom! to be visible again after unfolding")
	}
}
