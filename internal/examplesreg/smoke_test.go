package examplesreg_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/examplesreg"
	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func buildExample(t *testing.T, name string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), name)
	path := "codeberg.org/ubunatic/loom/examples/" + name
	if out, err := exec.Command("go", "build", "-o", bin, path).CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", name, err, out)
	}
	return bin
}

func waitForInitialScreen(t *testing.T, s *ptytest.Session) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, row := range s.Screen() {
			if strings.TrimSpace(row) != "" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("initial screen stayed blank")
}

func assertRowsFit(t *testing.T, s *ptytest.Session, phase string) {
	t.Helper()
	for row, line := range s.Screen() {
		if width := len([]rune(line)); width > 80 {
			t.Fatalf("%s row %d has width %d (>80): %q", phase, row, width, line)
		}
	}
}

// TestRegisteredExamplesPTYSmoke exercises the same minimal terminal contract
// for every registered example: draw at 80x24, accept input, and quit cleanly.
func TestRegisteredExamplesPTYSmoke(t *testing.T) {
	for _, example := range examplesreg.Registry {
		example := example
		t.Run(example.Name, func(t *testing.T) {
			bin := buildExample(t, example.Name)
			s := ptytest.Start(t, 80, 24, bin, example.DemoArgs...)
			waitForInitialScreen(t, s)
			assertRowsFit(t, s, "initial screen")

			// Exercise navigation before the standard q quit sequence.
			s.Send("\x1b[B")
			time.Sleep(50 * time.Millisecond)
			assertRowsFit(t, s, "interactive screen")
			quit := "q"
			if example.Name == "filebrowser" {
				quit = "\x11" // Ctrl-Q; q starts the file filter in this app.
			}
			s.Send(quit)
			if err := s.Wait(5 * time.Second); err != nil {
				t.Fatalf("clean exit: %v\nscreen:\n%s", err, strings.Join(s.Screen(), "\n"))
			}
		})
	}
}
