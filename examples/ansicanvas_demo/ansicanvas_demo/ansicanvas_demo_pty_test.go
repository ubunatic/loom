// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package ansicanvas_demo_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ubunatic.com/loom/internal/ptytest"
)

func buildAnsiCanvasDemo(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ansicanvas_demo")
	out, err := exec.Command("go", "build", "-o", bin, "ubunatic.com/loom/examples/ansicanvas_demo").CombinedOutput()
	if err != nil {
		t.Fatalf("build ansicanvas_demo: %v\n%s", err, out)
	}
	return bin
}

func TestAnsiCanvasDemoPTYStartupAndQuit(t *testing.T) {
	bin := buildAnsiCanvasDemo(t)

	s := ptytest.Start(t, 80, 24, bin)
	s.WaitFor("ANSI Canvas Demo", 5*time.Second)

	screen := strings.Join(s.Screen(), "\n")
	if !strings.Contains(screen, "Graphic Cell Buffer") {
		t.Fatalf("expected canvas box rendered, got:\n%s", screen)
	}

	// Exit with F10 (\x1b[21~) or q
	s.Send("\x1b[21~")
	if err := s.Wait(3 * time.Second); err != nil {
		s.Send("q")
		if err2 := s.Wait(2 * time.Second); err2 != nil {
			t.Fatalf("ansicanvas_demo did not exit on quit command: %v / %v", err, err2)
		}
	}
}

func TestAnsiCanvasDemoPTYNavigationAndTyping(t *testing.T) {
	bin := buildAnsiCanvasDemo(t)

	s := ptytest.Start(t, 80, 24, bin)
	s.WaitFor("ANSI Canvas Demo", 5*time.Second)

	// Send navigation keys
	for i := 0; i < 4; i++ {
		s.Send("\x1b[A") // Up
		s.Send("\x1b[B") // Down
		s.Send("\x1b[C") // Right
		s.Send("\x1b[D") // Left
	}

	time.Sleep(150 * time.Millisecond)

	// Ensure still running
	screen := strings.Join(s.Screen(), "\n")
	if !strings.Contains(screen, "ANSI Canvas Demo") {
		t.Fatalf("expected demo app still active, got:\n%s", screen)
	}

	// Quit with F10
	s.Send("\x1b[21~")
	if err := s.Wait(3 * time.Second); err != nil {
		s.Send("q")
		if err2 := s.Wait(2 * time.Second); err2 != nil {
			t.Fatalf("ansicanvas_demo did not exit: %v / %v", err, err2)
		}
	}
}
