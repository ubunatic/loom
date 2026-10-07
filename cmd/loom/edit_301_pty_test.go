package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

func clickEditPTY(t *testing.T, s *ptytest.Session, row int, label string) {
	t.Helper()
	screen := s.Screen()
	if row < 0 || row >= len(screen) || !strings.Contains(screen[row], label) {
		found := false
		for r := len(screen) - 1; r >= 0; r-- {
			if strings.Contains(screen[r], label) {
				row = r
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing click target %q in screen:\n%s", label, strings.Join(screen, "\n"))
		}
	}
	cells := []rune(screen[row])
	needle := []rune(label)
	for x := 0; x+len(needle) <= len(cells); x++ {
		if string(cells[x:x+len(needle)]) == label {
			s.Send(fmt.Sprintf("\x1b[<0;%d;%dM\x1b[<0;%d;%dm", x+1, row+1, x+1, row+1))
			return
		}
	}
	t.Fatalf("missing click target %q: %q", label, screen[row])
}

func waitEditPTY(t *testing.T, s *ptytest.Session, predicate func() bool) {
	t.Helper()
	until := time.Now().Add(3 * time.Second)
	for time.Now().Before(until) {
		if predicate() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("screen state did not arrive:\n%s", strings.Join(s.Screen(), "\n"))
}

func TestEditPTYBottomBarActionsAndToggles(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, path := startMouseEditPTY(t, "base\n")
	s.WaitFor("^O", 3*time.Second)
	// Browser's separator at column 25 proves both bindings toggle both ways.
	for _, key := range []string{"\x0f", "\x1bOQ"} {
		s.Send(key)
		waitEditPTY(t, s, func() bool { return s.Cell(25, 3).Rune == '│' })
		s.Send(key)
		waitEditPTY(t, s, func() bool { return s.Cell(25, 3).Rune != '│' })
	}
	clickEditPTY(t, s, 22, "Files")
	waitEditPTY(t, s, func() bool { return s.Cell(25, 3).Rune == '│' })
	clickEditPTY(t, s, 22, "^O")
	waitEditPTY(t, s, func() bool { return s.Cell(25, 3).Rune != '│' })
	clickEditPTY(t, s, 22, "Search")
	s.WaitFor("Regex", 3*time.Second)
	s.Send("\x06")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "Regex") })
	s.Send("hello")
	s.WaitFor("· Modified", 3*time.Second)
	clickEditPTY(t, s, 22, "Save")
	s.WaitFor("· Saved", 3*time.Second)
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "hello") {
		t.Fatal("click did not save")
	}
	clickEditPTY(t, s, 22, "Help")
	s.WaitFor("RichTextEdit Help", 3*time.Second)
	s.Send("\x1b")
	waitEditPTY(t, s, func() bool { return !strings.Contains(strings.Join(s.Screen(), "\n"), "RichTextEdit Help") })
	icons := loom.SpeccedDefaults.Editor.StatusIcons
	statusRow := len(s.Screen()) - 3
	before := s.Screen()[statusRow]
	clickEditPTY(t, s, statusRow, " "+icons.Theme+" ")
	waitEditPTY(t, s, func() bool { return s.Screen()[len(s.Screen())-3] != before })
	clickEditPTY(t, s, len(s.Screen())-3, " "+icons.AltOn+" ")
	waitEditPTY(t, s, func() bool { return strings.Contains(strings.Join(s.Screen(), "\n"), " "+icons.AltOff+" ") })
	waitEditPTY(t, s, func() bool { return strings.Contains(string(s.Raw()), "\x1b[?1049l") })
	clickEditPTY(t, s, len(s.Screen())-3, " "+icons.AltOff+" ")
	waitEditPTY(t, s, func() bool { return strings.Contains(strings.Join(s.Screen(), "\n"), " "+icons.AltOn+" ") })
	clickEditPTY(t, s, len(s.Screen())-2, "Quit")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestEditPTYMouseStatusToggle(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, _ := startMouseEditPTY(t, "base\n")
	icons := loom.SpeccedDefaults.Editor.StatusIcons
	clickEditPTY(t, s, len(s.Screen())-3, " "+icons.MouseOn+" ")
	waitEditPTY(t, s, func() bool { return strings.Contains(strings.Join(s.Screen(), "\n"), " "+icons.MouseOff+" ") })
	waitEditPTY(t, s, func() bool { return strings.Contains(string(s.Raw()), "\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l") })
	s.Send("\x11")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestEditPTYDrawKeyAndBoxSubmenu(t *testing.T) {
	s, _ := startMouseEditPTY(t, "base\n")
	s.Send("\x04\x1b[C\x1b[B\x04")
	s.WaitFor("─┐", 3*time.Second)
	s.Send("\x1a")
	s.WaitFor("· Saved", 3*time.Second)
	s.Send("\x00") // Ctrl-Space selects the word and opens its toolbar.
	s.WaitFor("#FG", 3*time.Second)
	for row, text := range s.Screen() {
		if strings.Contains(text, "#FG") {
			if strings.Contains(text, "Draw") {
				t.Fatal("Draw is top-level")
			}
			clickEditPTY(t, s, row, "Box")
			break
		}
	}
	s.WaitFor("Rounded", 3*time.Second)
	s.WaitFor("Draw", 3*time.Second)
}
