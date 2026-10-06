package loom

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFilePickerNavigatesFiltersAndSelects(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.go", "b.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var selected string
	picker, err := NewFilePicker(root, FilePickerOptions{Mode: FilePickerFiles, Patterns: []string{"*.go"}, OnSelect: func(path string) { selected = path }})
	if err != nil {
		t.Fatal(err)
	}
	if got := picker.Entries(); len(got) != 3 || got[0].Name != ".." || got[1].Name != "a.go" || got[2].Name != "child" {
		t.Fatalf("filtered entries = %#v", got)
	}
	picker.ConsumeKey(KeyEvent{Text: "ag"})
	if got, ok := picker.Selected(); !ok || got.Name != "a.go" {
		t.Fatalf("fuzzy filtered selection = %#v, %v", got, ok)
	}
	picker.ConsumeKey(KeyEvent{Key: "backspace"})
	picker.ConsumeKey(KeyEvent{Key: "backspace"})
	picker.List().SelectIndex(2)
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if picker.Directory().Path != child {
		t.Fatalf("directory = %q, want %q", picker.Directory().Path, child)
	}
	picker.ConsumeKey(KeyEvent{Key: "backspace"})
	if picker.Directory().Path != root {
		t.Fatalf("directory after parent = %q", picker.Directory().Path)
	}
	picker.List().SelectIndex(1)
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if selected != filepath.Join(root, "a.go") {
		t.Fatalf("selected = %q", selected)
	}
}

func TestFilePickerDirectoryModeAndCancel(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	var selected, cancelled string
	picker, err := NewFilePicker(root, FilePickerOptions{Mode: FilePickerDirectories, OnSelect: func(path string) { selected = path }, OnCancel: func() { cancelled = "yes" }})
	if err != nil {
		t.Fatal(err)
	}
	picker.ConsumeKey(KeyEvent{Key: "down"})
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if selected != child {
		t.Fatalf("selected = %q, want %q", selected, child)
	}
	picker, err = NewFilePicker(root, FilePickerOptions{Mode: FilePickerDirectories, OnCancel: func() { cancelled = "yes" }})
	if err != nil {
		t.Fatal(err)
	}
	picker.ConsumeKey(KeyEvent{Key: "esc"})
	if cancelled != "yes" {
		t.Fatal("cancel callback was not called")
	}
}

func TestFilePickerSaveModeSelectsEnteredFileName(t *testing.T) {
	root := t.TempDir()
	var selected string
	picker, err := NewFilePicker(root, FilePickerOptions{
		Mode:          FilePickerSave,
		FileName:      "new.ansi",
		FocusFileName: true,
		OnSelect:      func(path string) { selected = path },
	})
	if err != nil {
		t.Fatal(err)
	}
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if selected != filepath.Join(root, "new.ansi") {
		t.Fatalf("selected save path = %q, want %q", selected, filepath.Join(root, "new.ansi"))
	}
}

func TestFilePickerSaveModeAcceptsTypedDestinationAndCancels(t *testing.T) {
	root := t.TempDir()
	var selected, cancelled string
	picker, err := NewFilePicker(root, FilePickerOptions{
		Mode:     FilePickerSave,
		OnSelect: func(path string) { selected = path },
		OnCancel: func() { cancelled = "yes" },
	})
	if err != nil {
		t.Fatal(err)
	}
	if picker.nameFocus {
		t.Fatal("save picker stole focus from directory search")
	}
	picker.ConsumeKey(KeyEvent{Text: "typed"})
	if got := picker.List().Query(); got != "typed" {
		t.Fatalf("save directory query = %q, want typed", got)
	}
	picker.ConsumeKey(KeyEvent{Key: "tab"})
	for _, r := range "typed.ansi" {
		picker.ConsumeKey(KeyEvent{Text: string(r)})
	}
	if got := picker.FileName(); got != "typed.ansi" {
		t.Fatalf("entered file name = %q", got)
	}
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if selected != filepath.Join(root, "typed.ansi") {
		t.Fatalf("selected save path = %q", selected)
	}

	picker, err = NewFilePicker(root, FilePickerOptions{Mode: FilePickerSave, OnCancel: func() { cancelled = "yes" }})
	if err != nil {
		t.Fatal(err)
	}
	picker.ConsumeKey(KeyEvent{Key: "esc"})
	if cancelled != "yes" {
		t.Fatal("save picker cancel callback was not called")
	}
}

func TestFilePickerSaveModeNavigatesDirectories(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	picker, err := NewFilePicker(root, FilePickerOptions{Mode: FilePickerSave, FileName: "out.ansi"})
	if err != nil {
		t.Fatal(err)
	}
	picker.List().SelectIndex(1)
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if picker.Directory().Path != child {
		t.Fatalf("save picker directory = %q, want %q", picker.Directory().Path, child)
	}
}

func TestFilePickerSaveModeFilenameFocusIsExplicitAndSwitchable(t *testing.T) {
	picker, err := NewFilePicker(t.TempDir(), FilePickerOptions{
		Mode:     FilePickerSave,
		FileName: "seed.ansi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if picker.nameFocus {
		t.Fatal("seeded filename stole focus without explicit option")
	}
	picker.ConsumeKey(KeyEvent{Key: "tab"})
	if !picker.nameFocus {
		t.Fatal("Tab did not switch focus to filename editing")
	}
	picker.ConsumeKey(KeyEvent{Key: "shift-tab"})
	if picker.nameFocus {
		t.Fatal("Shift+Tab did not switch focus back to directory search")
	}
	explicit, err := NewFilePicker(t.TempDir(), FilePickerOptions{Mode: FilePickerSave, FocusFileName: true})
	if err != nil {
		t.Fatal(err)
	}
	if !explicit.nameFocus {
		t.Fatal("explicit filename focus option was ignored")
	}
}

func TestFilePickerSaveModeFocusMovesCursorAndSeparatesFilename(t *testing.T) {
	picker, err := NewFilePicker(t.TempDir(), FilePickerOptions{Mode: FilePickerSave, FileName: "report.ansi"})
	if err != nil {
		t.Fatal(err)
	}
	canvas := NewCanvas(40, 8)
	picker.Draw(canvas, Rect{W: 40, H: 8})
	if canvas.CursorY != 5 {
		t.Fatalf("search cursor row = %d, want 5", canvas.CursorY)
	}
	picker.ConsumeKey(KeyEvent{Key: "tab"})
	picker.Draw(canvas, Rect{W: 40, H: 8})
	if canvas.CursorY != 7 || canvas.CursorX < 6 {
		t.Fatalf("filename cursor = (%d,%d), want filename input row and column", canvas.CursorX, canvas.CursorY)
	}
	if got := canvas.Get(0, 6).Text; got != "─" {
		t.Fatalf("filename divider cell = %q, want horizontal divider", got)
	}
	picker.ConsumeKey(KeyEvent{Key: "shift-tab"})
	picker.Draw(canvas, Rect{W: 40, H: 8})
	if canvas.CursorY != 5 {
		t.Fatalf("search cursor row after Shift+Tab = %d, want 5", canvas.CursorY)
	}
}

func TestFilePickerSaveModeMouseSelectsFilenameAndDirectoryFocus(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "seed.ansi"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	picker, err := NewFilePicker(root, FilePickerOptions{Mode: FilePickerSave, FileName: "seed.ansi"})
	if err != nil {
		t.Fatal(err)
	}
	picker.Draw(NewCanvas(40, 8), Rect{W: 40, H: 8})
	picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 7})
	if !picker.nameFocus {
		t.Fatal("clicking filename row did not focus filename editing")
	}
	filenameCanvas := NewCanvas(40, 8)
	picker.Draw(filenameCanvas, Rect{W: 40, H: 8})
	if filenameCanvas.CursorY != 7 {
		t.Fatalf("filename cursor row after click = %d, want 7", filenameCanvas.CursorY)
	}
	picker.ConsumeKey(KeyEvent{Text: "x"})
	if got := picker.FileName(); got != "seed.ansi"+"x" {
		t.Fatalf("filename after click and type = %q", got)
	}
	picker.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft, X: 2, Y: 2})
	if picker.nameFocus {
		t.Fatal("clicking directory list did not restore search focus")
	}
	searchCanvas := NewCanvas(40, 8)
	picker.Draw(searchCanvas, Rect{W: 40, H: 8})
	if searchCanvas.CursorY != 5 {
		t.Fatalf("search cursor row after click = %d, want 5", searchCanvas.CursorY)
	}
}

func TestFilePickerEnterOnListedFileFocusesFilenameInput(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "sample.ansi")
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	savedPath := ""
	picker, err := NewFilePicker(root, FilePickerOptions{
		Mode: FilePickerSave,
		OnSave: func(path string) error {
			savedPath = path
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Press Enter on the listed file
	picker.ConsumeKey(KeyEvent{Key: "enter"})

	if savedPath != "" {
		t.Fatalf("first Enter saved file immediately to %q; expected focus move only", savedPath)
	}
	if !picker.nameFocus {
		t.Fatal("Enter on listed file did not set nameFocus = true")
	}
	if picker.FileName() != "sample.ansi" {
		t.Fatalf("picker.FileName() = %q, want sample.ansi", picker.FileName())
	}

	canvas := NewCanvas(40, 8)
	picker.Draw(canvas, Rect{W: 40, H: 8})
	if canvas.CursorY != 7 {
		t.Fatalf("canvas.CursorY = %d, want 7 (filename row)", canvas.CursorY)
	}
	// "Name: sample.ansi" length is 6 + 11 = 17 -> cursor at x = 17
	if canvas.CursorX != 17 {
		t.Fatalf("canvas.CursorX = %d, want 17 (end of filename)", canvas.CursorX)
	}

	// Second Enter while nameFocus is true saves the file
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if savedPath != filePath {
		t.Fatalf("second Enter saved path = %q, want %q", savedPath, filePath)
	}
}

func TestFilePickerMouseHoverDoesNotChangeFocus(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "seed.ansi"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	picker, err := NewFilePicker(root, FilePickerOptions{Mode: FilePickerSave, FileName: "seed.ansi"})
	if err != nil {
		t.Fatal(err)
	}
	picker.Draw(NewCanvas(40, 8), Rect{W: 40, H: 8})

	// Start with filename focus
	picker.nameFocus = true
	picker.updateFocus()

	// Hover over list area
	picker.ConsumeMouse(MouseEvent{Action: MouseHover, X: 5, Y: 2})
	if !picker.nameFocus {
		t.Fatal("MouseHover over list stole focus from filename input")
	}

	// Now switch to list focus
	picker.nameFocus = false
	picker.updateFocus()

	// Hover over filename row
	picker.ConsumeMouse(MouseEvent{Action: MouseHover, X: 5, Y: 7})
	if picker.nameFocus {
		t.Fatal("MouseHover over filename row stole focus from list")
	}
}

func TestFilePickerSaveModeKeepsOpenWhenSaveCallbackFails(t *testing.T) {
	wantErr := errors.New("write failed")
	root := t.TempDir()
	picker, err := NewFilePicker(root, FilePickerOptions{
		Mode:          FilePickerSave,
		FileName:      "document.ansi",
		FocusFileName: true,
		OnSave:        func(string) error { return wantErr },
	})
	if err != nil {
		t.Fatal(err)
	}
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if picker.done {
		t.Fatal("picker closed after save callback failed")
	}
}
