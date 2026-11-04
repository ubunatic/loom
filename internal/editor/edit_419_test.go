// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package editor

import (
	"os"
	"path/filepath"
	"testing"

	"ubunatic.com/loom"
)

func drawFirstFrame(t *testing.T, v *editView) {
	t.Helper()
	v.Draw(loom.NewCanvas(100, 24), loom.Rect{W: 100, H: 24})
}

func TestEdit419FilesOptionShowsFocusedBrowserInFirstFrame(t *testing.T) {
	v, err := newEditViewWith(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "plain"}, Options{ShowFiles: true})
	if err != nil {
		t.Fatal(err)
	}
	drawFirstFrame(t, v)
	if v.browserRect.W <= 0 || v.browserRect.H <= 0 {
		t.Fatalf("browser rect = %+v, want non-empty pane on first frame", v.browserRect)
	}
	if v.focused != focusBrowser {
		t.Fatalf("focus = %v on first frame, want focusBrowser", v.focused)
	}
}

func TestEdit419DefaultKeepsFilePaneClosed(t *testing.T) {
	v, err := newEditViewWith(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "plain"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	drawFirstFrame(t, v)
	if v.browserRect.W != 0 || v.browserRect.H != 0 {
		t.Fatalf("browser rect = %+v, want empty pane by default", v.browserRect)
	}
	if v.focused != focusEditor {
		t.Fatalf("focus = %v by default, want focusEditor", v.focused)
	}
}

func TestEdit419DirectoryArgumentRootsPickerAndOpensNoBuffer(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("hi\n"), 0600); err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}

	path, opts, err := resolveTarget(dir, Options{})
	if err != nil {
		t.Fatalf("resolveTarget(dir): %v", err)
	}
	if path != "" {
		t.Fatalf("resolveTarget(dir) path = %q, want no file", path)
	}
	if !opts.ShowFiles || opts.Dir != want {
		t.Fatalf("resolveTarget(dir) opts = %+v, want ShowFiles and Dir %q", opts, want)
	}

	v, err := newEditViewWith(loom.NewRichTextEdit(nil), path, loom.EditorConfig{Theme: "plain"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if got := v.filePicker.Directory().Path; got != want {
		t.Fatalf("picker root = %q, want %q", got, want)
	}
	if v.filePath != "" || v.edit.FilePath != "" {
		t.Fatalf("directory argument opened file %q / %q, want none", v.filePath, v.edit.FilePath)
	}
	drawFirstFrame(t, v)
	if v.browserRect.W <= 0 || v.focused != focusBrowser {
		t.Fatalf("browser rect = %+v focus = %v, want shown and focused", v.browserRect, v.focused)
	}
}

func TestEdit419FileArgumentKeepsDefaultsAndPicker(t *testing.T) {
	file := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(file, []byte("doc\n"), 0600); err != nil {
		t.Fatal(err)
	}
	path, opts, err := resolveTarget(file, Options{})
	if err != nil {
		t.Fatalf("resolveTarget(file): %v", err)
	}
	if path != file || opts.ShowFiles || opts.Dir != "" {
		t.Fatalf("resolveTarget(file) = %q, %+v; want unchanged file and zero options", path, opts)
	}
}

func TestEdit419ResolveTargetEmptyArgumentIsUntitled(t *testing.T) {
	path, opts, err := resolveTarget("", Options{})
	if err != nil || path != "" || opts.ShowFiles || opts.Dir != "" {
		t.Fatalf("resolveTarget(\"\") = %q, %+v, %v; want untitled defaults", path, opts, err)
	}
}
