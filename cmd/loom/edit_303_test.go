package main

import (
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

const mouseRelease303 = "\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l"

func waitMouseMode303(t *testing.T, s *ptytest.Session, from int, sequence string) {
	t.Helper()
	waitEditPTY(t, s, func() bool {
		return strings.Contains(string(s.Raw()[from:]), sequence)
	})
}

func Test303EditPTYTemporaryMouseGrab(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit")
	s.WaitFor("Ln 1, Col 1", 3*time.Second)
	for _, closeWith := range []string{"\x1b", "\x1b[<0;1;24M\x1b[<0;1;24m"} {
		from := len(s.Raw())
		s.Send("\x1bOP")
		s.WaitFor("RichTextEdit Help", 3*time.Second)
		waitMouseMode303(t, s, from, "\x1b[?1000h\x1b[?1006h")
		before := strings.Join(s.Screen(), "\n")
		s.Send("\x1b[<65;50;8M")
		waitEditPTY(t, s, func() bool { return strings.Join(s.Screen(), "\n") != before })
		from = len(s.Raw())
		s.Send(closeWith)
		waitMouseMode303(t, s, from, mouseRelease303)
		waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "RichTextEdit Help") })
	}
	from := len(s.Raw())
	s.Send("word\x1b[D\x00") // Select the word from inside it with Ctrl-Space.
	waitMouseMode303(t, s, from, "\x1b[?1000h\x1b[?1006h")
	s.WaitFor("[ B", 3*time.Second)
	clickEditPTY(t, s, 3, " B")
	waitEditPTY(t, s, func() bool { return s.Cell(1, 1).Style.Bold })
	s.WaitFor("Untitled · Modified", 3*time.Second)
	from = len(s.Raw())
	s.Send("\x1b[<0;1;24M\x1b[<0;1;24m")
	waitMouseMode303(t, s, from, mouseRelease303)
	// The same backdrop click must not activate the bottom bar or insert a report.
	from = len(s.Raw())
	s.Send("\x13")
	s.WaitFor("Name:", 3*time.Second)
	waitMouseMode303(t, s, from, "\x1b[?1000h\x1b[?1006h")
	from = len(s.Raw())
	s.Send("\x1b[<0;1;24M")
	waitMouseMode303(t, s, from, mouseRelease303)
	for _, button := range []bool{false, true} {
		from = len(s.Raw())
		s.Send("\x11")
		s.WaitFor("Save changes?", 3*time.Second)
		waitMouseMode303(t, s, from, "\x1b[?1000h\x1b[?1006h")
		from = len(s.Raw())
		if button {
			clickEditPTY(t, s, 12, "Cancel")
		} else {
			s.Send("\x1b[<0;1;24M")
		}
		waitMouseMode303(t, s, from, mouseRelease303)
	}
	s.Send("\x11")
	s.WaitFor("Save changes?", 3*time.Second)
	clickEditPTY(t, s, 12, "Discard")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func Test303EditPTYGlobalMouseGrabPersists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startMouseEditPTY(t, "base\n")
	from := len(s.Raw())
	s.Send("\x1bOP")
	s.WaitFor("RichTextEdit Help", 3*time.Second)
	s.Send("\x1b[<0;1;24M")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "RichTextEdit Help") })
	if raw := string(s.Raw()[from:]); strings.Contains(raw, "\x1b[?1006l") || strings.Contains(raw, "\x1b[?1003l") {
		t.Fatalf("closing help disabled global tracking: %q", raw)
	}
	clickEditPTY(t, s, 22, "Help")
	s.WaitFor("RichTextEdit Help", 3*time.Second)
	s.Send("\x1b")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "RichTextEdit Help") })
	s.Send("\x11")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func Test303EditPopoverBackdropBlocksHostAction(t *testing.T) {
	v, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "julia256"})
	if err != nil {
		t.Fatal(err)
	}
	c := loom.NewCanvas(100, 24)
	v.Draw(c, c.Bounds())
	v.ConsumeKey(loom.KeyEvent{Key: "ctrl-space"})
	v.Draw(c, c.Bounds())
	if !v.edit.ModalOpen() {
		t.Fatal("popover did not open")
	}
	v.ConsumeMouse(loom.MouseEvent{X: v.indicatorRect.X + 2, Y: v.indicatorRect.Y, Action: loom.MousePress, Button: loom.MouseLeft})
	if v.edit.ModalOpen() || v.config.Theme != "julia256" || v.config.MouseGrab {
		t.Fatal("popover backdrop activated a host control")
	}
}
