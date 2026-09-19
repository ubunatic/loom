// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package winch_test

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"codeberg.org/ubunatic/loom/internal/ptytest"
)

func buildWinch(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "winch")
	if out, err := exec.Command("go", "build", "-o", bin, "codeberg.org/ubunatic/loom/examples/winch").CombinedOutput(); err != nil {
		t.Fatalf("build winch: %v\n%s", err, out)
	}
	return bin
}

// TestWinchPTYResizeStream drives the real binary through a shrink/grow
// stream and checks what a terminal would show: every synchronized frame
// stays within the terminal width, and the settled screen after each resize
// has no content left over from a previous, wider frame.
func TestWinchPTYResizeStream(t *testing.T) {
	s := ptytest.Start(t, 100, 30, buildWinch(t))
	s.WaitFor("Resize Modes", 5*time.Second)

	sizes := [][2]int{{80, 24}, {50, 20}, {30, 12}, {50, 20}, {100, 30}}
	for _, sz := range sizes {
		s.Resize(sz[0], sz[1])
		time.Sleep(20 * time.Millisecond)
	}
	// Let the width guard settle (its default restore delay is one second).
	time.Sleep(1300 * time.Millisecond)
	s.WaitFor("Resize Modes", 3*time.Second)

	for i, frame := range s.Frames() {
		for y, row := range frame {
			if w := len([]rune(row)); w > 100 {
				t.Fatalf("frame %d row %d has %d runes:\n%s", i, y, w, strings.Join(frame, "\n"))
			}
		}
	}
	if final := s.Screen(); !strings.Contains(strings.Join(final, "\n"), "Guard:") {
		t.Fatalf("settled screen lacks the guard status line:\n%s", strings.Join(final, "\n"))
	}

	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("winch exit: %v", err)
	}
	if got := string(s.Raw()); !strings.Contains(got, "\x1b[?25h") {
		t.Fatalf("cursor not restored on exit; tail: %q", got[max(0, len(got)-80):])
	}
}

// TestWinchPTYAdaptiveGuard enables "Use WINCH speed", fires a rapid resize
// burst and expects the status line to report an adaptive guard at some
// frame, then to fall back to full width after the burst settles.
func TestWinchPTYAdaptiveGuard(t *testing.T) {
	s := ptytest.Start(t, 100, 30, buildWinch(t))
	s.WaitFor("Resize Modes", 5*time.Second)
	s.Send("a")
	s.WaitFor("[ON]  Use WINCH speed", 3*time.Second)

	for i := 0; i < 8; i++ {
		s.Resize(100-i*2, 30)
		time.Sleep(15 * time.Millisecond)
	}
	found := false
	for _, frame := range s.Frames() {
		if strings.Contains(strings.Join(frame, "\n"), "adaptive") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no frame reported an adaptive guard; screen:\n%s", strings.Join(s.Screen(), "\n"))
	}

	s.WaitFor("Guard: idle", 3*time.Second)
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("winch exit: %v", err)
	}
}

func framesMention(s *ptytest.Session, word string) bool {
	for _, frame := range s.Frames() {
		if strings.Contains(strings.Join(frame, "\n"), word) {
			return true
		}
	}
	return false
}

// TestWinchPTYSlowDragAndManualBurst covers the other two adaptive cases: a
// slow drag (events too sparse to measure a speed) must keep the manual n,
// and a rapid burst with the switch off must stay manual.
func TestWinchPTYSlowDragAndManualBurst(t *testing.T) {
	bin := buildWinch(t)

	slow := ptytest.Start(t, 100, 30, bin)
	slow.WaitFor("Resize Modes", 5*time.Second)
	slow.Send("a")
	slow.WaitFor("[ON]  Use WINCH speed", 3*time.Second)
	for i := 0; i < 3; i++ {
		slow.Resize(100-i*4, 30)
		time.Sleep(700 * time.Millisecond) // longer than the measurement window
	}
	if framesMention(slow, "adaptive") {
		t.Fatalf("slow drag reported an adaptive guard; screen:\n%s", strings.Join(slow.Screen(), "\n"))
	}
	slow.Send("q")
	if err := slow.Wait(3 * time.Second); err != nil {
		t.Fatalf("winch exit: %v", err)
	}

	manual := ptytest.Start(t, 100, 30, bin)
	manual.WaitFor("Resize Modes", 5*time.Second)
	for i := 0; i < 8; i++ {
		manual.Resize(100-i*2, 30)
		time.Sleep(15 * time.Millisecond)
	}
	manual.WaitFor("Guard: idle", 3*time.Second)
	if framesMention(manual, "adaptive") {
		t.Fatal("manual mode reported an adaptive guard during a rapid burst")
	}
	manual.Send("q")
	if err := manual.Wait(3 * time.Second); err != nil {
		t.Fatalf("winch exit: %v", err)
	}
}

// blankThenWrite matches a whole-line erase (CSI 2K) followed directly by
// content, the "clear row, then redraw it" pattern that makes lines flash
// blank between the two writes (see emojig issue 016).
var blankThenWrite = regexp.MustCompile(`\x1b\[2K[^\x1b]`)

// TestWinchPTYNoBlankThenRewrite asserts Loom never blanks a row before
// rewriting it: rows are overwritten in place and trimmed with CSI K, and a
// whole-line erase only ever clears rows below the canvas.
func TestWinchPTYNoBlankThenRewrite(t *testing.T) {
	s := ptytest.Start(t, 100, 30, buildWinch(t))
	s.WaitFor("Resize Modes", 5*time.Second)
	for i := 0; i < 6; i++ {
		s.Resize(100-i*6, 30-i)
		time.Sleep(15 * time.Millisecond)
	}
	s.Send("m") // toggle motion to force further redraws
	time.Sleep(300 * time.Millisecond)
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("winch exit: %v", err)
	}
	raw := s.Raw()
	if !strings.Contains(string(raw), "\x1b[K") {
		t.Fatal("expected per-row CSI K trims in the output")
	}
	if loc := blankThenWrite.FindIndex(raw); loc != nil {
		t.Fatalf("row blanked then rewritten near %q", raw[max(0, loc[0]-40):min(len(raw), loc[1]+40)])
	}
}

// TestWinchPTYAltScreen switches to the alternate screen, drags, and expects
// the app to leave it again on exit so the shell screen is untouched.
func TestWinchPTYAltScreen(t *testing.T) {
	s := ptytest.Start(t, 100, 30, buildWinch(t))
	s.WaitFor("Resize Modes", 5*time.Second)
	s.Send("b")
	s.WaitFor("[ON]  Alternate screen", 3*time.Second)
	for i := 0; i < 8; i++ {
		s.Resize(100-i*5, 30-i)
		time.Sleep(15 * time.Millisecond)
	}
	s.WaitFor("Guard: idle", 3*time.Second)
	s.Send("q")
	if err := s.Wait(3 * time.Second); err != nil {
		t.Fatalf("winch exit: %v", err)
	}
	raw := string(s.Raw())
	on, off := strings.Index(raw, "\x1b[?1049h"), strings.LastIndex(raw, "\x1b[?1049l")
	if on < 0 || off < on {
		t.Fatalf("alt screen not entered and left (on=%d off=%d)", on, off)
	}
}
