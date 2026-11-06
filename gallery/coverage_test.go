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

func TestCoverageDemoUnwrap(t *testing.T) {
	for _, name := range []string{"Router", "TextEntry"} {
		t.Run(name, func(t *testing.T) {
			w, err := New(name)
			if err != nil {
				t.Fatal(err)
			}
			wrapper, ok := w.(interface{ Unwrap() loom.Widget })
			if !ok || wrapper.Unwrap() == nil {
				t.Fatal("demo must expose its underlying widget")
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

func TestTextEntryDemoCtrlCPolicy(t *testing.T) {
	d := newTextEntryControls().(*textEntryControls)
	d.entry.SetValue("selected draft")

	// TextEntry gets first refusal so copying a selection never clears input.
	if !d.ConsumeKey(loom.KeyEvent{Key: "ctrl-a"}).Consumed {
		t.Fatal("select all was not consumed")
	}
	if result := d.ConsumeKey(loom.KeyEvent{Key: "ctrl-c"}); !result.Consumed || result.Quit || d.entry.Value() != "selected draft" {
		t.Fatalf("selected ctrl-c = %+v, value = %q", result, d.entry.Value())
	}

	// Collapse selection, then the gallery parent clears nonempty input.
	if !d.ConsumeKey(loom.KeyEvent{Key: "right"}).Consumed {
		t.Fatal("right arrow did not collapse selection")
	}
	if result := d.ConsumeKey(loom.KeyEvent{Key: "ctrl-c"}); !result.Consumed || result.Quit || d.entry.Value() != "" {
		t.Fatalf("unselected nonempty ctrl-c = %+v, value = %q", result, d.entry.Value())
	}

	// Empty input stays unhandled by the wrapper, preserving parent routing.
	if result := d.ConsumeKey(loom.KeyEvent{Key: "ctrl-c"}); result.Consumed || result.Quit {
		t.Fatalf("empty ctrl-c = %+v, want ignored", result)
	}
	if result := d.ConsumeKey(loom.KeyEvent{Text: "x"}); !result.Consumed || d.entry.Value() != "x" {
		t.Fatalf("typing after empty ctrl-c = %+v, value = %q", result, d.entry.Value())
	}
	if result := d.ConsumePaste(loom.PasteEvent{Text: "\n[y]"}); !result.Consumed || d.entry.Value() != "x\n[y]" {
		t.Fatalf("paste after empty ctrl-c = %+v, value = %q", result, d.entry.Value())
	}
}
