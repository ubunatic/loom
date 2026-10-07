package loom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test302DefaultTheme(t *testing.T) {
	if SpeccedDefaults.Editor.Theme != "julia256" || DefaultTheme() != Theme("julia256") || Theme("") != DefaultTheme() {
		t.Fatal("default theme is not julia256")
	}
	if DefaultChoiceStyle() != DefaultTheme().ChoiceStyle() || DefaultTableStyle() != DefaultTheme().TableStyle() {
		t.Fatal("widget defaults do not use the default theme")
	}
}

func Test302MouseDisableAndReenable(t *testing.T) {
	for _, mode := range []int{0, 1000, 1003} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tty")
			out, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			p := &Pane{tty: out, mouse: mode != 0, mouseMode: mode}
			p.DisableMouse()
			if p.mouse || p.mouseMode != 0 {
				t.Fatal("mouse reporting still enabled")
			}
			p.EnableMouse()
			if !p.mouse || p.mouseMode != 1003 {
				t.Fatal("mouse tracking did not reenable")
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l\x1b[?1003h\x1b[?1006h" {
				t.Fatalf("terminal sequences = %q", data)
			}
		})
	}
}

func Test302BoxEnterFinishesWithoutNewline(t *testing.T) {
	for _, key := range []string{"enter", "return"} {
		e := NewRichTextEdit(nil)
		e.ConsumeKey(KeyEvent{Key: "ctrl-d"})
		e.ConsumeKey(KeyEvent{Key: "right"})
		before := e.Document.ToANSI()
		if !e.BoxMode {
			t.Fatal("draw mode did not start")
		}
		res := e.ConsumeKey(KeyEvent{Key: key})
		if !res.Consumed || e.BoxMode || e.Document.ToANSI() != before {
			t.Fatalf("%s did not finish draw mode without editing", key)
		}
	}
}

func Test302PopupBackdropAndWheel(t *testing.T) {
	h := newRichTextEditHelp(nil)
	p := NewPopup("Help", h)
	p.Width, p.Height = 30, 8
	c := NewCanvas(80, 30)
	r := Rect{X: 5, Y: 3, W: 60, H: 20}
	p.Draw(c, r)
	x, y := p.innerRect.X-r.X, p.innerRect.Y-r.Y
	for range h.maxScroll + 5 {
		p.ConsumeMouse(MouseEvent{X: x, Y: y, Action: MouseScrollDown})
	}
	if h.scroll != h.maxScroll || h.scroll == 0 || !p.Open {
		t.Fatal("wheel did not scroll to end")
	}
	for range h.maxScroll + 5 {
		p.ConsumeMouse(MouseEvent{X: x, Y: y, Action: MouseScrollUp})
	}
	if h.scroll != 0 {
		t.Fatal("wheel did not clamp to start")
	}
	p.ConsumeMouse(MouseEvent{X: p.popupRect.X - r.X, Y: p.popupRect.Y - r.Y, Action: MousePress, Button: MouseLeft})
	if !p.Open {
		t.Fatal("border click dismissed popup")
	}
	p.ConsumeMouse(MouseEvent{Action: MouseHover})
	if !p.Open {
		t.Fatal("hover dismissed popup")
	}
	if res := p.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft}); !res.Consumed || p.Open {
		t.Fatal("backdrop did not dismiss")
	}
}

func Test302DialogBackdropCancels(t *testing.T) {
	for _, placed := range []bool{false, true} {
		d := NewDialog("Confirm", "Keep edits?", "Save", "Cancel", "Discard")
		if placed {
			d.Rect = Rect{X: 12, Y: 8, W: 36, H: 7}
		}
		calls := 0
		d.OnSelect = func(label string) {
			calls++
			if label != "Cancel" {
				t.Fatalf("selected %q", label)
			}
		}
		d.Draw(NewCanvas(80, 30), Rect{X: 5, Y: 3, W: 60, H: 20})
		d.ConsumeMouse(MouseEvent{Action: MouseScrollDown})
		if !d.Open || calls != 0 {
			t.Fatal("wheel cancelled dialog")
		}
		d.ConsumeMouse(MouseEvent{Action: MousePress, Button: MouseLeft})
		if d.Open || calls != 1 || d.SelectedButton() != "Cancel" {
			t.Fatal("backdrop did not invoke Cancel")
		}
		d.ConsumeMouse(MouseEvent{Action: MouseRelease, Button: MouseLeft})
		if calls != 1 {
			t.Fatal("release cancelled twice")
		}
	}
}

func Test302PaneHelpWheelAndBackdrop(t *testing.T) {
	h := newHelpWidget([]Cmd{{Name: "one"}, {Name: "two"}, {Name: "three"}})
	p := &Pane{help: NewPopup("Help", h)}
	p.help.Height = 4
	p.help.Draw(NewCanvas(40, 20), Rect{W: 40, H: 20})
	r := p.help.innerRect
	p.handleHelpMouse(MouseEvent{X: r.X, Y: r.Y, Action: MouseScrollDown})
	if h.scroll != 1 || p.help == nil {
		t.Fatal("pane help wheel failed")
	}
	p.handleHelpKey(KeyEvent{Key: "up"})
	if h.scroll != 0 || p.help == nil {
		t.Fatal("pane help navigation closed popup")
	}
	p.handleHelpMouse(MouseEvent{Action: MousePress, Button: MouseLeft})
	if p.help != nil {
		t.Fatal("pane retained dismissed help")
	}
}

func Test302SaveAsShortcutAbsent(t *testing.T) {
	e := NewRichTextEdit(nil)
	for _, fileBar := range []bool{false, true} {
		e.ShowFileBar = fileBar
		if e.ConsumeKey(KeyEvent{Key: "ctrl-alt-s"}).Consumed || e.ModalOpen() {
			t.Fatal("Save as shortcut still active")
		}
	}
	if strings.Contains(strings.Join(newRichTextEditHelp(nil).plainLines(72), "\n"), "⌃⌥S") {
		t.Fatal("Save as shortcut still in help")
	}
	e.ensureFileBar().menu.Menus[0].Items[3].Action()
	if !e.ModalOpen() {
		t.Fatal("menu Save as did not open picker")
	}
}
