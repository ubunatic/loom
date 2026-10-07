package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	s.WaitFor("[Saved]", 3*time.Second)
	s.Send("hello")
	s.WaitFor("[Modified]", 3*time.Second)
	s.Send("\x13") // ^S
	s.WaitFor("[Saved]", 3*time.Second)
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "hello") {
		t.Fatalf("^S did not write: %q", data)
	}

	// Modify again, then save with the file browser focused.
	s.Send("!")
	s.WaitFor("[Modified]", 3*time.Second)
	s.Send("\x1bOQ") // F2 opens the browser and focuses it
	time.Sleep(300 * time.Millisecond)
	s.Send("\x13")
	s.WaitFor("[Saved]", 3*time.Second)
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), "!") {
		t.Fatalf("^S with browser focused did not write: %q", data)
	}
}
