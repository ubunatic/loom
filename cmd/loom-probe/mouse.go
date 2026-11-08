// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"
	"ubunatic.com/loom"
)

// mouseOn enables any-motion tracking (so a pointer that comes back is seen
// even without a button) with SGR coordinates; mouseOff turns both off.
const (
	mouseOn  = "\x1b[?1003h\x1b[?1006h"
	mouseOff = "\x1b[?1006l\x1b[?1003l"
)

// mouseState is what the probe has detected so far. It is shown as fixed rows.
type mouseState struct {
	down     [3]bool // buttons that seem held
	stale    [3]bool // assumed still held, but the terminal now reports hover
	inferred [3]bool // held because a drag arrived, not because a press was seen
	x, y     int
	last     loom.MouseAction
	lastAt   time.Time
	counts   map[loom.MouseAction]int
	noPress  int    // release or drag with no press seen before
	lostRel  int    // hover while a button seemed held: no release came first
	lostNote string // where the last drag was, and what came after, at the last loss
	lastDrag string // position of the last drag report, and whether it was outside
	outside  int    // reports with coordinates outside the window
	gesture  struct {
		x, y, drags int
		at          time.Time
	}
	history []string // finished gestures, newest last

	wheelAt     time.Time     // time of the previous wheel report
	wheelGap    time.Duration // gap before the latest wheel report
	wheelMinGap time.Duration // smallest gap seen between two wheel reports
	wheelRun    int           // reports in the current burst (gaps under wheelBurstGap)
	wheelRunMax int           // longest burst seen
}

// wheelBurstGap is the longest pause that still counts as one continuous scroll.
const wheelBurstGap = 150 * time.Millisecond

func (s *mouseState) anyDown() bool { return s.down[0] || s.down[1] || s.down[2] }

func (s *mouseState) apply(me loom.MouseEvent, now time.Time, cols, rows int) {
	s.counts[me.Action]++
	s.x, s.y, s.last, s.lastAt = me.X, me.Y, me.Action, now
	if me.X < 1 || me.Y < 1 || me.X > cols || me.Y > rows {
		s.outside++
	}
	b := int(me.Button)
	switch me.Action {
	case loom.MousePress:
		if b < 3 {
			if !s.anyDown() {
				s.gesture.x, s.gesture.y, s.gesture.drags, s.gesture.at = me.X, me.Y, 0, now
			}
			s.down[b], s.inferred[b], s.stale[b] = true, false, false
		}
	case loom.MouseDrag:
		if b >= 3 {
			return
		}
		if !s.down[b] {
			s.noPress++
			if !s.anyDown() {
				s.gesture.x, s.gesture.y, s.gesture.drags, s.gesture.at = me.X, me.Y, 0, now
			}
			s.down[b], s.inferred[b], s.stale[b] = true, true, false
		}
		s.stale[b] = false
		s.gesture.drags++
		where := "inside"
		if me.X < 1 || me.Y < 1 || me.X > cols || me.Y > rows {
			where = "OUTSIDE"
		}
		s.lastDrag = fmt.Sprintf("x=%d y=%d %s", me.X, me.Y, where)
	case loom.MouseScrollUp, loom.MouseScrollDown:
		s.wheel(now)
	case loom.MouseHover:
		if s.anyDown() {
			s.lostRel++
			s.lostNote = fmt.Sprintf("last drag %s, then hover at x=%d y=%d with no release", s.lastDrag, me.X, me.Y)
			s.stale = s.down
		}
	case loom.MouseRelease:
		if b >= 3 || !s.down[b] {
			s.noPress++
			return
		}
		s.down[b], s.inferred[b], s.stale[b] = false, false, false
		if !s.anyDown() {
			s.history = append(s.history, fmt.Sprintf("press %d,%d -> %d drags -> release %d,%d after %dms",
				s.gesture.x, s.gesture.y, s.gesture.drags, me.X, me.Y, now.Sub(s.gesture.at).Milliseconds()))
			if len(s.history) > 3 {
				s.history = s.history[1:]
			}
		}
	}
}

// wheel tracks how fast wheel reports arrive, to see what a scroll gesture
// looks like at the terminal before any acceleration is designed.
func (s *mouseState) wheel(now time.Time) {
	if !s.wheelAt.IsZero() {
		s.wheelGap = now.Sub(s.wheelAt)
		if s.wheelMinGap == 0 || s.wheelGap < s.wheelMinGap {
			s.wheelMinGap = s.wheelGap
		}
	}
	if s.wheelAt.IsZero() || s.wheelGap > wheelBurstGap {
		s.wheelRun = 0
	}
	s.wheelRun++
	s.wheelRunMax = max(s.wheelRunMax, s.wheelRun)
	s.wheelAt = now
}

// buttons draws one box per button; a held one is reverse video (red with ! when the terminal ended its drag), and one held
// only because a drag arrived (no press seen) is marked with a question mark.
func (s *mouseState) buttons() string {
	out := ""
	for i, name := range []string{"LEFT", "MIDDLE", "RIGHT"} {
		switch {
		case s.stale[i]:
			out += "\x1b[7;31m " + name + "! \x1b[0m "
		case s.inferred[i]:
			out += "\x1b[7;33m " + name + "? \x1b[0m "
		case s.down[i]:
			out += "\x1b[7;32m " + name + " \x1b[0m "
		default:
			out += "\x1b[2m[" + name + "]\x1b[0m "
		}
	}
	return out
}

// render writes the fixed rows, each cleared to the line end, from the top.
func (s *mouseState) render(w io.Writer, cols, rows int, now time.Time) {
	silent := now.Sub(s.lastAt).Milliseconds()
	if s.lastAt.IsZero() {
		silent = 0
	}
	pointer := "no report yet"
	if !s.lastAt.IsZero() {
		pointer = fmt.Sprintf("x=%d y=%d in a %dx%d window", s.x, s.y, cols, rows)
		if s.x < 1 || s.y < 1 || s.x > cols || s.y > rows {
			pointer += "  \x1b[1;31mOUTSIDE\x1b[0m"
		}
	}
	verdict := "ok"
	switch {
	case s.anyDown() && silent > 300:
		verdict = "button seems HELD but no reports for >300ms: pointer left, or release lost"
	case s.stale[0] || s.stale[1] || s.stale[2]:
		verdict = "terminal ended the drag: hover reports while a button is assumed held (red !)"
	case s.noPress > 0:
		verdict = "release/drag for a press this probe never saw"
	}
	lines := []string{
		"loom-probe mouse   q quits, r resets   TERM=" + os.Getenv("TERM") + " " + os.Getenv("TERM_PROGRAM"),
		"",
		"buttons      " + s.buttons(),
		"pointer      " + pointer,
		fmt.Sprintf("silent for   %d ms", silent),
		fmt.Sprintf("last event   %v", s.last),
		fmt.Sprintf("counts       press %d  drag %d  release %d  hover %d",
			s.counts[loom.MousePress], s.counts[loom.MouseDrag], s.counts[loom.MouseRelease], s.counts[loom.MouseHover]),
		fmt.Sprintf("outside      %d reports with coordinates outside the window", s.outside),
		fmt.Sprintf("anomalies    no-press %d, drag->hover without release %d", s.noPress, s.lostRel),
		"last such    " + s.lostNote,
		fmt.Sprintf("wheel        up %d  down %d  last gap %d ms  min gap %d ms  burst %d (max %d)",
			s.counts[loom.MouseScrollUp], s.counts[loom.MouseScrollDown],
			s.wheelGap.Milliseconds(), s.wheelMinGap.Milliseconds(), s.wheelRun, s.wheelRunMax),
		"verdict      " + verdict,
		"",
		"gestures (last 3)",
	}
	for i := range 3 {
		line := ""
		if i < len(s.history) {
			line = s.history[i]
		}
		lines = append(lines, "  "+line)
	}
	out := "\x1b[H"
	for _, l := range lines {
		out += l + "\x1b[K\r\n"
	}
	_, _ = io.WriteString(w, out)
}

// runMouse shows the mouse state seen at the terminal, with no widget involved,
// so a drag that stops or a release the terminal invents when the pointer
// leaves the window is visible as a state change.
func runMouse(input *os.File, output io.Writer) error {
	inFD := int(input.Fd())
	outFD, ok := output.(*os.File)
	if !ok || !term.IsTerminal(inFD) || !term.IsTerminal(int(outFD.Fd())) {
		return fmt.Errorf("stdin and stdout must be attached to a real terminal")
	}
	oldState, err := term.MakeRaw(inFD)
	if err != nil {
		return fmt.Errorf("enable raw terminal input: %w", err)
	}
	defer func() {
		_, _ = io.WriteString(output, mouseOff+"\x1b[?25h")
		_ = term.Restore(inFD, oldState)
	}()
	if _, err := io.WriteString(output, "\x1b[2J\x1b[?25l"+mouseOn); err != nil {
		return fmt.Errorf("enable mouse tracking: %w", err)
	}
	newState := func() *mouseState { return &mouseState{counts: map[loom.MouseAction]int{}} }
	st := newState()
	for {
		cols, rows, _ := term.GetSize(int(outFD.Fd()))
		st.render(output, cols, rows, time.Now())
		chunk, err := readChunk(inFD, 100*time.Millisecond)
		if err != nil {
			return fmt.Errorf("read mouse: %w", err)
		}
		switch string(chunk) {
		case "q":
			return nil
		case "r":
			st = newState()
			continue
		}
		for rest := chunk; len(rest) > 0; {
			me, n, ok := loom.ScanMouse(rest)
			if !ok {
				break
			}
			st.apply(me, time.Now(), cols, rows)
			rest = rest[n:]
		}
	}
}
