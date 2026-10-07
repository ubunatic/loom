package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

func Test304EditBackdropDoesNotChangeEditorOrHost(t *testing.T) {
	for _, mouseGrab := range []bool{false, true} {
		e := loom.NewRichTextEdit(&loom.RichDocument{Lines: []loom.RichLine{
			{Spans: []loom.RichSpan{{Text: "first word"}}},
			{Spans: []loom.RichSpan{{Text: "second line"}}},
		}})
		v, err := newEditView(e, "", loom.EditorConfig{Theme: "julia256", MouseGrab: mouseGrab})
		if err != nil {
			t.Fatal(err)
		}
		c := loom.NewCanvas(100, 24)
		v.Draw(c, c.Bounds())
		v.ConsumeKey(loom.KeyEvent{Key: "ctrl-space"})
		v.Draw(c, c.Bounds())
		cursor, from, to := e.Cursor, e.SelectionFrom, e.SelectionTo
		document, config, focus := e.Document.ToANSI(), v.config, v.focused
		for _, action := range []loom.MouseAction{loom.MouseScrollDown, loom.MousePress, loom.MouseDrag, loom.MouseRelease} {
			got := v.ConsumeMouse(loom.MouseEvent{X: v.editorRect.X + 25, Y: v.editorRect.Y,
				Action: action, Button: loom.MouseLeft})
			if got != loom.Handled() {
				t.Fatalf("mousegrab=%v action=%v result=%+v", mouseGrab, action, got)
			}
			v.Draw(c, c.Bounds())
		}
		if e.ModalOpen() || e.Cursor != cursor || e.SelectionFrom != from || e.SelectionTo != to || !e.HasSelection ||
			e.Document.ToANSI() != document || e.IsModified() || v.config != config || v.focused != focus {
			t.Fatalf("mousegrab=%v backdrop changed editor or host", mouseGrab)
		}
	}
}

func Test304EditPTYPopoverBackdropIsolation(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	content := "first word\nsecond line\n"
	s, path := startMouseEditPTY(t, content)
	s.Send("\x1b[C\x1b[C\x00") // Move inside a word, then Ctrl-Space.
	s.WaitFor("#FG", 3*time.Second)
	cursorPattern := regexp.MustCompile(`Ln [0-9]+, Col [0-9]+`)
	cursor := cursorPattern.FindString(strings.Join(s.Screen(), "\n"))
	if cursor == "" {
		t.Fatal("missing cursor status")
	}
	selected := make([]ptytest.Style, len([]rune("first word")))
	for x := range selected {
		selected[x] = s.Cell(x+1, 1).Style
	}
	// Queue the full dismissal gesture together so Pane must isolate every
	// report even when it redraws and releases modal ownership between them.
	s.Send("\x1b[<65;26;2M\x1b[<0;26;2M\x1b[<32;10;3M\x1b[<0;10;3m")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "#FG") })
	// Saving gives an observable acknowledgement after all queued reports.
	s.Send("\x13\x1bOP")
	s.WaitFor("RichTextEdit Help", 3*time.Second)
	s.Send("\x1b")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "RichTextEdit Help") })
	if got := cursorPattern.FindString(strings.Join(s.Screen(), "\n")); got != cursor {
		t.Fatalf("backdrop moved cursor: %q -> %q", cursor, got)
	}
	for x, style := range selected {
		if got := s.Cell(x+1, 1).Style; got != style {
			t.Fatalf("backdrop changed selection cell %d: %+v -> %+v", x, style, got)
		}
	}
	// loom edit saves ANSI styling even for an input named doc.txt. Compare
	// its canonical serialization rather than expecting plain input bytes.
	var expected loom.RichDocument
	expected.FromANSI(content)
	if data, err := os.ReadFile(path); err != nil || string(data) != expected.ToANSI()+"\n" {
		t.Fatalf("backdrop changed saved document: %q, %v", data, err)
	}
	if raw := string(s.Raw()); !strings.Contains(raw, "\x1b[?1003h") {
		t.Fatal("regression did not run with mousegrab enabled")
	}
	s.Send("\x11")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}
