// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestEditViewUnsavedChangesDialog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.ansi")
	if err := os.WriteFile(path, []byte("Initial content\n"), 0600); err != nil {
		t.Fatal(err)
	}
	edit, err := loom.NewRichTextEditFromFile(path)
	if err != nil {
		t.Fatalf("NewRichTextEditFromFile: %v", err)
	}
	view, err := newEditView(edit, path, loom.EditorConfig{Theme: "plain", AltScreen: true})
	if err != nil {
		t.Fatalf("newEditView: %v", err)
	}
	canvas := loom.NewCanvas(80, 10)
	view.Draw(canvas, canvas.Bounds())
	screenText := canvasScreenText(canvas)
	if !strings.Contains(screenText, "Initial content") {
		t.Fatalf("editor did not render initial content: %q", screenText)
	}
	if !strings.Contains(screenText, "F10 Quit") {
		t.Fatalf("editor bottom bar omits F10 Quit: %q", screenText)
	}
	quits := countQuits(view)
	if got := view.closeRequest(loom.CloseReasonQuitKey); got != loom.CloseAllow {
		t.Fatalf("close request on unmodified view = %v, want allow", got)
	}
	view.edit.ConsumeKey(loom.KeyEvent{Text: "X"})
	if !view.edit.IsModified() {
		t.Fatal("document expected to be modified after typing")
	}
	if got := view.closeRequest(loom.CloseReasonQuitKey); got != loom.CloseVeto {
		t.Fatalf("close request on modified view = %v, want veto", got)
	}
	if view.unsavedDialog == nil || !view.unsavedDialog.Open {
		t.Fatal("close request on modified view did not open unsaved changes dialog")
	}
	canvas.Clear()
	view.Draw(canvas, canvas.Bounds())
	screenText = canvasScreenText(canvas)
	if !strings.Contains(screenText, "Save changes?") {
		t.Fatalf("dialog title missing in screen text: %q", screenText)
	}
	view.ConsumeKey(loom.KeyEvent{Key: "esc"})
	if view.unsavedDialog != nil {
		t.Fatal("esc did not dismiss unsaved changes dialog")
	}
	if *quits != 0 {
		t.Fatal("cancelling dialog must not quit")
	}
	view.closeRequest(loom.CloseReasonQuitKey)
	if view.unsavedDialog == nil {
		t.Fatal("dialog failed to reopen")
	}
	view.ConsumeKey(loom.KeyEvent{Key: "right"})
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if *quits != 1 {
		t.Fatalf("Discard quits = %d, want 1", *quits)
	}
	view.edit.ConsumeKey(loom.KeyEvent{Text: "Y"})
	view.closeRequest(loom.CloseReasonQuitKey)
	view.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if *quits != 2 {
		t.Fatalf("Save quits = %d, want 2", *quits)
	}
	savedData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(savedData), "Initial content") || !strings.Contains(string(savedData), "XY") {
		t.Fatalf("saved disk content = %q, want modified content", string(savedData))
	}
}
