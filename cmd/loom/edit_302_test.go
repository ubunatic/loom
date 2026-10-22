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
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit", "--mousegrab=false")
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
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startMouseEditPTY(t, "base\n")
	timeout := 3 * time.Second * ptytest.RaceFactor
	waitState := func(predicate func([]string) bool) {
		t.Helper()
		s.WaitForState(timeout, predicate)
	}
	waitUpdate := func(predicate func([]string) bool) {
		t.Helper()
		s.WaitForUpdateState(timeout, predicate)
	}
	helpReady := func(rows []string) bool {
		text := strings.Join(rows, "\n")
		return strings.Contains(text, "RichTextEdit Help") && strings.Contains(text, loom.SpeccedDefaults.RichTextEdit.HelpFooter)
	}
	inlineReady := func(rows []string) bool {
		return len(rows) == 24 && strings.Contains(rows[21], " a ") && strings.ContainsRune(rows[23], '┘')
	}
	waitState(edit302BackdropReady)
	s.Send("\x1bOP")
	waitUpdate(helpReady)
	waitState(helpReady)
	before := strings.Join(s.Screen(), "\n")
	scrolledHelpReady := func(rows []string) bool {
		return helpReady(rows) && strings.Join(rows, "\n") != before
	}
	s.Send("\x1b[<65;50;8M")
	waitUpdate(scrolledHelpReady)
	if !strings.Contains(cellGridText(s.Cells()), "RichTextEdit Help") {
		t.Fatal("wheel dismissed help")
	}
	waitState(scrolledHelpReady)
	s.Send("\x1b[<0;1;24M") // Outside editor and popup: modal routing must win.
	waitUpdate(edit302BackdropReady)
	waitState(edit302BackdropReady)
	clickEditPTY(t, s, 21, " A ")
	waitUpdate(inlineReady)
	if !cellRowHasRune(s.Cells()[23], '┘') {
		t.Fatalf("inline app leaves a blank bottom row: %q", s.Screen()[23])
	}
	waitState(inlineReady)
	s.Resize(100, 18)
	resizedReady := func(rows []string) bool {
		return len(rows) == 18 && strings.Contains(rows[15], " a ") && strings.ContainsRune(rows[17], '┘')
	}
	waitUpdate(resizedReady)
	waitState(resizedReady)
	s.Send("\x11")
	if err := s.Wait(timeout); err != nil {
		t.Fatal(err)
	}
}

func edit302BackdropReady(rows []string) bool {
	// Dismissal can clear the help before the status bar is restored. Wait for
	// the next click target, rather than accepting that intermediate redraw.
	return len(rows) == 24 && strings.Contains(rows[21], " A ") &&
		!strings.Contains(strings.Join(rows, "\n"), "RichTextEdit Help")
}

func Test302EditBackdropReadinessDuringRedraw(t *testing.T) {
	vt := ptytest.NewVT(100, 24)
	for _, step := range []struct {
		name, output string
		ready        bool
	}{
		{"help open", "\x1b[3;10HRichTextEdit Help\x1b[22;90H A │", false},
		{"help cleared before status restored", "\x1b[2J\x1b[1;1Hbase", false},
		{"status restored", "\x1b[22;90H A │", true},
	} {
		if _, err := vt.Write([]byte(step.output)); err != nil {
			t.Fatal(err)
		}
		if got := edit302BackdropReady(vt.Screen()); got != step.ready {
			t.Errorf("%s: backdrop ready = %v, want %v", step.name, got, step.ready)
		}
	}
}

func cellRowHasRune(row []ptytest.Cell, want rune) bool {
	for _, cell := range row {
		if cell.Rune == want {
			return true
		}
	}
	return false
}

func cellGridText(grid [][]ptytest.Cell) string {
	rows := make([]string, len(grid))
	for y, row := range grid {
		var text strings.Builder
		for _, cell := range row {
			if cell.Text != "" {
				text.WriteString(cell.Text)
			} else if cell.Rune != 0 {
				text.WriteRune(cell.Rune)
			}
		}
		rows[y] = text.String()
	}
	return strings.Join(rows, "\n")
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
