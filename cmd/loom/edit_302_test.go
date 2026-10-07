package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

func startMouseEditPTY(t *testing.T, content string) (*ptytest.Session, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit", "--mousegrab", path)
	s.WaitFor("Ln 1, Col 1", 3*time.Second)
	return s, path
}

func Test302EditArguments(t *testing.T) {
	cmd := editCommand()
	for _, args := range [][]string{nil, {"doc.ansi"}} {
		if err := cmd.Args(cmd, args); err != nil {
			t.Fatal(err)
		}
	}
	if err := cmd.Args(cmd, []string{"one", "two"}); err == nil {
		t.Fatal("accepted two files")
	}
}

func Test302EditPTYUntitledAndMouseOff(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit")
	s.WaitFor("Untitled · Unsaved", 3*time.Second)
	s.WaitFor("Ln 1, Col 1", 3*time.Second)
	raw := string(s.Raw())
	if !strings.Contains(raw, "\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l") || strings.Contains(raw, "\x1b[?1000h") || strings.Contains(raw, "\x1b[?1003h") {
		t.Fatal("mouse-off startup enabled terminal reporting")
	}
	s.Send("hello")
	s.WaitFor("Untitled · Modified", 3*time.Second)
	s.Send("\x13")
	s.WaitFor("Name:", 3*time.Second)
	s.Send("\x1b")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "Name:") })
	s.Send("\x11")
	s.WaitFor("Save changes?", 3*time.Second)
	s.Send("\x1b[C\r")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func Test302EditPTYHelpWheelBackdropAndInline(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startMouseEditPTY(t, "base\n")
	s.Send("\x1bOP")
	s.WaitFor("RichTextEdit Help", 3*time.Second)
	before := strings.Join(s.Screen(), "\n")
	s.Send("\x1b[<65;50;8M")
	waitEditPTY(t, s, func() bool { return strings.Join(s.Screen(), "\n") != before })
	if !strings.Contains(strings.Join(s.Screen(), "\n"), "RichTextEdit Help") {
		t.Fatal("wheel dismissed help")
	}
	s.Send("\x1b[<0;1;24M") // Outside editor and popup: modal routing must win.
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "RichTextEdit Help") })
	clickEditPTY(t, s, 21, " A ")
	waitEditPTY(t, s, func() bool { return strings.Contains(strings.Join(s.Screen(), "\n"), " a ") })
	if row := s.Screen()[23]; !strings.Contains(row, "┘") {
		t.Fatalf("inline app leaves a blank bottom row: %q", row)
	}
	s.Resize(100, 18)
	waitEditPTY(t, s, func() bool { return len(s.Screen()) == 18 && strings.Contains(s.Screen()[17], "┘") })
	s.Send("\x11")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func Test302EditModalBlocksHostActions(t *testing.T) {
	v, err := newEditView(loom.NewRichTextEdit(nil), "", loom.EditorConfig{Theme: "julia256"})
	if err != nil {
		t.Fatal(err)
	}
	v.Draw(loom.NewCanvas(100, 24), loom.Rect{W: 100, H: 24})
	v.ConsumeKey(loom.KeyEvent{Key: "f1"})
	v.Draw(loom.NewCanvas(100, 24), loom.Rect{W: 100, H: 24})
	v.ConsumeKey(loom.KeyEvent{Key: "ctrl-o"})
	if v.showSidePanel {
		t.Fatal("host action bypassed modal")
	}
	v.ConsumeMouse(loom.MouseEvent{X: 0, Y: 23, Action: loom.MousePress, Button: loom.MouseLeft})
	if v.edit.ModalOpen() {
		t.Fatal("host backdrop did not close help")
	}
}
