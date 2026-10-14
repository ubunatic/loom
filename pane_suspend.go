// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// paneRead is one result of the input reader: bytes read from the tty, or the
// error that ended reading.
type paneRead struct {
	data []byte
	err  error
}

// inputReader reads the pane tty on its own goroutine.
//
// Stopping it must be prompt and must not leave a Read pending: a pending Read
// would steal input from whatever owns the terminal next (the shell after
// Close, a child process during Suspend; see issue 066). os.File.Fd (called in
// New) puts the tty in blocking mode and detaches it from the runtime poller,
// so read deadlines do not work and closing the file cannot wake a blocked
// Read. The reader therefore blocks in poll(2) on the tty and a self-pipe and
// only reads after POLLIN, so Read never blocks; stop writes a byte to the
// pipe, which wakes poll at once without idle wakeups. Bytes it has read but
// not yet handed to the loop are kept in held and replayed by the next reader,
// so stopping loses no input.
type inputReader struct {
	stop     chan struct{}
	stopOnce sync.Once
	wakeR    int // self-pipe read end, owned and closed by loop; -1 without a pipe
	wakeW    int // self-pipe write end, closed by signal
	done     chan struct{}
	held     []byte // undelivered bytes; valid once done is closed
}

// readerFallbackPoll is the poll timeout (ms) used only if no self-pipe could
// be created; it bounds how long stopping the reader can take then.
const readerFallbackPoll = 50

// startReader starts an input reader on p.reads. initial, if not empty, is
// delivered first (bytes a previous reader read but could not deliver).
func (p *Pane) startReader(initial []byte) {
	r := &inputReader{stop: make(chan struct{}), done: make(chan struct{}), wakeR: -1, wakeW: -1}
	var fds [2]int
	if unix.Pipe(fds[:]) == nil {
		unix.CloseOnExec(fds[0])
		unix.CloseOnExec(fds[1])
		r.wakeR, r.wakeW = fds[0], fds[1]
	}
	p.reader = r
	p.readerDone = r.done
	go r.loop(p.tty, p.fd, p.reads, initial)
}

// stopReader asks the current reader to exit without waiting. close() joins it
// through readerDone.
func (p *Pane) stopReader() {
	if p.reader != nil {
		p.reader.signal()
	}
}

// haltReader stops the current reader, waits for it to exit and returns the
// bytes it read but did not deliver. It returns ok=false when no reader runs.
func (p *Pane) haltReader() (held []byte, ok bool) {
	r := p.reader
	if r == nil {
		return nil, false
	}
	select {
	case <-r.done:
		r.signal()           // release its self-pipe write end
		return r.held, false // already gone (input error or Run returned)
	default:
	}
	r.signal()
	<-r.done
	return r.held, true
}

// signal stops the reader. A write after loop closed the read end fails with
// EPIPE, which is harmless (Go ignores SIGPIPE for fds other than 1 and 2).
func (r *inputReader) signal() {
	r.stopOnce.Do(func() {
		close(r.stop)
		if r.wakeW >= 0 {
			_, _ = unix.Write(r.wakeW, []byte{0})
			_ = unix.Close(r.wakeW)
		}
	})
}

func (r *inputReader) loop(tty *os.File, fd int, reads chan<- paneRead, initial []byte) {
	defer close(r.done) // let close() and Suspend join us so no read is left racing
	if r.wakeR >= 0 {
		defer unix.Close(r.wakeR)
	}
	send := func(res paneRead) bool {
		select {
		case reads <- res:
			return true
		case <-r.stop:
			r.held = append(r.held, res.data...)
			return false
		}
	}
	if len(initial) > 0 && !send(paneRead{data: initial}) {
		return
	}
	pfd := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	timeout := readerFallbackPoll
	if r.wakeR >= 0 {
		pfd = append(pfd, unix.PollFd{Fd: int32(r.wakeR), Events: unix.POLLIN})
		timeout = -1
	}
	for {
		select {
		case <-r.stop:
			return
		default:
		}
		n, perr := unix.Poll(pfd, timeout)
		if perr != nil {
			if perr == unix.EINTR {
				continue // e.g. a delivered SIGWINCH
			}
			send(paneRead{err: perr})
			return
		}
		if n == 0 || pfd[0].Revents == 0 {
			continue // fallback timeout or a stop wakeup: re-check stop
		}
		buf := make([]byte, 64)
		m, rerr := tty.Read(buf)
		if m > 0 && !send(paneRead{data: buf[:m]}) {
			return
		}
		if rerr != nil || m == 0 {
			send(paneRead{err: rerr})
			return
		}
	}
}

// Suspend hands the terminal to fn, typically to run an interactive child
// process such as $EDITOR on the same tty, and takes it back afterwards. Call
// it from the event loop (a key or mouse handler, or a Post callback).
//
// Before fn runs, Suspend stops the input reader (unread input stays queued for
// the child; bytes the pane already read are replayed after resume), leaves the
// alternate screen, disables mouse tracking and bracketed paste, shows the
// cursor and restores the terminal mode the pane started from (cooked). After
// fn returns, it re-enters raw mode, restores the screen, mouse and paste
// modes, re-reads the terminal size, restarts the reader and repaints every
// row. An inline pane reserves its rows again below the cursor the child left.
// Resizes during fn are applied once on resume, and an interrupt (Ctrl-C) that
// reached the pane while fn ran is dropped, since it was meant for the child.
//
// The terminal is restored even if fn returns an error or panics; the panic is
// re-raised after the restore. Suspend returns fn's error, or the error of
// restoring raw mode.
func (p *Pane) Suspend(fn func() error) (err error) {
	if fn == nil {
		return nil
	}
	if p.tty == nil || p.restored {
		return fmt.Errorf("loom: Suspend needs an open pane")
	}
	if p.suspended {
		return fmt.Errorf("loom: Suspend called while already suspended")
	}
	rawState, err := term.GetState(p.fd)
	if err != nil {
		return fmt.Errorf("loom: suspend: %w", err)
	}
	p.suspended = true
	held, readerRan := p.haltReader()
	wasAlt, mouseMode, paste := p.altActive, p.mouseMode, p.pasteMode
	p.leaveTerminal()

	defer func() {
		r := recover()
		if p.restored {
			// fn closed the pane: the tty is gone, so do not resume onto it;
			// Run returns at its next loop turn.
			p.suspended = false
			p.Quit()
			if r != nil {
				panic(r)
			}
			return
		}
		if rerr := p.resumeTerminal(rawState, wasAlt, mouseMode, paste, held, readerRan); rerr != nil && err == nil {
			err = rerr
		}
		if r != nil {
			panic(r)
		}
	}()
	return fn()
}

// leaveTerminal hands the screen back in the state the shell expects: frame
// erased (or alternate screen left), modes off, cursor visible, cooked mode.
func (p *Pane) leaveTerminal() {
	p.releaseMouse()
	p.setBracketedPaste(false)
	var b strings.Builder
	_ = p.cursorShape.restore(&b)
	if p.altActive {
		p.tty.WriteString(b.String()) //nolint:errcheck
		b.Reset()
		p.switchAltScreen(false) // ?1049l puts the cursor back where the pane started
	} else {
		for i := 0; i < p.rows+p.staleRows; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", p.startRow+i)
		}
		p.staleRows = 0
		fmt.Fprintf(&b, "\x1b[%d;1H", p.startRow)
	}
	b.WriteString("\x1b[?25h\x1b[?7h") // cursor visible, auto-wrap on
	p.tty.WriteString(b.String())      //nolint:errcheck
	if p.oldState != nil {
		_ = term.Restore(p.fd, p.oldState)
	}
}

// resumeTerminal undoes leaveTerminal and restarts the reader.
func (p *Pane) resumeTerminal(rawState *term.State, wasAlt bool, mouseMode int, paste bool, held []byte, readerRan bool) error {
	defer func() { p.suspended = false }()
	err := term.Restore(p.fd, rawState)
	if err != nil {
		err = fmt.Errorf("loom: resume raw mode: %w", err)
	}
	if !p.fullActive {
		held = append(held, p.reserveAtCursor()...)
	}
	if !wasAlt {
		var b strings.Builder
		for i := 0; i < p.rows; i++ {
			fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", p.startRow+i)
		}
		p.tty.WriteString(b.String()) //nolint:errcheck
	} else {
		p.switchAltScreen(true)
	}
	if mouseMode != 0 {
		p.setMouseMode(mouseMode)
	}
	p.setBracketedPaste(paste)
	p.dropSuspendSignals()
	p.resumeRepaint = true
	if readerRan {
		p.startReader(held)
	}
	p.Invalidate()
	return err
}

// reserveAtCursor places the inline pane below the cursor the child left, as
// New does, scrolling when the rows do not fit. It returns input bytes that
// arrived ahead of the cursor report.
func (p *Pane) reserveAtCursor() []byte {
	cols, termRows := termSize(p.fd)
	p.cols = cols
	want := p.wantRows
	if want < 1 {
		want = p.rows
	}
	height := max(1, min(want, termRows))
	row, col, extra, err := queryCursorKeep(p.tty)
	if err != nil || row < 1 || row > termRows {
		p.startRow, p.rows = winchBounds(p.startRow, want, termRows)
		return extra
	}
	startRow, toScroll := reserveRegion(row, termRows, height)
	if toScroll > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "\x1b[%d;1H", termRows)
		for i := 0; i < toScroll; i++ {
			b.WriteString("\r\n")
		}
		p.tty.WriteString(b.String()) //nolint:errcheck
	}
	p.startRow, p.rows = startRow, height
	p.cursorRowOffset, p.cursorCol = toScroll, col
	return extra
}

// dropSuspendSignals discards resizes that arrived while suspended (the resume
// repaint reads the current size) and a SIGINT, which the terminal sent to the
// whole foreground process group for the child. Other signals stay queued.
func (p *Pane) dropSuspendSignals() {
	if p.winch != nil {
		for drained := false; !drained; {
			select {
			case <-p.winch:
			default:
				drained = true
			}
		}
	}
	if p.interrupts != nil {
		var keep []os.Signal
		for drained := false; !drained; {
			select {
			case sig := <-p.interrupts:
				if sig != os.Interrupt {
					keep = append(keep, sig)
				}
			default:
				drained = true
			}
		}
		for _, sig := range keep {
			select {
			case p.interrupts <- sig:
			default: // a new signal took the slot; it is handled as well
			}
		}
	}
}

var cursorReport = regexp.MustCompile(`\x1b\[(\d+);(\d+)R`)

// queryCursorKeep asks for the cursor position like queryCursor, but returns
// any other bytes read meanwhile (typeahead) so the caller can replay them.
// The tty must be in raw mode and no other goroutine may read it.
func queryCursorKeep(tty *os.File) (row, col int, extra []byte, err error) {
	if _, err = tty.WriteString("\x1b[6n"); err != nil {
		return 0, 0, nil, err
	}
	deadline := time.Now().Add(100 * time.Millisecond)
	poll := []unix.PollFd{{Fd: int32(tty.Fd()), Events: unix.POLLIN}}
	var buf []byte
	chunk := make([]byte, 64)
	for len(buf) < 4096 {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, 0, buf, fmt.Errorf("cursor query timed out")
		}
		ready, perr := unix.Poll(poll, max(1, int(remaining.Milliseconds())))
		if perr == unix.EINTR {
			continue
		}
		if perr != nil {
			return 0, 0, buf, perr
		}
		if ready == 0 {
			continue
		}
		m, rerr := tty.Read(chunk)
		buf = append(buf, chunk[:m]...)
		if loc := cursorReport.FindSubmatchIndex(buf); loc != nil {
			row, _ = strconv.Atoi(string(buf[loc[2]:loc[3]]))
			col, _ = strconv.Atoi(string(buf[loc[4]:loc[5]]))
			extra = append(append([]byte(nil), buf[:loc[0]]...), buf[loc[1]:]...)
			return row, col, extra, nil
		}
		if rerr != nil || m == 0 {
			return 0, 0, buf, fmt.Errorf("cursor query: input closed")
		}
	}
	return 0, 0, buf, fmt.Errorf("cursor query: response too long")
}
