package editor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func TestEditPTYMouseGrabFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte("base\n"), 0600); err != nil {
		t.Fatal(err)
	}

	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit", "--mousegrab", path)
	s.WaitFor("Ln 1, Col 1", 3*time.Second*ptytest.RaceFactor)

	if mode := s.VT().MouseMode; mode != 1002 && mode != 1003 {
		t.Fatalf("expected mouse grab to be enabled (1002 or 1003), got %v", mode)
	}

	s.Send("\x11") // Ctrl+Q

	if err := s.Wait(3 * time.Second * ptytest.RaceFactor); err != nil {
		t.Fatalf("child exit: %v", err)
	}

	if mode := s.VT().MouseMode; mode != 0 {
		t.Fatalf("expected mouse grab to be disabled on exit (0), got %v", mode)
	}
}

func TestEditPTYMouseGrabDisabledFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}

	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte("base\n"), 0600); err != nil {
		t.Fatal(err)
	}

	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit", "--mousegrab=false", path)
	s.WaitFor("Ln 1, Col 1", 3*time.Second*ptytest.RaceFactor)

	if mode := s.VT().MouseMode; mode != 0 {
		t.Fatalf("expected mouse grab to be disabled (0), got %v", mode)
	}

	s.Send("\x11") // Ctrl+Q

	if err := s.Wait(3 * time.Second * ptytest.RaceFactor); err != nil {
		t.Fatalf("child exit: %v", err)
	}

	if mode := s.VT().MouseMode; mode != 0 {
		t.Fatalf("expected mouse grab to be disabled on exit (0), got %v", mode)
	}
}
