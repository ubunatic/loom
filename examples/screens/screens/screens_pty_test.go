// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package screens_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func build(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "screens")
	if out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/screens").CombinedOutput(); err != nil {
		t.Fatalf("build screens: %v\n%s", err, out)
	}
	return bin
}

func quit(t *testing.T, s *ptytest.Session) {
	t.Helper()
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("exit: %v", err)
	}
}

// TestScreensSwitchInlineFullAlt walks inline -> full -> alt -> inline and
// checks each layout's footprint and that the alternate screen is left again.
func TestScreensSwitchInlineFullAlt(t *testing.T) {
	s := ptytest.Start(t, 100, 30, build(t), "-auto=false")
	s.WaitFor("Screen: inline", 5*time.Second)

	s.Send("n")
	s.WaitFor("Screen: primary full screen", 3*time.Second)
	s.Send("f")
	s.WaitFor("Screen: full screen (alt)", 3*time.Second)
	if !strings.Contains(string(s.Raw()), "\x1b[?1049h") {
		t.Fatal("alternate screen not entered")
	}
	s.Send("f")
	s.WaitFor("Screen: inline", 3*time.Second)
	if !strings.Contains(strings.Join(s.Screen(), "\n"), "┌") {
		t.Fatal("no border drawn")
	}
	if strings.Count(string(s.Raw()), "\x1b[?1049l") < 1 {
		t.Fatal("alternate screen not left")
	}
	quit(t, s)
}

// TestScreensAutoFullscreen grows the inline height to full height and
// expects the pane to promote itself to the alternate screen, and to fall
// back to inline when the terminal grows again.
func TestScreensAutoFullscreen(t *testing.T) {
	s := ptytest.Start(t, 100, 30, build(t), "-height=27")
	s.WaitFor("Screen: inline", 5*time.Second)
	s.Send("+") // 28 of 30 rows: still 2 short, margin_rows=1
	s.WaitFor("wanted height 28", 3*time.Second)
	if strings.Contains(strings.Join(s.Screen(), "\n"), "Screen: full screen (alt)") {
		t.Fatal("promoted one row too early")
	}
	s.Send("+") // 29 of 30: within margin
	s.WaitFor("Screen: full screen (alt) (auto)", 3*time.Second)

	s.Resize(100, 40) // 29 of 40 rows is no longer full height
	s.WaitFor("Screen: inline", 3*time.Second)
	quit(t, s)
}

// TestScreensAutoParamsChangeDetection tightens the margin to 0: 29 of 30
// rows is no longer full, 30 is.
func TestScreensAutoParamsChangeDetection(t *testing.T) {
	s := ptytest.Start(t, 100, 30, build(t), "-height=29", "-margin=0", "-auto-alt=false")
	s.WaitFor("Screen: inline", 5*time.Second)
	s.Send("+")
	s.WaitFor("Screen: primary full screen (auto)", 3*time.Second)
	if strings.Contains(string(s.Raw()), "\x1b[?1049h") {
		t.Fatal("auto-alt=false must stay on the primary screen")
	}
	quit(t, s)
}
