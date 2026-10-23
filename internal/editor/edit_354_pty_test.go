package editor

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestEditPTYFirstCtrlSAfterLoneEscape(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	s, path := startEditPTY(t, "base\n")
	s.Send("q")
	s.WaitFor("· Modified", 3*time.Second)
	s.Send("\x1b")
	time.Sleep(5 * time.Millisecond) // separate tty reads while still inside the Escape window
	s.Send("\x13")
	s.WaitForUpdate("· Saved", 3*time.Second)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "q") {
		t.Fatalf("first Ctrl+S after lone Escape did not save: %q", data)
	}
}

func TestEditPTYFirstCtrlSAfterEscapeClosesCtrlSpaceBar(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	s, path := startEditPTY(t, "base\n")
	s.Send("q")
	s.WaitFor("· Modified", 3*time.Second)
	s.Send("\x00") // Ctrl+Space opens the editor action bar.
	s.Send("\x1b")
	time.Sleep(5 * time.Millisecond)
	s.Send("\x13")
	s.WaitForUpdate("· Saved", 3*time.Second)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "q") {
		t.Fatalf("first Ctrl+S after Escape closed Ctrl+Space bar did not save: %q", data)
	}
}
