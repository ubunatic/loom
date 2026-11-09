package gallery

import (
	"strconv"
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
	if !strings.Contains(screen, `Submitted Markdown: "**a\\*b**"`) || !strings.Contains(screen, "Ctrl+B: bold") {
		t.Fatalf("gallery does not show serialized submit and bold control: %q", screen)
	}
}

func TestTextEntryDemoMarkdownPreviewEscapesWrapsAndBounds(t *testing.T) {
	d := newTextEntryControls().(*textEntryControls)
	d.entry.SetValue("1\n2\n3")
	if got, want := d.entry.MarkdownValue(), "1\n2\n3"; got != want {
		t.Fatalf("pre-submit markdown = %q, want raw multiline %q", got, want)
	}
	d.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if d.lastSubmittedPlain != "1\n2\n3" || d.lastSubmittedMD != "1\n2\n3" {
		t.Fatalf("submitted plain=%q markdown=%q", d.lastSubmittedPlain, d.lastSubmittedMD)
	}
	screen := strings.Join(loom.Render(d, 80, 8), "\n")
	if !strings.Contains(screen, `Submitted Markdown: "1\n2\n3"`) {
		t.Fatalf("quoted multiline preview missing: %q", screen)
	}
	if strings.Contains(screen, "Submitted Markdown: 123") {
		t.Fatalf("preview flattened submitted newlines: %q", screen)
	}

	d.lastSubmittedMD = "界\\\"\t\r\x01 **bold**"
	wantQuoted := "Submitted Markdown: " + strconv.Quote(d.lastSubmittedMD)
	if screen := strings.Join(loom.Render(d, 80, 8), "\n"); !strings.Contains(screen, wantQuoted) {
		t.Fatalf("quoted Unicode/control/quote/backslash/Markdown preview missing %q: %q", wantQuoted, screen)
	}
	d.lastSubmittedMD += " " + strings.Repeat("wide界 ", 30)
	c := loom.NewCanvas(30, 10)
	r := loom.Rect{X: 4, Y: 2, W: 12, H: 6}
	d.Draw(c, r)
	var boundedRows []string
	for y := 0; y < 10; y++ {
		for x := 0; x < 30; x++ {
			cell := c.Get(x, y)
			if (x < r.X || x >= r.X+r.W || y < r.Y || y >= r.Y+r.H) && cell.Text != " " {
				t.Fatalf("paint outside child at (%d,%d): %+v", x, y, cell)
			}
			if x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H && loom.StringWidth(cell.Text) > r.X+r.W-x {
				t.Fatalf("cell crosses child right edge at (%d,%d): %+v", x, y, cell)
			}
		}
		var row strings.Builder
		for x := r.X; x < r.X+r.W; x++ {
			row.WriteString(c.Get(x, y).Text)
		}
		boundedRows = append(boundedRows, row.String())
	}
	bounded := strings.Join(boundedRows, "\n")
	if !strings.Contains(bounded, "Submitted") || !strings.Contains(bounded, "[overflow]") {
		t.Fatalf("narrow preview lacks explicit overflow: %q", bounded)
	}

	tiny := loom.NewCanvas(5, 2)
	d.Draw(tiny, loom.Rect{X: 0, Y: 0, W: 5, H: 2})
	if cell := tiny.Get(0, 1); cell.Text != "C" {
		t.Fatalf("tiny layout lost controls row: %+v", cell)
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

func TestTextEntryDemoSampleKeysAndDebugView(t *testing.T) {
	d := newTextEntryControls().(*textEntryControls)
	d.entry.SetValue("")
	for _, key := range []string{"f5", "f6", "f7", "f8"} {
		before := d.entry.Value()
		if !d.ConsumeKey(loom.KeyEvent{Key: key}).Consumed || d.entry.Value() != before+textEntrySamples[key] {
			t.Fatalf("%s did not insert its sample at the caret: %q", key, d.entry.Value())
		}
	}
	c := loom.NewCanvas(80, 10)
	d.Draw(c, loom.Rect{W: 80, H: 10})
	var screen string
	for y := range c.Rows() {
		screen += c.Row(y) + "\n"
	}
	if !strings.Contains(screen, "debug: ") || !strings.Contains(screen, "F8 wide characters") {
		t.Fatalf("debug view or sample keys missing:\n%s", screen)
	}
}
