package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func Test343EditPTYMouseGrabIndicatorAndCtrlGRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY integration test")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path := filepath.Join(t.TempDir(), "doc.txt")
	if err := os.WriteFile(path, []byte("base\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s := ptytest.Start(t, 100, 24, buildLoomBinary(t), "edit", "--mousegrab=false", path)
	s.WaitFor("Ln 1, Col 1", 3*time.Second)
	waitEditPTY(t, s, func() bool { return hasStatusIcon343(s.Screen()[21], "m") })

	from := len(s.Raw())
	s.Send("\x07") // Ctrl+G must work while the terminal has mouse reporting disabled.
	waitMouseMode303(t, s, from, "\x1b[?1003h\x1b[?1006h")
	waitEditPTY(t, s, func() bool { return hasStatusIcon343(s.Screen()[21], "M") })

	s.Send("\x11")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatal(err)
	}
}
