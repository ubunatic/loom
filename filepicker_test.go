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
		Mode:     FilePickerSave,
		FileName: "new.ansi",
		OnSelect: func(path string) { selected = path },
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
	picker.ConsumeKey(KeyEvent{Key: "tab"})
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if picker.Directory().Path != child {
		t.Fatalf("save picker directory = %q, want %q", picker.Directory().Path, child)
	}
}

func TestFilePickerSaveModeKeepsOpenWhenSaveCallbackFails(t *testing.T) {
	wantErr := errors.New("write failed")
	root := t.TempDir()
	picker, err := NewFilePicker(root, FilePickerOptions{
		Mode:     FilePickerSave,
		FileName: "document.ansi",
		OnSave:   func(string) error { return wantErr },
	})
	if err != nil {
		t.Fatal(err)
	}
	picker.ConsumeKey(KeyEvent{Key: "enter"})
	if picker.done {
		t.Fatal("picker closed after save callback failed")
	}
}
