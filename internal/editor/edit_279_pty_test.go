package editor

import (
	"os"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func TestEditPTYF10QuitSaveDiscardCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY, media or subprocess integration test")
	}

	// 1) Test Cancel
	s, path := startEditPTY(t, "base\n")
	s.Send("hello")
	s.WaitFor("hello", 3*time.Second)
	s.Send("\x1b[21~") // F10
	s.WaitFor("Save changes?", 3*time.Second)

	s.Send("\x1b[C") // right to Discard
	s.Send("\x1b[C") // right to Cancel
	s.Send("\r")     // press Enter

	// Cancel must close the dialog and keep the editor with its edit.
	s.WaitForState(3*time.Second*ptytest.RaceFactor, func(rows []string) bool {
		text := strings.Join(rows, "\n")
		return !strings.Contains(text, "Save changes?") && strings.Contains(text, "hello")
	})

	// 2) Test Discard
	s.Send("\x1b[21~") // F10
	s.WaitFor("Save changes?", 3*time.Second)

	s.Send("\x1b[C") // right to Discard
	s.Send("\r")     // press Enter

	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("Discard did not close editor cleanly: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "base\n" {
		t.Fatalf("Discard changed file content: %q", data)
	}

	// 3) Test Save
	s2, path2 := startEditPTY(t, "base\n")
	s2.Send("hello")
	s2.WaitFor("hello", 3*time.Second)
	s2.Send("\x1b[21~") // F10
	s2.WaitFor("Save changes?", 3*time.Second)

	s2.Send("\r") // press Enter on Save (default selection)

	if err := s2.Wait(3 * time.Second); err != nil {
		t.Fatalf("Save did not close editor cleanly: %v", err)
	}
	if data, _ := os.ReadFile(path2); string(data) != "hellobase\n" {
		t.Fatalf("Save did not write expected file content: %q", data)
	}
}
