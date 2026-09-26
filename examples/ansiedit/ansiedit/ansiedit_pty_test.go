// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansiedit_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func buildAnsiEdit(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ansiedit")
	out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/ansiedit").CombinedOutput()
	if err != nil {
		t.Fatalf("build ansiedit: %v\n%s", err, out)
	}
	return bin
}

func TestAnsiEditPTYStartupAndQuit(t *testing.T) {
	file := filepath.Join(t.TempDir(), "test.ansi")
	bin := buildAnsiEdit(t)

	s := ptytest.Start(t, 80, 24, bin, file)
	s.WaitFor("Loom AnsiEdit", 5*time.Second)

	screen := strings.Join(s.Screen(), "\n")
	if !strings.Contains(screen, "Palette") && !strings.Contains(screen, "Info") {
		t.Fatalf("expected side panel rendered, got:\n%s", screen)
	}

	// Send F10 key sequence to quit: ESC [ 2 1 ~
	s.Send("\x1b[21~")
	if err := s.Wait(3 * time.Second); err != nil {
		// Fallback: send ctrl-q if F10 terminal sequence differs
		s.Send("\x11")
		if err2 := s.Wait(2 * time.Second); err2 != nil {
			t.Fatalf("ansiedit did not exit on quit command: %v / %v", err, err2)
		}
	}
}

func TestAnsiEditPTYTypingAndSaving(t *testing.T) {
	file := filepath.Join(t.TempDir(), "typed.ansi")
	bin := buildAnsiEdit(t)

	s := ptytest.Start(t, 80, 24, bin, file)
	s.WaitFor("Loom AnsiEdit", 5*time.Second)

	// Type characters
	s.Send("Hello ANSI")
	time.Sleep(100 * time.Millisecond)

	// Save with Ctrl-S (\x13)
	s.Send("\x13")
	time.Sleep(100 * time.Millisecond)

	// Quit with Ctrl-Q (\x11)
	s.Send("\x11")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("ansiedit did not exit: %v", err)
	}
}
