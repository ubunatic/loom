package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func TestEditorSpecDrivesLayoutAndActions(t *testing.T) {
	before := loom.SpeccedDefaults.Editor
	t.Cleanup(func() { loom.SpeccedDefaults.Editor = before })
	d := &loom.SpeccedDefaults.Editor
	d.Title = "Spec editor"
	d.BrowserMaxWidth = 19
	d.SearchMaxWidth = 39
	d.HotkeySearchBinding = "ctrl-alt-f"
	view, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: d.Theme})
	if err != nil {
		t.Fatal(err)
	}
	view.showSidePanel = true
	view.focused = focusEditor
	if !view.ConsumeKey(loom.KeyEvent{Key: d.HotkeySearchBinding}).Consumed || !view.showSearch {
		t.Fatal("spec search key ignored")
	}
	c := loom.NewCanvas(100, 24)
	view.Draw(c, c.Bounds())
	if view.browserRect.W != d.BrowserMaxWidth || view.searchRect.W != d.SearchMaxWidth {
		t.Fatalf("spec geometry ignored: browser=%+v search=%+v", view.browserRect, view.searchRect)
	}
	if !strings.Contains(canvasScreenText(c), d.Title) {
		t.Fatal("spec title ignored")
	}
	// The clickable hint must dispatch the same spec-defined search action.
	view.showSearch = false
	view.Draw(c, c.Bounds())
	bar := view.hints
	x := 1
	for _, entry := range bar.Entries {
		if entry.Binding == d.HotkeySearchBinding {
			break
		}
		single := loom.NewHintBar(entry)
		single.Style = bar.Style
		x += single.ContentWidth()
	}
	view.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: view.lastRect.X + 1 + x, Y: view.lastRect.Y + view.lastRect.H - 2})
	if !view.showSearch {
		t.Fatal("mouse hint ignored spec search action")
	}
}

func TestEditorFlagDefaultsComeFromSpec(t *testing.T) {
	d := loom.SpeccedDefaults.Editor
	cmd := editCommand()
	if got, _ := cmd.Flags().GetBool("altscreen"); got != d.AltScreen {
		t.Fatal("altscreen flag differs from spec")
	}
	if got, _ := cmd.Flags().GetBool("mousegrab"); got != d.MouseGrab {
		t.Fatal("mousegrab flag differs from spec")
	}
}

func TestEditorSpecDrivesScreenshotPolicy(t *testing.T) {
	before := loom.SpeccedDefaults.Editor
	t.Cleanup(func() { loom.SpeccedDefaults.Editor = before })
	d := &loom.SpeccedDefaults.Editor
	d.ScreenshotDirectory = []string{"captures"}
	d.ScreenshotMarker = "-test-"
	d.ScreenshotFirstNumber = 7
	d.ScreenshotNumberWidth = 3
	d.ScreenshotExtension = ".txt"
	d.ScreenshotUntitled = "empty"
	d.ScreenshotSavedPrefix = "Captured: "
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	v, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: d.Theme})
	if err != nil {
		t.Fatal(err)
	}
	if err := v.captureScreenshot(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, d.ScreenshotDirectory[0], "007-test-empty.txt")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if v.statusMessage != d.ScreenshotSavedPrefix+path {
		t.Fatalf("screenshot status = %q", v.statusMessage)
	}
}
