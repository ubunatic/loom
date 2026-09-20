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
	s := ptytest.Start(t, 100, 30, build(t), "--auto=false")
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
	s := ptytest.Start(t, 100, 30, build(t), "--height=27")
	s.WaitFor("Screen: inline", 5*time.Second)
	s.Send("+") // 28 of 30 rows: still 2 short, margin_rows=1
	s.WaitFor("Height [-] 28 [+]", 3*time.Second)
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
	s := ptytest.Start(t, 100, 30, build(t), "--height=29", "--margin=0", "--auto-alt=false")
	s.WaitFor("Screen: inline", 5*time.Second)
	s.Send("+")
	s.WaitFor("Screen: primary full screen (auto)", 3*time.Second)
	if strings.Contains(string(s.Raw()), "\x1b[?1049h") {
		t.Fatal("auto-alt=false must stay on the primary screen")
	}
	quit(t, s)
}

// TestScreensWidthButtonsThemeAstra presses the width key and expects
// the pane to grow up to the terminal width, cycles the theme and toggles Astra.
func TestScreensWidthButtonsThemeAstra(t *testing.T) {
	s := ptytest.Start(t, 100, 30, build(t), "--auto=false", "--width=50", "--astra=false")
	s.WaitFor("Width [-] 50/100 [+]", 5*time.Second)
	s.Send("w")
	s.WaitFor("Width [-] 60/100 [+]", 3*time.Second)
	for range 8 {
		s.Send("w")
		time.Sleep(30 * time.Millisecond) // one key per read
	}
	s.WaitFor("Width [-] 100/100 [+]", 3*time.Second)
	s.Send("W")
	s.WaitFor("Width [-] 90/100 [+]", 3*time.Second)

	s.Send("t")
	s.WaitFor("Theme [", 3*time.Second)
	s.Send("a")
	s.WaitFor("Astra [on]", 3*time.Second)
	quit(t, s)
}

// TestScreensAutoByWidth grows the pane to the terminal width: the width
// detector promotes it even though the height is small, and the height
// detector alone (--by-width=false) does not.
func TestScreensAutoByWidth(t *testing.T) {
	bin := build(t)
	s := ptytest.Start(t, 100, 30, bin, "--width=90")
	s.WaitFor("Screen: inline", 5*time.Second)
	s.Send("w") // 100 of 100 columns
	s.WaitFor("Screen: full screen (alt) (auto)", 3*time.Second)
	s.Resize(140, 30) // 100 of 140 columns is clearly narrower again
	s.WaitFor("Screen: inline", 3*time.Second)
	quit(t, s)

	off := ptytest.Start(t, 100, 30, bin, "--width=90", "--by-width=false")
	off.WaitFor("Screen: inline", 5*time.Second)
	off.Send("w")
	off.WaitFor("Width [-] 100/100 [+]", 3*time.Second)
	time.Sleep(200 * time.Millisecond)
	if strings.Contains(strings.Join(off.Screen(), "\n"), "(auto)") {
		t.Fatal("width detection is off but the pane was promoted")
	}
	quit(t, off)
}

// TestScreensShrinkErasesOldRows presses "-" and expects the rows below the
// smaller pane to be erased, leaving exactly one bottom border.
func TestScreensShrinkErasesOldRows(t *testing.T) {
	s := ptytest.Start(t, 100, 30, build(t), "--auto=false", "--height=14")
	s.WaitFor("Height [-] 14 [+]", 5*time.Second)
	for range 3 {
		s.Send("-")
		time.Sleep(40 * time.Millisecond)
	}
	s.WaitFor("Height [-] 11 [+]", 3*time.Second)
	time.Sleep(100 * time.Millisecond)
	borders := 0
	for _, row := range s.Screen() {
		if strings.HasPrefix(row, "└") {
			borders++
		}
	}
	if borders != 1 {
		t.Fatalf("want one bottom border, got %d; screen:\n%s", borders, strings.Join(s.Screen(), "\n"))
	}
	quit(t, s)
}
