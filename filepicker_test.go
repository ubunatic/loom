package loom

import (
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
	picker.HandleKey(KeyEvent{Text: "ag"})
	if got, ok := picker.Selected(); !ok || got.Name != "a.go" {
		t.Fatalf("fuzzy filtered selection = %#v, %v", got, ok)
	}
	picker.HandleKey(KeyEvent{Key: "backspace"})
	picker.HandleKey(KeyEvent{Key: "backspace"})
	picker.List().SelectIndex(2)
	picker.HandleKey(KeyEvent{Key: "enter"})
	if picker.Directory().Path != child {
		t.Fatalf("directory = %q, want %q", picker.Directory().Path, child)
	}
	picker.HandleKey(KeyEvent{Key: "backspace"})
	if picker.Directory().Path != root {
		t.Fatalf("directory after parent = %q", picker.Directory().Path)
	}
	picker.List().SelectIndex(1)
	picker.HandleKey(KeyEvent{Key: "enter"})
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
	picker.HandleKey(KeyEvent{Key: "down"})
	picker.HandleKey(KeyEvent{Key: "enter"})
	if selected != child {
		t.Fatalf("selected = %q, want %q", selected, child)
	}
	picker, err = NewFilePicker(root, FilePickerOptions{Mode: FilePickerDirectories, OnCancel: func() { cancelled = "yes" }})
	if err != nil {
		t.Fatal(err)
	}
	picker.HandleKey(KeyEvent{Key: "esc"})
	if cancelled != "yes" {
		t.Fatal("cancel callback was not called")
	}
}
