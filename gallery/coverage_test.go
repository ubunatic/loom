package gallery

import (
	"strings"
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

func TestTextEntryDemoCtrlUForwarding(t *testing.T) {
	d := newTextEntryControls().(*textEntryControls)
	d.entry.SetValue("draft")
	if result := d.ConsumeKey(loom.KeyEvent{Key: "ctrl-u"}); !result.Consumed || result.Quit || d.entry.Value() != "" {
		t.Fatalf("ctrl-u = %+v, value %q", result, d.entry.Value())
	}
}

func TestTextEntryDemoMarkdownSubmission(t *testing.T) {
	d := newTextEntryControls().(*textEntryControls)
	d.entry.SetValue("a*b")
	d.entry.SetCursorOffset(0)
	d.ConsumeKey(loom.KeyEvent{Key: "ctrl-a"})
	d.ConsumeKey(loom.KeyEvent{Key: "ctrl-b"})
	if d.entry.Value() != "a*b" || d.entry.MarkdownValue() != "**a\\*b**" {
		t.Fatalf("pre-submit entry value=%q markdown=%q", d.entry.Value(), d.entry.MarkdownValue())
	}
	if result := d.ConsumeKey(loom.KeyEvent{Key: "enter"}); !result.Consumed || result.Quit {
		t.Fatalf("submit result = %+v", result)
	}
	if d.lastSubmittedPlain != "a*b" || d.lastSubmittedMD != "**a\\*b**" || d.entry.Value() != "" {
		t.Fatalf("submission plain=%q markdown=%q remaining=%q", d.lastSubmittedPlain, d.lastSubmittedMD, d.entry.Value())
	}
	screen := strings.Join(loom.Render(d, 80, 8), "\n")
	if !strings.Contains(screen, "Submitted Markdown: **a\\*b**") || !strings.Contains(screen, "Ctrl+B: bold") {
		t.Fatalf("gallery does not show serialized submit and bold control: %q", screen)
	}
}

func TestTextEntryDemoPlaceholderKeys(t *testing.T) {
	for _, key := range []string{"tab", "end", "ctrl-e"} {
		d := newTextEntryControls().(*textEntryControls)
		d.ConsumeKey(loom.KeyEvent{Key: "f2"})
		if result := d.ConsumeKey(loom.KeyEvent{Key: key}); !result.Consumed || d.entry.Value() != "Write an update…" {
			t.Errorf("%s: result=%+v value=%q", key, result, d.entry.Value())
		}
	}
}
