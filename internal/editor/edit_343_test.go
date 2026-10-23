package editor

import (
	"strings"
	"testing"

	"ubunatic.com/loom"
)

func Test343MouseGrabHotkeyTogglesFromOffAndUpdatesIndicator(t *testing.T) {
	view, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "plain", MouseGrab: false})
	if err != nil {
		t.Fatal(err)
	}
	canvas := loom.NewCanvas(100, 24)
	view.Draw(canvas, canvas.Bounds())
	if !hasStatusIcon343(canvasPlainRow(canvas, 21), loom.SpeccedDefaults.Editor.StatusIcons.MouseOff) {
		t.Fatal("mouse-off indicator is missing before toggle")
	}

	for i, want := range []bool{true, false, true} {
		result := view.ConsumeKey(loom.KeyEvent{Key: loom.SpeccedDefaults.Editor.HotkeyMouseGrabBinding})
		if !result.Consumed || view.config.MouseGrab != want {
			t.Fatalf("toggle %d: result=%+v mousegrab=%v, want %v", i+1, result, view.config.MouseGrab, want)
		}
		view.Draw(canvas, canvas.Bounds())
		indicator := loom.SpeccedDefaults.Editor.StatusIcons.MouseOff
		if want {
			indicator = loom.SpeccedDefaults.Editor.StatusIcons.MouseOn
		}
		if !hasStatusIcon343(canvasPlainRow(canvas, 21), indicator) {
			t.Fatalf("toggle %d: indicator %q missing", i+1, indicator)
		}
	}
}

func hasStatusIcon343(row, icon string) bool {
	return strings.Contains(" "+strings.Join(strings.Fields(row), " ")+" ", " "+icon+" ")
}

func Test343MouseGrabHotkeyPreemptsEditorModals(t *testing.T) {
	view, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "plain", MouseGrab: false})
	if err != nil {
		t.Fatal(err)
	}
	view.Draw(loom.NewCanvas(100, 24), loom.Rect{W: 100, H: 24})
	view.ConsumeKey(loom.KeyEvent{Key: loom.SpeccedDefaults.RichTextEdit.HotkeyHelpBinding})
	if !view.edit.ModalOpen() {
		t.Fatal("help modal did not open")
	}
	result := view.ConsumeKey(loom.KeyEvent{Key: loom.SpeccedDefaults.Editor.HotkeyMouseGrabBinding})
	if !result.Consumed || !view.config.MouseGrab || !view.edit.ModalOpen() {
		t.Fatalf("hotkey did not toggle while preserving help: result=%+v mousegrab=%v modal=%v", result, view.config.MouseGrab, view.edit.ModalOpen())
	}
	view.ConsumeKey(loom.KeyEvent{Key: "esc"})
	view.edit.ConsumeKey(loom.KeyEvent{Text: "x"})
	view.edit.RequestClose()
	if view.unsavedDialog == nil || !view.unsavedDialog.Open {
		t.Fatal("unsaved dialog did not open")
	}
	result = view.ConsumeKey(loom.KeyEvent{Key: loom.SpeccedDefaults.Editor.HotkeyMouseGrabBinding})
	if !result.Consumed || view.config.MouseGrab || !view.unsavedDialog.Open {
		t.Fatalf("hotkey did not toggle while preserving dialog: result=%+v mousegrab=%v dialog=%v", result, view.config.MouseGrab, view.unsavedDialog.Open)
	}
}

func Test343MouseGrabStatusClickTogglesGrab(t *testing.T) {
	view, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "plain", MouseGrab: true})
	if err != nil {
		t.Fatal(err)
	}
	view.Draw(loom.NewCanvas(100, 24), loom.Rect{W: 100, H: 24})
	for x := view.indicatorRect.X; x < view.indicatorRect.X+view.indicatorRect.W; x++ {
		view.ConsumeMouse(loom.MouseEvent{X: x, Y: view.indicatorRect.Y, Action: loom.MousePress, Button: loom.MouseLeft})
		if !view.config.MouseGrab {
			return
		}
	}
	t.Fatal("status indicator click did not disable mouse grab")
}

func Test343MouseGrabHotkeyAppearsInHelpAndHintBar(t *testing.T) {
	defs := loom.SpeccedDefaults.Editor
	help := helpSection()
	found := false
	for _, entry := range help.Entries {
		if entry.Label == defs.HotkeyMouseGrabLabel && entry.Cap == defs.HotkeyMouseGrabKey {
			found = true
		}
	}
	if !found {
		t.Fatal("mousegrab hotkey missing from F1 help")
	}
	view, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "plain"})
	if err != nil {
		t.Fatal(err)
	}
	bar := view.hotkeyBar()
	found = false
	for _, entry := range bar.Entries {
		if entry.Binding == defs.HotkeyMouseGrabBinding && entry.Label == defs.HotkeyMouseGrabLabel {
			found = true
		}
	}
	if !found {
		t.Fatal("mousegrab hotkey missing from hint bar")
	}
	if got := bar.ConsumeKey(loom.KeyEvent{Key: defs.HotkeyMouseGrabBinding}); !got.Consumed || !view.config.MouseGrab {
		t.Fatalf("hint bar binding does not use mousegrab action: result=%+v mousegrab=%v", got, view.config.MouseGrab)
	}
}
