package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom"
	"ubunatic.com/loom/internal/ptytest"
)

func startEditPTY(t *testing.T, content string) (*ptytest.Session, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit", path)
	s.WaitFor("Ln 1, Col 1", 3*time.Second)
	return s, path
}

func TestEditPTYHeaderHasNoDuplicateFileLabel(t *testing.T) {
	s, _ := startEditPTY(t, "base\n")
	s.WaitFor("doc.txt · Saved", 3*time.Second)
	screen := s.Screen()
	files, names := 0, 0
	for _, row := range screen {
		files += strings.Count(row, " File ")
		names += strings.Count(row, "doc.txt ·")
	}
	if files != 1 || names != 1 || strings.Contains(screen[1], "theme:") || strings.Contains(screen[1], "File") || !strings.Contains(screen[20], "File") {
		t.Fatalf("File label x%d, file name x%d, want 1 each and file bar above bottom status:\n%s", files, names, strings.Join(screen, "\n"))
	}
	icons := loom.SpeccedDefaults.Editor.StatusIcons
	for _, icon := range []string{icons.Theme, icons.MouseOff, icons.AltOn} {
		if !strings.Contains(screen[21], " "+icon+" ") {
			t.Fatalf("bottom status missing %q: %q", icon, screen[21])
		}
	}
}

func TestEditPTYEscKeepsRunningAndQuitKeysAskFirst(t *testing.T) {
	s, path := startEditPTY(t, "base\n")
	s.Send("hello")
	s.WaitFor("hello", 3*time.Second)

	s.Send("\x1b")
	time.Sleep(300 * time.Millisecond)
	if text := strings.Join(s.Screen(), "\n"); !strings.Contains(text, "hello") || strings.Contains(text, "Save changes?") {
		t.Fatalf("Esc changed the editor state:\n%s", text)
	}

	for _, key := range []string{"\x11", "\x1b[21~"} { // ^Q, F10
		s.Send(key)
		s.WaitFor("Save changes?", 3*time.Second)
		s.Send("\x1b") // Esc closes only the dialog
		time.Sleep(300 * time.Millisecond)
		text := strings.Join(s.Screen(), "\n")
		if strings.Contains(text, "Save changes?") || !strings.Contains(text, "hello") {
			t.Fatalf("Esc on dialog (%q) left screen:\n%s", key, text)
		}
	}

	s.Send("\x11")
	s.WaitFor("Save changes?", 3*time.Second)
	s.Send("\x1b[C") // Discard
	s.Send("\r")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("loom edit exit: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "base\n" {
		t.Fatalf("Discard changed file: %q", data)
	}
}

func TestEditPTYSaveThenQuit(t *testing.T) {
	s, path := startEditPTY(t, "base\n")
	s.Send("hello")
	s.WaitFor("hello", 3*time.Second)
	s.Send("\x1b[21~")
	s.WaitFor("Save changes?", 3*time.Second)
	s.Send("\r") // Save is the default button
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("loom edit exit: %v", err)
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "hello") {
		t.Fatalf("Save did not write: %q", data)
	}
}

func TestEditPTYSpanDeletionSaveAndQuit(t *testing.T) {
	for _, dialog := range []bool{false, true} {
		t.Run(strconv.FormatBool(dialog), func(t *testing.T) {
			s, path := startEditPTY(t, "base\n")
			s.Send("\x1b[F\r") // End, then a new logical line.
			s.WaitFor("Ln 2, Col 1", 3*time.Second)
			s.Send("xy\x7f\x1b[H\x1b[3~") // type, Backspace, Home, Delete to empty spans.
			s.WaitFor("Ln 2, Col 1", 3*time.Second)
			s.WaitFor("· Modified", 3*time.Second)
			if dialog {
				s.Send("\x11")
				s.WaitFor("Save changes?", 3*time.Second)
				s.Send("\r")
			} else {
				s.Send("\x13")
				s.WaitFor("· Saved", 3*time.Second)
				s.Send("\x11")
			}
			if err := s.Wait(3 * time.Second); err != nil {
				t.Fatalf("Save did not allow quit: %v", err)
			}
			doc := &loom.RichDocument{}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			doc.FromANSI(string(data))
			if got := doc.ToPlainText(); got != "base\n" {
				t.Fatalf("saved text = %q, want trailing empty logical line", got)
			}
		})
	}
}

func TestEditPTYRightBorderSurvivesResizeSettling(t *testing.T) {
	s, _ := startEditPTY(t, "base\n")
	for _, size := range [][2]int{{60, 18}, {80, 24}, {40, 16}} {
		s.Resize(size[0], size[1])
		// Beyond the spec guard duration, the canvas uses the full width.
		time.Sleep(loom.SpeccedDefaults.Pane.GuardDuration + 300*time.Millisecond)
		screen := s.Screen()
		for y := 0; y < size[1]; y++ {
			want := '│'
			if y == 0 {
				want = '┐'
			} else if y == size[1]-1 {
				want = '┘'
			}
			if got := s.Cell(size[0]-1, y).Rune; got != want {
				t.Fatalf("settled %dx%d row %d right border = %q, want %q:\n%s", size[0], size[1], y, got, want, strings.Join(screen, "\n"))
			}
		}
	}
}

func TestEditPTYCleanQuitKeysExit(t *testing.T) {
	for _, key := range []string{"\x11", "\x1b[21~"} {
		s, _ := startEditPTY(t, "base\n")
		s.Send(key)
		if err := s.Wait(3 * time.Second); err != nil {
			t.Fatalf("key %q: loom edit exit: %v", key, err)
		}
	}
}

func TestEditPTYCtrlSFlipsHeaderToSaved(t *testing.T) {
	s, path := startEditPTY(t, "base\n")
	s.WaitFor("· Saved", 3*time.Second)
	s.Send("hello")
	s.WaitFor("· Modified", 3*time.Second)
	s.Send("\x13") // ^S
	s.WaitFor("· Saved", 3*time.Second)
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "hello") {
		t.Fatalf("^S did not write: %q", data)
	}

	// Modify again, then save with the file browser focused.
	s.Send("!")
	s.WaitFor("· Modified", 3*time.Second)
	s.Send("\x1bOQ") // F2 opens the browser and focuses it
	time.Sleep(300 * time.Millisecond)
	s.Send("\x13")
	s.WaitFor("· Saved", 3*time.Second)
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "!") {
		t.Fatalf("^S with browser focused did not write: %q", data)
	}
}

func TestEditPTYBrowserSingleClickBeyondTextSelectsRow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	for name, body := range map[string]string{"doc.txt": "base\n", "b.txt": "second file\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit", "--mousegrab", path)
	s.WaitFor("Ln 1, Col 1", 3*time.Second)
	s.Send("\x1bOQ") // F2 opens the browser
	s.WaitFor("b.txt", 3*time.Second)

	row, col := -1, -1
	for y, line := range s.Screen() {
		if i := strings.Index(line, "b.txt"); i >= 0 {
			row, col = y, len([]rune(line[:i]))+len("b.txt")+3
			break
		}
	}
	if row < 0 {
		t.Fatalf("b.txt not on screen:\n%s", strings.Join(s.Screen(), "\n"))
	}
	// SGR mouse press and release, 1-based, a few cells right of the name.
	s.Send("\x1b[<0;" + strconv.Itoa(col+1) + ";" + strconv.Itoa(row+1) + "M")
	s.Send("\x1b[<0;" + strconv.Itoa(col+1) + ";" + strconv.Itoa(row+1) + "m")
	time.Sleep(300 * time.Millisecond)
	s.Send("\r")
	s.WaitFor("second file", 3*time.Second)
}

func TestEditPTYCtrlAltSOpensSaveAs(t *testing.T) {
	for name, seq := range map[string]string{
		"legacy ESC ^S": "\x1b\x13",
		"CSI-u":         "\x1b[115;7u",
	} {
		t.Run(name, func(t *testing.T) {
			s, _ := startEditPTY(t, "base\n")
			s.Send(seq) // ⌃⌥S
			s.WaitFor("Name:", 3*time.Second)
			s.Send("\x1b")
		})
	}
}

func TestEditPTYCtrlOFocusesBrowserAndCtrlWDiscardsToUntitled(t *testing.T) {
	s, path := startEditPTY(t, "base\n")
	s.Send("\x0f") // ^O
	s.WaitFor("doc.txt", 3*time.Second)
	s.Send("\x0f") // again: stays open, does not toggle off
	time.Sleep(300 * time.Millisecond)
	s.WaitFor("│", 3*time.Second)

	s.Send("\x17") // ^W with the browser focused, clean buffer: resets at once
	s.WaitFor("· Unsaved", 3*time.Second)

	s.Send("\x1b") // leave the browser (cancel hides it)
	time.Sleep(300 * time.Millisecond)
	s.Send("hi")
	s.WaitFor("· Modified", 3*time.Second)
	s.Send("\x17")
	s.WaitFor("Save changes?", 3*time.Second)
	s.Send("\x1b[C") // Discard
	s.Send("\r")
	s.WaitFor("· Unsaved", 3*time.Second)
	if data, _ := os.ReadFile(path); string(data) != "base\n" {
		t.Fatalf("Discard changed the file: %q", data)
	}
}
