package gallery

import (
	"testing"
	"ubunatic.com/loom"
)

func TestCoverageDemos(t *testing.T) {
	for _, name := range []string{"AlignBox", "AnsiEditor", "Confirm", "Frame", "Gauge", "Notif", "Router", "Rows", "Settings", "SplashView", "Split", "Stack", "StyledRows"} {
		t.Run(name, func(t *testing.T) {
			w, err := New(name)
			if err != nil {
				t.Fatal(err)
			}
			if len(loom.Render(w, 80, 24)) == 0 {
				t.Fatal("empty render")
			}
		})
	}
}

func TestTextEntryDemoControls(t *testing.T) {
	d := newTextEntryControls().(*textEntryControls)
	d.Draw(loom.NewCanvas(80, 8), loom.Rect{W: 80, H: 8})
	if !d.ConsumeKey(loom.KeyEvent{Key: "f2"}).Consumed || d.entry.Value() != "" || d.entry.Placeholder() != "Write an update…" {
		t.Fatal("placeholder control failed")
	}
	if !d.ConsumeKey(loom.KeyEvent{Key: "f3"}).Consumed || d.entry.Value() != "Write an update…" {
		t.Fatal("accept control failed")
	}
	if !d.ConsumeKey(loom.KeyEvent{Key: "f4"}).Consumed || d.entry.CursorOffset() != 0 {
		t.Fatal("caret control failed")
	}
	d.Draw(loom.NewCanvas(80, 8), loom.Rect{W: 80, H: 8})
	if !d.ConsumeMouse(loom.MouseEvent{Action: loom.MousePress, Button: loom.MouseLeft, X: 5, Y: 0}).Consumed || d.entry.CursorOffset() == 0 {
		t.Fatal("mouse caret movement failed")
	}
}
