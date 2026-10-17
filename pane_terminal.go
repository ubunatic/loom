// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"os/signal"
	"strings"
	"time"
)

// New opens /dev/tty, enters raw mode, and reserves height rows below the
// current cursor position. Call Close (or defer it) to restore the terminal.
//
// Always uses /dev/tty — never os.Stdin — so the pane works inside shells
// and command substitution (e.g. `foo=$(bar)`), where stdin may be a pipe.
// See docs/TuiInput.md §1.
func New(height int) (*Pane, error) {
	if height < 1 {
		height = 1
	}
	if !claimPaneOwnership() {
		return nil, fmt.Errorf("loom: another pane already owns the terminal")
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		releasePaneOwnership()
		return nil, fmt.Errorf("loom: open /dev/tty: %w", err)
	}
	fd := int(tty.Fd())

	old, err := term.MakeRaw(fd)
	if err != nil {
		_ = tty.Close()
		releasePaneOwnership()
		return nil, fmt.Errorf("loom: raw mode: %w", err)
	}

	cols, termRows := termSize(fd)
	if cols < 1 {
		cols = 80
	}
	wantRows := height
	if height > termRows {
		height = termRows
	}

	if Debug {
		fmt.Fprintf(os.Stderr, "loom: [debug] terminal: %dx%d\r\n", cols, termRows)
	}

	// Query cursor position using DSR
	cy, cx, err := queryCursor(tty)
	if err != nil || cy < 1 || cy > termRows {
		cy = termRows
		cx = 1
	}

	actualCy := cy
	if Debug {
		// We print 2 more debug lines below, which will advance the cursor by 2 lines.
		// Adjust actualCy accordingly (capped to termRows) for the TUI reservation.
		actualCy = cy + 2
		if actualCy > termRows {
			actualCy = termRows
		}
	}

	startRow, toScroll := reserveRegion(actualCy, termRows, height)
	if Debug {
		fmt.Fprintf(os.Stderr, "loom: [debug] cursor: (%d,%d)\r\n", cx, cy)
		fmt.Fprintf(os.Stderr, "loom: [debug] placement: start_row: %d, scroll: %d\r\n", startRow, toScroll)
	}

	if toScroll > 0 {
		// Scroll with line feeds instead of CSI S. Many terminals discard lines
		// scrolled with CSI S instead of adding them to scrollback.
		tty.WriteString(fmt.Sprintf("\x1b[%d;1H", termRows))
		for i := 0; i < toScroll; i++ {
			tty.WriteString("\r\n")
		}
	}

	// Move cursor to startRow so drawing begins at the correct relative row
	tty.WriteString(fmt.Sprintf("\x1b[%d;1H", startRow))

	p := &Pane{
		tty:             tty,
		fd:              fd,
		oldState:        old,
		rows:            height,
		wantRows:        wantRows,
		startRow:        startRow,
		cursorRowOffset: toScroll,
		cursorCol:       cx,
		cols:            cols,
		MaxCols:         DefaultMaxCols,
		ResizeConfig:    DefaultResizeConfig(),
		ownsTTY:         true,
	}
	if env := strings.ToLower(strings.TrimSpace(os.Getenv("LOOM_DEBUG"))); env == "1" || env == "true" {
		p.DebugMode = true
	}
	if paneBeforeSignalHandler != nil {
		paneBeforeSignalHandler(p)
	}
	p.installSignalHandler()
	return p, nil
}

func claimPaneOwnership() bool {
	paneOwnership.Lock()
	defer paneOwnership.Unlock()
	if paneOwnership.active {
		return false
	}
	paneOwnership.active = true
	return true
}

func releasePaneOwnership() {
	paneOwnership.Lock()
	paneOwnership.active = false
	paneOwnership.Unlock()
}

func queryCursor(tty *os.File) (row, col int, err error) {
	if _, err = tty.WriteString("\x1b[6n"); err != nil {
		return 0, 0, err
	}
	// Poll synchronously: a timed-out reader goroutine would steal later keys.
	deadline := time.Now().Add(100 * time.Millisecond)
	var buf [32]byte
	for n := 0; n < len(buf); {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, 0, fmt.Errorf("cursor query timed out")
		}
		ready, err := waitTTY(int(tty.Fd()), remaining)
		if err != nil {
			if isInterrupted(err) {
				continue
			}
			return 0, 0, err
		}
		if !ready {
			continue
		}
		m, err := tty.Read(buf[n : n+1])
		if err != nil {
			if isInterrupted(err) {
				continue
			}
			return 0, 0, err
		}
		if m == 0 {
			return 0, 0, fmt.Errorf("cursor query: input closed")
		}
		n += m
		if buf[n-1] == 'R' {
			_, err := fmt.Sscanf(string(buf[:n]), "\x1b[%d;%dR", &row, &col)
			return row, col, err
		}
	}
	return 0, 0, fmt.Errorf("cursor query: response too long")
}

func reserveRegion(cy, rows, want int) (startRow, toScroll int) {
	startRow = cy
	if rows-cy < want {
		if d := want - (rows - cy) - 1; d > 0 {
			toScroll = d
			startRow -= d
		}
	}
	if startRow < 1 {
		startRow = 1
	}
	return startRow, toScroll
}

// Close tears down the pane: clears the reserved region, restores terminal
// state, and closes /dev/tty. Safe to call multiple times.
func (p *Pane) Close() { p.close() }

func (p *Pane) close() {
	if p.restored {
		return
	}
	p.restored = true
	p.postMu.Lock()
	p.postClosed = true
	p.posted = nil
	p.postMu.Unlock()
	if p.ownsTTY {
		releasePaneOwnership()
		p.ownsTTY = false
	}

	if p.winch != nil {
		signal.Stop(p.winch)
	}
	if p.interrupts != nil {
		signal.Stop(p.interrupts)
	}

	// Stop the reader and wait for it before the tty is drained and closed, so
	// it cannot consume a keystroke meant for the next terminal reader or poll
	// a closed (possibly reused) fd. Its self-pipe wakes it at once; the wait is
	// bounded so teardown can never hang. A Run still in progress (Close from a
	// handler) returns at its next loop turn.
	p.stopReader()
	if p.readerDone != nil {
		select {
		case <-p.readerDone:
		case <-time.After(300 * time.Millisecond):
		}
	}
	p.Quit()

	p.disableMouse()
	p.setBracketedPaste(false)
	p.setKittyKeyboard(false) // pop on the current screen, before leaving alt

	var b strings.Builder
	_ = p.cursorShape.restore(&b)
	if p.altActive {
		// Leaving the alternate screen restores the shell's screen and cursor.
		b.WriteString("\x1b[?1049l")
		p.altActive = false
	} else {
		// Clear reserved region.
		for i := 0; i < p.rows; i++ {
			b.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", p.startRow+i))
		}
		b.WriteString(fmt.Sprintf("\x1b[%d;%dH", p.startRow+p.cursorRowOffset, max(1, p.cursorCol)))
	}
	b.WriteString("\x1b[?25h")    // restore cursor visibility (a prompt-less frame may have hidden it)
	b.WriteString("\x1b[?7h")     // restore auto-wrap
	p.tty.WriteString(b.String()) //nolint:errcheck

	// Discard any unread input before leaving raw mode: keys pressed during the
	// session (the closing Enter, leftover mouse/DSR reports) otherwise sit in the
	// tty queue and leak into the shell or the next cooked reader as a stray line.
	p.drainInput()

	_ = term.Restore(p.fd, p.oldState)
	_ = p.tty.Close()
}

// drainInput flushes the terminal input queue (TCIFLUSH) so raw-mode leftovers
// don't bleed into whatever reads the terminal next.
func (p *Pane) drainInput() {
	drainTerminalInput(p.fd)
}

func (p *Pane) installSignalHandler() {
	// Room for each of the three signals, so dropping a SIGINT on resume
	// (Suspend) cannot cost a SIGTERM or SIGHUP that arrived alongside.
	p.interrupts = make(chan os.Signal, 3)
	// SIGWINCH goes to its own channel; the Run loop selects on it to reflow.
	// Cap 1 coalesces a burst of resizes (e.g. an interactive drag) into one.
	p.winch = make(chan os.Signal, 1)
	notifyPaneSignals(p.interrupts, p.winch)
}

// TerminalSize returns the dimensions of the controlling terminal.
// It returns an error when /dev/tty is unavailable or its size cannot be read.
func TerminalSize() (cols, rows int, err error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return 0, 0, fmt.Errorf("loom: open /dev/tty: %w", err)
	}
	defer tty.Close()
	cols, rows, err = term.GetSize(int(tty.Fd()))
	if err != nil {
		return 0, 0, fmt.Errorf("loom: terminal size: %w", err)
	}
	if cols < 1 || rows < 1 {
		return 0, 0, fmt.Errorf("loom: invalid terminal size %dx%d", cols, rows)
	}
	return cols, rows, nil
}

// termSize returns the terminal dimensions via x/term (TIOCGWINSZ under the
// hood on Unix), falling back to a conventional 80x24 on error -- this used
// to hand-roll its own syscall.Syscall(SYS_IOCTL, TIOCGWINSZ, ...) call
// even though term.GetSize (already imported in this file for MakeRaw/
// Restore) does the same thing more portably.
func termSize(fd int) (cols, rows int) {
	cols, rows, err := term.GetSize(fd)
	if err != nil || cols < 1 || rows < 1 {
		return 80, 24
	}
	return cols, rows
}
