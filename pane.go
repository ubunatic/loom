// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package loom is an inline terminal UI library.
//
// A Pane reserves a fixed-height region below the shell cursor, renders
// a Widget tree into a Canvas on every frame, and dispatches keyboard and
// mouse events. The terminal is always restored on exit, panic, or signal.
//
// Typical usage:
//
//	items := []loom.Item{{Name: "git"}, {Name: "make"}}
//	choice := loom.NewChoice(items)
//	pane, err := loom.New(12)
//	if err != nil { log.Fatal(err) }
//	defer pane.Close()
//	pane.Run(choice)
//	if item, ok := choice.Selected(); ok {
//	    fmt.Println(item.Name)
//	}
//
// See docs/TuiInput.md for the constraints that govern /dev/tty usage,
// EINTR handling, and ZSH widget compatibility.
package loom

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// DefaultMaxCols is the default maximum canvas width. Keeps the inline pane
// narrow enough to feel like a popup rather than a full-screen takeover.
const DefaultMaxCols = 50

// Pane manages an inline terminal region and drives the widget event loop.
type Pane struct {
	tty      *os.File
	fd       int
	oldState *term.State
	rows     int // reserved height (clamped to the current terminal)
	wantRows int // desired height; rows is restored toward this when space allows
	startRow int // 1-based terminal row of the pane top
	cols     int // terminal width at Open time
	MaxCols  int // canvas width cap; 0 = use terminal width
	restored bool

	// mouse tracking is enabled with EnableMouse.
	mouse      bool
	Resizeable bool

	// winch carries SIGWINCH notifications so the Run loop reflows on a
	// terminal window resize. Buffered (cap 1) to coalesce resize bursts.
	winch chan os.Signal

	// readerDone is closed by Run's input-reader goroutine when it exits. close()
	// waits on it so a lingering blocked Read cannot steal input meant for
	// whatever reads the terminal after the pane (e.g. the calling shell).
	readerDone chan struct{}
}

// New opens /dev/tty, enters raw mode, and reserves height rows below the
// current cursor position. Call Close (or defer it) to restore the terminal.
//
// Always uses /dev/tty — never os.Stdin — so the pane works inside ZSH
// command substitution (result=$(uzu)) where stdin may be a pipe.
// See docs/TuiInput.md §1.
func New(height int) (*Pane, error) {
	if height < 1 {
		height = 1
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("loom: open /dev/tty: %w", err)
	}
	fd := int(tty.Fd())

	old, err := term.MakeRaw(fd)
	if err != nil {
		_ = tty.Close()
		return nil, fmt.Errorf("loom: raw mode: %w", err)
	}

	cols, termRows := termSize(fd)
	if cols < 1 {
		cols = 80
	}
	wantRows := height
	if height > termRows-1 {
		height = termRows - 1
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
		tty.WriteString(fmt.Sprintf("\x1b[%dS", toScroll)) // scroll up toScroll lines
	}

	// Move cursor to startRow so drawing begins at the correct relative row
	tty.WriteString(fmt.Sprintf("\x1b[%d;1H", startRow))

	p := &Pane{
		tty:      tty,
		fd:       fd,
		oldState: old,
		rows:     height,
		wantRows: wantRows,
		startRow: startRow,
		cols:     cols,
		MaxCols:  DefaultMaxCols,
	}
	p.installSignalHandler()
	return p, nil
}

func queryCursor(tty *os.File) (row, col int, err error) {
	if _, err = tty.WriteString("\x1b[6n"); err != nil {
		return 0, 0, err
	}
	type res struct {
		row, col int
		err      error
	}
	ch := make(chan res, 1)
	go func() {
		var buf [32]byte
		n := 0
		for n < len(buf) {
			m, rerr := tty.Read(buf[n : n+1])
			if rerr != nil {
				ch <- res{err: rerr}
				return
			}
			n += m
			if m > 0 && buf[n-1] == 'R' {
				break
			}
		}
		var r, c int
		if _, serr := fmt.Sscanf(string(buf[:n]), "\x1b[%d;%dR", &r, &c); serr != nil {
			ch <- res{err: serr}
			return
		}
		ch <- res{row: r, col: c}
	}()
	select {
	case out := <-ch:
		return out.row, out.col, out.err
	case <-time.After(100 * time.Millisecond):
		return 0, 0, fmt.Errorf("cursor query timed out")
	}
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

// EnableMouse turns on SGR mouse tracking (button press, release, hover, scroll).
// Must be called before Run. Mouse events are delivered to the root Widget's
// HandleMouse method.
func (p *Pane) EnableMouse() {
	p.mouse = true
	p.tty.WriteString("\x1b[?1003h\x1b[?1006h") //nolint:errcheck
}

// Resize adjusts the pane height dynamically.
// If the height increases and overflows the terminal screen, it scrolls the terminal up
// to reserve the required space. If it decreases, it clears the abandoned lines.
func (p *Pane) Resize(newHeight int) {
	if newHeight < 1 {
		newHeight = 1
	}
	cols, termRows := termSize(p.fd)
	if cols < 1 {
		cols = 80
	}
	p.wantRows = newHeight
	if newHeight > termRows-1 {
		newHeight = termRows - 1
	}

	if newHeight == p.rows {
		return
	}

	if newHeight > p.rows {
		// Growing: check for overflow at the bottom of the screen
		overflow := (p.startRow + newHeight - 1) - termRows
		if overflow > 0 {
			// Scroll up to reserve space
			p.tty.WriteString(fmt.Sprintf("\x1b[%dS", overflow)) //nolint:errcheck
			p.startRow -= overflow
			if p.startRow < 1 {
				p.startRow = 1
			}
		}
	} else {
		// Shrinking: clear the abandoned lines at the bottom so they don't leave stale text
		var b strings.Builder
		for i := newHeight; i < p.rows; i++ {
			b.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", p.startRow+i))
		}
		p.tty.WriteString(b.String()) //nolint:errcheck
	}

	p.rows = newHeight
}

// winchBounds recomputes pane placement after a terminal window resize. Given
// the current top row and height and the new terminal row count, it clamps the
// height to fit (leaving the shell prompt line) and lifts the top row if the
// pane would overflow the new bottom. It is pure so the math can be unit-tested.
func winchBounds(startRow, rows, termRows int) (newStartRow, newRows int) {
	newRows = rows
	if newRows > termRows-1 {
		newRows = termRows - 1
	}
	if newRows < 1 {
		newRows = 1
	}
	newStartRow = startRow
	if overflow := (newStartRow + newRows - 1) - termRows; overflow > 0 {
		newStartRow -= overflow
	}
	if newStartRow < 1 {
		newStartRow = 1
	}
	return newStartRow, newRows
}

// applyWinch re-queries the terminal size after a SIGWINCH and updates the
// pane bounds and canvas width so the next frame reflows. It clears generously —
// from the topmost of the old/new pane row down to the bottom of the new screen —
// because narrowing makes the terminal rewrap our old wide lines and push the
// overflow down past the old region; a fixed old-region clear would miss that.
// The caller recreates the canvas at the returned width and redraws.
func (p *Pane) applyWinch(cols *int) {
	newCols, termRows := termSize(p.fd)
	if newCols < 1 {
		newCols = 80
	}

	// Clamp from the desired height, not the currently-clamped one, so the pane
	// grows back toward wantRows when the window is enlarged again (a tiny window
	// must not permanently pin the pane at one row).
	newStartRow, newRows := winchBounds(p.startRow, p.wantRows, termRows)

	// Clear from the highest pane top (old or new) down to the new screen bottom
	// so rewrapped overflow leaves no stale text. Inline panes live at the bottom,
	// so clearing to termRows does not eat unrelated content above the pane.
	top := p.startRow
	if newStartRow < top {
		top = newStartRow
	}
	var b strings.Builder
	for row := top; row <= termRows; row++ {
		fmt.Fprintf(&b, "\x1b[%d;1H\x1b[2K", row)
	}
	p.tty.WriteString(b.String()) //nolint:errcheck

	p.startRow, p.rows = newStartRow, newRows
	p.cols = newCols

	canvasCols := newCols
	if p.MaxCols > 0 && canvasCols > p.MaxCols {
		canvasCols = p.MaxCols
	}
	*cols = canvasCols
}

// Run renders root on every frame and dispatches events until the root signals
// quit or an unrecoverable error occurs.
func (p *Pane) Run(root Widget) error {
	defer func() {
		if r := recover(); r != nil {
			p.close()
			panic(r)
		}
	}()

	cols := p.cols
	if p.MaxCols > 0 && cols > p.MaxCols {
		cols = p.MaxCols
	}
	canvas := NewCanvas(cols, p.rows)

	// Read input in a goroutine and forward it on a channel so the main loop can
	// select between input and SIGWINCH. os.File.Read retries EINTR via Go's poll
	// layer (docs/TuiInput.md §2), so a signal alone cannot wake a blocking read —
	// hence the separate reader. The goroutine exits when the tty is closed (on
	// Close) or when done is signalled.
	type readResult struct {
		data []byte
		err  error
	}
	reads := make(chan readResult)
	done := make(chan struct{})
	defer close(done)
	p.readerDone = make(chan struct{})
	go func() {
		defer close(p.readerDone) // let close() join us so no read is left racing
		// Poll with a short timeout instead of blocking in Read: os.File.Fd (called
		// in New) detaches the tty from the runtime poller and puts it in blocking
		// mode, so Close cannot interrupt a bare Read — the goroutine would leak and
		// race the next terminal reader for input. Polling lets us observe done and
		// exit on teardown. EINTR (e.g. from a delivered SIGWINCH) just re-polls.
		pfd := []unix.PollFd{{Fd: int32(p.fd), Events: unix.POLLIN}}
		for {
			select {
			case <-done:
				return
			default:
			}
			n, perr := unix.Poll(pfd, 50)
			if perr != nil {
				if perr == unix.EINTR {
					continue
				}
				select {
				case reads <- readResult{err: perr}:
				case <-done:
				}
				return
			}
			if n == 0 {
				continue // timeout — loop back to re-check done
			}
			buf := make([]byte, 64)
			m, rerr := p.tty.Read(buf)
			if m > 0 {
				select {
				case reads <- readResult{data: buf[:m]}:
				case <-done:
					return
				}
			}
			if rerr != nil {
				select {
				case reads <- readResult{err: rerr}:
				case <-done:
				}
				return
			}
		}
	}()

	// redraw paints one frame. It is a closure so the mouse loop can repaint
	// after each event in a burst, letting the highlight follow the pointer
	// rather than jumping once after the whole burst is drained.
	redraw := func() {
		if canvas.Rows() != p.rows {
			canvas = NewCanvas(cols, p.rows)
		}
		canvas.Clear()
		root.Draw(canvas, canvas.Bounds())
		canvas.Flush(p.tty, p.startRow)
	}

	for {
		if p.Resizeable {
			if ch, ok := root.(ContentHeighter); ok {
				targetH := ch.ContentHeight()
				if targetH != p.rows {
					p.Resize(targetH)
				}
			}
		}

		redraw()

		select {
		case <-p.winch:
			// Terminal window resized: reflow width/bounds and repaint.
			p.applyWinch(&cols)
			canvas = NewCanvas(cols, p.rows)
			continue
		case rr := <-reads:
			if rr.err != nil || len(rr.data) == 0 {
				return nil
			}
			raw := rr.data

			// Try mouse first (SGR: \x1b[<…M/m). A single read may carry several
			// reports (1003 any-motion tracking floods them), so drain the whole
			// buffer report-by-report and dispatch each in order.
			if p.mouse && len(raw) >= 3 && raw[0] == 27 && raw[1] == '[' && raw[2] == '<' {
				quit := false
				for len(raw) > 0 {
					me, used, ok := scanMouse(raw)
					if !ok {
						break
					}
					// Translate absolute terminal coords to pane-relative 1-based:
					// the pane top (canvas row 0) is terminal row startRow, but widgets
					// hit-test as if their first row were Y=1. Without this, mouse only
					// lined up when the pane happened to sit at row 1.
					me.Y -= p.startRow - 1
					if root.HandleMouse(me) {
						quit = true
						break
					}
					raw = raw[used:]
					redraw() // repaint after each event so the pointer is followed live
				}
				if quit {
					return nil
				}
				continue
			}

			ke := DecodeKey(raw)
			if root.HandleKey(ke) {
				return nil
			}
		}
	}
}

// Close tears down the pane: clears the reserved region, restores terminal
// state, and closes /dev/tty. Safe to call multiple times.
func (p *Pane) Close() { p.close() }

func (p *Pane) close() {
	if p.restored {
		return
	}
	p.restored = true

	if p.winch != nil {
		signal.Stop(p.winch)
	}

	if p.mouse {
		p.tty.WriteString("\x1b[?1003l\x1b[?1006l") //nolint:errcheck
	}

	// Clear reserved region.
	var b strings.Builder
	for i := 0; i < p.rows; i++ {
		b.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", p.startRow+i))
	}
	b.WriteString(fmt.Sprintf("\x1b[%d;1H", p.startRow))
	b.WriteString("\x1b[?25h")    // restore cursor visibility (a prompt-less frame may have hidden it)
	p.tty.WriteString(b.String()) //nolint:errcheck

	// Discard any unread input before leaving raw mode: keys pressed during the
	// session (the closing Enter, leftover mouse/DSR reports) otherwise sit in the
	// tty queue and leak into the shell or the next cooked reader as a stray line.
	p.drainInput()

	_ = term.Restore(p.fd, p.oldState)
	_ = p.tty.Close()

	// Wait for the reader goroutine to notice done (it polls on a 50ms cycle) and
	// exit, so it cannot consume a keystroke meant for the next terminal reader.
	// Bounded so teardown can never hang.
	if p.readerDone != nil {
		select {
		case <-p.readerDone:
		case <-time.After(300 * time.Millisecond):
		}
	}
}

// drainInput flushes the terminal input queue (TCIFLUSH) so raw-mode leftovers
// don't bleed into whatever reads the terminal next.
func (p *Pane) drainInput() {
	const tcflsh = 0x540B // Linux TCFLSH; argument 0 selects TCIFLUSH (input queue)
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, uintptr(p.fd), tcflsh, 0)
}

func (p *Pane) installSignalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if _, ok := <-ch; ok {
			p.close()
			os.Exit(1)
		}
	}()

	// SIGWINCH goes to its own channel; the Run loop selects on it to reflow.
	// Cap 1 coalesces a burst of resizes (e.g. an interactive drag) into one.
	p.winch = make(chan os.Signal, 1)
	signal.Notify(p.winch, syscall.SIGWINCH)
}

// termSize returns the terminal dimensions via TIOCGWINSZ.
func termSize(fd int) (cols, rows int) {
	var ws struct{ Row, Col, X, Y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}
