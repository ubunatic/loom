package main

import (
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
)

func Test316EditCustomPopupKeysBeforePanelTab(t *testing.T) {
	e := loom.NewRichTextEdit(&loom.RichDocument{Lines: []loom.RichLine{{Spans: []loom.RichSpan{{Text: "hello world"}}}}})
	v, err := newEditView(e, "", loom.EditorConfig{Theme: "julia256"})
	if err != nil {
		t.Fatal(err)
	}
	v.Draw(loom.NewCanvas(100, 24), loom.Rect{W: 100, H: 24})
	e.SetSelection(loom.RichPosition{}, loom.RichPosition{Offset: 5})
	e.PopoverDefaults.AutoPopupOnSelection = false
	e.PopoverDefaults.PopupOpenKey = "f2"
	e.PopoverDefaults.MenuFocusKey = "f3"
	v.ConsumeKey(loom.KeyEvent{Key: "f2"})
	if !e.ModalOpen() {
		t.Fatal("custom open key not forwarded")
	}
	v.ConsumeKey(loom.KeyEvent{Key: "f3"})
	v.ConsumeKey(loom.KeyEvent{Key: "right"})
	v.ConsumeKey(loom.KeyEvent{Key: "enter"})
	if !e.Document.Lines[0].Spans[0].Style.Italic {
		t.Fatal("custom focus key not forwarded")
	}
}

func Test316EditPTYFocusDismissAndReopen(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startEditPTY(t, "hello world\n")
	s.Send("\x00")
	s.WaitFor("#FG", 3*time.Second)
	s.Send("\x1b[C")
	s.WaitFor("Ln 1, Col 6", 3*time.Second)
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "#FG") })
	s.Send("\x00")
	s.WaitFor("#FG", 3*time.Second)
	s.Send("\x1b[B\x1b[C\r")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "#FG") })
	s.Send("\x00")
	s.WaitFor("#FG", 3*time.Second)
	s.Send("\x1b")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "#FG") })
}
